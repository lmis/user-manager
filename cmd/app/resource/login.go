package resource

import (
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util"
	"user-manager/util/nullable"
	"user-manager/util/slices"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type LoginResource struct {
	securityLog                      domain_model.SecurityLog
	sessionCookieService             *service.SessionCookieService
	sessionRepository                *repository.SessionRepository
	userRepository                   *repository.UserRepository
	secondFactorThrottlingRepository *repository.SecondFactorThrottlingRepository
}

func ProvideLoginResource(
	securityLog domain_model.SecurityLog,
	sessionCookieService *service.SessionCookieService,
	sessionRepository *repository.SessionRepository,
	userRepository *repository.UserRepository,
	secondFactorThrottlingRepository *repository.SecondFactorThrottlingRepository,
) *LoginResource {
	return &LoginResource{securityLog, sessionCookieService, sessionRepository, userRepository, secondFactorThrottlingRepository}
}

func RegisterLoginResource(group *gin.RouterGroup) {
	group.POST("login", ginext.WrapEndpoint(InitializeLoginResource, (*LoginResource).Login))
	group.POST("login-with-second-factor", ginext.WrapEndpoint(InitializeLoginResource, (*LoginResource).LoginWithSecondFactor))
}

type LoginTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type LoginResponseStatus string

const (
	LOGIN_RESPONSE_LOGGED_IN           LoginResponseStatus = "logged-in"
	LOGIN_RESPONSE_2FA_REQUIRED        LoginResponseStatus = "second-factor-required"
	LOGIN_RESPONSE_INVALID_CREDENTIALS LoginResponseStatus = "invalid-credentials"
)

type LoginResponseTO struct {
	Status LoginResponseStatus `json:"status"`
}

func (r *LoginResource) Login(requestTO *LoginTO) (*LoginResponseTO, error) {
	securityLog := r.securityLog
	sessionCookieService := r.sessionCookieService
	userRepository := r.userRepository
	sessionRepository := r.sessionRepository

	maybeUser, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return nil, util.Wrap("error fetching user", err)
	}
	if maybeUser.IsEmpty() {
		securityLog.Info("Login attempt for non-existant user")
		return &LoginResponseTO{LOGIN_RESPONSE_INVALID_CREDENTIALS}, nil
	}

	user := maybeUser.OrPanic()

	hasNonUserRole := slices.Any(user.UserRoles, func(role domain_model.UserRole) bool { return role != domain_model.USER_ROLE_USER })
	if hasNonUserRole {
		securityLog.Info("Login attempt without second factor for non-user %d", user.AppUserID)
		return &LoginResponseTO{LOGIN_RESPONSE_INVALID_CREDENTIALS}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("Password mismatch for user %s", user.AppUserID)
		return &LoginResponseTO{LOGIN_RESPONSE_INVALID_CREDENTIALS}, nil
	}

	if user.SecondFactorToken.IsPresent {
		return &LoginResponseTO{LOGIN_RESPONSE_2FA_REQUIRED}, nil
	}

	securityLog.Info("Login")
	sessionId := util.MakeRandomURLSafeB64(21)
	if err = sessionRepository.InsertSession(sessionId, domain_model.USER_SESSION_TYPE_LOGIN, user.AppUserID, domain_model.LOGIN_SESSION_DURATION); err != nil {
		return nil, util.Wrap("error inserting session", err)
	}

	sessionCookieService.SetSessionCookie(nullable.Of(sessionId), domain_model.USER_SESSION_TYPE_LOGIN)
	return &LoginResponseTO{LOGIN_RESPONSE_LOGGED_IN}, nil
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

func (r *LoginResource) LoginWithSecondFactor(requestTO *LoginWithSecondFactorTO) (*LoginWithSecondFactorResponseTO, error) {
	sessionCookieService := r.sessionCookieService
	sessionRepository := r.sessionRepository
	userRepository := r.userRepository
	secondFactorThrottlingRepository := r.secondFactorThrottlingRepository
	securityLog := r.securityLog

	maybeUser, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return nil, util.Wrap("error finding user", err)
	}
	if maybeUser.IsEmpty() {
		securityLog.Info("Login attempt for non-existant user")
		return &LoginWithSecondFactorResponseTO{}, nil
	}
	user := maybeUser.OrPanic()

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("password missmatch")
		return &LoginWithSecondFactorResponseTO{}, nil
	}

	if requestTO.SecondFactor != "" {
		throttling, err := secondFactorThrottlingRepository.GetForUser(user.AppUserID)
		if err != nil {
			return nil, util.Wrap("error loading throttling", err)
		}

		if throttling.IsPresent && throttling.OrPanic().TimeoutUntil.IsPresent && throttling.OrPanic().TimeoutUntil.OrPanic().After(time.Now()) {
			securityLog.Info("Throttled 2FA attempted")
			return &LoginWithSecondFactorResponseTO{TimeoutUntil: throttling.OrPanic().TimeoutUntil.OrPanic()}, nil
		}

		tokenMatches := user.SecondFactorToken.IsPresent && totp.Validate(requestTO.SecondFactor, user.SecondFactorToken.OrPanic())

		if throttling.IsPresent {
			failedAttemptsSinceLastSuccess := int32(0)
			timeoutUntil := nullable.Empty[time.Time]()
			if !tokenMatches {
				failedAttemptsSinceLastSuccess = throttling.OrPanic().FailedAttemptsSinceLastSuccess + 1
				// TODO: Check this exponential timeout logic
				if failedAttemptsSinceLastSuccess%5 == 0 {
					timeoutUntil = nullable.Of(time.Now().Add(time.Minute * 3 * time.Duration(failedAttemptsSinceLastSuccess)))
				}
			}
			if err := secondFactorThrottlingRepository.Update(throttling.OrPanic().SecondFactorThrottlingID, failedAttemptsSinceLastSuccess, timeoutUntil); err != nil {
				return nil, util.Wrap("issue updating throttling in db", err)
			}
		} else if !tokenMatches {
			if err := secondFactorThrottlingRepository.Insert(user.AppUserID, 1); err != nil {
				return nil, util.Wrap("issue inserting throttling in db", err)
			}
		}

		if !tokenMatches {
			securityLog.Info("2FA mismatch")
			return &LoginWithSecondFactorResponseTO{}, nil
		}

		if requestTO.RememberDevice {
			securityLog.Info("2FA login with 'remember device' enabled, issuing device token")
			deviceSessionId := util.MakeRandomURLSafeB64(21)
			err = sessionRepository.InsertSession(deviceSessionId, domain_model.USER_SESSION_TYPE_LOGIN, user.AppUserID, domain_model.DEVICE_SESSION_DURATION)
			if err != nil {
				return nil, util.Wrap("error inserting device session", err)
			}

			sessionCookieService.SetSessionCookie(nullable.Of(deviceSessionId), domain_model.USER_SESSION_TYPE_REMEMBER_DEVICE)
		}

		securityLog.Info("Login passed with 2FA token")
	} else {
		maybeDeviceSessionId, err := sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_LOGIN)
		if err != nil {
			return nil, util.Wrap("issue reading device session cookie", err)
		}
		if maybeDeviceSessionId.IsEmpty() {
			return &LoginWithSecondFactorResponseTO{}, nil
		}

		deviceSession, err := sessionRepository.GetSessionAndUser(maybeDeviceSessionId.OrPanic(), domain_model.USER_SESSION_TYPE_REMEMBER_DEVICE)

		if err != nil {
			return nil, util.Wrap("fetching device session failed", err)
		}
		if deviceSession.IsEmpty() {
			return &LoginWithSecondFactorResponseTO{}, nil
		}
		securityLog.Info("Login passed with device token cookie")
	}

	sessionId := util.MakeRandomURLSafeB64(21)
	if err = sessionRepository.InsertSession(sessionId, domain_model.USER_SESSION_TYPE_LOGIN, user.AppUserID, domain_model.LOGIN_SESSION_DURATION); err != nil {
		return nil, util.Wrap("error inserting login session", err)
	}

	sessionCookieService.SetSessionCookie(nullable.Of(sessionId), domain_model.USER_SESSION_TYPE_LOGIN)
	return &LoginWithSecondFactorResponseTO{LoggedIn: true}, nil
}
