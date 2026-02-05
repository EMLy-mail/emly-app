// Package main provides viewer window functionality for EMLy.
// This file contains methods for opening attachments in viewer windows
// or with external applications.
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// =============================================================================
// Viewer Data Types
// =============================================================================

// ImageViewerData contains the data needed to display an image in the viewer window.
type ImageViewerData struct {
	// Data is the base64-encoded image data
	Data string `json:"data"`
	// Filename is the original filename of the image
	Filename string `json:"filename"`
}

// PDFViewerData contains the data needed to display a PDF in the viewer window.
type PDFViewerData struct {
	// Data is the base64-encoded PDF data
	Data string `json:"data"`
	// Filename is the original filename of the PDF
	Filename string `json:"filename"`
}

// ViewerData is a union type that contains either image or PDF viewer data.
// Used by the viewer page to determine which type of content to display.
type ViewerData struct {
	// ImageData is set when viewing an image (mutually exclusive with PDFData)
	ImageData *ImageViewerData `json:"imageData,omitempty"`
	// PDFData is set when viewing a PDF (mutually exclusive with ImageData)
	PDFData *PDFViewerData `json:"pdfData,omitempty"`
}

// =============================================================================
// Built-in Viewer Window Methods
// =============================================================================

// OpenEMLWindow opens an EML attachment in a new EMLy window.
// The EML data is saved to a temp file and a new EMLy instance is launched.
//
// This method tracks open EML files to prevent duplicate windows for the same file.
// The tracking is released when the viewer window is closed.
//
// Parameters:
//   - base64Data: Base64-encoded EML file content
//   - filename: The original filename of the EML attachment
//
// Returns:
//   - error: Error if the file is already open or if launching fails
func (a *App) OpenEMLWindow(base64Data string, filename string) error {
	// Check if this EML is already open
	a.openEMLsMux.Lock()
	if a.openEMLs[filename] {
		a.openEMLsMux.Unlock()
		return fmt.Errorf("eml '%s' is already open", filename)
	}
	a.openEMLs[filename] = true
	a.openEMLsMux.Unlock()

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Save to temp file with timestamp to avoid conflicts
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s_%s", "emly_attachment", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Launch new EMLy instance with the file path
	exe, err := os.Executable()
	if err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, tempFile)
	if err := cmd.Start(); err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to start viewer: %w", err)
	}

	// Monitor process in background to release lock when closed
	go func() {
		cmd.Wait()
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
	}()

	return nil
}

// OpenImageWindow opens an image attachment in a new EMLy viewer window.
// The image data is saved to a temp file and a new EMLy instance is launched
// with the --view-image flag.
//
// This method tracks open images to prevent duplicate windows for the same file.
//
// Parameters:
//   - base64Data: Base64-encoded image data
//   - filename: The original filename of the image
//
// Returns:
//   - error: Error if the image is already open or if launching fails
func (a *App) OpenImageWindow(base64Data string, filename string) error {
	// Check if this image is already open
	a.openImagesMux.Lock()
	if a.openImages[filename] {
		a.openImagesMux.Unlock()
		return fmt.Errorf("image '%s' is already open", filename)
	}
	a.openImages[filename] = true
	a.openImagesMux.Unlock()

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Save to temp file
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Launch new EMLy instance in image viewer mode
	exe, err := os.Executable()
	if err != nil {
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, "--view-image="+tempFile)
	if err := cmd.Start(); err != nil {
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
		return fmt.Errorf("failed to start viewer: %w", err)
	}

	// Monitor process in background to release lock when closed
	go func() {
		cmd.Wait()
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
	}()

	return nil
}

// OpenPDFWindow opens a PDF attachment in a new EMLy viewer window.
// The PDF data is saved to a temp file and a new EMLy instance is launched
// with the --view-pdf flag.
//
// This method tracks open PDFs to prevent duplicate windows for the same file.
//
// Parameters:
//   - base64Data: Base64-encoded PDF data
//   - filename: The original filename of the PDF
//
// Returns:
//   - error: Error if the PDF is already open or if launching fails
func (a *App) OpenPDFWindow(base64Data string, filename string) error {
	// Check if this PDF is already open
	a.openPDFsMux.Lock()
	if a.openPDFs[filename] {
		a.openPDFsMux.Unlock()
		return fmt.Errorf("pdf '%s' is already open", filename)
	}
	a.openPDFs[filename] = true
	a.openPDFsMux.Unlock()

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Save to temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Launch new EMLy instance in PDF viewer mode
	exe, err := os.Executable()
	if err != nil {
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, "--view-pdf="+tempFile)
	if err := cmd.Start(); err != nil {
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
		return fmt.Errorf("failed to start viewer: %w", err)
	}

	// Monitor process in background to release lock when closed
	go func() {
		cmd.Wait()
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
	}()

	return nil
}

