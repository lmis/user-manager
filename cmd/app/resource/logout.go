package resource

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

func RegisterLogoutResource(group *gin.RouterGroup) {
	group.POST("logout", ginext.WrapEndpointWithoutResponseBody(Logout))
}

type LogoutTO struct {
	ForgetDevice bool `json:"forgetDevice"`
}

func Logout(ctx *gin.Context, r *dm.RequestContext, request LogoutTO) error {
	logger := r.Logger

	logger.Info("Logout")

	err := forgetSession(ctx, r, dm.UserSessionTypeLogin)
	if err != nil {
		return errs.Wrap("issue while forgetting login session", err)
	}

	err = forgetSession(ctx, r, dm.UserSessionTypeSudo)
	if err != nil {
		return errs.Wrap("issue while forgetting sudo session", err)
	}

	if request.ForgetDevice {
		err := forgetSession(ctx, r, dm.UserSessionTypeRememberDevice)
		if err != nil {
			return errs.Wrap("issue while forgetting device session", err)
		}
	}

	return nil
}

func forgetSession(ctx *gin.Context, r *dm.RequestContext, sessionType dm.UserSessionType) error {
	sessionToken, err := auth.GetSessionCookie(ctx, sessionType)
	if err != nil {
		return errs.Wrap("issue reading session cookie", err)
	}
	if sessionToken != "" {
		auth.RemoveSessionCookie(ctx, r.Config, sessionType)
		if err := auth.DeleteSession(ctx, r.Database, sessionToken); err != nil {
			return errs.Wrap("issue while deleting session", err)
		}
	}
	return nil
}
