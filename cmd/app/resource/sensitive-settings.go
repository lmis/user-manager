package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"
	"user-manager/util/nullable"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

type SensitiveSettingsResource struct {
	securityLog      domain_model.SecurityLog
	mailQueueService *service.MailQueueService
	userSession      nullable.Nullable[domain_model.UserSession]
	userRepository   *repository.UserRepository
}

func ProvideSensitiveSettingsResource(
	securityLog domain_model.SecurityLog,
	mailQueueService *service.MailQueueService,
	userSession nullable.Nullable[domain_model.UserSession],
	userRepository *repository.UserRepository,
) *SensitiveSettingsResource {
	return &SensitiveSettingsResource{securityLog, mailQueueService, userSession, userRepository}
}

func RegisterSensitiveSettingsResource(group *gin.RouterGroup) {
	group.POST("initiate-email-change", ginext.WrapEndpointWithoutResponseBody(InitializeSensitiveSettingsResource, (*SensitiveSettingsResource).InitiateEmailChange))
}

type ChangeEmailTO struct {
	NewEmail string `json:"newEmail"`
}

func (r *SensitiveSettingsResource) InitiateEmailChange(requestTO *ChangeEmailTO) error {
	securityLog := r.securityLog
	userSession := r.userSession
	mailQueueService := r.mailQueueService
	userRepository := r.userRepository

	user := userSession.OrPanic().User
	nextEmail := requestTO.NewEmail

	securityLog.Info("Changing user email")

	verificationToken := random.MakeRandomURLSafeB64(21)
	if err := userRepository.SetNextEmail(user.AppUserID, nextEmail, verificationToken); err != nil {
		return errors.Wrap("issue setting next email for user", err)
	}

	if err := mailQueueService.SendChangeVerificationEmail(user.Language, nextEmail, verificationToken); err != nil {
		return errors.Wrap("error sending change verification email", err)
	}
	if err := mailQueueService.SendChangeNotificationEmail(user.Language, user.Email, nextEmail); err != nil {
		return errors.Wrap("error sending change notification email", err)
	}

	return nil
}
