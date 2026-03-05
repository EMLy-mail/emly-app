package main

import (
	"fmt"
	"time"

	pkglogger "emly/backend/logger"
)

// InitLogger initializes the shared logger.
func InitLogger() error { return pkglogger.InitLogger() }

// CloseLogger closes the log file.
func CloseLogger() { pkglogger.CloseLogger() }

// Log prints a timestamped, file:line tagged log line.
// Uses depth 2 so the reported location is the caller of this wrapper.
func Log(args ...any) { pkglogger.LogDepth(2, args...) }

// FrontendLog allows the frontend to send logs to the backend logger.
func (a *App) FrontendLog(level string, message string) {
	now := time.Now()
	date := now.Format("2006-01-02")
	tm := now.Format("15:04:05")
	pkglogger.Logger.Printf("[%s] - [%s] - [FRONTEND] - [%s] %s", date, tm, level, fmt.Sprint(message))
}
