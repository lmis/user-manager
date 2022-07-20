package userendpoints

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	userservice "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type EmailConfirmationTO struct {
	Token string `json:"token"`
}

type EmailConfirmationStatus string

const (
	AlreadyConfirmed EmailConfirmationStatus = "already-confirmed"
	NewlyConfirmed   EmailConfirmationStatus = "newly-confirmed"
	InvalidToken     EmailConfirmationStatus = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status EmailConfirmationStatus `json:"status"`
}

func PostConfirmEmail(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	emailConfirmationTO := EmailConfirmationTO{}
	if err := c.BindJSON(&emailConfirmationTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to emailConfirmationTO", err))
		return
	}
	user := requestContext.Authentication.AppUser

	if user.EmailVerified {
		securityLog.Info("Email already verified")
		c.JSON(http.StatusOK, EmailConfirmationResponseTO{
			Status: AlreadyConfirmed,
		})
		return
	}

	if !user.EmailVerificationToken.Valid {
		c.AbortWithError(http.StatusInternalServerError, util.Error("no verification token present on database"))
		return
	}

	if emailConfirmationTO.Token != user.EmailVerificationToken.String {
		securityLog.Info("Invalid email verification token")
		c.JSON(http.StatusOK, EmailConfirmationResponseTO{
			Status: InvalidToken,
		})
		return
	}

	user.EmailVerificationToken = null.StringFromPtr(nil)
	user.EmailVerified = true

	if err := userservice.UpdateUser(requestContext, &user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue persisting user", err))
		return
	}

	c.JSON(http.StatusOK, EmailConfirmationResponseTO{
		Status: NewlyConfirmed,
	})
}
