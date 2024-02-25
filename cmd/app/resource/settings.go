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
)

func RegisterSettingsResource(group *gin.RouterGroup) {
	group.POST("confirm-email-change", ginext.WrapEndpoint(ConfirmEmailChange))
	group.POST("enter-sudo-mode", ginext.WrapEndpoint(EnterSudoMode))
}

type SudoTO struct {
	Password []byte `json:"password"`
}

type SudoResponseTO struct {
	Success bool `json:"success"`
}

func EnterSudoMode(ctx *gin.Context, r *dm.RequestContext, requestTO *SudoTO) (*SudoResponseTO, error) {
	securityLog := r.SecurityLog
	user := r.User

	if !user.IsPresent() {
		return nil, errs.Error("no user")
	}

	if !auth.VerifyCredentials(requestTO.Password, user.Credentials) {
		securityLog.Info(fmt.Sprintf("Password mismatch in sudo attempt for user %s", user.ID()))
		return &SudoResponseTO{}, nil
	}

	securityLog.Info("Entering sudo mode")
	session := dm.UserSession{
		Token:     dm.UserSessionToken(random.MakeRandomURLSafeB64(21)),
		Type:      dm.UserSessionTypeSudo,
		TimeoutAt: time.Now().Add(dm.SudoSessionDuration),
	}
	if err := auth.InsertSession(ctx, r.Database, user.ID(), session); err != nil {
		return nil, errs.Wrap("error inserting session", err)
	}

	auth.SetSessionCookie(ctx, r.Config, string(session.Token), session.Type)
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

func ConfirmEmailChange(ctx *gin.Context, r *dm.RequestContext, request EmailChangeConfirmationTO) (EmailChangeConfirmationResponseTO, error) {
	securityLog := r.SecurityLog
	user := r.User

	if user.IsPresent() {
		return EmailChangeConfirmationResponseTO{}, errs.Error("no user")
	}

	if user.NextEmail == "" {
		return EmailChangeConfirmationResponseTO{EmailChangeResponseNoChangeInProgress}, nil
	}

	if user.EmailVerificationToken == "" {
		return EmailChangeConfirmationResponseTO{}, errs.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken {
		securityLog.Info("Invalid email verification token")
		return EmailChangeConfirmationResponseTO{EmailChangeResponseInvalidToken}, nil
	}

	if err := users.SetEmailAndClearNextEmail(ctx, r.Database, user.ID(), user.NextEmail); err != nil {
		return EmailChangeConfirmationResponseTO{}, errs.Wrap("issue setting email ", err)
	}

	return EmailChangeConfirmationResponseTO{EmailChangeResponseNewEmailConfirmed}, nil
}
