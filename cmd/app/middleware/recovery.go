package middleware

import (
	errs "errors"
	"net/http"
	"runtime/debug"
	"syscall"
	"user-manager/util/errors"

	"github.com/gin-gonic/gin"
)

func RegisterRecoveryMiddleware(app *gin.Engine) {
	app.Use(RecoveryMiddleware)
}

func RecoveryMiddleware(c *gin.Context) {
	defer func() {
		if p := recover(); p != nil {
			// Client interrupted connection
			if err, ok := p.(error); ok && (errs.Is(err, syscall.EPIPE) || errs.Is(err, syscall.ECONNRESET)) {
				c.AbortWithError(http.StatusBadRequest, errors.Wrap("client connection lost", err))
				return
			}
			c.AbortWithError(http.StatusInternalServerError, errors.WrapRecoveredPanic(p, debug.Stack()))
			return
		}
	}()
	c.Next()
}
