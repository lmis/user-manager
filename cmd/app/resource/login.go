package resource

import (
	"fmt"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"user-manager/util/random"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

func RegisterLoginResource(group *gin.RouterGroup) {
	group.POST("login", ginext.WrapEndpoint(Login))
	group.POST("login-with-second-factor", ginext.WrapEndpoint(LoginWithSecondFactor))
}

type LoginTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
	Sudo     bool   `json:"sudo"`
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

func Login(ctx *gin.Context, r *dm.RequestContext, requestTO LoginTO) (LoginResponseTO, error) {
	securityLog := r.SecurityLog

	loginDescription := "Login"
	if requestTO.Sudo {
		loginDescription = "Sudo-login"
		if !r.User.IsPresent() {
			securityLog.Info("Sudo login attempted without valid user.")
			return LoginResponseTO{LoginResponseInvalidCredentials}, nil
		}
	}

	user, err := users.
		GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return LoginResponseTO{}, errs.Wrap("error fetching user", err)
	}
	if !user.IsPresent() {
		securityLog.Info(loginDescription + " attempt for non-existent user")
		return LoginResponseTO{LoginResponseInvalidCredentials}, nil
	}

	for _, role := range user.UserRoles {
		if role != dm.UserRoleUser {
			securityLog.Info(fmt.Sprintf(loginDescription+" attempt without second factor for non-user %d", user.ID()))
			return LoginResponseTO{LoginResponseInvalidCredentials}, nil
		}
	}

	if !auth.VerifyCredentials(requestTO.Password, user.Credentials) {
		securityLog.Info(fmt.Sprintf("Password mismatch for user %s", user.ID()))
		return LoginResponseTO{LoginResponseInvalidCredentials}, nil
	}

	if user.SecondFactorToken != "" {
		return LoginResponseTO{LoginResponse2faRequired}, nil
	}

	securityLog.Info(loginDescription)
	session := dm.UserSession{
		Token:     dm.UserSessionToken(random.MakeRandomURLSafeB64(21)),
		Type:      dm.UserSessionTypeLogin,
		TimeoutAt: time.Now().Add(dm.LoginSessionDuration),
	}

	if requestTO.Sudo {
		session.Type = dm.UserSessionTypeSudo
		session.TimeoutAt = time.Now().Add(dm.SudoSessionDuration)
	}
	if err = auth.InsertSession(ctx, r.Database, user.ID(), session); err != nil {
		return LoginResponseTO{}, errs.Wrap("error inserting session", err)
	}

	auth.SetSessionCookie(ctx, r.Config, string(session.Token), session.Type)
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

func LoginWithSecondFactor(ctx *gin.Context, r *dm.RequestContext, requestTO LoginWithSecondFactorTO) (LoginWithSecondFactorResponseTO, error) {
	securityLog := r.SecurityLog

	loginDescription := "Login (2FA)"
	if requestTO.Sudo {
		loginDescription = "Sudo-login (2FA)"
		if !r.User.IsPresent() {
			securityLog.Info("Sudo 2FA attempted without valid session.")
			return LoginWithSecondFactorResponseTO{}, nil
		}
	}
	user, err := users.GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return LoginWithSecondFactorResponseTO{}, errs.Wrap("error finding user", err)
	}
	if !user.IsPresent() {
		securityLog.Info(loginDescription + " attempt for non-existent user")
		return LoginWithSecondFactorResponseTO{}, nil
	}

	if !auth.VerifyCredentials(requestTO.Password, user.Credentials) {
		securityLog.Info("password mismatch")
		return LoginWithSecondFactorResponseTO{}, nil
	}

	if requestTO.SecondFactor != "" {
		throttling := user.SecondFactorThrottling

		if throttling.FailuresSinceSuccess != 0 && throttling.TimeoutUntil.After(time.Now()) {
			securityLog.Info("Throttled 2FA attempted")
			return LoginWithSecondFactorResponseTO{TimeoutUntil: throttling.TimeoutUntil}, nil
		}

		tokenMatches := user.SecondFactorToken != "" && totp.Validate(requestTO.SecondFactor, user.SecondFactorToken)

		if tokenMatches {
			if err := auth.UpdateSecondFactorThrottling(ctx, r.Database, user.ID(), 0, nil); err != nil {
				return LoginWithSecondFactorResponseTO{}, errs.Wrap("issue resetting throttling in db", err)
			}
		} else {
			var maybeTimeoutUntil *time.Time
			failedAttemptsSinceLastSuccess := user.SecondFactorThrottling.FailuresSinceSuccess + 1
			// TODO: Check this exponential timeout logic
			if failedAttemptsSinceLastSuccess%5 == 0 {
				*maybeTimeoutUntil = time.Now().Add(time.Minute * 3 * time.Duration(failedAttemptsSinceLastSuccess))
			}
			if err := auth.UpdateSecondFactorThrottling(ctx, r.Database, user.ID(), failedAttemptsSinceLastSuccess, maybeTimeoutUntil); err != nil {
				return LoginWithSecondFactorResponseTO{}, errs.Wrap("issue updating throttling in db", err)
			}
			securityLog.Info("2FA mismatch")
			return LoginWithSecondFactorResponseTO{}, nil
		}

		if requestTO.RememberDevice {
			securityLog.Info("2FA login with 'remember device' enabled, issuing device token")
			deviceSession := dm.UserSession{
				Token:     dm.UserSessionToken(random.MakeRandomURLSafeB64(21)),
				Type:      dm.UserSessionTypeRememberDevice,
				TimeoutAt: time.Now().Add(dm.DeviceSessionDuration),
			}
			err = auth.InsertSession(ctx, r.Database, user.ID(), deviceSession)
			if err != nil {
				return LoginWithSecondFactorResponseTO{}, errs.Wrap("error inserting device session", err)
			}

			auth.SetSessionCookie(ctx, r.Config, string(deviceSession.Token), deviceSession.Type)
		}

		securityLog.Info("Login passed with 2FA token")
	} else {
		maybeDeviceSessionID, err := auth.GetSessionCookie(ctx, dm.UserSessionTypeLogin)
		if err != nil {
			return LoginWithSecondFactorResponseTO{}, errs.Wrap("issue reading device session cookie", err)
		}
		if maybeDeviceSessionID == "" {
			return LoginWithSecondFactorResponseTO{}, nil
		}

		deviceSession, err := auth.GetUserForSession(ctx, r.Database, maybeDeviceSessionID, dm.UserSessionTypeRememberDevice)

		if err != nil {
			return LoginWithSecondFactorResponseTO{}, errs.Wrap("fetching device session failed", err)
		}
		if !deviceSession.IsPresent() {
			return LoginWithSecondFactorResponseTO{}, nil
		}
		securityLog.Info("Login passed with device token cookie")
	}

	session := dm.UserSession{
		Token:     dm.UserSessionToken(random.MakeRandomURLSafeB64(21)),
		Type:      dm.UserSessionTypeLogin,
		TimeoutAt: time.Now().Add(dm.LoginSessionDuration),
	}
	if requestTO.Sudo {
		session.Type = dm.UserSessionTypeSudo
		session.TimeoutAt = time.Now().Add(dm.SudoSessionDuration)
	}
	if err = auth.InsertSession(ctx, r.Database, user.ID(), session); err != nil {
		return LoginWithSecondFactorResponseTO{}, errs.Wrap("error inserting login session", err)
	}

	auth.SetSessionCookie(ctx, r.Config, string(session.Token), session.Type)
	return LoginWithSecondFactorResponseTO{LoggedIn: true}, nil
}
