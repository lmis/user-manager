package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util"
	"user-manager/util/nullable"

	"github.com/gin-gonic/gin"
)

type EmailConfirmationResource struct {
	securityLog      domain_model.SecurityLog
	mailQueueService *service.MailQueueService
	userSession      nullable.Nullable[*domain_model.UserSession]
	userRepository   *repository.UserRepository
}

func ProvideEmailConfirmationResource(
	securityLog domain_model.SecurityLog,
	mailQueueService *service.MailQueueService,
	userSession nullable.Nullable[*domain_model.UserSession],
	userRepository *repository.UserRepository,
) *EmailConfirmationResource {
	return &EmailConfirmationResource{securityLog, mailQueueService, userSession, userRepository}
}

func RegisterEmailConfirmationResource(group *gin.RouterGroup) {
	group.POST("confirm-email", ginext.WrapEndpoint(InitializeEmailConfirmationResource, (*EmailConfirmationResource).Post))
	group.POST("re-trigger-confirmation-email", ginext.WrapEndpointWithoutRequestBody(InitializeEmailConfirmationResource, (*EmailConfirmationResource).RetriggerVerificationEmail))
}

type EmailConfirmationTO struct {
	Token string `json:"token"`
}

type EmailConfirmationStatus string

const (
	AlreadyConfirmed EmailConfirmationStatus = "already-confirmed"
	NewlyConfirmed   EmailConfirmationStatus = "newly-confirmed"
	InvalidToken     EmailConfirmationStatus = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status EmailConfirmationStatus `json:"status"`
}

func (r *EmailConfirmationResource) Post(request *EmailConfirmationTO) (*EmailConfirmationResponseTO, error) {
	securityLog := r.securityLog
	user := r.userSession.OrPanic().User
	userRepository := r.userRepository

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return &EmailConfirmationResponseTO{
			Status: AlreadyConfirmed,
		}, nil
	}

	if user.EmailVerificationToken.IsEmpty() {
		return nil, util.Error("no verification token present on database")
	}

	if request.Token != user.EmailVerificationToken.OrPanic() {
		securityLog.Info("Invalid email verification token")
		return &EmailConfirmationResponseTO{
			Status: InvalidToken,
		}, nil
	}

	if err := userRepository.SetEmailToVerified(user.AppUserID); err != nil {
		return nil, util.Wrap("issue setting email to verified", err)
	}

	return &EmailConfirmationResponseTO{
		Status: NewlyConfirmed,
	}, nil
}

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func (r *EmailConfirmationResource) RetriggerVerificationEmail() (*RetriggerConfirmationEmailResponseTO, error) {
	user := r.userSession.OrPanic().User
	securityLog := r.securityLog
	userRepository := r.userRepository
	mailQueueService := r.mailQueueService

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		return &RetriggerConfirmationEmailResponseTO{Sent: false}, nil
	}

	if err := userRepository.UpdateUserEmailVerificationToken(user.AppUserID, util.MakeRandomURLSafeB64(21)); err != nil {
		return nil, util.Wrap("issue updating token", err)
	}

	if err := mailQueueService.SendVerificationEmail(user.Language, user.Email, user.EmailVerificationToken.OrPanic()); err != nil {
		return nil, util.Wrap("error sending verification email", err)
	}

	return &RetriggerConfirmationEmailResponseTO{Sent: true}, nil
}
