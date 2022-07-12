package middlewares

import (
	"user-manager/config"
	ginext "user-manager/gin-extensions"

	"github.com/gin-gonic/gin"
)

func RequestContextMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &ginext.RequestContext{
			Config: config,
		}
		c.Set("ctx", ctx)
	}
}
