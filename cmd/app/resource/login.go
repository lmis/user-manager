package resource

import (
	"github.com/a-h/templ"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/router/render"
	"user-manager/cmd/app/service/auth"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"user-manager/util/random"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

func RegisterLoginResource(group *gin.RouterGroup) {
	group.GET("login", ginext.WrapTemplWithoutPayload(GetLogin))
	group.POST("login", ginext.WrapTempl(PostLogin))
	group.POST("login-with-second-factor", ginext.WrapEndpoint(SecondFactor))
}

type LoginTO struct {
	Email       string `form:"email"`
	Password    string `form:"password"`
	Sudo        bool   `form:"sudo"`
	RedirectURL string `form:"redirectUrl"`
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

func GetLogin(ctx *gin.Context, r *dm.RequestContext) (templ.Component, error) {
	redirectURL, ok := ctx.GetQuery("redirectUrl")
	if redirectURL == "" || !ok {
		redirectURL = "/"
	}
	return render.FullPage(ctx, r, "Login", render.LoginForm(redirectURL)), nil
}

func PostLogin(ctx *gin.Context, r *dm.RequestContext, requestTO LoginTO) (templ.Component, error) {
	logger := r.Logger

	loginDescription := "Login"
	if requestTO.Sudo {
		loginDescription = "Sudo-login"
		if !r.User.IsPresent() {
			logger.Info("Sudo login attempted without valid user.")
			return render.LoginFormError("Invalid credentials"), nil
		}
	}

	user, err := users.
		GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return nil, errs.Wrap("error fetching user", err)
	}
	if !user.IsPresent() {
		logger.Info(loginDescription + " attempt for non-existent user")
		return render.LoginFormError("Invalid credentials"), nil
	}

	for _, role := range user.UserRoles {
		if role != dm.UserRoleUser {
			logger.Info(loginDescription+" attempt without second factor for non-user", "userID", user.IDHex())
			return render.LoginFormError("Invalid credentials"), nil
		}
	}

	if !auth.VerifyCredentials([]byte(requestTO.Password), user.Credentials) {
		logger.Info("Password mismatch for user", "userID", user.IDHex())
		return render.LoginFormError("Invalid credentials"), nil
	}

	if user.SecondFactorToken != "" {
		ctx.Header("HX-Reswap", "innerHTML")
		return render.Login2FA(), nil
	}

	logger.Info(loginDescription)
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
		return nil, errs.Wrap("error inserting session", err)
	}

	auth.SetSessionCookie(ctx, r.Config, string(session.Token), session.Type)
	// TODO: Redirect URL is user input. Rethink this.
	ginext.HXLocationOrRedirect(ctx, requestTO.RedirectURL)
	return nil, nil
}

type SecondFactorTO struct {
	SecondFactor   string `json:"secondFactor"`
	RememberDevice bool   `json:"rememberDevice"`
	Sudo           bool   `json:"sudo"`
}

type LoginWithSecondFactorResponseTO struct {
	LoggedIn     bool      `json:"loggedIn"`
	TimeoutUntil time.Time `json:"timeoutUntil,omitempty"`
}

func SecondFactor(ctx *gin.Context, r *dm.RequestContext, requestTO SecondFactorTO) (LoginWithSecondFactorResponseTO, error) {
	logger := r.Logger

	loginDescription := "Login (2FA)"
	sessionType := dm.UserSessionTypeLogin
	if requestTO.Sudo {
		loginDescription = "Sudo-login (2FA)"
		sessionType = dm.UserSessionTypeSudo
	}
	sessionToken, err := auth.GetSessionCookie(ctx, sessionType)

	if err != nil {
		return LoginWithSecondFactorResponseTO{}, errs.Wrap("error getting session cookie for second factor auth", err)
	}
	if sessionToken == "" {
		logger.Info(loginDescription + " attempt without session token")
		return LoginWithSecondFactorResponseTO{}, nil
	}

	user, err := auth.GetUserForSessionForSecondFactorVerification(ctx, r.Database, sessionToken, sessionType)
	if err != nil {
		return LoginWithSecondFactorResponseTO{}, errs.Wrap("error fetching user for session token", err)
	}
	if !user.IsPresent() {
		logger.Info(loginDescription + " attempt for non-existent user")
		return LoginWithSecondFactorResponseTO{}, nil
	}

	if requestTO.SecondFactor != "" {
		throttling := user.SecondFactorThrottling

		if throttling.FailuresSinceSuccess != 0 && throttling.TimeoutUntil.After(time.Now()) {
			logger.Info("Throttled 2FA attempted")
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
			logger.Info("2FA mismatch")
			return LoginWithSecondFactorResponseTO{}, nil
		}

		if requestTO.RememberDevice {
			logger.Info("2FA login with 'remember device' enabled, issuing device token")
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

		logger.Info("Login passed with 2FA token")
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
		logger.Info("Login passed with device token cookie")
	}

	if err = auth.SetSecondFactorVerifiedForSession(ctx, r.Database, sessionToken); err != nil {
		return LoginWithSecondFactorResponseTO{}, errs.Wrap("issue setting second factor verified in db", err)
	}

	return LoginWithSecondFactorResponseTO{LoggedIn: true}, nil
}
