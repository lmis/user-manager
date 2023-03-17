package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

type SignUpResource struct {
	userRepository   *repository.UserRepository
	mailQueueService *service.MailQueueService
	authService      *service.AuthService
	securityLog      domain_model.SecurityLog
}

func ProvideSignUpResource(
	userRepository *repository.UserRepository,
	mailQueueService *service.MailQueueService,
	authService *service.AuthService,
	securityLog domain_model.SecurityLog,
) *SignUpResource {
	return &SignUpResource{userRepository, mailQueueService, authService, securityLog}
}

func RegisterSignUpResource(group *gin.RouterGroup) {
	group.POST("sign-up", ginext.WrapEndpointWithoutResponseBody(InitializeSignUpResource, (*SignUpResource).SignUp))
}

type SignUpTO struct {
	UserName string `json:"userName"`
	Language string `json:"language"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func (r *SignUpResource) SignUp(requestTO SignUpTO) error {
	securityLog := r.securityLog
	userRepository := r.userRepository
	mailQueueService := r.mailQueueService
	authService := r.authService

	user, err := userRepository.GetUserForEmail(requestTO.Email)
	if err != nil {
		return errors.Wrap("error fetching user", err)
	}
	if user.AppUserID != 0 {
		securityLog.Info("User already exists")
		if err = mailQueueService.SendSignUpAttemptEmail(user.Language, user.Email); err != nil {
			return errors.Wrap("error sending signup attempted email", err)
		}
		return nil
	}

	hash, err := authService.Hash(requestTO.Password)
	if err != nil {
		return errors.Wrap("error hashing password", err)
	}

	language := domain_model.UserLanguage(requestTO.Language)
	if !language.IsValid() {
		return errors.Errorf("unsupported language \"%s\"", string(language))
	}

	verificationToken := random.MakeRandomURLSafeB64(21)
	if err = userRepository.Insert(domain_model.USER_ROLE_USER, requestTO.UserName, requestTO.Email, false, verificationToken, hash, language); err != nil {
		return errors.Wrap("error inserting user", err)
	}

	if err = mailQueueService.SendVerificationEmail(language, requestTO.Email, verificationToken); err != nil {
		return errors.Wrap("error sending verification email", err)
	}

	return nil
}
