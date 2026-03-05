package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	Logger  = log.New(os.Stdout, "", 0)
	logFile *os.File
)

// InitLogger initializes the logger to write to both stdout and a file in AppData.
func InitLogger() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	logsDir := filepath.Join(configDir, "EMLy", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logsDir, "app.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = file

	multi := io.MultiWriter(os.Stdout, file)
	Logger = log.New(multi, "", 0)

	Log("Logger initialized. Writing to: " + logPath)
	return nil
}

// CloseLogger closes the log file.
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogDepth prints a timestamped log line. skip is passed to runtime.Caller to
// determine the source location: 1 = direct caller, 2 = caller's caller, etc.
func LogDepth(skip int, args ...any) {
	now := time.Now()
	date := now.Format("2006-01-02")
	tm := now.Format("15:04:05")

	_, file, line, ok := runtime.Caller(skip)
	loc := "unknown"
	if ok {
		loc = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	msg := fmt.Sprintln(args...)
	Logger.Printf("[%s] - [%s] - [%s] - %s", date, tm, loc, strings.TrimRight(msg, "\n"))
}

// Log prints a timestamped, file:line tagged log line.
func Log(args ...any) {
	LogDepth(2, args...)
}
