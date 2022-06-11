package middlewares

import (
	"user-manager/db/generated/models"
	ginext "user-manager/gin-extensions"
	"user-manager/util"

	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(c *gin.Context) {
	// Add logger to context
	var latency time.Duration
	logger := makeLogger("REQUEST", c, &latency)
	securityLogger := makeLogger("SECURITY", c, &latency)

	requestContext := ginext.GetRequestContext(c)
	requestContext.Log = logger
	requestContext.SecurityLog = securityLogger

	logger.Info("Starting request")

	// Start timer
	start := time.Now()

	// Process request
	c.Next()

	// Stop timer
	latency = time.Since(start)

	logger.Info("Finished request. Status: %d", c.Writer.Status())
}

type LogPayload struct {
	Topic        string          `json:"topic"`
	Latency      time.Duration   `json:"latency,omitempty"`
	Path         string          `json:"path,omitempty"`
	UserID       int             `json:"userID,omitempty"`
	Role         models.UserRole `json:"role,omitempty"`
	ClientIP     string          `json:"clientIP,omitempty"`
	Method       string          `json:"method,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	BodySize     int             `json:"bodySize,omitempty"`
	Status       int             `json:"status,omitempty"`
}

func makeLogger(topic string, c *gin.Context, latency *time.Duration) *util.Logger {
	return util.LogWithPayloadMaker(
		topic,
		func() interface{} {
			requestContext := ginext.GetRequestContext(c)
			authentication := requestContext.Authentication
			path := c.Request.URL.Path
			raw := c.Request.URL.RawQuery
			if raw != "" {
				// TODO: Anonymize query values
				path = path + "?" + raw
			}

			jsonPayload := LogPayload{
				Topic:        topic,
				Path:         path,
				ClientIP:     c.ClientIP(),
				Method:       c.Request.Method,
				Status:       c.Writer.Status(),
				ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
				BodySize:     c.Writer.Size(),
			}
			if authentication != nil {
				jsonPayload.UserID = authentication.UserID
				jsonPayload.Role = authentication.Role
			}
			if latency != nil {
				jsonPayload.Latency = *latency
			}
			return jsonPayload
		})
}
