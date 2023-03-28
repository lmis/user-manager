package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

type EmailConfirmationResource struct {
	securityLog      dm.SecurityLog
	mailQueueService *service.MailQueueService
	userSession      dm.UserSession
	userRepository   *repository.UserRepository
}

func ProvideEmailConfirmationResource(
	securityLog dm.SecurityLog,
	mailQueueService *service.MailQueueService,
	userSession dm.UserSession,
	userRepository *repository.UserRepository,
) *EmailConfirmationResource {
	return &EmailConfirmationResource{securityLog, mailQueueService, userSession, userRepository}
}

func RegisterEmailConfirmationResource(group *gin.RouterGroup) {
	group.POST("confirm-email", ginext.WrapEndpoint(InitializeEmailConfirmationResource, (*EmailConfirmationResource).ConfirmEmail))
	group.POST("re-trigger-confirmation-email", ginext.WrapEndpointWithoutRequestBody(InitializeEmailConfirmationResource, (*EmailConfirmationResource).RetriggerVerificationEmail))
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

func (r *EmailConfirmationResource) ConfirmEmail(request EmailConfirmationTO) (EmailConfirmationResponseTO, error) {
	securityLog := r.securityLog
	user := r.userSession.User
	userRepository := r.userRepository

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

	if err := userRepository.SetEmailToVerified(user.AppUserID); err != nil {
		return EmailConfirmationResponseTO{}, errors.Wrap("issue setting email to verified", err)
	}

	return EmailConfirmationResponseTO{EmailConfirmationResponseNewlyConfirmed}, nil
}

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func (r *EmailConfirmationResource) RetriggerVerificationEmail() (RetriggerConfirmationEmailResponseTO, error) {
	user := r.userSession.User
	securityLog := r.securityLog
	userRepository := r.userRepository
	mailQueueService := r.mailQueueService

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return RetriggerConfirmationEmailResponseTO{Sent: false}, nil
	}

	if err := userRepository.UpdateUserEmailVerificationToken(user.AppUserID, random.MakeRandomURLSafeB64(21)); err != nil {
		return RetriggerConfirmationEmailResponseTO{}, errors.Wrap("issue updating token", err)
	}

	if user.EmailVerificationToken == "" {
		return RetriggerConfirmationEmailResponseTO{}, errors.Errorf("missing email verification token")
	}
	if err := mailQueueService.SendVerificationEmail(user.Language, user.Email, user.EmailVerificationToken); err != nil {
		return RetriggerConfirmationEmailResponseTO{}, errors.Wrap("error sending verification email", err)
	}

	return RetriggerConfirmationEmailResponseTO{Sent: true}, nil
}
