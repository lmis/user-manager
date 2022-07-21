package auth

import (
	"database/sql"
	"fmt"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/pquerna/otp/totp"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
)

type ValidateLoginError struct {
	Message                          string
	ResubmitWithSecondFactorRequired bool
	TimeoutUntil                     time.Time
}

func (e *ValidateLoginError) Error() string {
	return e.Message
}

func ValidateLogin(
	requestContext *ginext.RequestContext,
	user *models.AppUser,
	password []byte,
	secondFactor string,
) error {
	tx := requestContext.Tx

	if user.Role != models.UserRoleUSER && !user.TwoFactorToken.Valid {
		return util.Errorf("missing two factor token in DB for non-USER (id: %v)", user.AppUserID)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), password); err != nil {
		return &ValidateLoginError{Message: "password mismatch"}
	}

	if !user.TwoFactorToken.Valid {
		return nil
	}

	if secondFactor == "" {
		if user.Role == models.UserRoleUSER {
			return &ValidateLoginError{
				Message:                          "2FA token missing",
				ResubmitWithSecondFactorRequired: true,
			}
		} else {
			return &ValidateLoginError{
				Message: "2FA token missing",
			}
		}
	}

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	throttling, err := models.TwoFactorThrottlings(
		models.TwoFactorThrottlingWhere.AppUserID.EQ(user.AppUserID),
	).One(ctx, tx)
	if err != nil && err != sql.ErrNoRows {
		return util.Wrap("error loading throttling", err)
	}
	if throttling != nil && throttling.TimeoutUntil.Valid && throttling.TimeoutUntil.Time.After(time.Now()) {
		return &ValidateLoginError{Message: "Throttled 2FA attempted", TimeoutUntil: throttling.TimeoutUntil.Time}
	}

	tokenMatches := totp.Validate(secondFactor, user.TwoFactorToken.String)
	if tokenMatches {
		throttling.FailedAttemptsSinceLastSuccess = 0
		throttling.TimeoutUntil = null.TimeFromPtr(nil)
	} else {
		throttling.FailedAttemptsSinceLastSuccess += 1
		// TODO: Check this exponential timeout logic
		if throttling.FailedAttemptsSinceLastSuccess%5 == 0 {
			throttling.TimeoutUntil = null.TimeFrom(time.Now().Add(time.Minute * 3 * time.Duration(throttling.FailedAttemptsSinceLastSuccess)))
		}
	}
	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := throttling.Update(ctx, tx, boil.Infer())
	if err != nil {
		return util.Wrap("issue updating throttling in db", err)
	}
	if rows != 1 {
		return util.Wrap(fmt.Sprintf("wrong number of rows affected by throttling update: %d", rows), err)
	}

	if !tokenMatches {
		return &ValidateLoginError{
			Message: "2FA mismatch",
		}
	}
	return nil
}
