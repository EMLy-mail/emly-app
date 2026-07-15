# Validazione semver di GUI_SEMVER in LoadConfig Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Validate `EMLy.GUISemver` in `LoadConfig` with `golang.org/x/mod/semver`; if the value isn't a valid semantic version, show a native error messagebox and terminate EMLy.

**Architecture:** Add `golang.org/x/mod` as a new module dependency (vendored, like every other dependency in this repo). Extract a pure, unit-testable `isValidGUIVersion(version string) bool` helper in `backend/utils/ini-reader.go`. Wire it into `LoadConfig` right after the config struct is populated: on failure, call a new `fatalInvalidGUIVersion` helper that shows a native Win32 `MessageBoxW` (no Wails `context.Context` needed — the first `LoadConfig` call happens in `main.go` before `wails.Run`, so no Wails runtime exists yet) and then `os.Exit(1)`. Because the check lives inside `LoadConfig` itself, it applies uniformly to the startup path and every runtime reload/binding call (`app.go:GetConfig`, `app.go:ReloadEMLyConfig`, `app_settings.go:ReloadConfig`, `app_heartbeat.go`, `tray.go`, `backend/utils/machine-identifier.go`) without touching any of those call sites.

**Tech Stack:** Go 1.26, `golang.org/x/mod/semver`, `gopkg.in/ini.v1` (existing), Win32 `user32.dll` via `syscall` (existing pattern in `backend/utils/screenshot_windows.go`).

## Global Constraints

