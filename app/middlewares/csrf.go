package middlewares

import (
	"net/http"
	ginext "user-manager/gin-extensions"
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
			ginext.LogAndAbortWithError(c, http.StatusInternalServerError, util.Wrap("CsrfMiddleware", "getting CSRF cookie failed", err))
			return
		}
		header := c.GetHeader("X-CSRF-Token")
		if header == "" || cookie == "" {
			ginext.LogAndAbortWithError(c, http.StatusBadRequest, util.Error("CsrfMiddleware", "missing tokens"))
		}

		if header != cookie {
			ginext.LogAndAbortWithError(c, http.StatusBadRequest, util.Error("CsrfMiddleware", "mismatching csrf tokens"))
			return
		}

		c.Next()
	}
}
