package resource

import (
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"user-manager/util/random"

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

	user, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return nil, errors.Wrap("error fetching user", err)
	}
	if user.AppUserID == 0 {
		securityLog.Info("Login attempt for non-existant user")
		return &LoginResponseTO{LOGIN_RESPONSE_INVALID_CREDENTIALS}, nil
	}

	for _, role := range user.UserRoles {
		if role != domain_model.USER_ROLE_USER {
			securityLog.Info("Login attempt without second factor for non-user %d", user.AppUserID)
			return &LoginResponseTO{LOGIN_RESPONSE_INVALID_CREDENTIALS}, nil
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("Password mismatch for user %s", user.AppUserID)
		return &LoginResponseTO{LOGIN_RESPONSE_INVALID_CREDENTIALS}, nil
	}

	if user.SecondFactorToken != "" {
		return &LoginResponseTO{LOGIN_RESPONSE_2FA_REQUIRED}, nil
	}

	securityLog.Info("Login")
	sessionID := random.MakeRandomURLSafeB64(21)
	if err = sessionRepository.InsertSession(sessionID, domain_model.USER_SESSION_TYPE_LOGIN, user.AppUserID, domain_model.LOGIN_SESSION_DURATION); err != nil {
		return nil, errors.Wrap("error inserting session", err)
	}

	sessionCookieService.SetSessionCookie(sessionID, domain_model.USER_SESSION_TYPE_LOGIN)
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

	user, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return nil, errors.Wrap("error finding user", err)
	}
	if user.AppUserID == 0 {
		securityLog.Info("Login attempt for non-existant user")
		return &LoginWithSecondFactorResponseTO{}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("password missmatch")
		return &LoginWithSecondFactorResponseTO{}, nil
	}

	if requestTO.SecondFactor != "" {
		throttling, err := secondFactorThrottlingRepository.GetForUser(user.AppUserID)
		if err != nil {
			return nil, errors.Wrap("error loading throttling", err)
		}

		if throttling.AppUserID != 0 && throttling.TimeoutUntil.After(time.Now()) {
			securityLog.Info("Throttled 2FA attempted")
			return &LoginWithSecondFactorResponseTO{TimeoutUntil: throttling.TimeoutUntil}, nil
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
			if err := secondFactorThrottlingRepository.Update(throttling.SecondFactorThrottlingID, failedAttemptsSinceLastSuccess, maybeTimeoutUntil); err != nil {
				return nil, errors.Wrap("issue updating throttling in db", err)
			}
		} else if !tokenMatches {
			if err := secondFactorThrottlingRepository.Insert(user.AppUserID, 1); err != nil {
				return nil, errors.Wrap("issue inserting throttling in db", err)
			}
		}

		if !tokenMatches {
			securityLog.Info("2FA mismatch")
			return &LoginWithSecondFactorResponseTO{}, nil
		}

		if requestTO.RememberDevice {
			securityLog.Info("2FA login with 'remember device' enabled, issuing device token")
			deviceSessionID := random.MakeRandomURLSafeB64(21)
			err = sessionRepository.InsertSession(deviceSessionID, domain_model.USER_SESSION_TYPE_LOGIN, user.AppUserID, domain_model.DEVICE_SESSION_DURATION)
			if err != nil {
				return nil, errors.Wrap("error inserting device session", err)
			}

			sessionCookieService.SetSessionCookie(deviceSessionID, domain_model.USER_SESSION_TYPE_REMEMBER_DEVICE)
		}

		securityLog.Info("Login passed with 2FA token")
	} else {
		maybeDeviceSessionID, err := sessionCookieService.GetSessionCookie(domain_model.USER_SESSION_TYPE_LOGIN)
		if err != nil {
			return nil, errors.Wrap("issue reading device session cookie", err)
		}
		if maybeDeviceSessionID == "" {
			return &LoginWithSecondFactorResponseTO{}, nil
		}

		deviceSession, err := sessionRepository.GetSessionAndUser(maybeDeviceSessionID, domain_model.USER_SESSION_TYPE_REMEMBER_DEVICE)

		if err != nil {
			return nil, errors.Wrap("fetching device session failed", err)
		}
		if deviceSession.UserSessionID == "" {
			return &LoginWithSecondFactorResponseTO{}, nil
		}
		securityLog.Info("Login passed with device token cookie")
	}

	sessionID := random.MakeRandomURLSafeB64(21)
	if err = sessionRepository.InsertSession(sessionID, domain_model.USER_SESSION_TYPE_LOGIN, user.AppUserID, domain_model.LOGIN_SESSION_DURATION); err != nil {
		return nil, errors.Wrap("error inserting login session", err)
	}

	sessionCookieService.SetSessionCookie(sessionID, domain_model.USER_SESSION_TYPE_LOGIN)
	return &LoginWithSecondFactorResponseTO{LoggedIn: true}, nil
}
