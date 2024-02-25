package middleware

import (
	"user-manager/util/errs"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterVerifiedEmailAuthorizationMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		r := GetRequestContext(ctx)
		securityLog := r.SecurityLog
		user := r.User

		if !user.IsPresent() || !user.EmailVerified {
			securityLog.Info("Email not verified")
			_ = ctx.AbortWithError(http.StatusForbidden, errs.Error("email not verified"))
			return
		}
	})
}
