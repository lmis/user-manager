package usersettingsendpoints

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	userservice "user-manager/cmd/app/services/user"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type EmailChangeConfirmationTO struct {
	Token string `json:"token"`
}

type EmailChangeStatus string

const (
	NoEmailChangeInProgress EmailChangeStatus = "no-change-in-progress"
	InvalidToken            EmailChangeStatus = "invalid-token"
	NewEmailConfirmed       EmailChangeStatus = "new-email-confirmed"
)

type EmailChangeConfirmationResponseTO struct {
	Status EmailChangeStatus `json:"status"`
	Email  string            `json:"email"`
}

func PostConfirmEmailChange(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	emailChangeConfirmationTO := EmailChangeConfirmationTO{}
	if err := c.BindJSON(&emailChangeConfirmationTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to emailChangeConfirmationTO", err))
		return
	}
	user := requestContext.Authentication.AppUser

	if !user.NewEmail.Valid {
		c.JSON(http.StatusOK, EmailChangeConfirmationResponseTO{
			Status: NoEmailChangeInProgress,
			Email:  user.Email,
		})
		return
	}

	if !user.EmailVerificationToken.Valid {
		c.AbortWithError(http.StatusInternalServerError, util.Error("no verification token present on database"))
		return
	}

	if emailChangeConfirmationTO.Token != user.EmailVerificationToken.String {
		securityLog.Info("Invalid email verification token")
		c.JSON(http.StatusOK, EmailChangeConfirmationResponseTO{
			Status: InvalidToken,
			Email:  user.Email,
		})
		return
	}

	user.EmailVerificationToken = null.StringFromPtr(nil)
	user.Email = user.NewEmail.String
	user.NewEmail = null.StringFromPtr(nil)

	if err := userservice.UpdateUser(requestContext, &user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue persisting user", err))
		return
	}

	c.JSON(http.StatusOK, EmailChangeConfirmationResponseTO{
		Status: NewEmailConfirmed,
		Email:  user.Email,
	})
}
