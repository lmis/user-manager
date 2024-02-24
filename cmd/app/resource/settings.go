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
	user := r.User

	if !user.IsPresent() {
		return errors.Error("no user")
	}

	language := requestTO.Language
	if !language.IsValid() {
		return errors.Errorf("invalid language %s", language)
	}

	if err := repository.SetLanguage(ctx, r.Database, user.ID, language); err != nil {
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
	user := r.User

	if !user.IsPresent() {
		return nil, errors.Error("no user")
	}

	if !service.VerifyCredentials(requestTO.Password, user.Credentials) {
		securityLog.Info("Password mismatch in sudo attempt for user %s", user.ID)
		return &SudoResponseTO{}, nil
	}

	securityLog.Info("Entering sudo mode")
	session := dm.UserSession{
		Token:     dm.UserSessionToken(random.MakeRandomURLSafeB64(21)),
		Type:      dm.UserSessionTypeSudo,
		TimeoutAt: time.Now().Add(dm.SudoSessionDuration),
	}
	if err := repository.InsertSession(ctx, r.Database, user.ID, session); err != nil {
		return nil, errors.Wrap("error inserting session", err)
	}

	service.SetSessionCookie(ctx, r.Config, string(session.Token), session.Type)
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
	user := r.User

	if user.IsPresent() {
		return EmailChangeConfirmationResponseTO{}, errors.Error("no user")
	}

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

	if err := repository.SetEmailAndClearNextEmail(ctx, r.Database, user.ID, user.NextEmail); err != nil {
		return EmailChangeConfirmationResponseTO{}, errors.Wrap("issue setting email ", err)
	}

	return EmailChangeConfirmationResponseTO{EmailChangeResponseNewEmailConfirmed}, nil
}
