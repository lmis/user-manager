package util

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var logJson bool

type Logger struct {
	getMetadata func() interface{}
}

type LogData struct {
	Level       string      `json:"level"`
	Message     string      `json:"message"`
	JsonPayload interface{} `json:"jsonPayload"`
	Timestamp   time.Time   `json:"timestamp"`
	LineNumber  string      `json:"lineNumber"`
	FileName    string      `json:"fileName"`
}

func Log(topic string) *Logger {
	return &Logger{
		getMetadata: func() interface{} {
			return struct {
				Topic string `json:"topic"`
			}{
				Topic: topic,
			}
		},
	}
}

func LogWithMetadata(getMetadata func() interface{}) *Logger {
	return &Logger{
		getMetadata: getMetadata,
	}
}

func (logger *Logger) log(level string, format string, args ...interface{}) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Panic while printing log line. Level=%s format=%s", level, format)
		}
	}()
	message := fmt.Sprintf(format, args...)
	_, file, number, _ := runtime.Caller(2)

	fileParts := strings.Split(file, "/")
	fileName := fileParts[len(fileParts)-1]
	lineNumber := strconv.Itoa(number)
	metadata := logger.getMetadata()

	if logJson {
		logData, err := json.Marshal(&LogData{
			Level:       level,
			JsonPayload: &metadata,
			Timestamp:   time.Now(),
			Message:     message,
			FileName:    fileName,
			LineNumber:  lineNumber,
		})

		if err == nil {
			fmt.Println(string(logData))
			return
		}
	}

	log.Printf("%s:%s [%s]%v %s", fileName, lineNumber, level, metadata, message)
}

func (logger *Logger) Info(format string, args ...interface{}) {
	logger.log("INFO", format, args...)
}
func (logger *Logger) Warn(format string, args ...interface{}) {
	logger.log("WARN", format, args...)
}
func (logger *Logger) Err(e error) {
	logger.log("ERROR", e.Error())
}

func SetLogJSON(enable bool) {
	logJson = enable
}
