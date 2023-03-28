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

type ResetPasswordResource struct {
	securityLog      dm.SecurityLog
	authService      *service.AuthService
	mailQueueService *service.MailQueueService
	userRepository   *repository.UserRepository
}

func ProvideResetPasswordResource(
	securityLog dm.SecurityLog,
	authService *service.AuthService,
	mailQueueService *service.MailQueueService,
	userRepository *repository.UserRepository,
) *ResetPasswordResource {
	return &ResetPasswordResource{
		securityLog,
		authService,
		mailQueueService,
		userRepository,
	}
}

func RegisterResetPasswordResource(group *gin.RouterGroup) {
	group.POST("request-password-reset", ginext.WrapEndpointWithoutResponseBody(InitializeResetPasswordResource, (*ResetPasswordResource).RequestPasswordReset))
	group.POST("reset-password", ginext.WrapEndpoint(InitializeResetPasswordResource, (*ResetPasswordResource).ResetPassword))
}

type PasswordResetRequestTO struct {
	Email string `json:"email"`
}

func (r *ResetPasswordResource) RequestPasswordReset(requestTO *PasswordResetRequestTO) error {
	securityLog := r.securityLog
	mailQueueService := r.mailQueueService
	userRepository := r.userRepository

	securityLog.Info("Password reset requested")

	user, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return errors.Wrap("error finding user for email", err)
	}
	if user.AppUserID == 0 {
		securityLog.Info("Password reset request for non-existing email")
		return nil
	}

	token := random.MakeRandomURLSafeB64(21)
	if err := userRepository.SetPasswordResetToken(user.AppUserID, token, time.Now().Add(dm.PasswordResetTokenDuration)); err != nil {
		return errors.Wrap("issue persisting password reset token", err)
	}

	if err := mailQueueService.SendResetPasswordEmail(user.Language, user.Email, token); err != nil {
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

func (r *ResetPasswordResource) ResetPassword(requestTO ResetPasswordTO) (ResetPasswordResponseTO, error) {
	securityLog := r.securityLog
	userRepository := r.userRepository
	authService := r.authService
	securityLog.Info("Password reset")

	user, err := userRepository.GetUserForEmail(requestTO.Email)
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

	hash, err := authService.Hash(requestTO.NewPassword)
	if err != nil {
		return ResetPasswordResponseTO{}, errors.Wrap("issue making password hash", err)
	}

	if err := userRepository.SetPasswordHash(user.AppUserID, hash); err != nil {
		return ResetPasswordResponseTO{}, errors.Wrap("issue setting password hash", err)
	}

	return ResetPasswordResponseTO{ResetPasswordResponseSuccess}, nil
}
