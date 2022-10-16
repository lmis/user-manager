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

type RetriggerConfirmationEmailResource struct {
	userRepository   *repository.UserRepository
	mailQueueService *service.MailQueueService
	securityLog      domain_model.SecurityLog
	userSession      nullable.Nullable[*domain_model.UserSession]
}

func ProvideRetriggerConfirmationEmailResource(
	userRepository *repository.UserRepository,
	mailQueueService *service.MailQueueService,
	securityLog domain_model.SecurityLog,
	userSession nullable.Nullable[*domain_model.UserSession],
) *RetriggerConfirmationEmailResource {
	return &RetriggerConfirmationEmailResource{userRepository, mailQueueService, securityLog, userSession}
}

func RegisterRetriggerConfirmationEmailResource(group *gin.RouterGroup) {
	group.POST("re-trigger-confirmation-email", ginext.WrapEndpointWithoutRequestBody(InitializeRetriggerConfirmationEmailResource, (*RetriggerConfirmationEmailResource).Retrigger))
}

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func (r *RetriggerConfirmationEmailResource) Retrigger() (*RetriggerConfirmationEmailResponseTO, error) {
	user := r.userSession.OrPanic().User

	if user.EmailVerified {
		r.securityLog.Info("Email already verified")
		return &RetriggerConfirmationEmailResponseTO{Sent: false}, nil
	}

	if err := r.userRepository.UpdateUserEmailVerificationToken(user.AppUserID, util.MakeRandomURLSafeB64(21)); err != nil {
		return nil, util.Wrap("issue updating token", err)
	}

	if err := r.mailQueueService.SendVerificationEmail(user); err != nil {
		return nil, util.Wrap("error sending verification email", err)
	}

	return &RetriggerConfirmationEmailResponseTO{Sent: true}, nil
}
