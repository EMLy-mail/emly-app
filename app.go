package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"emly/backend/utils"
	internal "emly/backend/utils/mail"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                 context.Context
	StartupFilePath     string
	CurrentMailFilePath string // Tracks the currently loaded mail file (from startup or file dialog)
	openImagesMux       sync.Mutex
	openImages          map[string]bool
	openPDFsMux         sync.Mutex
	openPDFs            map[string]bool
	openEMLsMux         sync.Mutex
	openEMLs            map[string]bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		openImages: make(map[string]bool),
		openPDFs:   make(map[string]bool),
		openEMLs:   make(map[string]bool),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set CurrentMailFilePath to StartupFilePath if a file was opened via command line
	if a.StartupFilePath != "" {
		a.CurrentMailFilePath = a.StartupFilePath
	}

	isViewer := false
	for _, arg := range os.Args {
		if strings.Contains(arg, "--view-image") || strings.Contains(arg, "--view-pdf") {
			isViewer = true
			break
		}
	}

	if isViewer {
		Log("Second instance launch")
	} else {
		Log("Wails startup")
	}
}

func (a *App) GetConfig() *utils.Config {
	cfgPath := utils.DefaultConfigPath()
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		Log("Failed to load config for version:", err)
		return nil
	}
	return cfg
}

func (a *App) SaveConfig(cfg *utils.Config) error {
	cfgPath := utils.DefaultConfigPath()
	if err := utils.SaveConfig(cfgPath, cfg); err != nil {
		Log("Failed to save config:", err)
		return err
	}
	return nil
}

func (a *App) shutdown(ctx context.Context) {
	// Best-effort cleanup.
}

func (a *App) QuitApp() {
	runtime.Quit(a.ctx)
	// Generate exit code 138
	os.Exit(133) // 133 + 5 (SIGTRAP)
}

func (a *App) GetMachineData() *utils.MachineInfo {
	data, _ := utils.GetMachineInfo()
	return data
}

// GetStartupFile returns the file path if the app was opened with a file argument
func (a *App) GetStartupFile() string {
	return a.StartupFilePath
}

// SetCurrentMailFilePath sets the path of the currently loaded mail file
func (a *App) SetCurrentMailFilePath(filePath string) {
	a.CurrentMailFilePath = filePath
}

// GetCurrentMailFilePath returns the path of the currently loaded mail file
func (a *App) GetCurrentMailFilePath() string {
	return a.CurrentMailFilePath
}

// ReadEML reads a .eml file and returns the email data
func (a *App) ReadEML(filePath string) (*internal.EmailData, error) {
	return internal.ReadEmlFile(filePath)
}

// ReadPEC reads a PEC .eml file and returns the inner email data
func (a *App) ReadPEC(filePath string) (*internal.EmailData, error) {
	return internal.ReadPecInnerEml(filePath)
}

// ReadMSG reads a .msg file and returns the email data
func (a *App) ReadMSG(filePath string, useExternalConverter bool) (*internal.EmailData, error) {
	if useExternalConverter {
		return internal.ReadMsgFile(filePath)
	}
	return internal.ReadMsgFile(filePath)
}

// ReadMSGOSS reads a .msg file and returns the email data
func (a *App) ReadMSGOSS(filePath string) (*internal.EmailData, error) {
	return internal.ReadMsgFile(filePath)
}

// ShowOpenFileDialog shows the file open dialog for EML files
func (a *App) ShowOpenFileDialog() (string, error) {
	return internal.ShowFileDialog(a.ctx)
}

