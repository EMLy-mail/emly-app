// Package main provides self-hosted update functionality for EMLy.
// This file contains methods for checking, downloading, and installing updates
// from a corporate network share without relying on third-party services.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// Update System Types
// =============================================================================

// UpdateManifest represents the version.json file structure on the network share
type UpdateManifest struct {
	StableVersion   string            `json:"stableVersion"`
	BetaVersion     string            `json:"betaVersion"`
	StableDownload  string            `json:"stableDownload"`
	BetaDownload    string            `json:"betaDownload"`
	SHA256Checksums map[string]string `json:"sha256Checksums"`
	ReleaseNotes    map[string]string `json:"releaseNotes,omitempty"`
}

// UpdateStatus represents the current state of the update system
type UpdateStatus struct {
	CurrentVersion   string `json:"currentVersion"`
	AvailableVersion string `json:"availableVersion"`
	UpdateAvailable  bool   `json:"updateAvailable"`
	Checking         bool   `json:"checking"`
	Downloading      bool   `json:"downloading"`
	DownloadProgress int    `json:"downloadProgress"`
	Ready            bool   `json:"ready"`
	InstallerPath    string `json:"installerPath"`
	ErrorMessage     string `json:"errorMessage"`
	ReleaseNotes     string `json:"releaseNotes,omitempty"`
	LastCheckTime    string `json:"lastCheckTime"`
}

// Global update state
var updateStatus = UpdateStatus{
	CurrentVersion:   "",
	AvailableVersion: "",
	UpdateAvailable:  false,
	Checking:         false,
	Downloading:      false,
	DownloadProgress: 0,
	Ready:            false,
	InstallerPath:    "",
	ErrorMessage:     "",
}

// =============================================================================
// Update Check Methods
// =============================================================================

// CheckForUpdates checks the configured network share for available updates.
// Compares the manifest version with the current GUI version based on release channel.
//
// Returns:
//   - UpdateStatus: Current update state including available version
//   - error: Error if check fails (network, parsing, etc.)
func (a *App) CheckForUpdates() (UpdateStatus, error) {
	// Reset status
	updateStatus.Checking = true
	updateStatus.ErrorMessage = ""
	updateStatus.LastCheckTime = time.Now().Format("2006-01-02 15:04:05")
	runtime.EventsEmit(a.ctx, "update:status", updateStatus)

	defer func() {
		updateStatus.Checking = false
		runtime.EventsEmit(a.ctx, "update:status", updateStatus)
	}()

	// Get current version from config
	config := a.GetConfig()
	if config == nil {
		updateStatus.ErrorMessage = "Failed to load configuration"
		return updateStatus, fmt.Errorf("failed to load config")
	}

	updateStatus.CurrentVersion = config.EMLy.GUISemver
	currentChannel := config.EMLy.GUIReleaseChannel

	// Check if updates are enabled
	if config.EMLy.UpdateCheckEnabled != "true" {
		updateStatus.ErrorMessage = "Update checking is disabled"
		return updateStatus, fmt.Errorf("update checking is disabled in config")
	}

	// Validate update path
	updatePath := strings.TrimSpace(config.EMLy.UpdatePath)
	if updatePath == "" {
		updateStatus.ErrorMessage = "Update path not configured"
		return updateStatus, fmt.Errorf("UPDATE_PATH is empty in config.ini")
	}

	// Load manifest from network share
	manifest, err := a.loadUpdateManifest(updatePath)
	if err != nil {
		updateStatus.ErrorMessage = fmt.Sprintf("Failed to load manifest: %v", err)
		return updateStatus, err
	}

	// Determine target version based on release channel
	var targetVersion string
	if currentChannel == "beta" {
		targetVersion = manifest.BetaVersion
	} else {
		targetVersion = manifest.StableVersion
	}

	updateStatus.AvailableVersion = targetVersion

	// Compare versions
	comparison := compareSemanticVersions(updateStatus.CurrentVersion, targetVersion)
	if comparison < 0 {
		// New version available
		updateStatus.UpdateAvailable = true
		updateStatus.InstallerPath = "" // Reset installer path
		updateStatus.Ready = false

		// Get release notes if available
		if notes, ok := manifest.ReleaseNotes[targetVersion]; ok {
			updateStatus.ReleaseNotes = notes
		}

		log.Printf("Update available: %s -> %s (%s channel)",
			updateStatus.CurrentVersion, targetVersion, currentChannel)
	} else {
		updateStatus.UpdateAvailable = false
		updateStatus.InstallerPath = ""
		updateStatus.Ready = false
		updateStatus.ReleaseNotes = ""
		log.Printf("Already on latest version: %s (%s channel)",
			updateStatus.CurrentVersion, currentChannel)
	}

	return updateStatus, nil
}

