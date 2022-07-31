package sensitive_settings

import (
	ginext "user-manager/cmd/app/gin-extensions"
	email_service "user-manager/cmd/app/services/email"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type ChangeEmailTO struct {
	NewEmail string `json:"newEmail"`
}

func PostChangeEmail(requestContext *ginext.RequestContext, requestTO *ChangeEmailTO, _ *gin.Context) error {
	securityLog := requestContext.SecurityLog
	securityLog.Info("Changing user email")
	user := requestContext.Authentication.AppUser

	user.EmailVerificationToken = null.StringFrom(util.MakeRandomURLSafeB64(21))
	user.NewEmail = null.StringFrom(requestTO.NewEmail)

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		return util.Wrap("issue persisting user", err)
	}

	if err := email_service.SendChangeVerificationEmail(requestContext, user); err != nil {
		return util.Wrap("error sending change verification email", err)
	}
	if err := email_service.SendChangeNotificationEmail(requestContext, user); err != nil {
		return util.Wrap("error sending change notification email", err)
	}

	return nil
}
