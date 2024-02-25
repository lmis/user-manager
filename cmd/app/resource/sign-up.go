package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	"user-manager/cmd/app/service/mail"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

func RegisterSignUpResource(group *gin.RouterGroup) {
	group.POST("sign-up", ginext.WrapEndpointWithoutResponseBody(SignUp))
}

type SignUpTO struct {
	UserName string `json:"userName"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func SignUp(ctx *gin.Context, r *dm.RequestContext, requestTO SignUpTO) error {
	securityLog := r.SecurityLog

	user, err := users.GetUserForEmail(ctx, r.Database, requestTO.Email)
	if err != nil {
		return errs.Wrap("error fetching user", err)
	}
	if user.IsPresent() {
		securityLog.Info("User already exists")
		if err = mail.SendSignUpAttemptEmail(ctx, r, user.Email); err != nil {
			return errs.Wrap("error sending signup attempted email", err)
		}
		return nil
	}

	credentials, err := auth.MakeCredentials(requestTO.Password)
	if err != nil {
		return errs.Wrap("error hashing password", err)
	}

	verificationToken := random.MakeRandomURLSafeB64(21)
	if err = users.InsertUser(ctx, r.Database, dm.UserInsert{
		UserName:               requestTO.UserName,
		Credentials:            credentials,
		Email:                  requestTO.Email,
		EmailVerificationToken: verificationToken,
		UserRoles:              []dm.UserRole{dm.UserRoleUser},
	}); err != nil {
		return errs.Wrap("error inserting user", err)
	}

	if err = mail.SendVerificationEmail(ctx, r, requestTO.Email, verificationToken); err != nil {
		return errs.Wrap("error sending verification email", err)
	}

	return nil
}
