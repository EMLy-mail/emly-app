// Package main provides the core EMLy application.
// EMLy is a desktop email viewer for .eml and .msg files built with Wails v2.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// App - Core Application Structure
// =============================================================================

// App is the main application struct that holds the application state and
// provides methods that are exposed to the frontend via Wails bindings.
//
// The struct manages:
//   - Application context for Wails runtime calls
//   - File paths for startup and currently loaded emails
//   - Tracking of open viewer windows to prevent duplicates
type App struct {
	// ctx is the Wails application context, used for runtime calls like dialogs
	ctx context.Context

	// StartupFilePath is set when the app is launched with an email file argument
	StartupFilePath string

	// CurrentMailFilePath tracks the currently loaded mail file path
	// Used for bug reports to include the relevant email file
	CurrentMailFilePath string

	// openImages tracks which images are currently open in viewer windows
	// The key is the filename, preventing duplicate viewers for the same file
	openImagesMux sync.Mutex
	openImages    map[string]bool

	// openPDFs tracks which PDFs are currently open in viewer windows
	openPDFsMux sync.Mutex
	openPDFs    map[string]bool

	// openEMLs tracks which EML attachments are currently open in viewer windows
	openEMLsMux sync.Mutex
	openEMLs    map[string]bool

	// httpClient is a shared HTTP client with custom User-Agent for all
	// outgoing requests (heartbeat, bug report upload, etc.)
	httpClient *http.Client
}

// =============================================================================
// Constructor & Lifecycle
// =============================================================================

// NewApp creates and initializes a new App instance.
// userAgent is injected into the shared HTTP client so every outgoing
// request identifies itself as "EMLy/<version>".
func NewApp(userAgent string) *App {
	return &App{
		openImages: make(map[string]bool),
		openPDFs:   make(map[string]bool),
		openEMLs:   make(map[string]bool),
		httpClient: &http.Client{
			Transport: &userAgentTransport{
				ua:   userAgent,
				base: http.DefaultTransport,
			},
		},
	}
}

// userAgentTransport is an http.RoundTripper that injects a custom
// User-Agent header into every outgoing HTTP request.
type userAgentTransport struct {
	ua   string
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.ua)
	return t.base.RoundTrip(req)
}

// startup is called by Wails when the application starts.
// It receives the application context which is required for Wails runtime calls.
//
// This method:
//   - Saves the context for later use
//   - Syncs CurrentMailFilePath with StartupFilePath if a file was opened via CLI
//   - Applies LOG_LEVEL from config.ini
//   - Logs the startup mode (main app vs viewer window)
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Sync CurrentMailFilePath with StartupFilePath if a file was opened via command line
	if a.StartupFilePath != "" {
		a.CurrentMailFilePath = a.StartupFilePath
	}

	// Apply log level from config (overrides env var if set)
	if cfg := a.GetConfig(); cfg != nil && cfg.EMLy.LogLevel != "" {
		pkglogger.SetLevelFromString(cfg.EMLy.LogLevel)
		pkglogger.Info("log level set from config", "level", cfg.EMLy.LogLevel)
	}

	// Check if this instance is running as a viewer (image/PDF) rather than main app
	isViewer := false
	for _, arg := range os.Args {
		if strings.Contains(arg, "--view-image") || strings.Contains(arg, "--view-pdf") {
			isViewer = true
			break
		}
	}

	if isViewer {
		pkglogger.Info("viewer instance started")
	} else {
		pkglogger.Info("EMLy main application started")

		// Automatic update check on startup (if enabled)
		go func() {
			// Wait 5 seconds after startup to avoid blocking the UI
			time.Sleep(5 * time.Second)

			config := a.GetConfig()
			if config == nil {
				pkglogger.Warn("failed to load config for auto-update check")
				return
			}

			// Check if auto-update is enabled
			if config.EMLy.UpdateAutoCheck == "true" && config.EMLy.UpdateCheckEnabled == "true" {
				pkglogger.Info("performing automatic update check")
				status, err := a.CheckForUpdates()
				if err != nil {
					pkglogger.Error("auto-update check failed", "error", err.Error())
					return
				}

				// Emit event if update is available
				if status.UpdateAvailable {
					pkglogger.Info("update available",
						"current", status.CurrentVersion,
						"available", status.AvailableVersion,
					)
					runtime.EventsEmit(ctx, "update:available", status)
				} else {
					pkglogger.Info("no updates available")
				}
			}
		}()
	}
}

