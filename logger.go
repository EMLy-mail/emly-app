package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var logger = log.New(os.Stdout, "", 0)

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
