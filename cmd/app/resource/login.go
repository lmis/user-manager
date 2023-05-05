package resource

import (
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"user-manager/util/random"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func RegisterLoginResource(group *gin.RouterGroup) {
	group.POST("login", ginext.WrapEndpoint(Login))
	group.POST("login-with-second-factor", ginext.WrapEndpoint(LoginWithSecondFactor))
}

type LoginTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type LoginResponseStatus string

const (
	LoginResponseLoggedIn           LoginResponseStatus = "logged-in"
	LoginResponse2faRequired        LoginResponseStatus = "second-factor-required"
	LoginResponseInvalidCredentials LoginResponseStatus = "invalid-credentials"
)

type LoginResponseTO struct {
	Status LoginResponseStatus `json:"status"`
}

func Login(ctx *gin.Context, r *ginext.RequestContext, requestTO LoginTO) (LoginResponseTO, error) {
	securityLog := r.SecurityLog

	user, err := repository.GetUserForEmail(ctx, r.Tx, requestTO.Email)
	if err != nil {
		return LoginResponseTO{}, errors.Wrap("error fetching user", err)
	}
	if user.AppUserID == 0 {
		securityLog.Info("Login attempt for non-existent user")
		return LoginResponseTO{LoginResponseInvalidCredentials}, nil
	}

	for _, role := range user.UserRoles {
		if role != dm.UserRoleUser {
			securityLog.Info("Login attempt without second factor for non-user %d", user.AppUserID)
			return LoginResponseTO{LoginResponseInvalidCredentials}, nil
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("Password mismatch for user %s", user.AppUserID)
		return LoginResponseTO{LoginResponseInvalidCredentials}, nil
	}

	if user.SecondFactorToken != "" {
		return LoginResponseTO{LoginResponse2faRequired}, nil
	}

	securityLog.Info("Login")
	sessionID := random.MakeRandomURLSafeB64(21)
	if err = repository.InsertSession(ctx, r.Tx, sessionID, dm.UserSessionTypeLogin, user.AppUserID, dm.LoginSessionDuration); err != nil {
		return LoginResponseTO{}, errors.Wrap("error inserting session", err)
	}

	service.SetSessionCookie(ctx, r.Config, sessionID, dm.UserSessionTypeLogin)
	return LoginResponseTO{LoginResponseLoggedIn}, nil
}

type LoginWithSecondFactorTO struct {
	LoginTO
	SecondFactor   string `json:"secondFactor"`
	RememberDevice bool   `json:"rememberDevice"`
}

type LoginWithSecondFactorResponseTO struct {
	LoggedIn     bool      `json:"loggedIn"`
	TimeoutUntil time.Time `json:"timeoutUntil,omitempty"`
}

func LoginWithSecondFactor(ctx *gin.Context, r *ginext.RequestContext, requestTO LoginWithSecondFactorTO) (LoginWithSecondFactorResponseTO, error) {
	securityLog := r.SecurityLog

	user, err := repository.GetUserForEmail(ctx, r.Tx, requestTO.Email)
	if err != nil {
		return LoginWithSecondFactorResponseTO{}, errors.Wrap("error finding user", err)
	}
	if user.AppUserID == 0 {
		securityLog.Info("Login attempt for non-existent user")
		return LoginWithSecondFactorResponseTO{}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("password mismatch")
		return LoginWithSecondFactorResponseTO{}, nil
	}

	if requestTO.SecondFactor != "" {
		throttling, err := repository.GetSecondFactorThrottlingForUser(ctx, r.Tx, user.AppUserID)
		if err != nil {
			return LoginWithSecondFactorResponseTO{}, errors.Wrap("error loading throttling", err)
		}

		if throttling.AppUserID != 0 && throttling.TimeoutUntil.After(time.Now()) {
			securityLog.Info("Throttled 2FA attempted")
			return LoginWithSecondFactorResponseTO{TimeoutUntil: throttling.TimeoutUntil}, nil
		}

		tokenMatches := user.SecondFactorToken != "" && totp.Validate(requestTO.SecondFactor, user.SecondFactorToken)

		if throttling.AppUserID != 0 {
			failedAttemptsSinceLastSuccess := int32(0)
			var maybeTimeoutUntil *time.Time
			if !tokenMatches {
				failedAttemptsSinceLastSuccess = throttling.FailedAttemptsSinceLastSuccess + 1
				// TODO: Check this exponential timeout logic
				if failedAttemptsSinceLastSuccess%5 == 0 {
					*maybeTimeoutUntil = time.Now().Add(time.Minute * 3 * time.Duration(failedAttemptsSinceLastSuccess))
				}
			}
			if err := repository.UpdateSecondFactorThrottling(ctx, r.Tx, throttling.SecondFactorThrottlingID, failedAttemptsSinceLastSuccess, maybeTimeoutUntil); err != nil {
				return LoginWithSecondFactorResponseTO{}, errors.Wrap("issue updating throttling in db", err)
			}
		} else if !tokenMatches {
			if err := repository.InsertSecondFactorThrottling(ctx, r.Tx, user.AppUserID, 1); err != nil {
				return LoginWithSecondFactorResponseTO{}, errors.Wrap("issue inserting throttling in db", err)
			}
		}

		if !tokenMatches {
			securityLog.Info("2FA mismatch")
			return LoginWithSecondFactorResponseTO{}, nil
		}

		if requestTO.RememberDevice {
			securityLog.Info("2FA login with 'remember device' enabled, issuing device token")
			deviceSessionID := random.MakeRandomURLSafeB64(21)
			err = repository.InsertSession(ctx, r.Tx, deviceSessionID, dm.UserSessionTypeLogin, user.AppUserID, dm.DeviceSessionDuration)
			if err != nil {
				return LoginWithSecondFactorResponseTO{}, errors.Wrap("error inserting device session", err)
			}

			service.SetSessionCookie(ctx, r.Config, deviceSessionID, dm.UserSessionTypeRememberDevice)
		}

		securityLog.Info("Login passed with 2FA token")
	} else {
		maybeDeviceSessionID, err := service.GetSessionCookie(ctx, dm.UserSessionTypeLogin)
		if err != nil {
			return LoginWithSecondFactorResponseTO{}, errors.Wrap("issue reading device session cookie", err)
		}
		if maybeDeviceSessionID == "" {
			return LoginWithSecondFactorResponseTO{}, nil
		}

		deviceSession, err := repository.GetSessionAndUser(ctx, r.Tx, maybeDeviceSessionID, dm.UserSessionTypeRememberDevice)

		if err != nil {
			return LoginWithSecondFactorResponseTO{}, errors.Wrap("fetching device session failed", err)
		}
		if deviceSession.UserSessionID == "" {
			return LoginWithSecondFactorResponseTO{}, nil
		}
		securityLog.Info("Login passed with device token cookie")
	}

	sessionID := random.MakeRandomURLSafeB64(21)
	if err = repository.InsertSession(ctx, r.Tx, sessionID, dm.UserSessionTypeLogin, user.AppUserID, dm.LoginSessionDuration); err != nil {
		return LoginWithSecondFactorResponseTO{}, errors.Wrap("error inserting login session", err)
	}

	service.SetSessionCookie(ctx, r.Config, sessionID, dm.UserSessionTypeLogin)
	return LoginWithSecondFactorResponseTO{LoggedIn: true}, nil
}
