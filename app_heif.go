package main

import (
	"path/filepath"
	"sync"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils/heif"
)

// heifToDisplayJPEG converts HEIC/HEIF bytes (detected by content) to JPEG
// for display in the WebView, which cannot render HEIC. Anything else, a
// missing/broken libheif.dll, or a decode failure returns data unchanged with
// converted=false: the frontend then falls back to its libheif-js (WASM)
// decoder, so a broken or exotic file is never worse off than before.
//
// Only for display paths: saving or opening an attachment externally must
// keep the original bytes.
func heifToDisplayJPEG(data []byte, filename string) (out []byte, converted bool) {
	if !heif.IsHEIF(data) {
		return data, false
	}
	// Loads libheif.dll on the first HEIC only, so the DLL costs nothing
	// for users who never open one.
	logHeifStatusOnce.Do(logHeifStatus)
	if !heif.Available() {
		return data, false
	}

	start := time.Now()
	jpg, err := heif.ToJPEG(data, heif.DefaultJPEGQuality)
	if err != nil {
		pkglogger.Warn("heif decode failed, serving original bytes",
			"file", filepath.Base(filename),
			"size", len(data),
			"error", err.Error(),
		)
		return data, false
	}
	pkglogger.Debug("heif converted to jpeg",
		"file", filepath.Base(filename),
		"in_size", len(data),
		"out_size", len(jpg),
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return jpg, true
}

var logHeifStatusOnce sync.Once

func logHeifStatus() {
	if err := heif.LoadError(); err != nil {
		pkglogger.Warn("libheif unavailable, HEIC falls back to the frontend decoder",
			"error", err.Error())
		return
	}
	pkglogger.Info("libheif loaded", "version", heif.Version())
}
