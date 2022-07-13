package middlewares

import (
	config "user-manager/cmd/app/config"
	ginext "user-manager/cmd/app/gin-extensions"

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
