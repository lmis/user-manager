package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterTimingObfuscationMiddleware(group *gin.RouterGroup, minTime time.Duration) {
	group.Use(func(c *gin.Context) {
		// TODO: Rethink and document this
		start := time.Now()
		c.Next()
		waitTime := minTime - time.Since(start)
		time.Sleep(waitTime)
	})
}
