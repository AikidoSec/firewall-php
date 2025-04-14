package log

import (
	"errors"
	"fmt"
	"log"
	"main/globals"
	"os"
	"sync/atomic"
	"time"
)

const (
	DebugLevel int32 = 0
	InfoLevel  int32 = 1
	WarnLevel  int32 = 2
	ErrorLevel int32 = 3
)

var (
	currentLogLevel int32 = ErrorLevel
	logger                = log.New(os.Stdout, "", 0)
	logFile         *os.File
)

type AikidoFormatter struct{}

func (f *AikidoFormatter) Format(level int32, message string) string {
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

	logMessage := fmt.Sprintf("[AIKIDO][%s][%s] %s\n", levelStr, time.Now().Format("15:04:05"), message)
	return logMessage
}

func logMessage(level int32, args ...interface{}) {
	if level >= atomic.LoadInt32(&currentLogLevel) {
		formatter := &AikidoFormatter{}
		message := fmt.Sprint(args...)
		formattedMessage := formatter.Format(level, message)
		logger.Print(formattedMessage)
	}
}

func logMessagef(level int32, format string, args ...interface{}) {
	if level >= atomic.LoadInt32(&currentLogLevel) {
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
	levelInt := ErrorLevel
	switch level {
	case "DEBUG":
		levelInt = DebugLevel
	case "INFO":
		levelInt = InfoLevel
	case "WARN":
		levelInt = WarnLevel
	case "ERROR":
		levelInt = ErrorLevel
	default:
		return errors.New("invalid log level")
	}
	atomic.StoreInt32(&currentLogLevel, levelInt)
	return nil
}

func Init() {
	if !globals.EnvironmentConfig.DiskLogs {
		return
	}
	currentTime := time.Now()
	timeStr := currentTime.Format("20060102150405")
	logFilePath := fmt.Sprintf("/var/log/aikido-%s/aikido-agent-%s-%d.log", globals.Version, timeStr, os.Getpid())

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	logger.SetOutput(logFile)
}

func Uninit() {
	if !globals.EnvironmentConfig.DiskLogs {
		return
	}
	logger.SetOutput(os.Stdout)
	logFile.Close()
}
