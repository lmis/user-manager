package auth

import (
	"database/sql"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	email_service "user-manager/cmd/app/services/email"
	user_service "user-manager/cmd/app/services/user"
	"user-manager/db"
	"user-manager/db/generated/models"
	"user-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v8"
)

type PasswordResetRequestTO struct {
	Email string `json:"email"`
}

func PostRequestPasswordReset(requestContext *ginext.RequestContext, requestTO *PasswordResetRequestTO, _ *gin.Context) error {
	securityLog := requestContext.SecurityLog
	tx := requestContext.Tx
	securityLog.Info("Password reset requested")

	ctx, cancelTimeout := db.DefaultQueryContext()
	defer cancelTimeout()

	user, err := models.AppUsers(
		models.AppUserWhere.Email.EQ(requestTO.Email),
	).One(ctx, tx)
	if err != nil {
		if err == sql.ErrNoRows {
			securityLog.Info("Password reset attempt for non-existant email")
			return nil
		}
		return util.Wrap("error finding user", err)
	}

	user.PasswordResetToken = null.StringFrom(util.MakeRandomURLSafeB64(21))
	user.PasswordResetTokenValidUntil = null.TimeFrom(time.Now().Add(1 * time.Hour))

	if err := user_service.UpdateUser(requestContext, user); err != nil {
		return util.Wrap("issue persisting user", err)
	}
	if err := email_service.SendResetPasswordEmail(requestContext, user); err != nil {
		return util.Wrap("error sending password reset email", err)
	}
	return nil
}
