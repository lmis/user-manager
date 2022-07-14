package middlewares

import (
	"errors"
	"net/http"
	"runtime/debug"
	"syscall"
	"user-manager/util"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware(c *gin.Context) {
	defer func() {
		if p := recover(); p != nil {
			// Client interrupted connection
			if err, ok := p.(error); ok && (errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET)) {
				c.AbortWithError(http.StatusBadRequest, util.Wrap("client connection lost", err))
				return
			}
			c.AbortWithError(http.StatusInternalServerError, util.WrapRecoveredPanic(p, debug.Stack()))
		}
	}()
	c.Next()
}
