package middleware

import (
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/util/errors"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterVerifiedEmailAuthorizationMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := ginext.GetRequestContext(ctx)
		securityLog := r.SecurityLog
		user := r.User

		if !user.IsPresent() || !user.EmailVerified {
			securityLog.Info("Email not verified")
			_ = ctx.AbortWithError(http.StatusForbidden, errors.Error("email not verified"))
			return
		}
	})
}
