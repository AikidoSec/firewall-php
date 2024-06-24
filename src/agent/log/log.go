package log

import (
	"errors"
	"fmt"
	"log"
	"main/globals"
	"os"
	"time"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var (
	currentLogLevel = ErrorLevel
	logger          = log.New(os.Stdout, "", 0)
	logFile         *os.File
)

type AikidoFormatter struct{}

func (f *AikidoFormatter) Format(level LogLevel, message string) string {
	var levelStr string
	switch level {
	case DebugLevel:
		levelStr = "DEBUG"
	case InfoLevel:
		levelStr = "INFO"
	case WarnLevel:
		levelStr = "WARN"
	case ErrorLevel:
		levelStr = "ERROR"
	default:
		return "invalid log level"
	}

	logMessage := fmt.Sprintf("[AIKIDO][%s][GO][%s] %s\n", levelStr, time.Now().Format("15:04:05"), message)
	return logMessage
}

func logMessage(level LogLevel, args ...interface{}) {
	if level >= currentLogLevel {
		formatter := &AikidoFormatter{}
		message := fmt.Sprint(args...)
		formattedMessage := formatter.Format(level, message)
		logger.Print(formattedMessage)
	}
}

func logMessagef(level LogLevel, format string, args ...interface{}) {
	if level >= currentLogLevel {
		formatter := &AikidoFormatter{}
		message := fmt.Sprintf(format, args...)
		formattedMessage := formatter.Format(level, message)
		logger.Print(formattedMessage)
	}
}

func Debug(args ...interface{}) {
	logMessage(DebugLevel, args...)
}

func Info(args ...interface{}) {
	logMessage(InfoLevel, args...)
}

func Warn(args ...interface{}) {
	logMessage(WarnLevel, args...)
}

func Error(args ...interface{}) {
	logMessage(ErrorLevel, args...)
}

func Debugf(format string, args ...interface{}) {
	logMessagef(DebugLevel, format, args...)
}

func Infof(format string, args ...interface{}) {
	logMessagef(InfoLevel, format, args...)
}

func Warnf(format string, args ...interface{}) {
	logMessagef(WarnLevel, format, args...)
}

func Errorf(format string, args ...interface{}) {
	logMessagef(ErrorLevel, format, args...)

}

func SetLogLevel(level string) error {
	switch level {
	case "DEBUG":
		currentLogLevel = DebugLevel
	case "INFO":
		currentLogLevel = InfoLevel
	case "WARN":
		currentLogLevel = WarnLevel
	case "ERROR":
		currentLogLevel = ErrorLevel
	default:
		return errors.New("invalid log level")
	}
	return nil
}

func Init() {
	logFile, err := os.OpenFile(globals.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger.SetOutput(logFile)
}

func Uninit() {
	logger.SetOutput(os.Stdout)
	logFile.Close()
}