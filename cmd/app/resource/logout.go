package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

func RegisterLogoutResource(group *gin.RouterGroup) {
	group.POST("logout", ginext.WrapEndpointWithoutResponseBody(Logout))
}

type LogoutTO struct {
	ForgetDevice bool `json:"forgetDevice"`
}

func Logout(ctx *gin.Context, r *ginext.RequestContext, request LogoutTO) error {
	securityLog := r.SecurityLog

	securityLog.Info("Logout")

	err := forgetSession(ctx, r, dm.UserSessionTypeLogin)
	if err != nil {
		return errors.Wrap("issue while forgetting login session", err)
	}

	err = forgetSession(ctx, r, dm.UserSessionTypeSudo)
	if err != nil {
		return errors.Wrap("issue while forgetting sudo session", err)
	}

	if request.ForgetDevice {
		err := forgetSession(ctx, r, dm.UserSessionTypeRememberDevice)
		if err != nil {
			return errors.Wrap("issue while forgetting device session", err)
		}
	}

	return nil
}

func forgetSession(ctx *gin.Context, r *ginext.RequestContext, sessionType dm.UserSessionType) error {
	sessionToken, err := service.GetSessionCookie(ctx, sessionType)
	if err != nil {
		return errors.Wrap("issue reading session cookie", err)
	}
	if sessionToken != "" {
		service.RemoveSessionCookie(ctx, r.Config, sessionType)
		if err := repository.DeleteSession(ctx, r.Database, sessionToken); err != nil {
			return errors.Wrap("issue while deleting session", err)
		}
	}
	return nil
}
