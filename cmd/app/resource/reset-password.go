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
)

type ResetPasswordResource struct {
	securityLog      domain_model.SecurityLog
	authService      *service.AuthService
	mailQueueService *service.MailQueueService
	userRepository   *repository.UserRepository
}

func ProvideResetPasswordResource(
	securityLog domain_model.SecurityLog,
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
	if err := userRepository.SetPasswordResetToken(user.AppUserID, token, time.Now().Add(domain_model.PASSWORD_RESET_TOKEN_DURATION)); err != nil {
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
	RESET_PASSWORD_RESPONSE_SUCCESS ResetPasswordStatus = "success"
	RESET_PASSWORD_RESPONSE_INVALID ResetPasswordStatus = "invalid-token"
)

type ResetPasswordResponseTO struct {
	Status ResetPasswordStatus `json:"status"`
}

func (r *ResetPasswordResource) ResetPassword(requestTO *ResetPasswordTO) (*ResetPasswordResponseTO, error) {
	securityLog := r.securityLog
	userRepository := r.userRepository
	authService := r.authService
	securityLog.Info("Password reset")

	user, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return nil, errors.Wrap("error finding user", err)
	}

	if user.AppUserID == 0 {
		securityLog.Info("Password reset attempt for non-existing email")
		return &ResetPasswordResponseTO{RESET_PASSWORD_RESPONSE_INVALID}, nil
	}
	if user.PasswordResetToken == "" || user.PasswordResetToken != requestTO.Token {
		securityLog.Info("Password reset attempt with wrong token")
		return &ResetPasswordResponseTO{RESET_PASSWORD_RESPONSE_INVALID}, nil
	}
	if user.PasswordResetTokenValidUntil.Before(time.Now()) {
		securityLog.Info("Password reset attempt with expired token")
		return &ResetPasswordResponseTO{RESET_PASSWORD_RESPONSE_INVALID}, nil
	}

	hash, err := authService.Hash(requestTO.NewPassword)
	if err != nil {
		return nil, errors.Wrap("issue making password hash", err)
	}

	if err := userRepository.SetPasswordHash(user.AppUserID, hash); err != nil {
		return nil, errors.Wrap("issue setting password hash", err)
	}

	return &ResetPasswordResponseTO{RESET_PASSWORD_RESPONSE_SUCCESS}, nil
}
