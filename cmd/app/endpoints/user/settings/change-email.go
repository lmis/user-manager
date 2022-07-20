package usersettingsendpoints

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	emailservice "user-manager/cmd/app/services/email"
	userservice "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type ChangeEmailTO struct {
	NewEmail string `json:"newEmail"`
}

func PostChangeEmail(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	securityLog.Info("Changing user email")
	changeEmailTO := ChangeEmailTO{}
	if err := c.BindJSON(&changeEmailTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to changeEmailTO", err))
		return
	}
	user := requestContext.Authentication.AppUser

	user.EmailVerificationToken = null.StringFrom(util.MakeRandomURLSafeB64(21))
	user.NewEmail = null.StringFrom(changeEmailTO.NewEmail)

	if err := userservice.UpdateUser(requestContext, &user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue persisting user", err))
		return
	}

	if err := emailservice.SendChangeVerificationEmail(requestContext, &user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error sending change verification email", err))
		return
	}
	if err := emailservice.SendChangeNotificationEmail(requestContext, &user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error sending change notification email", err))
		return
	}

	c.JSON(http.StatusOK, EmailChangeConfirmationResponseTO{
		Status: NewEmailConfirmed,
		Email:  user.Email,
	})
}
