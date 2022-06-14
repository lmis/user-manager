package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var logJson bool
var out *log.Logger = log.New(os.Stdout, "", 0)

type Logger interface {
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Err(e error)
}

type LogData struct {
	Level       string      `json:"level"`
	Message     string      `json:"message"`
	JsonPayload interface{} `json:"jsonPayload"`
	Timestamp   time.Time   `json:"timestamp"`
	LineNumber  string      `json:"lineNumber"`
	FileName    string      `json:"fileName"`
}

type SimpleLogger struct {
	Topic string `json:"topic"`
}

func (logger *SimpleLogger) Info(format string, args ...interface{}) {
	WriteLog(logger, "INFO", format, args...)
}

func (logger *SimpleLogger) Warn(format string, args ...interface{}) {
	WriteLog(logger, "WARN", format, args...)
}

func (logger *SimpleLogger) Err(e error) {
	WriteLog(logger, "ERROR", e.Error())
}

func Log(topic string) Logger {
	return &SimpleLogger{
		topic,
	}
}

func WriteLog(metadata interface{}, level string, format string, args ...interface{}) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("Panic while printing log line. Level=%s format=%s", level, format)
		}
	}()
	message := fmt.Sprintf(format, args...)
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

	log.Printf("%s:%s [%s]%v %s", fileName, lineNumber, level, metadata, message)
}

func SetLogJSON(enable bool) {
	logJson = enable
}