- Validated field is `EMLy.GUISemver` (config key `GUI_SEMVER`) only — `SDK_DECODER_SEMVER` is out of scope (it's the mail-parsing SDK version, not "the current version" of EMLy).
- `GUI_SEMVER` is stored **without** a `v` prefix (e.g. `config.ini:4` has `GUI_SEMVER = 2.0.1`); `x/mod/semver.IsValid` requires a `v` prefix. Normalize by prepending `v` only if the value doesn't already start with `v` — don't double-prefix.
- The check must run inside `LoadConfig` itself (`backend/utils/ini-reader.go`), not at individual call sites, so it covers both startup (`main.go`, before `wails.Run` — no Wails `context.Context` available yet) and every runtime reload/binding call uniformly.
- On invalid version: show a native Win32 `MessageBoxW` (not `runtime.MessageDialog`, which needs a Wails context that doesn't exist at the first call site) with Italian text, then `os.Exit(1)`. `LoadConfig` never returns a recoverable error for this case — the process terminates before returning.
- Error messagebox text is in Italian, matching the existing UI dialog convention (see `app_viewer.go:395-399`).
- `vendor/` is gitignored in this repo — never `git add` anything under `vendor/`; only `go.mod`/`go.sum` are committed for the new dependency.
- Repo's vendored dependencies (including Wails v2 patches) must be regenerated via `powershell -File scripts/vendor-wails-patch.ps1`, not a bare `go mod vendor`, so the existing patches in `patches/*.patch` get reapplied.
- Go files in this repo's working tree are CRLF (and some have a UTF-8 BOM) due to `core.autocrlf=true`; **do not run `gofmt -w`** on existing files (it produces a massive spurious line-ending diff). New files created via the Write tool are LF/no-BOM and already gofmt-clean.
- `go` commands that touch the module graph (`go get`, `go list -m`) need `-mod=mod` in this repo — the vendor directory otherwise forces `-mod=vendor` automatically and those commands fail with `can't determine available versions using the vendor directory`. Plain build/test/vet commands don't need the flag.

---

### Task 1: Add `golang.org/x/mod` dependency and `isValidGUIVersion` helper (TDD)

**Files:**
- Modify: `backend/utils/ini-reader.go` (imports, new `isValidGUIVersion` function)
- Test: `backend/utils/ini-reader_test.go` (new)
- Modify: `go.mod`, `go.sum` (new dependency)

**Interfaces:**
- Produces: `isValidGUIVersion(version string) bool` in package `utils` — takes the raw `GUI_SEMVER` value (no `v` prefix expected, but tolerates one), returns whether it's a valid semver. Used by Task 2's `LoadConfig` wiring.

- [ ] **Step 1: Write the failing test**

Create `backend/utils/ini-reader_test.go`:

```go
package utils

import "testing"

func TestIsValidGUIVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{"plain semver", "2.0.1", true},
		{"semver with prerelease", "2.0.1-beta.1", true},
		{"empty string is rejected", "", false},
		{"non-numeric is rejected", "latest", false},
		{"already v-prefixed is tolerated", "v2.0.1", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidGUIVersion(tc.version)
			if got != tc.want {
				t.Errorf("isValidGUIVersion(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/utils/... -run TestIsValidGUIVersion -v`
Expected: FAIL — build error `undefined: isValidGUIVersion` (the function doesn't exist yet).

- [ ] **Step 3: Add the `golang.org/x/mod` dependency**

Run: `go get -mod=mod golang.org/x/mod@v0.38.0`
Expected: `go.mod` gains a `require golang.org/x/mod v0.38.0` line; `go.sum` gains matching entries.

- [ ] **Step 4: Implement `isValidGUIVersion`**

In `backend/utils/ini-reader.go`, replace the import block (current lines 3-10):

```go
import (
	"os"
	"path/filepath"

	"emly/backend/logger"

	"gopkg.in/ini.v1"
)
```

with:

```go
import (
	"os"
	"path/filepath"
	"strings"

	"emly/backend/logger"

	"golang.org/x/mod/semver"
	"gopkg.in/ini.v1"
)
```

Then add this function after the `EMLyConfig` struct (after current line 28, before `// LoadConfig reads...`):

```go
// isValidGUIVersion reports whether version — the raw GUI_SEMVER value
// from config.ini, normally stored without a "v" prefix (e.g. "2.0.1") —
// is a valid semantic version per golang.org/x/mod/semver, which requires
// the "v" prefix. A value that already starts with "v" is used as-is so
// this stays a no-op for callers that pass an already-prefixed string.
func isValidGUIVersion(version string) bool {
	v := version
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return semver.IsValid(v)
}
```

- [ ] **Step 5: Vendor the new dependency**

Run: `powershell -File scripts/vendor-wails-patch.ps1`
Expected: ends with `Done.` — this regenerates `vendor/` (now including `vendor/golang.org/x/mod/semver/`, since `isValidGUIVersion` imports it) and reapplies the three existing Wails patches.

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./backend/utils/... -run TestIsValidGUIVersion -v`
Expected: PASS, all 5 subtests green.

- [ ] **Step 7: Verify the wider build still compiles**

Run: `go build ./...`
Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add go.mod go.sum backend/utils/ini-reader.go backend/utils/ini-reader_test.go
git commit -m "feat: add semver validation helper for GUI_SEMVER"
```

(`vendor/` is gitignored — do not `git add` anything under it.)

---

### Task 2: Wire the check into `LoadConfig` with a native error messagebox

**Files:**
- Modify: `backend/utils/ini-reader.go` (imports, `LoadConfig`, new `fatalInvalidGUIVersion`)

**Interfaces:**
- Consumes: `isValidGUIVersion(version string) bool` from Task 1.
- Consumes: package-level `user32 = syscall.NewLazyDLL("user32.dll")`, already declared in `backend/utils/screenshot_windows.go` (same package `utils`) — reused here instead of opening a second handle to the same DLL.
- Produces: `fatalInvalidGUIVersion(version string)` — shows the error messagebox and calls `os.Exit(1)`; never returns.

- [ ] **Step 1: Add the native messagebox helper and wire it into `LoadConfig`**

In `backend/utils/ini-reader.go`, update the import block from Task 1 to add `fmt`, `syscall`, `unsafe`:

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"emly/backend/logger"

	"golang.org/x/mod/semver"
	"gopkg.in/ini.v1"
)
```

Replace the current `LoadConfig` function:

```go
// LoadConfig reads the config.ini file at the given path and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	logger.Log("LoadConfig path:", path)
	cfg, err := ini.Load(path)
	if err != nil {
		logger.Log("Fail to read file:", err)
		return nil, err
	}

	config := new(Config)
	if err := cfg.MapTo(config); err != nil {
		logger.Log("Fail to map config:", err)
		return nil, err
	}

	return config, nil
}
```

with:

```go
// LoadConfig reads the config.ini file at the given path and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	logger.Log("LoadConfig path:", path)
	cfg, err := ini.Load(path)
	if err != nil {
		logger.Log("Fail to read file:", err)
		return nil, err
	}

	config := new(Config)
	if err := cfg.MapTo(config); err != nil {
		logger.Log("Fail to map config:", err)
		return nil, err
	}

	if !isValidGUIVersion(config.EMLy.GUISemver) {
		logger.Log("Invalid GUI_SEMVER in config:", config.EMLy.GUISemver)
		fatalInvalidGUIVersion(config.EMLy.GUISemver)
	}

	return config, nil
}

