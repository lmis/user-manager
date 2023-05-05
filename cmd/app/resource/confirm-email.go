package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"
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

func ConfirmEmail(ctx *gin.Context, r *ginext.RequestContext, request EmailConfirmationTO) (EmailConfirmationResponseTO, error) {
	securityLog := r.SecurityLog
	user := r.UserSession.User

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return EmailConfirmationResponseTO{EmailConfirmationResponseAlreadyConfirmed}, nil
	}

	if user.EmailVerificationToken == "" {
		return EmailConfirmationResponseTO{}, errors.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken {
		securityLog.Info("Invalid email verification token")
		return EmailConfirmationResponseTO{EmailConfirmationResponseInvalidToken}, nil
	}

	if err := repository.SetEmailToVerified(ctx, r.Tx, user.AppUserID); err != nil {
		return EmailConfirmationResponseTO{}, errors.Wrap("issue setting email to verified", err)
	}

	return EmailConfirmationResponseTO{EmailConfirmationResponseNewlyConfirmed}, nil
}

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func RetriggerVerificationEmail(ctx *gin.Context, r *ginext.RequestContext) (RetriggerConfirmationEmailResponseTO, error) {
	user := r.UserSession.User
	securityLog := r.SecurityLog

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return RetriggerConfirmationEmailResponseTO{Sent: false}, nil
	}

	if err := repository.UpdateUserEmailVerificationToken(ctx, r.Tx, user.AppUserID, random.MakeRandomURLSafeB64(21)); err != nil {
		return RetriggerConfirmationEmailResponseTO{}, errors.Wrap("issue updating token", err)
	}

	if user.EmailVerificationToken == "" {
		return RetriggerConfirmationEmailResponseTO{}, errors.Errorf("missing email verification token")
	}
	if err := service.SendVerificationEmail(ctx, r, user.Language, user.Email, user.EmailVerificationToken); err != nil {
		return RetriggerConfirmationEmailResponseTO{}, errors.Wrap("error sending verification email", err)
	}

	return RetriggerConfirmationEmailResponseTO{Sent: true}, nil
}
