package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"user-manager/util/random"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SettingsResource struct {
	securityLog          dm.SecurityLog
	sessionCookieService *service.SessionCookieService
	userSession          dm.UserSession
	userRepository       *repository.UserRepository
	sessionRepository    *repository.SessionRepository
}

func ProvideSettingsResource(
	securityLog dm.SecurityLog,
	sessionCookieService *service.SessionCookieService,
	userSession dm.UserSession,
	userRepository *repository.UserRepository,
	sessionRepository *repository.SessionRepository,
) *SettingsResource {
	return &SettingsResource{securityLog, sessionCookieService, userSession, userRepository, sessionRepository}
}

func RegisterSettingsResource(group *gin.RouterGroup) {
	group.POST("language", ginext.WrapEndpointWithoutResponseBody(InitializeSettingsResource, (*SettingsResource).SetLanguage))
	group.POST("confirm-email-change", ginext.WrapEndpoint(InitializeSettingsResource, (*SettingsResource).ConfirmEmailChange))
	group.POST("enter-sudo-mode", ginext.WrapEndpoint(InitializeSettingsResource, (*SettingsResource).EnterSudoMode))
}

type LanguageTO struct {
	Language dm.UserLanguage `json:"language"`
}

func (r *SettingsResource) SetLanguage(requestTO *LanguageTO) error {
	userRepository := r.userRepository
	userSession := r.userSession

	if userSession.UserSessionID == "" {
		return errors.Error("no user")
	}

	language := requestTO.Language
	if !language.IsValid() {
		return errors.Errorf("invalid language %s", language)
	}

	if err := userRepository.SetLanguage(userSession.User.AppUserID, language); err != nil {
		return errors.Wrap("error updating language", err)
	}
	return nil
}

type SudoTO struct {
	Password []byte `json:"password"`
}

type SudoResponseTO struct {
	Success bool `json:"success"`
}

func (r *SettingsResource) EnterSudoMode(requestTO *SudoTO) (*SudoResponseTO, error) {
	securityLog := r.securityLog
	userSession := r.userSession
	sessionRepository := r.sessionRepository
	sessionCookieService := r.sessionCookieService

	if userSession.UserSessionID == "" {
		return nil, errors.Error("no user")
	}

	user := userSession.User
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.Password); err != nil {
		securityLog.Info("Password mismatch in sudo attempt for user %s", user.AppUserID)
		return &SudoResponseTO{}, nil
	}

	securityLog.Info("Entering sudo mode")
	sessionID := random.MakeRandomURLSafeB64(21)
	if err := sessionRepository.InsertSession(sessionID, dm.UserSessionTypeSudo, user.AppUserID, dm.SudoSessionDuration); err != nil {
		return nil, errors.Wrap("error inserting session", err)
	}

	sessionCookieService.SetSessionCookie(sessionID, dm.UserSessionTypeSudo)
	return &SudoResponseTO{Success: true}, nil
}

type EmailChangeConfirmationTO struct {
	Token string `json:"token"`
}

type EmailChangeStatus string

const (
	EmailChangeResponseNoChangeInProgress EmailChangeStatus = "no-change-in-progress"
	EmailChangeResponseInvalidToken       EmailChangeStatus = "invalid-token"
	EmailChangeResponseNewEmailConfirmed  EmailChangeStatus = "new-email-confirmed"
)

type EmailChangeConfirmationResponseTO struct {
	Status EmailChangeStatus `json:"status"`
}

func (r *SettingsResource) ConfirmEmailChange(request EmailChangeConfirmationTO) (EmailChangeConfirmationResponseTO, error) {
	securityLog := r.securityLog
	userSession := r.userSession
	userRepository := r.userRepository

	if userSession.UserSessionID == "" {
		return EmailChangeConfirmationResponseTO{}, errors.Error("no user")
	}
	user := userSession.User

	if user.NextEmail == "" {
		return EmailChangeConfirmationResponseTO{EmailChangeResponseNoChangeInProgress}, nil
	}

	if user.EmailVerificationToken == "" {
		return EmailChangeConfirmationResponseTO{}, errors.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken {
		securityLog.Info("Invalid email verification token")
		return EmailChangeConfirmationResponseTO{EmailChangeResponseInvalidToken}, nil
	}

	if err := userRepository.SetEmailAndClearNextEmail(user.AppUserID, user.NextEmail); err != nil {
		return EmailChangeConfirmationResponseTO{}, errors.Wrap("issue setting email ", err)
	}

	return EmailChangeConfirmationResponseTO{EmailChangeResponseNewEmailConfirmed}, nil
}
