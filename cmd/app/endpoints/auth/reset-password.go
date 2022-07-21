package authendpoints

import (
	"database/sql"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	passwordservice "user-manager/cmd/app/services/password"
	userservice "user-manager/cmd/app/services/user"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type ResetPasswordTO struct {
	Email       string `json:"email"`
	Token       string `json:"token"`
	NewPassword []byte `json:"newPassword"`
}

type ResetPasswordStatus string

const (
	Success      ResetPasswordStatus = "success"
	InvalidToken ResetPasswordStatus = "invalid-token"
)

type ResetPasswordResponseTO struct {
	Status ResetPasswordStatus `json:"status"`
}

func PostResetPassword(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	tx := requestContext.Tx
	securityLog.Info("Password reset requested")

	var resetPasswordTO ResetPasswordTO
	if err := c.BindJSON(&resetPasswordTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to resetPasswordTO", err))
		return
	}

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(
		models.AppUserWhere.Email.EQ(resetPasswordTO.Email),
		models.AppUserWhere.PasswordResetToken.EQ(null.StringFrom(resetPasswordTO.Token)),
		models.AppUserWhere.PasswordResetTokenValidUntil.GT(null.TimeFrom(time.Now())),
	).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			securityLog.Info("Invalid password reset attempt")
			c.JSON(http.StatusOK, ResetPasswordResponseTO{Status: InvalidToken})
			return
		} else {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("error finding user", err))
		}
		return
	}

	hash, err := passwordservice.Hash(resetPasswordTO.NewPassword)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue making password hash", err))
	}

	user.PasswordResetToken = null.StringFromPtr(nil)
	user.PasswordResetTokenValidUntil = null.TimeFromPtr(nil)
	user.PasswordHash = hash

	if err := userservice.UpdateUser(requestContext, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue persisting user", err))
		return
	}

	c.JSON(http.StatusOK, ResetPasswordResponseTO{Status: Success})
}
