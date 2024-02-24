package logger

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var logJson bool
var out *log.Logger = log.New(os.Stdout, "", 0)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

type Logger interface {
	Info(message string)
	Warn(message string)
	Err(e error)
}

type LogData struct {
	Level       LogLevel    `json:"level"`
	Message     string      `json:"message"`
	JsonPayload interface{} `json:"jsonPayload"`
	Timestamp   time.Time   `json:"timestamp"`
	LineNumber  string      `json:"lineNumber"`
	FileName    string      `json:"fileName"`
}

const (
	termCodeGreen   = "\033[0;32m"
	termCodeWhite   = "\033[0;37m"
	termCodeYellow  = "\033[0;33m"
	termCodeRed     = "\033[0;31m"
	termCodeBlue    = "\033[0;34m"
	termCodeMagenta = "\033[0;35m"
	termCodeCyan    = "\033[0;36m"
	termCodeReset   = "\033[0m"
)

type SimpleLogger struct {
	Topic string
}

func (logger SimpleLogger) Info(message string) {
	WriteLog(logger.Topic, LogLevelInfo, message)
}

func (logger SimpleLogger) Warn(message string) {
	WriteLog(logger.Topic, LogLevelWarn, message)
}

func (logger SimpleLogger) Err(e error) {
	WriteLog(logger.Topic, LogLevelError, e.Error())

}

func NewLogger(topic string) Logger {
	return SimpleLogger{
		topic,
	}
}

func WriteLog(metadata interface{}, level LogLevel, message string) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("Panic while printing log line. Level=%s message=%s", level, message)
		}
	}()
	_, file, number, _ := runtime.Caller(2)

	fileParts := strings.Split(file, "/")
	fileName := fileParts[len(fileParts)-1]
	lineNumber := strconv.Itoa(number)

	if logJson {
		logData, err := json.Marshal(&LogData{
			Level:       level,
			JsonPayload: metadata,
			Timestamp:   time.Now(),
			Message:     message,
			FileName:    fileName,
			LineNumber:  lineNumber,
		})

		if err == nil {
			out.Println(string(logData))
			return
		}
	}

	colorStart := ""
	colorEnd := ""
	switch level {
	case LogLevelInfo:
	case LogLevelWarn:
		colorStart = termCodeYellow
	case LogLevelError:
		colorStart = termCodeRed
	}
	if colorStart != "" {
		colorEnd = termCodeReset
	}
	log.Printf(
		"%s%s:%s [%s - %v] %s%s",
		colorStart,
		fileName,
		lineNumber,
		level,
		metadata,
		message,
		colorEnd,
	)
}

func SetLogJSON(enable bool) {
	logJson = enable
}
