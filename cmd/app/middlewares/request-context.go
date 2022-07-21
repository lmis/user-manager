package middleware

import (
	config "user-manager/cmd/app/config"
	ginext "user-manager/cmd/app/gin-extensions"

	"github.com/gin-gonic/gin"
)

func RequestContextMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ginext.RequestContextKey, &ginext.RequestContext{
			Config: config,
		})
	}
}
