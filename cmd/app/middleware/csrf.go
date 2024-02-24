package middleware

import (
	errs "errors"
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

func RegisterCsrfMiddleware(group *gin.RouterGroup) {
	group.Use(func(ctx *gin.Context) {
		config := ginext.GetRequestContext(ctx).Config

		cookieName := "__Host-CSRF-Token"
		if config.IsLocalEnv() {
			cookieName = "CSRF-Token"
		}
		cookie, err := ctx.Cookie(cookieName)
		if err != nil && !errs.Is(err, http.ErrNoCookie) {
			_ = ctx.AbortWithError(http.StatusInternalServerError, errors.Wrap("getting CSRF cookie failed", err))
			return
		}
		header := ctx.GetHeader("X-CSRF-Token")
		if header == "" || cookie == "" {
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.Error("missing tokens"))
			return
		}

		if header != cookie {
			_ = ctx.AbortWithError(http.StatusBadRequest, errors.Error("mismatching csrf tokens"))
			return
		}
	})
}