// OpenEMLWindow saves EML to temp and opens a new instance of the app
func (a *App) OpenEMLWindow(base64Data string, filename string) error {
	a.openEMLsMux.Lock()
	if a.openEMLs[filename] {
		a.openEMLsMux.Unlock()
		return fmt.Errorf("eml '%s' is already open", filename)
	}
	a.openEMLs[filename] = true
	a.openEMLsMux.Unlock()

	// 1. Decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// 2. Save to temp file
	tempDir := os.TempDir()
	// Use timestamp or unique ID to avoid conflicts if multiple files have same name
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s_%s", "emly_attachment", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 3. Launch new instance
	exe, err := os.Executable()
	if err != nil {
		a.openEMLsMux.Lock()
		delete(a.openEMLs, filename)
		a.openEMLsMux.Unlock()
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Run EMLy with the file path as argument
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

// OpenImageWindow opens a new window instance to display the image
func (a *App) OpenImageWindow(base64Data string, filename string) error {
	a.openImagesMux.Lock()
	if a.openImages[filename] {
		a.openImagesMux.Unlock()
		return fmt.Errorf("image '%s' is already open", filename)
	}
	a.openImages[filename] = true
	a.openImagesMux.Unlock()

	// 1. Decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// 2. Save to temp file
	tempDir := os.TempDir()
	// Use timestamp to make unique
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		a.openImagesMux.Lock()
		delete(a.openImages, filename)
		a.openImagesMux.Unlock()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 3. Launch new instance
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

// OpenPDFWindow opens a new window instance to display the PDF
func (a *App) OpenPDFWindow(base64Data string, filename string) error {
	a.openPDFsMux.Lock()
	if a.openPDFs[filename] {
		a.openPDFsMux.Unlock()
		return fmt.Errorf("pdf '%s' is already open", filename)
	}
	a.openPDFs[filename] = true
	a.openPDFsMux.Unlock()

	// 1. Decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// 2. Save to temp file
	tempDir := os.TempDir()
	// Use timestamp to make unique
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s", filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		a.openPDFsMux.Lock()
		delete(a.openPDFs, filename)
		a.openPDFsMux.Unlock()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 3. Launch new instance
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

// OpenPDF saves PDF to temp and opens with default app
func (a *App) OpenPDF(base64Data string, filename string) error {
	if base64Data == "" {
		return fmt.Errorf("no data provided")
	}
	// 1. Decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// 2. Save to temp file
	tempDir := os.TempDir()
	// Use timestamp to make unique
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 3. Open with default app (Windows)
	cmd := exec.Command("cmd", "/c", "start", "", tempFile)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}

// OpenImage saves image to temp and opens with default app (Windows)
func (a *App) OpenImage(base64Data string, filename string) error {
	if base64Data == "" {
		return fmt.Errorf("no data provided")
	}
	// 1. Decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// 2. Save to temp file
	tempDir := os.TempDir()
	// Use timestamp to make unique
	timestamp := time.Now().Format("20060102_150405")
	tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// 3. Open with default app (Windows)
	cmd := exec.Command("cmd", "/c", "start", "", tempFile)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}

type ImageViewerData struct {
	Data     string `json:"data"`
	Filename string `json:"filename"`
}

type PDFViewerData struct {
	Data     string `json:"data"`
	Filename string `json:"filename"`
}

type ViewerData struct {
	ImageData *ImageViewerData `json:"imageData,omitempty"`
	PDFData   *PDFViewerData   `json:"pdfData,omitempty"`
}

// GetImageViewerData checks CLI args and returns image data if in viewer mode
func (a *App) GetImageViewerData() (*ImageViewerData, error) {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--view-image=") {
			filePath := strings.TrimPrefix(arg, "--view-image=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read text file: %w", err)
			}
			// Return encoded base64 so frontend can handle it same way
			encoded := base64.StdEncoding.EncodeToString(data)
			return &ImageViewerData{
				Data:     encoded,
				Filename: filepath.Base(filePath),
			}, nil
		}
	}
	return nil, nil
}

// GetPDFViewerData checks CLI args and returns pdf data if in viewer mode
func (a *App) GetPDFViewerData() (*PDFViewerData, error) {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--view-pdf=") {
			filePath := strings.TrimPrefix(arg, "--view-pdf=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read text file: %w", err)
			}
			// Return encoded base64 so frontend can handle it same way
			encoded := base64.StdEncoding.EncodeToString(data)
			return &PDFViewerData{
				Data:     encoded,
				Filename: filepath.Base(filePath),
			}, nil
		}
	}
	return nil, nil
}

func (a *App) GetViewerData() (*ViewerData, error) {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--view-image=") {
			filePath := strings.TrimPrefix(arg, "--view-image=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read text file: %w", err)
			}
			// Return encoded base64 so frontend can handle it same way
			encoded := base64.StdEncoding.EncodeToString(data)
			return &ViewerData{
				ImageData: &ImageViewerData{
					Data:     encoded,
					Filename: filepath.Base(filePath),
				},
			}, nil
		}
		if strings.HasPrefix(arg, "--view-pdf=") {
			filePath := strings.TrimPrefix(arg, "--view-pdf=")
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read text file: %w", err)
			}
			// Return encoded base64 so frontend can handle it same way
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

