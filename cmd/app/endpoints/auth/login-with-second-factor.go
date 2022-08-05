package auth

import (
	"context"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	session_service "user-manager/cmd/app/services/session"
	"user-manager/db"
	"user-manager/db/generated/models"
	app_user "user-manager/domain-model/id/appUser"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/crypto/bcrypt"
)

type LoginWithSecondFactorTO struct {
	LoginTO
	SecondFactor   string `json:"secondFactor"`
	RememberDevice bool   `json:"rememberDevice"`
}

type LoginWithSecondFactorResponseTO struct {
	LoggedIn     bool      `json:"loggedIn"`
	TimeoutUntil time.Time `json:"timeoutUntil,omitempty"`
}

func PostLoginWithSecondFactor(requestContext *ginext.RequestContext, requestTO *LoginWithSecondFactorTO, c *gin.Context) (*LoginWithSecondFactorResponseTO, error) {
	tx := requestContext.Tx
	securityLog := requestContext.SecurityLog

	user, err := db.Fetch(func(ctx context.Context) (*models.AppUser, error) {
		return models.AppUsers(
			models.AppUserWhere.Email.EQ(requestTO.Email),
			qm.Load(models.AppUserRels.AppUserRoles),
			qm.Load(models.AppUserRels.TwoFactorThrottling),
		).One(ctx, tx)
	})
	if err != nil {
		return nil, util.Wrap("error finding user", err)
	}
	if user == nil {
		securityLog.Info("Login attempt for non-existant user")
		return &LoginWithSecondFactorResponseTO{}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("password missmatch")
		return &LoginWithSecondFactorResponseTO{}, nil
	}

	if requestTO.SecondFactor != "" {
		throttling, err := db.Fetch(func(ctx context.Context) (*models.TwoFactorThrottling, error) {
			return models.TwoFactorThrottlings(
				models.TwoFactorThrottlingWhere.AppUserID.EQ(user.AppUserID),
			).One(ctx, tx)
		})
		if err != nil {
			return nil, util.Wrap("error loading throttling", err)
		}

		if throttling != nil && throttling.TimeoutUntil.Valid && throttling.TimeoutUntil.Time.After(time.Now()) {
			securityLog.Info("Throttled 2FA attempted")
			return &LoginWithSecondFactorResponseTO{TimeoutUntil: throttling.TimeoutUntil.Time}, nil
		}

		tokenMatches := user.TwoFactorToken.Valid && totp.Validate(requestTO.SecondFactor, user.TwoFactorToken.String)
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
		if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return throttling.Update(ctx, tx, boil.Infer()) }); err != nil {
			return nil, util.Wrap("issue updating throttling in db", err)
		}

		if !tokenMatches {
			securityLog.Info("2FA mismatch")
			return &LoginWithSecondFactorResponseTO{}, nil
		}

		if requestTO.RememberDevice {
			securityLog.Info("2FA login with 'remember device' enabled, issuing device token")
			deviceSessionId, err := session_service.InsertSession(requestContext, models.UserSessionTypeLOGIN, app_user.ID(user.AppUserID), session_service.DeviceSessionDuration)
			if err != nil {
				return nil, util.Wrap("error inserting device session", err)
			}

			session_service.SetSessionCookie(c, deviceSessionId, models.UserSessionTypeREMEMBER_DEVICE)
		}

		securityLog.Info("Login passed with 2FA token")
	} else {
		deviceSessionId, err := session_service.GetSessionCookie(c, models.UserSessionTypeLOGIN)
		if err != nil {
			return nil, util.Wrap("issue reading device session cookie", err)
		}
		if deviceSessionId == "" {
			return &LoginWithSecondFactorResponseTO{}, nil
		}
		deviceSession, err := db.Fetch(func(ctx context.Context) (*models.UserSession, error) {
			return models.UserSessions(models.UserSessionWhere.UserSessionID.EQ(deviceSessionId),
				models.UserSessionWhere.TimeoutAt.GT(time.Now()),
				models.UserSessionWhere.UserSessionType.EQ(models.UserSessionTypeREMEMBER_DEVICE),
			).One(ctx, requestContext.Tx)
		})

		if err != nil {
			return nil, util.Wrap("getting session failed", err)
		}
		if deviceSession == nil {
			return &LoginWithSecondFactorResponseTO{}, nil
		}
		securityLog.Info("Login passed with device token cookie")
	}

	sessionId, err := session_service.InsertSession(requestContext, models.UserSessionTypeLOGIN, app_user.ID(user.AppUserID), time.Minute*60)
	if err != nil {
		return nil, util.Wrap("error inserting login session", err)
	}

	session_service.SetSessionCookie(c, sessionId, models.UserSessionTypeLOGIN)
	return &LoginWithSecondFactorResponseTO{LoggedIn: true}, nil
}
