package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"user-manager/util/random"

	"github.com/gin-gonic/gin"
)

func RegisterSensitiveSettingsResource(group *gin.RouterGroup) {
	group.POST("initiate-email-change", ginext.WrapEndpointWithoutResponseBody(InitiateEmailChange))
}

type ChangeEmailTO struct {
	NewEmail string `json:"newEmail"`
}

func InitiateEmailChange(ctx *gin.Context, r *ginext.RequestContext, requestTO ChangeEmailTO) error {
	securityLog := r.SecurityLog
	user := r.User

	nextEmail := requestTO.NewEmail

	if !user.IsPresent() {
		return errors.Error("missing user")

	}
	securityLog.Info("Changing user email")

	verificationToken := random.MakeRandomURLSafeB64(21)
	if err := repository.SetNextEmail(ctx, r.Database, user.ID, nextEmail, verificationToken); err != nil {
		return errors.Wrap("issue setting next email for user", err)
	}

	if err := service.SendChangeVerificationEmail(ctx, r, user.Language, nextEmail, verificationToken); err != nil {
		return errors.Wrap("error sending change verification email", err)
	}
	if err := service.SendChangeNotificationEmail(ctx, r, user.Language, user.Email, nextEmail); err != nil {
		return errors.Wrap("error sending change notification email", err)
	}

	return nil
}
