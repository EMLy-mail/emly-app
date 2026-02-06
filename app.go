// Package main provides the core EMLy application.
// EMLy is a desktop email viewer for .eml and .msg files built with Wails v2.
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

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
}

// =============================================================================
// Constructor & Lifecycle
// =============================================================================

// NewApp creates and initializes a new App instance.
// This is called from main.go before the Wails application starts.
func NewApp() *App {
	return &App{
		openImages: make(map[string]bool),
		openPDFs:   make(map[string]bool),
		openEMLs:   make(map[string]bool),
	}
}

// startup is called by Wails when the application starts.
// It receives the application context which is required for Wails runtime calls.
//
// This method:
//   - Saves the context for later use
//   - Syncs CurrentMailFilePath with StartupFilePath if a file was opened via CLI
//   - Logs the startup mode (main app vs viewer window)
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Sync CurrentMailFilePath with StartupFilePath if a file was opened via command line
	if a.StartupFilePath != "" {
		a.CurrentMailFilePath = a.StartupFilePath
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
		Log("Viewer instance started")
	} else {
		Log("EMLy main application started")

		// Automatic update check on startup (if enabled)
		go func() {
			// Wait 5 seconds after startup to avoid blocking the UI
			time.Sleep(5 * time.Second)

			config := a.GetConfig()
			if config == nil {
				log.Printf("Failed to load config for auto-update check")
				return
			}

			// Check if auto-update is enabled
			if config.EMLy.UpdateAutoCheck == "true" && config.EMLy.UpdateCheckEnabled == "true" {
				log.Println("Performing automatic update check...")
				status, err := a.CheckForUpdates()
				if err != nil {
					log.Printf("Auto-update check failed: %v", err)
					return
				}

				// Emit event if update is available
				if status.UpdateAvailable {
					log.Printf("Update available: %s -> %s", status.CurrentVersion, status.AvailableVersion)
					runtime.EventsEmit(ctx, "update:available", status)
				} else {
					log.Println("No updates available")
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

// =============================================================================
// Configuration Management
// =============================================================================

// GetConfig loads and returns the application configuration from config.ini.
// Returns nil if the configuration cannot be loaded.
func (a *App) GetConfig() *utils.Config {
	cfgPath := utils.DefaultConfigPath()
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		Log("Failed to load config:", err)
		return nil
	}
	return cfg
}

// SaveConfig persists the provided configuration to config.ini.
// Returns an error if saving fails.
func (a *App) SaveConfig(cfg *utils.Config) error {
	cfgPath := utils.DefaultConfigPath()
	if err := utils.SaveConfig(cfgPath, cfg); err != nil {
		Log("Failed to save config:", err)
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
