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

func RegisterExtractLoginSessionMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)

		sessionID, err := service.GetSessionCookie(ctx, dm.UserSessionTypeLogin)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting session cookie failed", err))
			return
		}

		if sessionID == "" {
			return
		}

		session, err := repository.GetSessionAndUser(ctx, r.Tx, sessionID, dm.UserSessionTypeLogin)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("fetching session failed", err))
			return
		}

		if session.UserSessionID != "" {

			ginext.GetRequestContext(ctx).UserSession = session

			if err := repository.UpdateSessionTimeout(ctx, r.Tx, session.UserSessionID, time.Now().Add(dm.LoginSessionDuration)); err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("issue updating session timeout in db", err))
				return
			}
		}
	})
}
