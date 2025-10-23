package log

import (
	"fmt"
	"log"
	"main/constants"
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

type AikidoLogger struct {
	level   int32
	logger  *log.Logger
	logFile *os.File
}

var MainLogger *AikidoLogger = nil

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

func getCurrentLogger(serverLogger *AikidoLogger) *AikidoLogger {
	if serverLogger != nil {
		return serverLogger
	}
	return MainLogger
}

func logMessage(serverLogger *AikidoLogger, level int32, args ...interface{}) {
	currentLogger := getCurrentLogger(serverLogger)

	if level >= atomic.LoadInt32(&currentLogger.level) {
		formatter := &AikidoFormatter{}
		message := fmt.Sprint(args...)
		formattedMessage := formatter.Format(level, message)
		currentLogger.logger.Print(formattedMessage)
	}
}

func logMessagef(serverLogger *AikidoLogger, level int32, format string, args ...interface{}) {
	currentLogger := getCurrentLogger(serverLogger)

	if level >= atomic.LoadInt32(&currentLogger.level) {
		formatter := &AikidoFormatter{}
		message := fmt.Sprintf(format, args...)
		formattedMessage := formatter.Format(level, message)
		currentLogger.logger.Print(formattedMessage)
	}
}

func Debug(serverLogger *AikidoLogger, args ...interface{}) {
	logMessage(serverLogger, DebugLevel, args...)
}

func Info(serverLogger *AikidoLogger, args ...interface{}) {
	logMessage(serverLogger, InfoLevel, args...)
}

func Warn(serverLogger *AikidoLogger, args ...interface{}) {
	logMessage(serverLogger, WarnLevel, args...)
}

func Error(serverLogger *AikidoLogger, args ...interface{}) {
	logMessage(serverLogger, ErrorLevel, args...)
}

func Debugf(serverLogger *AikidoLogger, format string, args ...interface{}) {
	logMessagef(serverLogger, DebugLevel, format, args...)
}

func Infof(serverLogger *AikidoLogger, format string, args ...interface{}) {
	logMessagef(serverLogger, InfoLevel, format, args...)
}

func Warnf(serverLogger *AikidoLogger, format string, args ...interface{}) {
	logMessagef(serverLogger, WarnLevel, format, args...)
}

func Errorf(serverLogger *AikidoLogger, format string, args ...interface{}) {
	logMessagef(serverLogger, ErrorLevel, format, args...)
}

func GetIntLogLevel(level string) int32 {
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
	}
	return levelInt
}

func CreateLogger(tag string, logLevel string, diskLogs bool) *AikidoLogger {
	currentLogger := &AikidoLogger{
		level:   GetIntLogLevel(logLevel),
		logger:  log.New(os.Stdout, "", 0),
		logFile: nil,
	}

	if !diskLogs {
		return currentLogger
	}

	currentTime := time.Now()
	timeStr := currentTime.Format("20060102150405")
	logFilePath := fmt.Sprintf("/var/log/aikido-%s/aikido-agent-%s-%d-%s.log", constants.Version, timeStr, os.Getpid(), tag)

	var err error
	currentLogger.logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		currentLogger.logFile = nil
		return currentLogger
	}

	currentLogger.logger.SetOutput(currentLogger.logFile)
	return currentLogger
}

func DestroyLogger(currentLogger *AikidoLogger) {
	if currentLogger.logFile == nil {
		return
	}
	currentLogger.logger.SetOutput(os.Stdout)
	currentLogger.logFile.Close()
	currentLogger.logFile = nil
}

func Init() {
	MainLogger = CreateLogger("main", "INFO", true)
}

func Uninit() {
	DestroyLogger(MainLogger)
}
