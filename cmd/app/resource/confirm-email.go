package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/mail"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

func RegisterEmailConfirmationResource(group *gin.RouterGroup) {
	group.POST("confirm-email", ginext.WrapEndpoint(ConfirmEmail))
	group.POST("re-trigger-confirmation-email", ginext.WrapEndpointWithoutRequestBody(RetriggerVerificationEmail))
}

type EmailConfirmationTO struct {
	Token string `json:"token"`
}

type EmailConfirmationStatus string

const (
	EmailConfirmationResponseAlreadyConfirmed EmailConfirmationStatus = "already-confirmed"
	EmailConfirmationResponseNewlyConfirmed   EmailConfirmationStatus = "newly-confirmed"
	EmailConfirmationResponseInvalidToken     EmailConfirmationStatus = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status EmailConfirmationStatus `json:"status"`
}

func ConfirmEmail(ctx *gin.Context, r *dm.RequestContext, request EmailConfirmationTO) (EmailConfirmationResponseTO, error) {
	securityLog := r.SecurityLog
	user := r.User

	if !user.IsPresent() {
		return EmailConfirmationResponseTO{}, errs.Error("no user")
	}

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return EmailConfirmationResponseTO{EmailConfirmationResponseAlreadyConfirmed}, nil
	}

	if user.EmailVerificationToken == "" {
		return EmailConfirmationResponseTO{}, errs.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken {
		securityLog.Info("Invalid email verification token")
		return EmailConfirmationResponseTO{EmailConfirmationResponseInvalidToken}, nil
	}

	if err := users.SetEmailToVerified(ctx, r.Database, user.ID()); err != nil {
		return EmailConfirmationResponseTO{}, errs.Wrap("issue setting email to verified", err)
	}

	return EmailConfirmationResponseTO{EmailConfirmationResponseNewlyConfirmed}, nil
}

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func RetriggerVerificationEmail(ctx *gin.Context, r *dm.RequestContext) (RetriggerConfirmationEmailResponseTO, error) {
	user := r.User
	securityLog := r.SecurityLog

	if !user.IsPresent() {
		return RetriggerConfirmationEmailResponseTO{}, errs.Error("no user")
	}

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return RetriggerConfirmationEmailResponseTO{Sent: false}, nil
	}

	if err := users.UpdateUserEmailVerificationToken(ctx, r.Database, user.ID(), random.MakeRandomURLSafeB64(21)); err != nil {
		return RetriggerConfirmationEmailResponseTO{}, errs.Wrap("issue updating token", err)
	}

	if user.EmailVerificationToken == "" {
		return RetriggerConfirmationEmailResponseTO{}, errs.Errorf("missing email verification token")
	}
	if err := mail.SendVerificationEmail(ctx, r, user.Email, user.EmailVerificationToken); err != nil {
		return RetriggerConfirmationEmailResponseTO{}, errs.Wrap("error sending verification email", err)
	}

	return RetriggerConfirmationEmailResponseTO{Sent: true}, nil
}
