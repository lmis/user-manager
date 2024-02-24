package resource

import (
	"fmt"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errs"

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
	user := r.User

	if !user.IsPresent() {
		return errs.Error("missing user")
	}
	securityLog.Info("Changing user password")

	if !service.VerifyCredentials(requestTO.OldPassword, user.Credentials) {
		securityLog.Info(fmt.Sprintf("Password mismatch for user %s trying to change password", user.ID()))
		ctx.AbortWithStatus(http.StatusBadRequest)
		return nil
	}

	newCredentials, err := service.MakeCredentials(requestTO.NewPassword)
	if err != nil {
		return errs.Wrap("error making credentials from new password", err)
	}

	if err := repository.SetCredentials(ctx, r.Database, user.ID(), newCredentials); err != nil {
		return errs.Wrap("issue setting new password hash for user", err)
	}

	return nil
}
