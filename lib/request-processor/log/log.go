package log

import (
	"errors"
	"fmt"
	"syscall"

	"main/globals"
	"main/instance"
	"os"
	"time"
)

type AikidoFormatter struct{}

func (f *AikidoFormatter) Format(level globals.LogLevel, threadID uint64, message string) string {
	var levelStr string
	switch level {
	case globals.LogDebugLevel:
		levelStr = "DEBUG"
	case globals.LogInfoLevel:
		levelStr = "INFO"
	case globals.LogWarnLevel:
		levelStr = "WARN"
	case globals.LogErrorLevel:
		levelStr = "ERROR"
	default:
		return "invalid log level"
	}

	if len(message) > 1024 {
		message = message[:1024] + "... [truncated]"
	}

	globals.LogMutex.RLock()
	isCliLogging := globals.CliLogging
	globals.LogMutex.RUnlock()

	if isCliLogging {
		return fmt.Sprintf("[AIKIDO][%s][tid:%d] %s\n", levelStr, threadID, message)
	}
	return fmt.Sprintf("[AIKIDO][%s][tid:%d][%s] %s\n", levelStr, threadID, time.Now().Format("15:04:05"), message)
}

func initLogFile() {
	globals.LogMutex.Lock()
	defer globals.LogMutex.Unlock()

	if globals.CliLogging {
		return
	}
	if globals.LogFile != nil {
		return
	}
	var err error
	globals.LogFile, err = os.OpenFile(globals.LogFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	globals.Logger.SetOutput(globals.LogFile)
}

func logMessage(instance *instance.RequestProcessorInstance, level globals.LogLevel, args ...interface{}) {
	globals.LogMutex.RLock()
	lvl := globals.CurrentLogLevel
	globals.LogMutex.RUnlock()

	if level >= lvl {
		initLogFile()
		formatter := &AikidoFormatter{}
		message := fmt.Sprint(args...)
		threadID := uint64(0)
		if instance != nil {
			threadID = instance.GetThreadID()
		} else {
			threadID = uint64(syscall.Gettid())
		}
		formattedMessage := formatter.Format(level, threadID, message)
		globals.Logger.Print(formattedMessage)
	}
}

func logMessagef(instance *instance.RequestProcessorInstance, level globals.LogLevel, format string, args ...interface{}) {
	globals.LogMutex.RLock()
	lvl := globals.CurrentLogLevel
	globals.LogMutex.RUnlock()

	if level >= lvl {
		initLogFile()
		formatter := &AikidoFormatter{}
		message := fmt.Sprintf(format, args...)
		threadID := uint64(0)
		if instance != nil {
			threadID = instance.GetThreadID()
		} else {
			threadID = uint64(syscall.Gettid())
		}
		formattedMessage := formatter.Format(level, threadID, message)
		globals.Logger.Print(formattedMessage)
	}
}

func Debug(instance *instance.RequestProcessorInstance, args ...interface{}) {
	logMessage(instance, globals.LogDebugLevel, args...)
}

func Info(instance *instance.RequestProcessorInstance, args ...interface{}) {
	logMessage(instance, globals.LogInfoLevel, args...)
}

func Warn(instance *instance.RequestProcessorInstance, args ...interface{}) {
	logMessage(instance, globals.LogWarnLevel, args...)
}

func Error(instance *instance.RequestProcessorInstance, args ...interface{}) {
	logMessage(instance, globals.LogErrorLevel, args...)
}

func Debugf(instance *instance.RequestProcessorInstance, format string, args ...interface{}) {
	logMessagef(instance, globals.LogDebugLevel, format, args...)
}

func Infof(instance *instance.RequestProcessorInstance, format string, args ...interface{}) {
	logMessagef(instance, globals.LogInfoLevel, format, args...)
}

func Warnf(instance *instance.RequestProcessorInstance, format string, args ...interface{}) {
	logMessagef(instance, globals.LogWarnLevel, format, args...)
}

func Errorf(instance *instance.RequestProcessorInstance, format string, args ...interface{}) {
	logMessagef(instance, globals.LogErrorLevel, format, args...)
}

// SetLogLevel changes the current log level (thread-safe)
func SetLogLevel(level string) error {
	var newLevel globals.LogLevel

	switch level {
	case "DEBUG":
		newLevel = globals.LogDebugLevel
	case "INFO":
		newLevel = globals.LogInfoLevel
	case "WARN":
		newLevel = globals.LogWarnLevel
	case "ERROR":
		newLevel = globals.LogErrorLevel
	default:
		return errors.New("invalid log level")
	}

	globals.LogMutex.Lock()
	defer globals.LogMutex.Unlock()
	globals.CurrentLogLevel = newLevel
	return nil
}

func Init(diskLogs bool) {
	globals.LogMutex.Lock()
	defer globals.LogMutex.Unlock()

	if !diskLogs {
		globals.CliLogging = true
		return
	}
	globals.CliLogging = false
	currentTime := time.Now()
	timeStr := currentTime.Format("20060102150405")
	globals.LogFilePath = fmt.Sprintf("/var/log/aikido-"+globals.Version+"/aikido-request-processor-%s-%d.log", timeStr, os.Getpid())
}

func Uninit() {
	globals.LogMutex.Lock()
	defer globals.LogMutex.Unlock()

	if globals.LogFile != nil {
		globals.LogFile.Close()
		globals.LogFile = nil
	}
}
