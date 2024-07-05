package middleware

import (
	"net/http"
	"time"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/cmd/app/service/auth"
	dm "user-manager/domain-model"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

func RegisterExtractLoginSessionMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)

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
			r := ginext.GetRequestContext(ctx)
			r.User = user
			r.Logger = r.Logger.With("userID", user.IDHex())
			r.Logger.Info("User session found", "roles", user.UserRoles)

			if err := auth.UpdateSessionTimeout(ctx, r.Database, sessionToken, time.Now().Add(dm.LoginSessionDuration)); err != nil {
				_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("issue updating session timeout in db", err))
				return
			}
		}
	})
}
