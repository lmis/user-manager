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

func RegisterSignUpResource(group *gin.RouterGroup) {
	group.POST("sign-up", ginext.WrapEndpointWithoutResponseBody(SignUp))
}

type SignUpTO struct {
	UserName string `json:"userName"`
	Language string `json:"language"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func SignUp(ctx *gin.Context, r *ginext.RequestContext, requestTO SignUpTO) error {
	securityLog := r.SecurityLog

	user, err := repository.GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return errors.Wrap("error fetching user", err)
	}
	if !user.IsPresent() {
		securityLog.Info("User already exists")
		if err = service.SendSignUpAttemptEmail(ctx, r, user.Language, user.Email); err != nil {
			return errors.Wrap("error sending signup attempted email", err)
		}
		return nil
	}

	credentials, err := service.MakeCredentials(requestTO.Password)
	if err != nil {
		return errors.Wrap("error hashing password", err)
	}

	language := dm.UserLanguage(requestTO.Language)
	if !language.IsValid() {
		return errors.Errorf("unsupported language \"%s\"", string(language))
	}

	verificationToken := random.MakeRandomURLSafeB64(21)
	if err = repository.InsertUser(ctx, r.Database, dm.UserInsert{
		Language:               language,
		UserName:               requestTO.UserName,
		Credentials:            credentials,
		Email:                  requestTO.Email,
		EmailVerificationToken: verificationToken,
		UserRoles:              []dm.UserRole{dm.UserRoleUser},
	}); err != nil {
		return errors.Wrap("error inserting user", err)
	}

	if err = service.SendVerificationEmail(ctx, r, language, requestTO.Email, verificationToken); err != nil {
		return errors.Wrap("error sending verification email", err)
	}

	return nil
}
