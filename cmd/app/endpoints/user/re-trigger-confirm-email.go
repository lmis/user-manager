package user

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	email_service "user-manager/cmd/app/services/email"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type RetriggerConfirmationEmailStatus string

type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}

func PostRetriggerConfirmationEmail(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	user := requestContext.Authentication.AppUser

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		c.JSON(http.StatusOK, RetriggerConfirmationEmailResponseTO{Sent: false})
		return
	}

	user.EmailVerificationToken = null.StringFrom(util.MakeRandomURLSafeB64(21))

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue persisting user", err))
		return
	}

	if err := email_service.SendVerificationEmail(requestContext, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error sending verification email", err))
		return
	}

	c.JSON(http.StatusOK, RetriggerConfirmationEmailResponseTO{Sent: true})
}
