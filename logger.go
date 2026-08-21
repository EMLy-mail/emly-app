package main

import (
	"time"

	pkglogger "emly/backend/logger"
)

// InitLogger initialises the global structured logger.
func InitLogger() error { return pkglogger.InitLogger() }

// CloseLogger closes the log file handle.
func CloseLogger() { pkglogger.CloseLogger() }

// Log emits an unstructured Info-level message (legacy compatibility).
// Prefer structured helpers (pkglogger.Info, etc.) for new code.
func Log(args ...any) { pkglogger.LogDepth(2, args...) }

// ---------------------------------------------------------------------------
// Frontend → Backend log bridge
// ---------------------------------------------------------------------------

// FrontendLog receives a log entry from the frontend and writes it using slog.
// The frontend should call this via the Wails binding:
//
//	await FrontendLog("INFO", "email loaded", '{"url":"/","userAgent":"..."}')
func (a *App) FrontendLog(level string, message string, contextJSON string) {
	args := []any{"source", "frontend"}
	if contextJSON != "" {
		args = append(args, "context", contextJSON)
	}

	switch level {
	case "DEBUG":
		pkglogger.Debug(message, args...)
	case "WARN":
		pkglogger.Warn(message, args...)
	case "ERROR":
		pkglogger.Error(message, args...)
	default:
		pkglogger.Info(message, args...)
	}
}

// ---------------------------------------------------------------------------
// Startup trace bridge
// ---------------------------------------------------------------------------

// TraceStartupStep records a checkpoint in the startup trace file (see
// backend/logger/trace.go, startup-trace.log next to app.log) - a plain-text
// timeline separate from the regular log, built to show exactly how long
// each step of app launch and mail loading took. The frontend calls this
// from the SvelteKit load path so its steps (GetStartupFile, ReadEML/MSG,
// DOMPurify sanitize, host integrity check, mount) land in the same
// chronological file as the Go-side steps.
//
//	await TraceStartupStep("fe_read_mail_done", "attachments=3")
//
// step should be a short machine-friendly name; detail is optional free text
// (sizes, counts, format) - pass "" when there's nothing to add. No-op if
// tracing wasn't initialised for this process (e.g. viewer windows).
func (a *App) TraceStartupStep(step string, detail string) {
	if detail != "" {
		pkglogger.TraceStep(step, detail)
	} else {
		pkglogger.TraceStep(step)
	}
}

// ---------------------------------------------------------------------------
// Canonical Log Line decorator
// ---------------------------------------------------------------------------

// canonicalLog emits a canonical log line at the end of a Wails-bound method.
// Usage inside any App method:
//
//	func (a *App) ReadEML(filePath string) (*mailfmt.EmailData, error) {
//	    start := time.Now()
//	    defer func() { canonicalLog("ReadEML", start, err) }()
//	    ...
//	}
func canonicalLog(fn string, start time.Time, err error) {
	fields := pkglogger.CanonicalFields(fn, start, err)
	if err != nil {
		pkglogger.Error("canonical_line", fields...)
	} else {
		pkglogger.Info("canonical_line", fields...)
	}
}
