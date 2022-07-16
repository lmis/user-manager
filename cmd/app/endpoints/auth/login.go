package authendpoints

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

type CredentialsTO struct {
	Email        string `json:"email"`
	Password     []byte `json:"password"`
	SecondFactor string `json:"secondFactor"`
}

type LoginResponseTO struct {
	LoggedIn                         bool      `json:"loggedIn"`
	ResubmitWithSecondFactorRequired bool      `json:"resubmitWithSecondFactorRequired"`
	TimeoutUntil                     time.Time `json:"timeoutUntil,omitempty"`
}

func PostLogin(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog
	loginResponseTO := LoginResponseTO{}
	var credentialsTO CredentialsTO
	if err := c.BindJSON(&credentialsTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to credentialsTO", err))
		return
	}

	var user *models.AppUser
	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(
		models.AppUserWhere.Email.EQ(credentialsTO.Email),
		qm.Load(models.AppUserRels.TwoFactorThrottling),
	).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			// Avoid 401 etc, to keep browsers from throwing out basic auth
			securityLog.Info("Failed login attempt")
			c.JSON(http.StatusOK, loginResponseTO)
		} else {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("error finding user", err))
		}
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), credentialsTO.Password); err != nil {
		securityLog.Info("Password mismatch")
		// Avoid 401 etc, to keep browsers from throwing out basic auth
		c.JSON(http.StatusOK, loginResponseTO)
		return
	}

	if user.TwoFactorToken.Valid {
		throttling := user.R.TwoFactorThrottling
		if throttling != nil && throttling.TimeoutUntil.Valid && throttling.TimeoutUntil.Time.After(time.Now()) {
			securityLog.Info("Throttled 2FA attemped")
			loginResponseTO.TimeoutUntil = throttling.TimeoutUntil.Time
			c.JSON(http.StatusOK, loginResponseTO)
			return
		}

		tokenMatches := totp.Validate(credentialsTO.SecondFactor, user.TwoFactorToken.String)
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
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue updating throttling in db", err))
			return
		}
		if rows != 1 {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap(fmt.Sprintf("wrong number of rows affected: %d", rows), err))
			return
		}
		if !tokenMatches {
			securityLog.Info("2FA token mismatch")
			if user.Role == models.UserRoleUSER {
				loginResponseTO.ResubmitWithSecondFactorRequired = true
			}
			c.JSON(http.StatusOK, loginResponseTO)
			return
		}
	} else if user.Role != models.UserRoleUSER {
		c.AbortWithError(http.StatusBadRequest, util.Errorf("missing two factor token in DB for non-USER (id: %v)", user.AppUserID))
		return
	}

	sessionID := util.MakeRandomURLSafeB64(21)

	session := models.UserSession{
		UserSessionID: sessionID,
		AppUserID:     user.AppUserID,
		TimeoutAt:     time.Now().Add(time.Minute * 60),
	}
	ctx, cancelTimeout = db.DefaultQueryContext()
	defer cancelTimeout()
	if err = session.Insert(ctx, tx, boil.Infer()); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("cannot insert session", err))
		return
	}

	securityLog.Info("Login")
	SetSessionCookie(c, sessionID)
	loginResponseTO.LoggedIn = true
	c.JSON(http.StatusOK, loginResponseTO)
}
