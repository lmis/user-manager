package middleware

import (
	"net/http"
	config "user-manager/cmd/app/config"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func CsrfMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieName := "__Host-CSRF-Token"
		if config.IsLocalEnv() {
			cookieName = "CSRF-Token"
		}
		cookie, err := c.Cookie(cookieName)
		if err != nil && err != http.ErrNoCookie {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("getting CSRF cookie failed", err))
			return
		}
		header := c.GetHeader("X-CSRF-Token")
		if header == "" || cookie == "" {
			c.AbortWithError(http.StatusBadRequest, util.Error("missing tokens"))
			return
		}

		if header != cookie {
			c.AbortWithError(http.StatusBadRequest, util.Error("mismatching csrf tokens"))
			return
		}

	}
}
