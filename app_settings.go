// Package main provides settings import/export functionality for EMLy.
// This file contains methods for exporting and importing application settings
// as JSON files.
package main

import (
	"fmt"
	"os"
	"strings"

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

// SetUpdateCheckerEnabled updates the UPDATE_CHECK_ENABLED setting in config.ini
// based on the user's preference from the GUI settings.
//
// Parameters:
//   - enabled: true to enable update checking, false to disable
//
// Returns:
//   - error: Error if loading or saving config fails
func (a *App) SetUpdateCheckerEnabled(enabled bool) error {
	// Load current config
	config := a.GetConfig()
	if config == nil {
		return fmt.Errorf("failed to load config")
	}

	// Update the setting
	if enabled {
		config.EMLy.UpdateCheckEnabled = "true"
	} else {
		config.EMLy.UpdateCheckEnabled = "false"
	}

	// Save config back to disk
	if err := a.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
