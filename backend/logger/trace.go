// Startup tracing: a plain-text, human-readable timeline of the app's
// launch and mail-loading path, kept in its own file (startup-trace.log)
// separate from the structured app.log. Where app.log answers "what
// happened", this file answers "how long did each step take" - built to
// pinpoint exactly which step a heavy .eml/.msg gets stuck on, instead of
// guessing from log noise.
package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// StartupTracer writes timestamped checkpoints to a plain-text file.
// Every write is fsync'd immediately: if the app hangs mid-load - the exact
// symptom this is built to diagnose - the trace up to that point is still
// on disk, instead of sitting lost in an OS write buffer.
type StartupTracer struct {
	mu   sync.Mutex
	f    *os.File
	t0   time.Time
	last time.Time
}

// trace is the process-wide tracer. Left nil (and every TraceStep call a
// silent no-op) unless InitStartupTrace was called - viewer windows
// (image/PDF) run as separate processes and never call it, so they don't
// race the main window to truncate the same file.
var trace *StartupTracer

// InitStartupTrace creates (truncating) the trace file at path. t0 becomes
// the reference point every step's "total elapsed" is measured against -
// pass the timestamp taken at the very top of main(), before config/logger
// setup, so the trace covers the true process start, not just the point
// this function happens to be called from.
//
// The file is truncated fresh on every launch: unlike app.log it never
// needs history, only the most recent startup - the one being reproduced
// right now.
func InitStartupTrace(path string, t0 time.Time) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("logger: create startup trace: %w", err)
	}
	trace = &StartupTracer{f: f, t0: t0, last: t0}
	fmt.Fprintf(f, "# EMLy startup trace - %s\n", t0.Format("2006-01-02 15:04:05.000"))
	fmt.Fprintf(f, "# %8s  %8s  step (detail)\n", "total", "delta")
	fmt.Fprintf(f, "# total = elapsed since process start; delta = elapsed since previous step (the one to look at)\n")
	_ = f.Sync()
	return nil
}

// TraceStep appends one checkpoint: elapsed time since process start and
// since the previous step, plus an optional detail (sizes, counts, file
// names - whatever helps explain why a step was slow). No-op if tracing
// was never initialised for this process.
func TraceStep(step string, detail ...any) {
	if trace == nil {
		return
	}
	trace.mu.Lock()
	defer trace.mu.Unlock()

	now := time.Now()
	totalMs := now.Sub(trace.t0).Milliseconds()
	deltaMs := now.Sub(trace.last).Milliseconds()
	trace.last = now

	line := fmt.Sprintf("%8dms  %8dms  %s", totalMs, deltaMs, step)
	if len(detail) > 0 {
		parts := make([]string, 0, len(detail))
		for _, d := range detail {
			if d == nil {
				continue
			}
			if s := fmt.Sprint(d); s != "" {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			line += "  (" + strings.Join(parts, " ") + ")"
		}
	}
	fmt.Fprintln(trace.f, line)
	_ = trace.f.Sync()
}

// CloseStartupTrace closes the trace file handle. Safe to call even if
// tracing was never initialised.
func CloseStartupTrace() {
	if trace != nil {
		_ = trace.f.Close()
	}
}
