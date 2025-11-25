package log

import (
	"errors"
	"fmt"
	"log"
	"main/globals"
	"os"
	"sync"
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
	Logger          = log.New(os.Stdout, "", 0)
	cliLogging      = true
	logFilePath     = ""
	logMutex        sync.RWMutex
)
var LogFile *os.File

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

	if len(message) > 1024 {
		message = message[:1024] + "... [truncated]"
	}

	logMutex.RLock()
	isCliLogging := cliLogging
	logMutex.RUnlock()

	if isCliLogging {
		return fmt.Sprintf("[AIKIDO][%s] %s\n", levelStr, message)
	}
	return fmt.Sprintf("[AIKIDO][%s][%s] %s\n", levelStr, time.Now().Format("15:04:05"), message)
}

func initLogFile() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if cliLogging {
		return
	}
	if LogFile != nil {
		return
	}
	var err error
	LogFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	Logger.SetOutput(LogFile)
}

func logMessage(level LogLevel, args ...interface{}) {
	logMutex.RLock()
	lvl := currentLogLevel
	logMutex.RUnlock()

	if level >= lvl {
		initLogFile()
		formatter := &AikidoFormatter{}
		message := fmt.Sprint(args...)
		formattedMessage := formatter.Format(level, message)
		Logger.Print(formattedMessage)
	}
}

func logMessagef(level LogLevel, format string, args ...interface{}) {
	logMutex.RLock()
	lvl := currentLogLevel
	logMutex.RUnlock()

	if level >= lvl {
		initLogFile()
		formatter := &AikidoFormatter{}
		message := fmt.Sprintf(format, args...)
		formattedMessage := formatter.Format(level, message)
		Logger.Print(formattedMessage)
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

// SetLogLevel changes the current log level (thread-safe)
func SetLogLevel(level string) error {
	var newLevel LogLevel

	switch level {
	case "DEBUG":
		newLevel = DebugLevel
	case "INFO":
		newLevel = InfoLevel
	case "WARN":
		newLevel = WarnLevel
	case "ERROR":
		newLevel = ErrorLevel
	default:
		return errors.New("invalid log level")
	}

	logMutex.Lock()
	defer logMutex.Unlock()
	currentLogLevel = newLevel
	return nil
}

func Init(diskLogs bool) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if !diskLogs {
		cliLogging = true
		return
	}
	cliLogging = false
	currentTime := time.Now()
	timeStr := currentTime.Format("20060102150405")
	logFilePath = fmt.Sprintf("/var/log/aikido-"+globals.Version+"/aikido-request-processor-%s-%d.log", timeStr, os.Getpid())
}

func Uninit() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if LogFile != nil {
		LogFile.Close()
		LogFile = nil
	}
}
