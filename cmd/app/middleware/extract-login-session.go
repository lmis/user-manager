package middleware

import (
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/repository"
	"user-manager/service"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

func RegisterExtractLoginSessionMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)

		sessionToken, err := service.GetSessionCookie(ctx, dm.UserSessionTypeLogin)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("getting session cookie failed", err))
			return
		}

		if sessionToken == "" {
			return
		}

		user, err := repository.GetUserForSession(ctx, r.Database, sessionToken, dm.UserSessionTypeLogin)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("fetching user for sessionToken failed", err))
			return
		}

		if user.IsPresent() {
			r := ginext.GetRequestContext(ctx)
			r.User = user

			if err := repository.UpdateSessionTimeout(ctx, r.Database, sessionToken, time.Now().Add(dm.LoginSessionDuration)); err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("issue updating session timeout in db", err))
				return
			}
		}
	})
}
