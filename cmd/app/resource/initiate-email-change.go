package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/mail"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

func RegisterSensitiveSettingsResource(group *gin.RouterGroup) {
	group.POST("initiate-email-change", ginext.WrapEndpointWithoutResponseBody(InitiateEmailChange))
}

type ChangeEmailTO struct {
	NewEmail string `json:"newEmail"`
}

func InitiateEmailChange(ctx *gin.Context, r *dm.RequestContext, requestTO ChangeEmailTO) error {
	securityLog := r.SecurityLog
	user := r.User

	nextEmail := requestTO.NewEmail

	if !user.IsPresent() {
		return errs.Error("missing user")

	}
	securityLog.Info("Changing user email")

	verificationToken := random.MakeRandomURLSafeB64(21)
	if err := users.SetNextEmail(ctx, r.Database, user.ID(), nextEmail, verificationToken); err != nil {
		return errs.Wrap("issue setting next email for user", err)
	}

	if err := mail.SendChangeVerificationEmail(ctx, r, nextEmail, verificationToken); err != nil {
		return errs.Wrap("error sending change verification email", err)
	}
	if err := mail.SendChangeNotificationEmail(ctx, r, user.Email, nextEmail); err != nil {
		return errs.Wrap("error sending change notification email", err)
	}

	return nil
}
