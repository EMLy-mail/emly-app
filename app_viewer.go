// Package main provides viewer window functionality for EMLy.
// This file contains methods for opening attachments in viewer windows
// or with external applications.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

// openInViewerWindow decodes base64Data, writes it to a uniquely-named temp
// file, and launches a new EMLy instance to display it, tracking the open
// file in tracker (guarded by mux) to prevent duplicate windows for the same
// filename. The tracking entry is released once the launched process exits.
//
// tempNamer builds the temp filename from a timestamp and the original
// filename; argsBuilder builds the CLI arguments passed to the relaunched
// EMLy executable from the temp file path.
func (a *App) openInViewerWindow(
	mux *sync.Mutex,
	tracker map[string]bool,
	kind string,
	base64Data string,
	filename string,
	tempNamer func(timestamp, filename string) string,
	argsBuilder func(tempFile string) []string,
) error {
	mux.Lock()
	if tracker[filename] {
		mux.Unlock()
		return fmt.Errorf("%s '%s' is already open", kind, filename)
	}
	tracker[filename] = true
	mux.Unlock()

	release := func() {
		mux.Lock()
		delete(tracker, filename)
		mux.Unlock()
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		release()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	tempDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, tempNamer(timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		release()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		release()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, argsBuilder(tempFile)...)
	if err := cmd.Start(); err != nil {
		release()
		return fmt.Errorf("failed to start viewer: %w", err)
	}

	// Monitor process in background to release the tracking entry when closed.
	go func() {
		cmd.Wait()
		release()
	}()

	return nil
}

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
	return a.openInViewerWindow(
		&a.openEMLsMux, a.openEMLs, "eml",
		base64Data, filename,
		func(timestamp, filename string) string {
			return fmt.Sprintf("%s_%s_%s", "emly_attachment", timestamp, filename)
		},
		func(tempFile string) []string { return []string{tempFile} },
	)
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
	return a.openInViewerWindow(
		&a.openImagesMux, a.openImages, "image",
		base64Data, filename,
		func(timestamp, filename string) string { return fmt.Sprintf("%s_%s", timestamp, filename) },
		func(tempFile string) []string { return []string{"--view-image=" + tempFile} },
	)
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
	return a.openInViewerWindow(
		&a.openPDFsMux, a.openPDFs, "pdf",
		base64Data, filename,
		func(timestamp, filename string) string { return fmt.Sprintf("%s_%s", timestamp, filename) },
		func(tempFile string) []string { return []string{"--view-pdf=" + tempFile} },
	)
}

// =============================================================================
// External Application Methods
// =============================================================================

// openWithDefaultApp saves data to a uniquely-named temp file and opens it
// with the system's default application for its file type (via "cmd /c
// start"). Used by OpenPDF, OpenImage and OpenDocument, which differ only in
// which file type they hand to the user - the launch mechanism is identical.
func openWithDefaultApp(base64Data string, filename string) error {
	if base64Data == "" {
		return fmt.Errorf("no data provided")
	}

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
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}

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
	return openWithDefaultApp(base64Data, filename)
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
	return openWithDefaultApp(base64Data, filename)
}

// OpenDocument saves a DOC/DOCX (or any Office document) to temp and opens it
// with the system's default application for that file type.
//
// Parameters:
//   - base64Data: Base64-encoded document data
//   - filename: The original filename of the document
//
// Returns:
//   - error: Error if saving or launching fails
func (a *App) OpenDocument(base64Data string, filename string) error {
	return openWithDefaultApp(base64Data, filename)
}

// =============================================================================
// Viewer Mode Detection
// =============================================================================

// imageFileExtensions lists the raster image extensions EMLy registers itself
// as a handler for, so double-clicking one of these in Explorer opens EMLy's
// built-in image viewer instead of (or alongside) the system default.
var imageFileExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".heic", ".heif"}

// isImageFilePath reports whether path has one of imageFileExtensions.
func isImageFilePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(imageFileExtensions, ext)
}

// findImageViewerPath scans os.Args for the file EMLy should display in image
// viewer mode: either the explicit --view-image= flag (used when EMLy
// relaunches itself for an email attachment) or a bare path with an image
// extension (used when Windows launches EMLy via file association after a
// double click on a .jpg/.png/etc file). Returns "" if neither is found.
func findImageViewerPath() string {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--view-image=") {
			return strings.TrimPrefix(arg, "--view-image=")
		}
		if isImageFilePath(arg) {
			return arg
		}
	}
	return ""
}

// GetImageViewerData checks CLI arguments and returns image data if running in image viewer mode.
// This is called by the viewer page on startup to get the image to display.
//
// Returns:
//   - *ImageViewerData: Image data if in viewer mode, nil otherwise
//   - error: Error if reading the image file fails
func (a *App) GetImageViewerData() (*ImageViewerData, error) {
	filePath := findImageViewerPath()
	if filePath == "" {
		return nil, nil
	}
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
	// Check for image viewer mode
	if filePath := findImageViewerPath(); filePath != "" {
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

	for _, arg := range os.Args {
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

// beforeClose is called by Wails before the main window closes.
// It prevents closing if any image or PDF viewer sub-processes are still open,
// showing an informational dialog to the user.
func (a *App) beforeClose(ctx context.Context) bool {
	a.openImagesMux.Lock()
	imagesOpen := len(a.openImages) > 0
	a.openImagesMux.Unlock()

	a.openPDFsMux.Lock()
	pdfsOpen := len(a.openPDFs) > 0
	a.openPDFsMux.Unlock()

	if imagesOpen || pdfsOpen {
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "EMLy",
			Message: "Ci sono ancora finestre di visualizzazione immagini o PDF aperte. Chiudile prima di uscire da EMLy.",
		})
		return true
	}
	return false
}
