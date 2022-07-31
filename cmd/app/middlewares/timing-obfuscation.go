package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// TODO: Rething and document this
func TimingObfuscationMiddleware(minTime time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		waitTime := minTime - time.Since(start)
		time.Sleep(waitTime)
	}
}
