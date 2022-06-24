package middlewares

import (
	"net/http"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func CsrfMiddleware(environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieName := "__Host-XSRF-Token"
		if environment == "local" {
			cookieName = "XSRF-Token"
		}
		cookie, err := c.Cookie(cookieName)
		if err != nil && err != http.ErrNoCookie {
			c.AbortWithError(http.StatusInternalServerError, util.Wrap("CsrfMiddleware", "getting CSRF cookie failed", err))
			return
		}
		header := c.GetHeader("X-CSRF-Token")
		if header == "" || cookie == "" {
			c.AbortWithError(http.StatusBadRequest, util.Error("CsrfMiddleware", "missing tokens"))
		}

		if header != cookie {
			c.AbortWithError(http.StatusBadRequest, util.Error("CsrfMiddleware", "mismatching csrf tokens"))
			return
		}

		c.Next()
	}
}
