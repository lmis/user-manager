package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errors"
)

func RegisterRequireSudoModeMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		sudoSessionToken, err := service.GetSessionCookie(ctx, dm.UserSessionTypeSudo)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting session cookie failed", err))
			return
		}

		if sudoSessionToken == "" {
			_ = ctx.AbortWithError(http.StatusForbidden, errors.Error("sudo session cookie missing"))
			return
		}

		sessionIsValid := false
		for _, session := range r.User.Sessions {
			if session.Token == sudoSessionToken && session.Type == dm.UserSessionTypeSudo && session.TimeoutAt.After(time.Now()) {
				sessionIsValid = true
				break
			}
		}

		if !sessionIsValid {
			_ = ctx.AbortWithError(http.StatusForbidden, errors.Error("sudo session not valid"))
			return
		}

		if err := repository.UpdateSessionTimeout(ctx, r.Database, sudoSessionToken, time.Now().Add(dm.SudoSessionDuration)); err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("issue updating session timeout in db", err))
			return
		}
	})
}
