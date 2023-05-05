package middleware

import (
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

func RegisterRequireSudoModeMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		sudoSessionID, err := service.GetSessionCookie(ctx, dm.UserSessionTypeSudo)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting session cookie failed", err))
			return
		}

		if sudoSessionID == "" {
			_ = ctx.AbortWithError(http.StatusForbidden, errors.Error("sudo session cookie missing"))
			return
		}

		session, err := repository.GetSessionAndUser(ctx, r.Tx, sudoSessionID, dm.UserSessionTypeSudo)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting sudo session failed", err))
			return
		}

		if session.UserSessionID == "" {
			_ = ctx.AbortWithError(http.StatusForbidden, errors.Error("sudo session not found on db"))
			return
		}

		if err := repository.UpdateSessionTimeout(ctx, r.Tx, session.UserSessionID, time.Now().Add(dm.SudoSessionDuration)); err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("issue updating session timeout in db", err))
			return
		}
	})
}