// CheckIsDefaultEMLHandler verifies if the current executable is the default handler for .eml files
func (a *App) CheckIsDefaultEMLHandler() (bool, error) {
	// 1. Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return false, err
	}
	// Normalize path for comparison
	exePath = strings.ToLower(exePath)

	// 2. Open UserChoice key for .eml
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\.eml\UserChoice`, registry.QUERY_VALUE)
	if err != nil {
		// Key doesn't exist implies user hasn't made a specific choice or system default is active (not us usually)
		return false, nil
	}
	defer k.Close()

	// 3. Get ProgId
	progId, _, err := k.GetStringValue("ProgId")
	if err != nil {
		return false, err
	}

	// 4. Find the command for this ProgId
	classKeyPath := fmt.Sprintf(`%s\shell\open\command`, progId)
	classKey, err := registry.OpenKey(registry.CLASSES_ROOT, classKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("unable to find command for ProgId %s", progId)
	}
	defer classKey.Close()

	cmd, _, err := classKey.GetStringValue("")
	if err != nil {
		return false, err
	}

	// 5. Compare command with our executable
	cmdLower := strings.ToLower(cmd)

	// Basic check: does the command contain our executable name?
	// In a real scenario, parsing the exact path respecting quotes would be safer,
	// but checking if our specific exe path is present is usually sufficient.
	if strings.Contains(cmdLower, strings.ToLower(filepath.Base(exePath))) {
		// More robust: escape backslashes and check presence
		// cleanExe := strings.ReplaceAll(exePath, `\`, `\\`)
		// For now, depending on how registry stores it (short path vs long path),
		// containment of the filename is a strong indicator if the filename is unique enough (emly.exe)
		return true, nil
	}

	return false, nil
}

// OpenDefaultAppsSettings opens the Windows default apps settings page
func (a *App) OpenDefaultAppsSettings() error {
	cmd := exec.Command("cmd", "/c", "start", "ms-settings:defaultapps")
	return cmd.Start()
}

func (a *App) IsDebuggerRunning() bool {
	if a == nil {
		return false
	}
	return utils.IsDebugged()
}

func (a *App) ConvertToUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	// If invalid UTF-8, assume Windows-1252 (superset of ISO-8859-1)
	decoder := charmap.Windows1252.NewDecoder()
	decoded, _, err := transform.String(decoder, s)
	if err != nil {
		return s // Return as-is if decoding fails
	}
	return decoded
}

// ScreenshotResult contains the screenshot data and metadata
type ScreenshotResult struct {
	Data     string `json:"data"`     // Base64-encoded PNG data
	Width    int    `json:"width"`    // Image width in pixels
	Height   int    `json:"height"`   // Image height in pixels
	Filename string `json:"filename"` // Suggested filename
}

// TakeScreenshot captures the current Wails application window and returns it as base64 PNG
func (a *App) TakeScreenshot() (*ScreenshotResult, error) {
	// Get the window title to find our window
	windowTitle := "EMLy - EML Viewer for 3gIT"

	// Check if we're in viewer mode
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

	img, err := utils.CaptureWindowByTitle(windowTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to capture window: %w", err)
	}

	base64Data, err := utils.ScreenshotToBase64PNG(img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode screenshot: %w", err)
	}

	bounds := img.Bounds()
	timestamp := time.Now().Format("20060102_150405")

	return &ScreenshotResult{
		Data:     base64Data,
		Width:    bounds.Dx(),
		Height:   bounds.Dy(),
		Filename: fmt.Sprintf("emly_screenshot_%s.png", timestamp),
	}, nil
}

// SaveScreenshot captures and saves the screenshot to a file, returning the file path
func (a *App) SaveScreenshot() (string, error) {
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

// SaveScreenshotAs opens a save dialog and saves the screenshot to the selected location
func (a *App) SaveScreenshotAs() (string, error) {
	result, err := a.TakeScreenshot()
	if err != nil {
		return "", err
	}

	// Open save dialog
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

	if savePath == "" {
		return "", nil // User cancelled
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode screenshot data: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	return savePath, nil
}

// BugReportResult contains paths to the bug report files
type BugReportResult struct {
	FolderPath     string `json:"folderPath"`     // Path to the bug report folder
	ScreenshotPath string `json:"screenshotPath"` // Path to the screenshot file
	MailFilePath   string `json:"mailFilePath"`   // Path to the copied mail file (empty if no mail)
}

// CreateBugReportFolder creates a folder in temp with screenshot and optionally the current mail file
func (a *App) CreateBugReportFolder() (*BugReportResult, error) {
	// Create timestamp for unique folder name
	timestamp := time.Now().Format("20060102_150405")
	folderName := fmt.Sprintf("emly_bugreport_%s", timestamp)

	// Create folder in temp directory
	tempDir := os.TempDir()
	bugReportFolder := filepath.Join(tempDir, folderName)

	if err := os.MkdirAll(bugReportFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bug report folder: %w", err)
	}

	result := &BugReportResult{
		FolderPath: bugReportFolder,
	}

	// Take and save screenshot
	screenshotResult, err := a.TakeScreenshot()
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	screenshotData, err := base64.StdEncoding.DecodeString(screenshotResult.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	screenshotPath := filepath.Join(bugReportFolder, screenshotResult.Filename)
	if err := os.WriteFile(screenshotPath, screenshotData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save screenshot: %w", err)
	}
	result.ScreenshotPath = screenshotPath

	// Check if there's a mail file loaded and copy it
	if a.CurrentMailFilePath != "" {
		// Read the original mail file
		mailData, err := os.ReadFile(a.CurrentMailFilePath)
		if err != nil {
			// Log but don't fail - screenshot is still valid
			Log("Failed to read mail file for bug report:", err)
		} else {
			// Get the original filename
			mailFilename := filepath.Base(a.CurrentMailFilePath)
			mailFilePath := filepath.Join(bugReportFolder, mailFilename)

			if err := os.WriteFile(mailFilePath, mailData, 0644); err != nil {
				Log("Failed to copy mail file for bug report:", err)
			} else {
				result.MailFilePath = mailFilePath
			}
		}
	}

	return result, nil
}

// BugReportInput contains the user-provided bug report details
type BugReportInput struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Description    string `json:"description"`
	ScreenshotData string `json:"screenshotData"` // Base64-encoded PNG screenshot
}

// SubmitBugReportResult contains the result of submitting a bug report
type SubmitBugReportResult struct {
	ZipPath    string `json:"zipPath"`    // Path to the created zip file
	FolderPath string `json:"folderPath"` // Path to the bug report folder
}

// SubmitBugReport creates a complete bug report with user input, saves it, and zips the folder
func (a *App) SubmitBugReport(input BugReportInput) (*SubmitBugReportResult, error) {
	// Create timestamp for unique folder name
	timestamp := time.Now().Format("20060102_150405")
	folderName := fmt.Sprintf("emly_bugreport_%s", timestamp)

	// Create folder in temp directory
	tempDir := os.TempDir()
	bugReportFolder := filepath.Join(tempDir, folderName)

	if err := os.MkdirAll(bugReportFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bug report folder: %w", err)
	}

	// Save the pre-captured screenshot
	if input.ScreenshotData != "" {
		screenshotData, err := base64.StdEncoding.DecodeString(input.ScreenshotData)
		if err != nil {
			Log("Failed to decode screenshot:", err)
		} else {
			screenshotPath := filepath.Join(bugReportFolder, fmt.Sprintf("emly_screenshot_%s.png", timestamp))
			if err := os.WriteFile(screenshotPath, screenshotData, 0644); err != nil {
				Log("Failed to save screenshot:", err)
			}
		}
	}

	// Copy the mail file if one is loaded
	if a.CurrentMailFilePath != "" {
		mailData, err := os.ReadFile(a.CurrentMailFilePath)
		if err != nil {
			Log("Failed to read mail file for bug report:", err)
		} else {
			mailFilename := filepath.Base(a.CurrentMailFilePath)
			mailFilePath := filepath.Join(bugReportFolder, mailFilename)
			if err := os.WriteFile(mailFilePath, mailData, 0644); err != nil {
				Log("Failed to copy mail file for bug report:", err)
			}
		}
	}

	// Create the report.txt file with user's description
	reportContent := fmt.Sprintf(`EMLy Bug Report
