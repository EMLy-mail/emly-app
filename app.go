package main

import (
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
	ctx             context.Context
	StartupFilePath string
	openImagesMux   sync.Mutex
	openImages      map[string]bool
	openPDFsMux     sync.Mutex
	openPDFs        map[string]bool
	openEMLsMux     sync.Mutex
	openEMLs        map[string]bool
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
