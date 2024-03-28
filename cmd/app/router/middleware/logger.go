package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

func RegisterLoggerMiddleware(app *gin.Engine) {
	app.Use(LoggerMiddleware)
}

func LoggerMiddleware(c *gin.Context) {
	requestContext := GetRequestContext(c)

	// Add requestLogger to context
	requestLogger := slog.Default().With(
		"correlationID", c.GetHeader("X-Correlation-ID"),
		"request", fmt.Sprintf("%s %s (%s)", c.Request.Method, c.Request.URL.Path, requestContext.RequestID))

	requestContext.Logger = requestLogger

	requestLogger.Info("Starting request", "clientIP", c.ClientIP(), "bodySize", c.Writer.Size())

	// Start timer
	start := time.Now()

	// Process request
	c.Next()

	// Stop timer
	latency := time.Since(start)
	status := c.Writer.Status()

	if status >= 400 {
		requestLogger.Info("Request failed", "status", status, "latency", latency, "errors", c.Errors.String())
	} else {
		requestLogger.Info("Request finished", "status", status, "latency", latency)
	}

	// Trigger alerts
	if status == http.StatusInternalServerError {
		for _, err := range c.Errors {
			requestLogger.Error(err.Err.Error())
		}
	}
}