// =============================================================================
// External Application Methods
// =============================================================================

// OpenPDF saves a PDF to temp and opens it with the system's default PDF application.
// This is used when the user prefers external viewers over the built-in viewer.
//
// Parameters:
//   - base64Data: Base64-encoded PDF data
//   - filename: The original filename of the PDF
//
// Returns:
//   - error: Error if saving or launching fails
func (a *App) OpenPDF(base64Data string, filename string) error {
	if base64Data == "" {
		return fmt.Errorf("no data provided")
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Save to temp file with timestamp for uniqueness
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Open with Windows default application
	cmd := exec.Command("cmd", "/c", "start", "", tempFile)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}

// OpenImage saves an image to temp and opens it with the system's default image viewer.
// This is used when the user prefers external viewers over the built-in viewer.
//
// Parameters:
//   - base64Data: Base64-encoded image data
//   - filename: The original filename of the image
//
// Returns:
//   - error: Error if saving or launching fails
func (a *App) OpenImage(base64Data string, filename string) error {
	if base64Data == "" {
		return fmt.Errorf("no data provided")
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Save to temp file with timestamp for uniqueness
	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Open with Windows default application
	cmd := exec.Command("cmd", "/c", "start", "", tempFile)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}

// =============================================================================
// Viewer Mode Detection
// =============================================================================

// GetImageViewerData checks CLI arguments and returns image data if running in image viewer mode.
// This is called by the viewer page on startup to get the image to display.
//
// Returns:
//   - *ImageViewerData: Image data if in viewer mode, nil otherwise
//   - error: Error if reading the image file fails
func (a *App) GetImageViewerData() (*ImageViewerData, error) {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--view-image=") {
			filePath := strings.TrimPrefix(arg, "--view-image=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read image file: %w", err)
			}
			// Return as base64 for consistent frontend handling
			encoded := base64.StdEncoding.EncodeToString(data)
			return &ImageViewerData{
				Data:     encoded,
				Filename: filepath.Base(filePath),
			}, nil
		}
	}
	return nil, nil
}

// GetPDFViewerData checks CLI arguments and returns PDF data if running in PDF viewer mode.
// This is called by the viewer page on startup to get the PDF to display.
//
// Returns:
//   - *PDFViewerData: PDF data if in viewer mode, nil otherwise
//   - error: Error if reading the PDF file fails
func (a *App) GetPDFViewerData() (*PDFViewerData, error) {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--view-pdf=") {
			filePath := strings.TrimPrefix(arg, "--view-pdf=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read PDF file: %w", err)
			}
			// Return as base64 for consistent frontend handling
			encoded := base64.StdEncoding.EncodeToString(data)
			return &PDFViewerData{
				Data:     encoded,
				Filename: filepath.Base(filePath),
			}, nil
		}
	}
	return nil, nil
}

// GetViewerData checks CLI arguments and returns viewer data for any viewer mode.
// This is a unified method that detects both image and PDF viewer modes.
//
// Returns:
//   - *ViewerData: Contains either ImageData or PDFData depending on mode
//   - error: Error if reading the file fails
func (a *App) GetViewerData() (*ViewerData, error) {
	for _, arg := range os.Args {
		// Check for image viewer mode
		if strings.HasPrefix(arg, "--view-image=") {
			filePath := strings.TrimPrefix(arg, "--view-image=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read image file: %w", err)
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			return &ViewerData{
				ImageData: &ImageViewerData{
					Data:     encoded,
					Filename: filepath.Base(filePath),
				},
			}, nil
		}

		// Check for PDF viewer mode
		if strings.HasPrefix(arg, "--view-pdf=") {
			filePath := strings.TrimPrefix(arg, "--view-pdf=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read PDF file: %w", err)
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			return &ViewerData{
				PDFData: &PDFViewerData{
					Data:     encoded,
					Filename: filepath.Base(filePath),
				},
			}, nil
		}
	}
	return nil, nil
}
