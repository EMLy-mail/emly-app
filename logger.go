package main

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
	logger  = log.New(os.Stdout, "", 0)
	logFile *os.File
)

// InitLogger initializes the logger to write to both stdout and a file in AppData
func InitLogger() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(configDir, "EMLy")
	logsDir := filepath.Join(appDir, "logs")

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logsDir, "app.log")
	// Open file in Append mode
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = file

	// MultiWriter to write to both stdout and file
	multi := io.MultiWriter(os.Stdout, file)
	logger = log.New(multi, "", 0)

	Log("Logger initialized. Writing to: " + logPath)
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// Log prints a timestamped, file:line tagged log line.
func Log(args ...any) {
	now := time.Now()
	date := now.Format("2006-01-02")
	tm := now.Format("15:04:05")

	_, file, line, ok := runtime.Caller(1)
	loc := "unknown"
	if ok {
		loc = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	msg := fmt.Sprintln(args...)
	logger.Printf("[%s] - [%s] - [%s] - %s", date, tm, loc, strings.TrimRight(msg, "\n"))
}

// FrontendLog allows the frontend to send logs to the backend logger
func (a *App) FrontendLog(level string, message string) {
	now := time.Now()
	date := now.Format("2006-01-02")
	tm := now.Format("15:04:05")

	// We don't use runtime.Caller here because it would point to this function
	// Instead we tag it as [FRONTEND]
	logger.Printf("[%s] - [%s] - [FRONTEND] - [%s] %s", date, tm, level, message)
}
