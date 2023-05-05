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

func RegisterResetPasswordResource(group *gin.RouterGroup) {
	group.POST("request-password-reset", ginext.WrapEndpointWithoutResponseBody(RequestPasswordReset))
	group.POST("reset-password", ginext.WrapEndpoint(ResetPassword))
}

type PasswordResetRequestTO struct {
	Email string `json:"email"`
}

func RequestPasswordReset(ctx *gin.Context, r *ginext.RequestContext, requestTO *PasswordResetRequestTO) error {
	securityLog := r.SecurityLog

	securityLog.Info("Password reset requested")

	user, err := repository.GetUserForEmail(ctx, r.Tx, requestTO.Email)
	if err != nil {
		return errors.Wrap("error finding user for email", err)
	}
	if user.AppUserID == 0 {
		securityLog.Info("Password reset request for non-existing email")
		return nil
	}

	token := random.MakeRandomURLSafeB64(21)
	if err := repository.SetPasswordResetToken(ctx, r.Tx, user.AppUserID, token, time.Now().Add(dm.PasswordResetTokenDuration)); err != nil {
		return errors.Wrap("issue persisting password reset token", err)
	}

	if err := service.SendResetPasswordEmail(ctx, r, user.Language, user.Email, token); err != nil {
		return errors.Wrap("error sending password reset email", err)
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

func ResetPassword(ctx *gin.Context, r *ginext.RequestContext, requestTO ResetPasswordTO) (ResetPasswordResponseTO, error) {
	securityLog := r.SecurityLog
	securityLog.Info("Password reset")

	user, err := repository.GetUserForEmail(ctx, r.Tx, requestTO.Email)
	if err != nil {
		return ResetPasswordResponseTO{}, errors.Wrap("error finding user", err)
	}

	if user.AppUserID == 0 {
		securityLog.Info("Password reset attempt for non-existing email")
		return ResetPasswordResponseTO{ResetPasswordResponseInvalid}, nil
	}
	if user.PasswordResetToken == "" || user.PasswordResetToken != requestTO.Token {
		securityLog.Info("Password reset attempt with wrong token")
		return ResetPasswordResponseTO{ResetPasswordResponseInvalid}, nil
	}
	if user.PasswordResetTokenValidUntil.Before(time.Now()) {
		securityLog.Info("Password reset attempt with expired token")
		return ResetPasswordResponseTO{ResetPasswordResponseInvalid}, nil
	}

	hash, err := service.HashPassword(requestTO.NewPassword)
	if err != nil {
		return ResetPasswordResponseTO{}, errors.Wrap("issue making password hash", err)
	}

	if err := repository.SetPasswordHash(ctx, r.Tx, user.AppUserID, hash); err != nil {
		return ResetPasswordResponseTO{}, errors.Wrap("issue setting password hash", err)
	}

	return ResetPasswordResponseTO{ResetPasswordResponseSuccess}, nil
}