================

Name: %s
Email: %s

Description:
%s

Generated: %s
`, input.Name, input.Email, input.Description, time.Now().Format("2006-01-02 15:04:05"))

	reportPath := filepath.Join(bugReportFolder, "report.txt")
	if err := os.WriteFile(reportPath, []byte(reportContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to save report file: %w", err)
	}

	// Get machine info and save it
	machineInfo, err := utils.GetMachineInfo()
	if err == nil && machineInfo != nil {
		sysInfoContent := fmt.Sprintf(`System Information
==================

Hostname: %s
OS: %s
Version: %s
Hardware ID: %s
External IP: %s
`, machineInfo.Hostname, machineInfo.OS, machineInfo.Version, machineInfo.HWID, machineInfo.ExternalIP)

		sysInfoPath := filepath.Join(bugReportFolder, "system_info.txt")
		if err := os.WriteFile(sysInfoPath, []byte(sysInfoContent), 0644); err != nil {
			Log("Failed to save system info:", err)
		}
	}

	// Zip the folder
	zipPath := bugReportFolder + ".zip"
	if err := zipFolder(bugReportFolder, zipPath); err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}

	return &SubmitBugReportResult{
		ZipPath:    zipPath,
		FolderPath: bugReportFolder,
	}, nil
}

// zipFolder creates a zip archive of the given folder
func zipFolder(sourceFolder, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the folder and add all files to the zip
	return filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root folder itself
		if path == sourceFolder {
			return nil
		}

		// Get relative path for the zip entry
		relPath, err := filepath.Rel(sourceFolder, path)
		if err != nil {
			return err
		}

		// Skip directories (they'll be created implicitly)
		if info.IsDir() {
			return nil
		}

		// Create the file in the zip
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Read the file content
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Write to zip
		_, err = writer.Write(fileContent)
		return err
	})
}

// OpenFolderInExplorer opens the specified folder in Windows Explorer
func (a *App) OpenFolderInExplorer(folderPath string) error {
	cmd := exec.Command("explorer", folderPath)
	return cmd.Start()
}
