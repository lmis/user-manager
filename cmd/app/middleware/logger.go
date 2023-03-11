package middleware

import (
	"net/http"
	ginext "user-manager/cmd/app/gin-extensions"
	domain_model "user-manager/domain-model"
	"user-manager/util/logger"

	"time"

	"github.com/gin-gonic/gin"
)

func RegisterLoggerMiddleware(app *gin.Engine) {
	app.Use(LoggerMiddlware)
}

func LoggerMiddlware(c *gin.Context) {
	// Add logger to context
	logger := &RequestLogger{topic: "REQUEST", context: c}
	securityLogger := &RequestLogger{topic: "SECURITY", context: c}

	requestContext := ginext.GetRequestContext(c)
	requestContext.Log = domain_model.Log(logger)
	requestContext.SecurityLog = domain_model.SecurityLog(securityLogger)

	logger.Info("Starting request")

	// Start timer
	start := time.Now()

	// Process request
	c.Next()

	// Stop timer
	logger.latency = time.Since(start)

	status := c.Writer.Status()
	logger.status = status
	if status >= 400 {
		securityLogger.Info("Request failed. Status: %d", status)
	} else {
		logger.Info("Finished request. Status: %d", status)
	}

	// Trigger alerts
	if status == http.StatusInternalServerError {
		if err := c.Errors.Last(); err != nil {
			securityLogger.Err(err.Err)
		}
	}
}

type LogMetadata struct {
	Topic         string                  `json:"topic"`
	CorrelationId string                  `json:"correlationId"`
	Latency       time.Duration           `json:"latency,omitempty"`
	Path          string                  `json:"path,omitempty"`
	UserID        int                     `json:"userID,omitempty"`
	Roles         []domain_model.UserRole `json:"role,omitempty"`
	ClientIP      string                  `json:"clientIP,omitempty"`
	Method        string                  `json:"method,omitempty"`
	ErrorMessage  string                  `json:"errorMessage,omitempty"`
	BodySize      int                     `json:"bodySize,omitempty"`
	Status        int                     `json:"status,omitempty"`
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
	userSession := requestContext.UserSession
	path := c.FullPath()

	metadata := LogMetadata{
		Topic:         topic,
		CorrelationId: c.GetHeader("X-Correlation-Id"),
		Path:          path,
		ClientIP:      c.ClientIP(),
		Method:        c.Request.Method,
		Status:        logger.status,
		ErrorMessage:  c.Errors.String(),
		BodySize:      c.Writer.Size(),
		Latency:       logger.latency,
	}
	if userSession.IsPresent {
		metadata.UserID = int(userSession.OrPanic().User.AppUserID)
		metadata.Roles = userSession.OrPanic().User.UserRoles
	}
	return &metadata
}

func (r *RequestLogger) Info(format string, args ...interface{}) {
	logger.WriteLog(getMetadata(r), logger.LOG_LEVEL_INFO, format, args...)
}

func (r *RequestLogger) Warn(format string, args ...interface{}) {
	logger.WriteLog(getMetadata(r), logger.LOG_LEVEL_WARN, format, args...)
}

func (r *RequestLogger) Err(e error) {
	logger.WriteLog(getMetadata(r), logger.LOG_LEVEL_ERROR, "%s", e.Error())
}
