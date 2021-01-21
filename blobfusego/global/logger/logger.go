package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel : Level of logging
type LogLevel int

var logChannel chan (string)
var logWorkerdone sync.WaitGroup
var logChannelLock sync.RWMutex

// Severity levels of log messages.
const (
	LogLvlDebug LogLevel = 1 + iota
	LogLvlInfo
	LogLvlWarn
	LogLvlErr
	LogLvlCrit
)

// LogLevelNames : Names of different log levels
var LogLevelNames = []string{
	"LOG_DEBUG", "LOG_INFO", "LOG_WARN", "LOG_ERR", "LOG_CRIT",
}

// GetLogLevel : From strin get the log level
func getLogLevel(lvl string) LogLevel {
	lvl = strings.ToUpper(strings.Trim(lvl, " "))

	switch lvl {
	case "LOG_DEBUG":
		return LogLvlDebug
	case "LOG_INFO":
		return LogLvlInfo
	case "LOG_WARN":
		return LogLvlWarn
	case "LOG_WARNING":
		return LogLvlWarn
	case "LOG_ERR":
		return LogLvlErr
	case "LOG_CRIT":
		return LogLvlCrit
	default:
		fmt.Println("Invalid logging level, default to LOG_WARN")
		return LogLvlWarn
	}
}

// getLogString : Convert lov level to its corrosponding string
func getLogString(lvl LogLevel) string {
	return LogLevelNames[lvl-1]
}

// LogConfig : Configuration to be provided to logging infra
type LogConfig struct {
	LogLevel     string
	LogFile      string
	LogSizeMB    int
	LogFileCount int
}

// Logger : Global logger structure holding the logging configuration
var Logger struct {
	level   LogLevel
	logger  *log.Logger
	LogFile io.WriteCloser
	ProcPID int
}

// StartLogger : Initialize the global logger
func StartLogger(cfg LogConfig) {
	Logger.level = getLogLevel(cfg.LogLevel)
	Logger.ProcPID = os.Getpid()

	// If a path is specified create a handle to the writer.
	if cfg.LogFile != "" {

		var err error
		Logger.LogFile, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			Logger.LogFile = os.Stdout
		}
	} else {
		Logger.LogFile = os.Stdout
	}

	Logger.logger = log.New(Logger.LogFile, "", 0)
	logChannel = make(chan string)
	go logDumper(1, logChannel)
}

// StopLogger : Deinit the logger
func StopLogger() error {
	close(logChannel)
	logWorkerdone.Wait()

	if err := Logger.LogFile.Close(); err != nil {
		return err
	}
	return nil
}

func logDumper(id int, logChannel <-chan string) {
	defer logWorkerdone.Done()

	//fmt.Println("Log Dumper started")

	for j := range logChannel {
		Logger.logger.Println(j)
	}

	//fmt.Println("Log Dumper closed")
}

// EnqeueLog : Dump the log to screen or file as configured
func EnqeueLog(fomat string, lvl LogLevel, args ...interface{}) {
	// Only log if the log level matches the log request
	if lvl >= Logger.level {
		_, fn, ln, _ := runtime.Caller(2)

		msg := fmt.Sprintf(fomat, args...)
		msg = fmt.Sprintf("%s : %d : %s [%s (%d)]: %s",
			time.Now().Format(time.UnixDate),
			Logger.ProcPID,
			getLogString(lvl),
			filepath.Base(fn), ln,
			msg)

		logChannelLock.Lock()
		logChannel <- msg
		logChannelLock.Unlock()

	}
}

// LogDebug : Debug message logging
func LogDebug(msg string, args ...interface{}) {
	EnqeueLog(msg, LogLvlDebug, args...)
}

// LogInfo : Info message logging
func LogInfo(msg string, args ...interface{}) {
	EnqeueLog(msg, LogLvlInfo, args...)
}

// LogWarn : Warning message logging
func LogWarn(msg string, args ...interface{}) {
	EnqeueLog(msg, LogLvlWarn, args...)
}

// LogErr : Error message logging
func LogErr(msg string, args ...interface{}) {
	EnqeueLog(msg, LogLvlErr, args...)
}

// LogCrit : Critical message logging
func LogCrit(msg string, args ...interface{}) {
	EnqeueLog(msg, LogLvlCrit, args...)
}

// GetLoggerObj : Get the logger object for other libs
func GetLoggerObj() *log.Logger {
	return Logger.logger
}
