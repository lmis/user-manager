package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	dm "user-manager/domain-model"
	"user-manager/util/errs"
)

func RegisterRequireSudoModeMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		sudoSessionToken, err := auth.GetSessionCookie(ctx, dm.UserSessionTypeSudo)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("getting session cookie failed", err))
			return
		}

		if sudoSessionToken == "" {
			_ = ctx.AbortWithError(http.StatusForbidden, errs.Error("sudo session cookie missing"))
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
			_ = ctx.AbortWithError(http.StatusForbidden, errs.Error("sudo session not valid"))
			return
		}

		if err := auth.UpdateSessionTimeout(ctx, r.Database, sudoSessionToken, time.Now().Add(dm.SudoSessionDuration)); err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("issue updating session timeout in db", err))
			return
		}
	})
}
