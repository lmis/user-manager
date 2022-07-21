package authendpoints

import (
	"database/sql"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	emailservice "user-manager/cmd/app/services/email"
	userservice "user-manager/cmd/app/services/user"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type PasswordResetRequestTO struct {
	Email string `json:"email"`
}

func PostRequestPasswordReset(c *gin.Context) {
	requestContext := ginext.GetRequestContext(c)
	securityLog := requestContext.SecurityLog
	tx := requestContext.Tx
	securityLog.Info("Password reset requested")

	var passwordResetRequestTO PasswordResetRequestTO
	if err := c.BindJSON(&passwordResetRequestTO); err != nil {
		c.AbortWithError(http.StatusBadRequest, util.Wrap("cannot bind to passwordResetRequestTO", err))
		return
	}

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(
		models.AppUserWhere.Email.EQ(passwordResetRequestTO.Email),
	).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			securityLog.Info("Password reset attempt for non-existant email")
			c.Status(http.StatusOK)
			return
		} else {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("error finding user", err))
		}
		return
	}

	user.PasswordResetToken = null.StringFrom(util.MakeRandomURLSafeB64(21))
	user.PasswordResetTokenValidUntil = null.TimeFrom(time.Now().Add(1 * time.Hour))

	if err := userservice.UpdateUser(requestContext, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("issue persisting user", err))
		return
	}
	if err := emailservice.SendResetPasswordEmail(requestContext, user); err != nil {
		c.AbortWithError(http.StatusInternalServerError, util.Wrap("error sending password reset email", err))
		return
	}
	c.Status(http.StatusOK)
}
