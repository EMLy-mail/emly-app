// Package main provides settings import/export functionality for EMLy.
// This file contains methods for exporting and importing application settings
// as JSON files.
package main

import (
	"fmt"
	"os"
	"strings"

	"emly/backend/utils"
	"emly/backend/utils/mail"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// Settings Export/Import Methods
// =============================================================================

// ExportSettings opens a save dialog and exports the provided settings JSON
// to the selected file location.
//
// The dialog is pre-configured with:
//   - Default filename: emly_settings.json
//   - Filter for JSON files
//
// Parameters:
//   - settingsJSON: The JSON string containing all application settings
//
// Returns:
//   - string: The path where settings were saved, or empty if cancelled
//   - error: Error if dialog or file operations fail
func (a *App) ExportSettings(settingsJSON string) (string, error) {
	// Open save dialog with JSON filter
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "emly_settings.json",
		Title:           "Export Settings",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}

	// User cancelled
	if savePath == "" {
		return "", nil
	}

	// Ensure .json extension
	if !strings.HasSuffix(strings.ToLower(savePath), ".json") {
		savePath += ".json"
	}

	// Write the settings file
	if err := os.WriteFile(savePath, []byte(settingsJSON), 0644); err != nil {
		return "", fmt.Errorf("failed to write settings file: %w", err)
	}

	return savePath, nil
}

// ImportSettings opens a file dialog for the user to select a settings JSON file
// and returns its contents.
//
// The dialog is configured to only show JSON files.
//
// Returns:
//   - string: The JSON content of the selected file, or empty if cancelled
//   - error: Error if dialog or file operations fail
func (a *App) ImportSettings() (string, error) {
	// Open file dialog with JSON filter
	openPath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Settings",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}

	// User cancelled
	if openPath == "" {
		return "", nil
	}

	// Read the settings file
	data, err := os.ReadFile(openPath)
	if err != nil {
		return "", fmt.Errorf("failed to read settings file: %w", err)
	}

	return string(data), nil
}

// ReloadConfig re-reads config.ini from disk and returns the current configuration.
// Useful to reflect any manual edits to config.ini without restarting the app.
//
// Returns:
//   - *utils.Config: The freshly loaded configuration
//   - error: Error if loading config fails
func (a *App) ReloadConfig() (*utils.Config, error) {
	cfgPath := utils.DefaultConfigPath()
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to reload config: %w", err)
	}
	return cfg, nil
}

// SetExportAttachmentFolder updates the EXPORT_ATTACHMENT_FOLDER setting in
// config.ini. An empty string resets to the default (Downloads folder).
//
// Parameters:
//   - folderPath: The folder where downloaded attachments should be saved
//
// Returns:
//   - error: Error if loading or saving config fails
func (a *App) SetExportAttachmentFolder(folderPath string) error {
	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("failed to load config")
	}
	config.EMLy.ExportAttachmentFolder = strings.TrimSpace(folderPath)
	if err := a.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// GetExportAttachmentFolder returns the EXPORT_ATTACHMENT_FOLDER setting from
// config.ini, or an empty string if not set (meaning the Downloads folder).
func (a *App) GetExportAttachmentFolder() string {
	config := a.GetConfig()
	if config == nil {
		return ""
	}
	return config.EMLy.ExportAttachmentFolder
}

// CheckFolderWritable verifies EMLy can actually create files in the given
// folder, by writing a probe file there and removing it again. Used before
// accepting a folder picked in Settings, and at startup to validate the
// folder already stored in config.ini.
//
// Parameters:
//   - folderPath: The folder to test; empty is always valid (Downloads default)
//
// Returns:
//   - error: Error if the folder cannot be written to
func (a *App) CheckFolderWritable(folderPath string) error {
	return internal.CheckFolderWritable(folderPath)
}

// validReleaseChannels are the values GUI_RELEASE_CHANNEL accepts, mirroring
// the ReleaseChannel union in the frontend's types.d.ts.
var validReleaseChannels = map[string]bool{
	"stable": true,
	"beta":   true,
	"next":   true,
}

// SetGUIReleaseChannel updates the GUI_RELEASE_CHANNEL setting in config.ini.
// The value is validated here rather than trusted from the frontend, since an
// unknown channel written to config.ini would survive restarts and be read
// back by everything that consumes the config (updater, tray, about screens).
//
// Parameters:
//   - channel: one of "stable", "beta" or "next"
//
// Returns:
//   - error: Error if the channel is unknown, or if loading or saving config fails
func (a *App) SetGUIReleaseChannel(channel string) error {
	normalized := strings.ToLower(strings.TrimSpace(channel))
	if !validReleaseChannels[normalized] {
		return fmt.Errorf("invalid release channel: %q", channel)
	}

	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("failed to load config")
	}
	config.EMLy.GUIReleaseChannel = normalized
	if err := a.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// SetTrayIconEnabled updates the DISABLE_TRAY_ICON setting in config.ini.
// The system tray icon is only created at startup (see main.go), so this
// takes effect after the next restart (see RestartApp).
//
// Parameters:
//   - enabled: whether the system tray icon should be shown on next startup
//
// Returns:
//   - error: Error if loading or saving config fails
func (a *App) SetTrayIconEnabled(enabled bool) error {
	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("failed to load config")
	}
	config.EMLy.DisableTrayIcon = !enabled
	if err := a.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// SetLogStartupTrace updates the LOG_STARTUP_TRACE setting in config.ini.
// The trace file is only opened at process startup (see main.go), so this
// takes effect after the next restart (see RestartApp) - same as the tray
// icon toggle above.
//
// Parameters:
//   - enabled: whether startup-trace.log should be written on next startup
//
// Returns:
//   - error: Error if loading or saving config fails
func (a *App) SetLogStartupTrace(enabled bool) error {
	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("failed to load config")
	}
	config.EMLy.LogStartupTrace = enabled
	if err := a.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// SetOldAttachmentPreload updates the OLD_ATTACHMENT_PRELOAD setting in
// config.ini - an escape hatch back to the pre-fix behaviour of sending
// every attachment's full bytes in the initial mail parse response (see
// oldAttachmentPreloadEnabled in app_mail.go), kept only for experiments
// and regression testing. Unlike the tray icon and startup trace toggles,
// this is read fresh on every ReadEML/ReadMSG/ReadPEC/ReadAuto call, so it
// takes effect immediately - no restart needed.
//
// Parameters:
//   - enabled: whether to revert to eager, full-byte attachment preloading
//
// Returns:
//   - error: Error if loading or saving config fails
func (a *App) SetOldAttachmentPreload(enabled bool) error {
	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("failed to load config")
	}
	config.EMLy.OldAttachmentPreload = enabled
	if err := a.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
