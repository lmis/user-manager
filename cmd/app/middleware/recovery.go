package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"
	"syscall"
	errs "user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

func RegisterRecoveryMiddleware(app *gin.Engine) {
	app.Use(RecoveryMiddleware)
}

func RecoveryMiddleware(c *gin.Context) {
	defer func() {
		if p := recover(); p != nil {
			// Client interrupted connection
			if err, ok := p.(error); ok && (errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET)) {
				c.AbortWithError(http.StatusBadRequest, errs.Wrap("client connection lost", err))
				return
			}
			c.AbortWithError(http.StatusInternalServerError, errs.WrapRecoveredPanic(p, debug.Stack()))
			return
		}
	}()
	c.Next()
}
