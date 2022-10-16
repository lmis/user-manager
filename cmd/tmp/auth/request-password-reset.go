package auth

// import (
// 	"context"
// 	"time"
// 	ginext "user-manager/cmd/app/gin-extensions"
// 	email_service "user-manager/cmd/app/services/email"
// 	"user-manager/db"
// 	"user-manager/db/generated/models"
// 	"user-manager/repository"
// 	"user-manager/util"

// 	"github.com/gin-gonic/gin"
// 	"github.com/volatiletech/null/v8"
// )

// type PasswordResetRequestTO struct {
// 	Email string `json:"email"`
// }

// func PostRequestPasswordReset(requestContext *ginext.RequestContext, requestTO *PasswordResetRequestTO, _ *gin.Context) error {
// 	userRepository := repository.NewUserRepository(requestContext.Tx)
// 	securityLog := requestContext.SecurityLog
// 	tx := requestContext.Tx
// 	securityLog.Info("Password reset requested")

// 	user, err := db.Fetch(func(ctx context.Context) (*models.AppUser, error) {
// 		return models.AppUsers(
// 			models.AppUserWhere.Email.EQ(requestTO.Email),
// 		).One(ctx, tx)
// 	})
// 	if err != nil {
// 		return util.Wrap("error finding user", err)
// 	}
// 	if user.IsNil() {
// 		securityLog.Info("Password reset attempt for non-existant email")
// 		return nil
// 	}

// 	user.Val.PasswordResetToken = null.StringFrom(util.MakeRandomURLSafeB64(21))
// 	user.Val.PasswordResetTokenValidUntil = null.TimeFrom(time.Now().Add(1 * time.Hour))

// 	if err := userRepository.UpdateUser(user.Val); err != nil {
// 		return util.Wrap("issue persisting user", err)
// 	}
// 	if err := email_service.SendResetPasswordEmail(requestContext, user.Val); err != nil {
// 		return util.Wrap("error sending password reset email", err)
// 	}
// 	return nil
// }
