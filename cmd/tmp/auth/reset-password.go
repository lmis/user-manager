package auth

// import (
// 	"context"
// 	"time"
// 	ginext "user-manager/cmd/app/gin-extensions"
// 	auth_service "user-manager/cmd/app/services/auth"
// 	"user-manager/db"
// 	"user-manager/db/generated/models"
// 	"user-manager/repository"
// 	"user-manager/util"

// 	"github.com/gin-gonic/gin"
// 	"github.com/volatiletech/null/v8"
// )

// type ResetPasswordTO struct {
// 	Email       string `json:"email"`
// 	Token       string `json:"token"`
// 	NewPassword []byte `json:"newPassword"`
// }

// type ResetPasswordStatus string

// const (
// 	Success      ResetPasswordStatus = "success"
// 	InvalidToken ResetPasswordStatus = "invalid-token"
// )

// type ResetPasswordResponseTO struct {
// 	Status ResetPasswordStatus `json:"status"`
// }

// func PostResetPassword(requestContext *ginext.RequestContext, requestTO *ResetPasswordTO, _ *gin.Context) (*ResetPasswordResponseTO, error) {
// 	userRepository := repository.NewUserRepository(requestContext.Tx)
// 	securityLog := requestContext.SecurityLog
// 	tx := requestContext.Tx
// 	securityLog.Info("Password reset requested")

// 	user, err := db.Fetch(func(ctx context.Context) (*models.AppUser, error) {
// 		return models.AppUsers(
// 			models.AppUserWhere.Email.EQ(requestTO.Email),
// 			models.AppUserWhere.PasswordResetToken.EQ(null.StringFrom(requestTO.Token)),
// 			models.AppUserWhere.PasswordResetTokenValidUntil.GT(null.TimeFrom(time.Now())),
// 		).One(ctx, tx)
// 	})
// 	if err != nil {
// 		return nil, util.Wrap("error finding user", err)
// 	}
// 	if user.IsNil() {
// 		securityLog.Info("Invalid password reset attempt")
// 		return &ResetPasswordResponseTO{Status: InvalidToken}, nil
// 	}

// 	hash, err := auth_service.Hash(requestTO.NewPassword)
// 	if err != nil {
// 		return nil, util.Wrap("issue making password hash", err)
// 	}

// 	user.Val.PasswordResetToken = null.StringFromPtr(nil)
// 	user.Val.PasswordResetTokenValidUntil = null.TimeFromPtr(nil)
// 	user.Val.PasswordHash = hash

// 	if err := userRepository.UpdateUser(user.Val); err != nil {
// 		return nil, util.Wrap("issue persisting user", err)
// 	}

// 	return &ResetPasswordResponseTO{Status: Success}, nil
// }
