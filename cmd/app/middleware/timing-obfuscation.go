package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterTimingObfuscationMiddleware(group *gin.RouterGroup, minTime time.Duration) {
	group.Use(TimingObfuscationMiddleware(minTime))
}

// TODO: Rethink and document this
func TimingObfuscationMiddleware(minTime time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		waitTime := minTime - time.Since(start)
		time.Sleep(waitTime)
	}
}
