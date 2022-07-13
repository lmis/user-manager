package middlewares

import (
	"fmt"
	ginext "user-manager/cmd/app/gin-extensions"
	"user-manager/db/generated/models"
	"user-manager/util"

	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(c *gin.Context) {
	// Add logger to context
	logger := &RequestLogger{topic: "REQUEST", context: c}
	securityLogger := &RequestLogger{topic: "SECURITY", context: c}

	requestContext := ginext.GetRequestContext(c)
	requestContext.Log = logger
	requestContext.SecurityLog = securityLogger

	logger.Info("Starting request")

	// Start timer
	start := time.Now()

	// Process request
	c.Next()

	// Stop timer
	logger.latency = time.Since(start)

	status := c.Writer.Status()
	if status >= 400 {
		securityLogger.Info("Request failed. Status: %d", status)
	} else {
		logger.Info("Finished request. Status: %d", status)
	}
}

type LogMetadata struct {
	Topic         string          `json:"topic"`
	CorrelationId string          `json:"correlationId"`
	Latency       time.Duration   `json:"latency,omitempty"`
	Path          string          `json:"path,omitempty"`
	UserID        int             `json:"userID,omitempty"`
	Role          models.UserRole `json:"role,omitempty"`
	ClientIP      string          `json:"clientIP,omitempty"`
	Method        string          `json:"method,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
	BodySize      int             `json:"bodySize,omitempty"`
	Status        int             `json:"status,omitempty"`
}

func (m *LogMetadata) String() string {
	return fmt.Sprintf("%s, %s %s (%d) | u=(%s %d) e=%s", m.Topic, m.Method, m.Path, m.Status, m.Role, m.UserID, m.ErrorMessage)
}

type RequestLogger struct {
	topic   string
	context *gin.Context
	latency time.Duration
}

func getMetadata(logger *RequestLogger) *LogMetadata {
	topic := logger.topic
	c := logger.context
	latency := logger.latency
	requestContext := ginext.GetRequestContext(c)
	authentication := requestContext.Authentication
	path := c.FullPath()
	status := 0
	if c.Writer.Written() {
		status = c.Writer.Status()
	}

	metadata := LogMetadata{
		Topic:         topic,
		CorrelationId: c.GetHeader("X-Correlation-Id"),
		Path:          path,
		ClientIP:      c.ClientIP(),
		Method:        c.Request.Method,
		Status:        status,
		ErrorMessage:  c.Errors.ByType(gin.ErrorTypePrivate).String(),
		BodySize:      c.Writer.Size(),
		Latency:       latency,
	}
	if authentication != nil {
		metadata.UserID = int(authentication.AppUser.AppUserID)
		metadata.Role = authentication.AppUser.Role
	}
	return &metadata
}

func (logger *RequestLogger) Info(format string, args ...interface{}) {
	util.WriteLog(getMetadata(logger), "INFO", format, args...)
}

func (logger *RequestLogger) Warn(format string, args ...interface{}) {
	util.WriteLog(getMetadata(logger), "WARN", format, args...)
}

func (logger *RequestLogger) Err(e error) {
	util.WriteLog(getMetadata(logger), "ERROR", e.Error())
}
func (logger *RequestLogger) Recovery(p interface{}, stack []byte) {
	util.WriteLog(logger, "ERROR", fmt.Sprintf("panic: %v\n%v", p, string(stack)))
}
