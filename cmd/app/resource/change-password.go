package resource

import (
	"golang.org/x/crypto/bcrypt"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

func RegisterChangePasswordResource(group *gin.RouterGroup) {
	group.POST("change-password", ginext.WrapEndpointWithoutResponseBody(ChangePassword))
}

type ChangePasswordTO struct {
	OldPassword []byte `json:"oldPassword"`
	NewPassword []byte `json:"newPassword"`
}

func ChangePassword(ctx *gin.Context, r *ginext.RequestContext, requestTO ChangePasswordTO) error {
	securityLog := r.SecurityLog
	userSession := r.UserSession

	user := userSession.User

	if user.AppUserID == 0 {
		return errors.Error("missing user")
	}
	securityLog.Info("Changing user password")

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), requestTO.OldPassword); err != nil {
		securityLog.Info("Password mismatch for user %s trying to change password", user.AppUserID)
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return nil
	}

	newHash, err := service.HashPassword(requestTO.NewPassword)
	if err != nil {
		return errors.Wrap("error hashing new password", err)
	}

	if err := repository.SetPasswordHash(ctx, r.Tx, user.AppUserID, newHash); err != nil {
		return errors.Wrap("issue setting new password hash for user", err)
	}

	return nil
}
