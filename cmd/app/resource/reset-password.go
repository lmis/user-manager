package resource

import (
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	"user-manager/cmd/app/service/mail"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

func RegisterResetPasswordResource(group *gin.RouterGroup) {
	group.POST("request-password-reset", ginext.WrapEndpointWithoutResponseBody(RequestPasswordReset))
	group.POST("reset-password", ginext.WrapEndpoint(ResetPassword))
}

type PasswordResetRequestTO struct {
	Email string `json:"email"`
}

func RequestPasswordReset(ctx *gin.Context, r *dm.RequestContext, requestTO *PasswordResetRequestTO) error {
	logger := r.Logger

	logger.Info("Password reset requested")

	user, err := users.GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return errs.Wrap("error finding user for email", err)
	}
	if !user.IsPresent() {
		logger.Info("Password reset request for non-existing email")
		return nil
	}

	token := random.MakeRandomURLSafeB64(21)
	if err := users.SetPasswordResetToken(ctx, r.Database, user.ID(), token, time.Now().Add(dm.PasswordResetTokenDuration)); err != nil {
		return errs.Wrap("issue persisting password reset token", err)
	}

	if err := mail.SendResetPasswordEmail(ctx, r, user.Email, token); err != nil {
		return errs.Wrap("error sending password reset email", err)
	}
	return nil
}

type ResetPasswordTO struct {
	Email       string `json:"email"`
	Token       string `json:"token"`
	NewPassword []byte `json:"newPassword"`
}

type ResetPasswordStatus string

const (
	ResetPasswordResponseSuccess ResetPasswordStatus = "success"
	ResetPasswordResponseInvalid ResetPasswordStatus = "invalid-token"
)

type ResetPasswordResponseTO struct {
	Status ResetPasswordStatus `json:"status"`
}

func ResetPassword(ctx *gin.Context, r *dm.RequestContext, requestTO ResetPasswordTO) (ResetPasswordResponseTO, error) {
	logger := r.Logger
	logger.Info("Password reset")

	user, err := users.GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return ResetPasswordResponseTO{}, errs.Wrap("error finding user", err)
	}

	if !user.IsPresent() {
		logger.Info("Password reset attempt for non-existing email")
		return ResetPasswordResponseTO{ResetPasswordResponseInvalid}, nil
	}
	if user.PasswordResetToken == "" || user.PasswordResetToken != requestTO.Token {
		logger.Info("Password reset attempt with wrong token")
		return ResetPasswordResponseTO{ResetPasswordResponseInvalid}, nil
	}
	if user.PasswordResetTokenValidUntil.Before(time.Now()) {
		logger.Info("Password reset attempt with expired token")
		return ResetPasswordResponseTO{ResetPasswordResponseInvalid}, nil
	}

	hash, err := auth.MakeCredentials(requestTO.NewPassword)
	if err != nil {
		return ResetPasswordResponseTO{}, errs.Wrap("issue making password hash", err)
	}

	if err := users.SetCredentials(ctx, r.Database, user.ID(), hash); err != nil {
		return ResetPasswordResponseTO{}, errs.Wrap("issue setting password hash", err)
	}

	return ResetPasswordResponseTO{ResetPasswordResponseSuccess}, nil
}