// fatalInvalidGUIVersion shows a native Win32 error messagebox and
// terminates the process. LoadConfig is called both before Wails starts
// (main.go, before wails.Run) and from every runtime reload/binding call
// (Settings, tray, heartbeat), so a Wails context.Context isn't always
// available — hence a native dialog instead of runtime.MessageDialog.
// Reuses the user32.dll handle already opened in screenshot_windows.go.
func fatalInvalidGUIVersion(version string) {
	const (
		mbOK        = 0x00000000
		mbIconError = 0x00000010
	)

	title := "EMLy - Errore di configurazione"
	message := fmt.Sprintf(
		"La versione configurata (%q) non è un formato semver valido.\nControllare il campo GUI_SEMVER in config.ini.\n\nEMLy verrà chiuso.",
		version,
	)

	titlePtr, errTitle := syscall.UTF16PtrFromString(title)
	messagePtr, errMessage := syscall.UTF16PtrFromString(message)
	if errTitle == nil && errMessage == nil {
		messageBoxW := user32.NewProc("MessageBoxW")
		messageBoxW.Call(0, uintptr(unsafe.Pointer(messagePtr)), uintptr(unsafe.Pointer(titlePtr)), uintptr(mbOK|mbIconError))
	}

	os.Exit(1)
}
```

- [ ] **Step 2: Verify it builds and existing tests still pass**

Run: `go build ./...`
Expected: no errors.

Run: `go test ./backend/utils/... -v`
Expected: PASS (`TestIsValidGUIVersion` from Task 1 still green; no other test in this package touches `LoadConfig`'s new exit path).

- [ ] **Step 3: Manual verification of the exit path (not automatable via `go test` — it terminates the process)**

1. Back up `config.debug.ini`: `Copy-Item config.debug.ini config.debug.ini.bak`
2. Edit `config.debug.ini`, set `GUI_SEMVER = notaversion`
3. Build and run the debug binary (use this repo's existing dev/run workflow, e.g. `wails dev` or the project's `run` skill)
4. Confirm: a native messagebox titled "EMLy - Errore di configurazione" appears with the Italian message quoting `notaversion`, and the app process exits after dismissing it
5. Restore the config: `Copy-Item config.debug.ini.bak config.debug.ini -Force; Remove-Item config.debug.ini.bak`
6. Repeat with `GUI_SEMVER` left at a valid value (e.g. `2.0.1`) and confirm the app starts normally with no messagebox

- [ ] **Step 4: Commit**

```bash
git add backend/utils/ini-reader.go
git commit -m "feat: close EMLy with error messagebox on invalid GUI_SEMVER"
```
