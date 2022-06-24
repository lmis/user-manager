package user

import (
	"fmt"
	"net/http"
	"user-manager/db"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type EmailConfirmationTO struct {
	Token string `json:"token"`
}

const (
	AlreadyConfirmed string = "already-confirmed"
	NewlyConfirmed   string = "newly-confirmed"
	InvalidToken     string = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status string `json:"status"`
}

func PostConfirmEmail(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	emailConfirmationTO := EmailConfirmationTO{}
	err := c.BindJSON(&emailConfirmationTO)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("PostConfirmationEmail", "cannot bind to emailConfirmationTO", err))
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

	// TODO: token valid until?
	if !user.EmailVerificationToken.Valid {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostConfirmationEmail", "no verification token present on database", err))
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

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()
	rows, err := user.Update(ctx, requestContext.Tx, boil.Infer())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostConfirmEmail", "issue updating user in db", err))
		return
	}
	if rows != 1 {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("PostConfirmEmail", fmt.Sprintf("wrong number of rows affected: %d", rows), err))
		return
	}
}