// loadUpdateManifest reads and parses version.json from the network share
func (a *App) loadUpdateManifest(updatePath string) (*UpdateManifest, error) {
	// Resolve path (handle UNC paths, file:// URLs, local paths)
	manifestPath, err := resolveUpdatePath(updatePath, "version.json")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve manifest path: %w", err)
	}

	log.Printf("Loading update manifest from: %s", manifestPath)

	// Read manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse JSON
	var manifest UpdateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	// Validate manifest
	if manifest.StableVersion == "" || manifest.StableDownload == "" {
		return nil, fmt.Errorf("invalid manifest: missing stable version or download")
	}

	return &manifest, nil
}

// =============================================================================
// Download Methods
// =============================================================================

// DownloadUpdate downloads the installer from the network share to a temporary location.
// Verifies SHA256 checksum if provided in the manifest.
//
// Returns:
//   - string: Path to the downloaded installer
//   - error: Error if download or verification fails
func (a *App) DownloadUpdate() (string, error) {
	if !updateStatus.UpdateAvailable {
		return "", fmt.Errorf("no update available")
	}

	updateStatus.Downloading = true
	updateStatus.DownloadProgress = 0
	updateStatus.ErrorMessage = ""
	runtime.EventsEmit(a.ctx, "update:status", updateStatus)

	defer func() {
		updateStatus.Downloading = false
		runtime.EventsEmit(a.ctx, "update:status", updateStatus)
	}()

	// Get config
	config := a.GetConfig()
	if config == nil {
		updateStatus.ErrorMessage = "Failed to load configuration"
		return "", fmt.Errorf("failed to load config")
	}

	updatePath := strings.TrimSpace(config.EMLy.UpdatePath)
	currentChannel := config.EMLy.GUIReleaseChannel

	// Reload manifest to get download filename
	manifest, err := a.loadUpdateManifest(updatePath)
	if err != nil {
		updateStatus.ErrorMessage = "Failed to load manifest"
		return "", err
	}

	// Determine download filename
	var downloadFilename string
	if currentChannel == "beta" {
		downloadFilename = manifest.BetaDownload
	} else {
		downloadFilename = manifest.StableDownload
	}

	// Resolve source path
	sourcePath, err := resolveUpdatePath(updatePath, downloadFilename)
	if err != nil {
		updateStatus.ErrorMessage = "Failed to resolve installer path"
		return "", fmt.Errorf("failed to resolve installer path: %w", err)
	}

	log.Printf("Downloading installer from: %s", sourcePath)

	// Create temp directory for download
	tempDir := filepath.Join(os.TempDir(), "emly_update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		updateStatus.ErrorMessage = "Failed to create temp directory"
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Destination path
	destPath := filepath.Join(tempDir, downloadFilename)

	// Copy file with progress
	if err := a.copyFileWithProgress(sourcePath, destPath); err != nil {
		updateStatus.ErrorMessage = "Download failed"
		return "", fmt.Errorf("failed to copy installer: %w", err)
	}

	// Verify checksum if available
	if checksum, ok := manifest.SHA256Checksums[downloadFilename]; ok {
		log.Printf("Verifying checksum for %s", downloadFilename)
		if err := verifyChecksum(destPath, checksum); err != nil {
			updateStatus.ErrorMessage = "Checksum verification failed"
			// Delete corrupted file
			os.Remove(destPath)
			return "", fmt.Errorf("checksum verification failed: %w", err)
		}
		log.Printf("Checksum verified successfully")
	} else {
		log.Printf("Warning: No checksum available for %s", downloadFilename)
	}

	updateStatus.InstallerPath = destPath
	updateStatus.Ready = true
	updateStatus.DownloadProgress = 100
	log.Printf("Update downloaded successfully to: %s", destPath)

	return destPath, nil
}

// copyFileWithProgress copies a file and emits progress events
func (a *App) copyFileWithProgress(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get file size
	stat, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	totalSize := stat.Size()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy with progress tracking
	buffer := make([]byte, 1024*1024) // 1MB buffer
	var copiedSize int64 = 0

	for {
		n, err := sourceFile.Read(buffer)
		if n > 0 {
			if _, writeErr := destFile.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			copiedSize += int64(n)

			// Update progress (avoid too many events)
			progress := int((copiedSize * 100) / totalSize)
			if progress != updateStatus.DownloadProgress {
				updateStatus.DownloadProgress = progress
				runtime.EventsEmit(a.ctx, "update:status", updateStatus)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================
// Install Methods
// =============================================================================

// InstallUpdate launches the downloaded installer with elevated privileges
// and optionally quits the application.
//
// Parameters:
//   - quitAfterLaunch: If true, exits EMLy after launching the installer
//
// Returns:
//   - error: Error if installer launch fails
func (a *App) InstallUpdate(quitAfterLaunch bool) error {
	if !updateStatus.Ready || updateStatus.InstallerPath == "" {
		return fmt.Errorf("no installer ready to install")
	}

	installerPath := updateStatus.InstallerPath

	// Verify installer exists
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		updateStatus.ErrorMessage = "Installer file not found"
		updateStatus.Ready = false
		return fmt.Errorf("installer not found: %s", installerPath)
	}

	log.Printf("Launching installer: %s", installerPath)

	// Launch installer with UAC elevation using ShellExecute
	if err := shellExecuteAsAdmin(installerPath); err != nil {
		updateStatus.ErrorMessage = fmt.Sprintf("Failed to launch installer: %v", err)
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	log.Printf("Installer launched successfully")

	// Quit application if requested
	if quitAfterLaunch {
		time.Sleep(500 * time.Millisecond) // Brief delay to ensure installer starts
		runtime.Quit(a.ctx)
	}

	return nil
}

// shellExecuteAsAdmin launches an executable with UAC elevation on Windows
func shellExecuteAsAdmin(exePath string) error {
	verb := "runas" // Triggers UAC elevation
	exe := syscall.StringToUTF16Ptr(exePath)
	verbPtr := syscall.StringToUTF16Ptr(verb)

	var hwnd uintptr = 0
	var operation = verbPtr
	var file = exe
	var parameters uintptr = 0
	var directory uintptr = 0
	var showCmd int32 = 1 // SW_SHOWNORMAL

	// Load shell32.dll
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	ret, _, err := shellExecute.Call(
		hwnd,
		uintptr(unsafe.Pointer(operation)),
		uintptr(unsafe.Pointer(file)),
		parameters,
		directory,
		uintptr(showCmd),
	)

	// ShellExecuteW returns a value > 32 on success
	if ret <= 32 {
		return fmt.Errorf("ShellExecuteW failed with code %d: %v", ret, err)
	}

	return nil
}

// =============================================================================
// Status Methods
// =============================================================================

// GetUpdateStatus returns the current update system status.
// This can be polled by the frontend to update UI state.
//
// Returns:
//   - UpdateStatus: Current state of the update system
func (a *App) GetUpdateStatus() UpdateStatus {
	return updateStatus
}

// =============================================================================
// Utility Functions
// =============================================================================

// resolveUpdatePath resolves a network share path or file:// URL to a local path.
// Handles UNC paths (\\server\share), file:// URLs, and local paths.
func resolveUpdatePath(basePath, filename string) (string, error) {
	basePath = strings.TrimSpace(basePath)

	// Handle file:// URL
	if strings.HasPrefix(strings.ToLower(basePath), "file://") {
		u, err := url.Parse(basePath)
		if err != nil {
			return "", fmt.Errorf("invalid file URL: %w", err)
		}
		// Convert file URL to local path
		basePath = filepath.FromSlash(u.Path)
		// Handle Windows drive letters (file:///C:/path -> C:/path)
		if len(basePath) > 0 && basePath[0] == '/' && len(basePath) > 2 && basePath[2] == ':' {
			basePath = basePath[1:]
		}
	}

	// Join with filename
	fullPath := filepath.Join(basePath, filename)

	// Verify path is accessible
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("path not accessible: %w", err)
	}

	return fullPath, nil
}

// compareSemanticVersions compares two semantic version strings.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareSemanticVersions(v1, v2 string) int {
	// Strip beta/alpha suffixes for comparison
	v1Clean := strings.Split(v1, "-")[0]
	v2Clean := strings.Split(v2, "-")[0]

	parts1 := strings.Split(v1Clean, ".")
	parts2 := strings.Split(v2Clean, ".")

	// Compare each version component
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int

		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	// If base versions are equal, check beta/stable
	if v1 != v2 {
		// Version with beta suffix is considered "older" than without
		if strings.Contains(v1, "-beta") && !strings.Contains(v2, "-beta") {
			return -1
		}
		if !strings.Contains(v1, "-beta") && strings.Contains(v2, "-beta") {
			return 1
		}
	}

	return 0
}

// verifyChecksum verifies the SHA256 checksum of a file
func verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}
