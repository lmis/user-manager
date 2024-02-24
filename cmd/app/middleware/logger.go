package middleware

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	dm "user-manager/domain-model"
	"user-manager/util/logger"

	"time"

	"github.com/gin-gonic/gin"
)

func RegisterLoggerMiddleware(app *gin.Engine) {
	app.Use(LoggerMiddleware)
}

func LoggerMiddleware(c *gin.Context) {
	// Add requestLogger to context
	requestLogger := &RequestLogger{topic: "REQUEST", context: c}
	securityLogger := &RequestLogger{topic: "SECURITY", context: c}

	requestContext := ginext.GetRequestContext(c)
	requestContext.Log = dm.Log(requestLogger)
	requestContext.SecurityLog = dm.SecurityLog(securityLogger)

	requestLogger.Info("Starting request")

	// Start timer
	start := time.Now()

	// Process request
	c.Next()

	// Stop timer
	requestLogger.latency = time.Since(start)

	status := c.Writer.Status()
	requestLogger.status = status
	if status >= 400 {
		securityLogger.Info("Request failed. Status: %d", status)
	} else {
		requestLogger.Info("Finished request. Status: %d", status)
	}

	// Trigger alerts
	if status == http.StatusInternalServerError {
		for _, err := range c.Errors {
			securityLogger.Err(err.Err)
		}
	}
}

type LogMetadata struct {
	Topic         string        `json:"topic"`
	CorrelationID string        `json:"correlationID"`
	Latency       time.Duration `json:"latency,omitempty"`
	Path          string        `json:"path,omitempty"`
	UserID        string        `json:"userID,omitempty"`
	Roles         []dm.UserRole `json:"role,omitempty"`
	ClientIP      string        `json:"clientIP,omitempty"`
	Method        string        `json:"method,omitempty"`
	ErrorMessage  string        `json:"errorMessage,omitempty"`
	BodySize      int           `json:"bodySize,omitempty"`
	Status        int           `json:"status,omitempty"`
}

type RequestLogger struct {
	topic   string
	context *gin.Context
	latency time.Duration
	status  int
}

func getMetadata(logger *RequestLogger) *LogMetadata {
	topic := logger.topic
	c := logger.context
	requestContext := ginext.GetRequestContext(c)
	user := requestContext.User
	path := c.FullPath()

	metadata := LogMetadata{
		Topic:         topic,
		CorrelationID: c.GetHeader("X-Correlation-ID"),
		Path:          path,
		ClientIP:      c.ClientIP(),
		Method:        c.Request.Method,
		Status:        logger.status,
		ErrorMessage:  c.Errors.String(),
		BodySize:      c.Writer.Size(),
		Latency:       logger.latency,
	}
	if user.IsPresent() {
		metadata.UserID = user.ID.ToString()
		metadata.Roles = user.UserRoles
	}
	return &metadata
}

func (r *RequestLogger) Info(format string, args ...interface{}) {
	logger.WriteLog(getMetadata(r), logger.LogLevelInfo, format, args...)
}

func (r *RequestLogger) Warn(format string, args ...interface{}) {
	logger.WriteLog(getMetadata(r), logger.LogLevelWarn, format, args...)
}

func (r *RequestLogger) Err(e error) {
	logger.WriteLog(getMetadata(r), logger.LogLevelError, "%s", e.Error())
}
