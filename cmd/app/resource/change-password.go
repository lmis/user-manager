package resource

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	"user-manager/cmd/app/service/users"
	dm "user-manager/domain-model"
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

func ChangePassword(ctx *gin.Context, r *dm.RequestContext, requestTO ChangePasswordTO) error {
	logger := r.Logger
	user := r.User

	if !user.IsPresent() {
		return errs.Error("missing user")
	}
	logger.Info("Changing user password")

	if !auth.VerifyCredentials(requestTO.OldPassword, user.Credentials) {
		logger.Info("Password mismatch for user trying to change password", "userID", user.IDHex())
		ctx.AbortWithStatus(http.StatusBadRequest)
		return nil
	}

	newCredentials, err := auth.MakeCredentials(requestTO.NewPassword)
	if err != nil {
		return errs.Wrap("error making credentials from new password", err)
	}

	if err := users.SetCredentials(ctx, r.Database, user.ID(), newCredentials); err != nil {
		return errs.Wrap("issue setting new password hash for user", err)
	}

	return nil
}
