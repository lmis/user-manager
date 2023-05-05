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
	userSession := r.UserSession

	securityLog.Info("Logout")

	service.RemoveSessionCookie(ctx, r.Config, dm.UserSessionTypeLogin)
	if userSession.UserSessionID != "" {
		if err := repository.DeleteSession(ctx, r.Tx, userSession.UserSessionID); err != nil {
			return errors.Wrap("issue while deleting login session", err)
		}
	}

	sudoSessionID, err := service.GetSessionCookie(ctx, dm.UserSessionTypeSudo)
	if err != nil {
		return errors.Wrap("issue reading sudo session cookie", err)
	}
	if sudoSessionID != "" {
		service.RemoveSessionCookie(ctx, r.Config, dm.UserSessionTypeSudo)
		if err := repository.DeleteSession(ctx, r.Tx, sudoSessionID); err != nil {
			return errors.Wrap("issue while deleting sudo session", err)
		}
	}
	if request.ForgetDevice {
		deviceSessionID, err := service.GetSessionCookie(ctx, dm.UserSessionTypeRememberDevice)
		if err != nil {
			return errors.Wrap("issue reading device session cookie", err)
		}
		if deviceSessionID != "" {
			service.RemoveSessionCookie(ctx, r.Config, dm.UserSessionTypeRememberDevice)
			if err := repository.DeleteSession(ctx, r.Tx, deviceSessionID); err != nil {
				return errors.Wrap("issue while deleting device session", err)
			}
		}
	}

	return nil
}
