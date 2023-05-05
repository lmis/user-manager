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

func RegisterSettingsResource(group *gin.RouterGroup) {
	group.POST("language", ginext.WrapEndpointWithoutResponseBody(SetLanguage))
	group.POST("confirm-email-change", ginext.WrapEndpoint(ConfirmEmailChange))
	group.POST("enter-sudo-mode", ginext.WrapEndpoint(EnterSudoMode))
}

type LanguageTO struct {
	Language dm.UserLanguage `json:"language"`
}

func SetLanguage(ctx *gin.Context, r *ginext.RequestContext, requestTO *LanguageTO) error {
	userSession := r.UserSession

	if userSession.UserSessionID == "" {
		return errors.Error("no user")
	}

	language := requestTO.Language
	if !language.IsValid() {
		return errors.Errorf("invalid language %s", language)
	}

	if err := repository.SetLanguage(ctx, r.Tx, userSession.User.AppUserID, language); err != nil {
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

func EnterSudoMode(ctx *gin.Context, r *ginext.RequestContext, requestTO *SudoTO) (*SudoResponseTO, error) {
	securityLog := r.SecurityLog
	userSession := r.UserSession

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
	if err := repository.InsertSession(ctx, r.Tx, sessionID, dm.UserSessionTypeSudo, user.AppUserID, dm.SudoSessionDuration); err != nil {
		return nil, errors.Wrap("error inserting session", err)
	}

	service.SetSessionCookie(ctx, r.Config, sessionID, dm.UserSessionTypeSudo)
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

func ConfirmEmailChange(ctx *gin.Context, r *ginext.RequestContext, request EmailChangeConfirmationTO) (EmailChangeConfirmationResponseTO, error) {
	securityLog := r.SecurityLog
	userSession := r.UserSession

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

	if err := repository.SetEmailAndClearNextEmail(ctx, r.Tx, user.AppUserID, user.NextEmail); err != nil {
		return EmailChangeConfirmationResponseTO{}, errors.Wrap("issue setting email ", err)
	}

	return EmailChangeConfirmationResponseTO{EmailChangeResponseNewEmailConfirmed}, nil
}
