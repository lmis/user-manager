package middleware

import (
	"errors"
	"net/http"
	"user-manager/util/errs"

	"github.com/gin-gonic/gin"
)

func RegisterCsrfMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		config := GetRequestContext(ctx).Config

		cookieName := "__Host-CSRF-Token"
		if config.IsLocalEnv() {
			cookieName = "CSRF-Token"
		}
		cookie, err := ctx.Cookie(cookieName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errs.Wrap("getting CSRF cookie failed", err))
			return
		}
		header := ctx.GetHeader("X-CSRF-Token")
		if header == "" || cookie == "" {
			_ = ctx.AbortWithError(http.StatusBadRequest, errs.Error("missing tokens"))
			return
		}

		if header != cookie {
			_ = ctx.AbortWithError(http.StatusBadRequest, errs.Error("mismatching csrf tokens"))
			return
		}
	})
}
