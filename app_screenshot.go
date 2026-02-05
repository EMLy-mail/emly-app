// Package main provides screenshot functionality for EMLy.
// This file contains methods for capturing, saving, and exporting screenshots
// of the application window.
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"emly/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// Screenshot Types
// =============================================================================

// ScreenshotResult contains the captured screenshot data and metadata.
type ScreenshotResult struct {
	// Data is the base64-encoded PNG image data
	Data string `json:"data"`
	// Width is the image width in pixels
	Width int `json:"width"`
	// Height is the image height in pixels
	Height int `json:"height"`
	// Filename is the suggested filename for saving
	Filename string `json:"filename"`
}

// =============================================================================
// Screenshot Methods
// =============================================================================

// TakeScreenshot captures the current EMLy application window and returns it as base64 PNG.
// This uses Windows GDI API to capture the window contents, handling DWM composition
// for proper rendering of modern Windows applications.
//
// The method automatically detects whether the app is in main mode or viewer mode
// and captures the appropriate window.
//
// Returns:
//   - *ScreenshotResult: Contains base64 PNG data, dimensions, and suggested filename
//   - error: Error if window capture or encoding fails
func (a *App) TakeScreenshot() (*ScreenshotResult, error) {
	// Determine window title based on current mode
	windowTitle := "EMLy - EML Viewer for 3gIT"

	// Check if running in viewer mode
	for _, arg := range os.Args {
		if strings.Contains(arg, "--view-image") {
			windowTitle = "EMLy Image Viewer"
			break
		}
		if strings.Contains(arg, "--view-pdf") {
			windowTitle = "EMLy PDF Viewer"
			break
		}
	}

	// Capture the window using Windows GDI API
	img, err := utils.CaptureWindowByTitle(windowTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to capture window: %w", err)
	}

	// Encode to PNG and convert to base64
	base64Data, err := utils.ScreenshotToBase64PNG(img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode screenshot: %w", err)
	}

	// Build result with metadata
	bounds := img.Bounds()
	timestamp := time.Now().Format("20060102_150405")

	return &ScreenshotResult{
		Data:     base64Data,
		Width:    bounds.Dx(),
		Height:   bounds.Dy(),
		Filename: fmt.Sprintf("emly_screenshot_%s.png", timestamp),
	}, nil
}

// SaveScreenshot captures and saves the screenshot to the system temp directory.
// This is a convenience method that captures and saves in one step.
//
// Returns:
//   - string: The full path to the saved screenshot file
//   - error: Error if capture or save fails
func (a *App) SaveScreenshot() (string, error) {
	// Capture the screenshot
	result, err := a.TakeScreenshot()
	if err != nil {
		return "", err
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode screenshot data: %w", err)
	}

	// Save to temp directory
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, result.Filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	return filePath, nil
}

// SaveScreenshotAs captures a screenshot and opens a save dialog for the user
// to choose where to save it.
//
// Returns:
//   - string: The selected save path, or empty string if cancelled
//   - error: Error if capture, dialog, or save fails
func (a *App) SaveScreenshotAs() (string, error) {
	// Capture the screenshot first
	result, err := a.TakeScreenshot()
	if err != nil {
		return "", err
	}

	// Open save dialog with PNG filter
	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: result.Filename,
		Title:           "Save Screenshot",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "PNG Images (*.png)",
				Pattern:     "*.png",
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

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode screenshot data: %w", err)
	}

	// Save to selected location
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	return savePath, nil
}