// shutdown is called by Wails when the application is closing.
// Used for cleanup operations.
func (a *App) shutdown(ctx context.Context) {
	// Best-effort cleanup - currently no resources require explicit cleanup
}

// QuitApp terminates the application.
// It first calls Wails Quit to properly close the window,
// then forces an exit with a specific code.
func (a *App) QuitApp() {
	runtime.Quit(a.ctx)
	// Exit with code 133 (133 + 5 = 138, SIGTRAP-like exit)
	os.Exit(133)
}

// RestartApp performs a full application restart, including the Go backend.
// It schedules a new process via PowerShell with a short delay to ensure the
// single-instance lock is released before the new instance starts, then exits.
func (a *App) RestartApp() error {
	start := time.Now()
	var err error
	defer func() { canonicalLog("RestartApp", start, err) }()

	exe, err := os.Executable()
	if err != nil {
		pkglogger.Error("RestartApp: failed to get executable path", "error", err.Error())
		return err
	}

	// Escape single quotes in the path for PowerShell string literal
	safePath := strings.ReplaceAll(exe, "'", "''")
	script := fmt.Sprintf(`Start-Sleep -Seconds 1; Start-Process '%s'`, safePath)

	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if err = cmd.Start(); err != nil {
		pkglogger.Error("RestartApp: failed to schedule restart", "error", err.Error())
		return err
	}

	pkglogger.Info("RestartApp: scheduled restart, quitting current instance")
	runtime.Quit(a.ctx)
	return nil
}

// =============================================================================
// Configuration Management
// =============================================================================

// GetConfig loads and returns the application configuration from config.ini.
// Returns nil if the configuration cannot be loaded.
func (a *App) GetConfig() *utils.Config {
	cfgPath := utils.DefaultConfigPath()
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		pkglogger.Error("failed to load config", "error", err.Error())
		return nil
	}
	return cfg
}

func (a *App) ReloadEMLyConfig() (utils.EMLyConfig, error) {
	cfg, err := a.ReloadConfig()
	if cfg == nil {
		return utils.EMLyConfig{}, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg.EMLy, nil
}

// SaveConfig persists the provided configuration to config.ini.
// Returns an error if saving fails.
func (a *App) SaveConfig(cfg *utils.Config) error {
	cfgPath := utils.DefaultConfigPath()
	if err := utils.SaveConfig(cfgPath, cfg); err != nil {
		pkglogger.Error("failed to save config", "error", err.Error())
		return err
	}
	return nil
}

// =============================================================================
// Startup File Management
// =============================================================================

// GetStartupFile returns the file path if the app was launched with an email file argument.
// Returns an empty string if no file was specified at startup.
func (a *App) GetStartupFile() string {
	return a.StartupFilePath
}

// SetCurrentMailFilePath updates the path of the currently loaded mail file.
// This is called when the user opens a file via the file dialog.
func (a *App) SetCurrentMailFilePath(filePath string) {
	a.CurrentMailFilePath = filePath
}

// GetCurrentMailFilePath returns the path of the currently loaded mail file.
// Used by bug reports to include the relevant email file.
func (a *App) GetCurrentMailFilePath() string {
	return a.CurrentMailFilePath
}

// =============================================================================
// System Information
// =============================================================================

// GetMachineData retrieves system information about the current machine.
// Returns hostname, OS version, hardware ID, etc.
func (a *App) GetMachineData() *utils.MachineInfo {
	data, _ := utils.GetMachineInfo()
	return data
}

// IsDebuggerRunning checks if a debugger is attached to the application.
// Used for anti-debugging protection in production builds.
func (a *App) IsDebuggerRunning() bool {
	if a == nil {
		return false
	}
	return utils.IsDebugged()
}

func (a *App) IsAppInDebugMode() bool {
	if a == nil {
		return false
	}
	return utils.IsRunningInDebugMode()
}
