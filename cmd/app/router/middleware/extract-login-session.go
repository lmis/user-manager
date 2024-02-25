package middleware

import (
	"net/http"
	"time"
	"user-manager/cmd/app/service/auth"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

func RegisterExtractLoginSessionMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := GetRequestContext(ctx)

		sessionToken, err := auth.GetSessionCookie(ctx, dm.UserSessionTypeLogin)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("getting session cookie failed", err))
			return
		}

		if sessionToken == "" {
			return
		}

		user, err := auth.GetUserForSession(ctx, r.Database, sessionToken, dm.UserSessionTypeLogin)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("fetching user for sessionToken failed", err))
			return
		}

		if user.IsPresent() {
			r := GetRequestContext(ctx)
			r.User = user

			if err := auth.UpdateSessionTimeout(ctx, r.Database, sessionToken, time.Now().Add(dm.LoginSessionDuration)); err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("issue updating session timeout in db", err))
				return
			}
		}
	})
}
