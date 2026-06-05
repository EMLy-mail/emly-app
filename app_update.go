// Package main provides self-hosted update functionality for EMLy.
// This file contains methods for checking, downloading, and installing updates
// from a corporate network share without relying on third-party services.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	pkglogger "emly/backend/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// Update System Types
// =============================================================================

// DetailedNote holds per-version structured release note data from the API manifest
type DetailedNote struct {
	SeverityType string            `json:"severityType"`
	Description  map[string]string `json:"description"`
}

// UpdateManifest represents the version.json / API manifest structure
type UpdateManifest struct {
	StableVersion        string                  `json:"stableVersion"`
	BetaVersion          string                  `json:"betaVersion"`
	StableDownload       string                  `json:"stableDownload"`
	BetaDownload         string                  `json:"betaDownload"`
	IsCritical           bool                    `json:"isCritical"`
	SHA256Checksums      map[string]string       `json:"sha256Checksums"`
	ReleaseNotes         map[string]string       `json:"releaseNotes,omitempty"`
	DetailedReleaseNotes map[string]DetailedNote `json:"detailedReleaseNotes,omitempty"`
}

// UpdateStatus represents the current state of the update system
type UpdateStatus struct {
	CurrentVersion   string `json:"currentVersion"`
	AvailableVersion string `json:"availableVersion"`
	UpdateAvailable  bool   `json:"updateAvailable"`
	IsCritical       bool   `json:"isCritical"`
	Checking         bool   `json:"checking"`
	Downloading      bool   `json:"downloading"`
	DownloadProgress int    `json:"downloadProgress"`
	Ready            bool   `json:"ready"`
	InstallerPath    string `json:"installerPath"`
	ErrorMessage     string `json:"errorMessage"`
	ReleaseNotes     string `json:"releaseNotes,omitempty"`
	SeverityType     string `json:"severityType,omitempty"`
	LastCheckTime    string `json:"lastCheckTime"`
	Channel          string `json:"channel,omitempty"`
}

// Global update state
var updateStatus = UpdateStatus{
	CurrentVersion:   "",
	AvailableVersion: "",
	UpdateAvailable:  false,
	IsCritical:       false,
	Checking:         false,
	Downloading:      false,
	DownloadProgress: 0,
	Ready:            false,
	InstallerPath:    "",
	ErrorMessage:     "",
}

// cachedManifest holds the last successfully loaded manifest so DownloadUpdate
// can reuse it without re-fetching. cachedManifestIsAPI is true when the manifest
// came from the HTTP API (download URLs are full URLs, checksums keyed by version).
var cachedManifest *UpdateManifest
var cachedManifestIsAPI bool

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

	// Get current version from config
	config := a.GetConfig()
	if config == nil {
		updateStatus.ErrorMessage = "Failed to load configuration"
		updateStatus.Checking = false
		return updateStatus, fmt.Errorf("failed to load config")
	}

	updateStatus.CurrentVersion = config.EMLy.GUISemver
	currentChannel := config.EMLy.GUIReleaseChannel

	// Check if updates are enabled
	if config.EMLy.UpdateCheckEnabled != "true" {
		updateStatus.ErrorMessage = "Update checking is disabled"
		updateStatus.Checking = false
		return updateStatus, fmt.Errorf("update checking is disabled in config")
	}

	// Determine update source: "api" uses the HTTP API, anything else uses the network share.
	// An empty UPDATE_SOURCE defaults to "unc" for backwards compatibility.
	updateSource := strings.ToLower(strings.TrimSpace(config.EMLy.UpdateSource))
	if updateSource == "" {
		updateSource = "unc"
	}

	var manifest *UpdateManifest
	var manifestErr error
	isAPI := false

	switch updateSource {
	case "api":
		apiBaseURL := strings.TrimSpace(config.EMLy.BugReportAPIURL)
		if apiBaseURL == "" {
			updateStatus.ErrorMessage = "BUGREPORT_API_URL not configured"
			updateStatus.Checking = false
			return updateStatus, fmt.Errorf("UPDATE_SOURCE is 'api' but BUGREPORT_API_URL is empty")
		}
		manifest, manifestErr = a.loadUpdateManifestFromAPI(apiBaseURL)
		if manifestErr != nil {
			updateStatus.ErrorMessage = fmt.Sprintf("API manifest failed: %v", manifestErr)
			updateStatus.Checking = false
			return updateStatus, manifestErr
		}
		isAPI = true
	default: // "unc"
		updatePath := strings.TrimSpace(config.EMLy.UpdatePath)
		if updatePath == "" {
			updateStatus.ErrorMessage = "UPDATE_PATH not configured"
			updateStatus.Checking = false
			return updateStatus, fmt.Errorf("UPDATE_PATH is empty in config.ini")
		}
		manifest, manifestErr = a.loadUpdateManifest(updatePath)
		pkglogger.Debug("loadUpdateManifest result", "error", fmt.Sprintf("%v", manifestErr))
		if manifestErr != nil {
			updateStatus.ErrorMessage = fmt.Sprintf("Failed to load manifest: %v", manifestErr)
			updateStatus.Checking = false
			return updateStatus, manifestErr
		}
	}

	cachedManifest = manifest
	cachedManifestIsAPI = isAPI

	// Determine target version based on release channel
	var targetVersion string
	if currentChannel == "beta" {
		targetVersion = manifest.BetaVersion
	} else {
		targetVersion = manifest.StableVersion
	}

	updateStatus.Channel = currentChannel
	updateStatus.AvailableVersion = targetVersion

	// Compare versions
	comparison := compareSemanticVersions(updateStatus.CurrentVersion, targetVersion)
	if comparison < 0 {
		// New version available
		updateStatus.UpdateAvailable = true
		updateStatus.IsCritical = manifest.IsCritical
		updateStatus.InstallerPath = "" // Reset installer path
		updateStatus.Ready = false

		// Get release notes: prefer detailed (current language), then simple string
		updateStatus.ReleaseNotes = resolveReleaseNotes(manifest, targetVersion, config.EMLy.Language)

		// Detailed release notes: severity type + localized description
		if detailed, ok := manifest.DetailedReleaseNotes[targetVersion]; ok {
			updateStatus.SeverityType = detailed.SeverityType
			lang := "en"
			if config.EMLy.Language != "" {
				lang = config.EMLy.Language
			}
			if desc, ok := detailed.Description[lang]; ok {
				updateStatus.ReleaseNotes = desc
			} else if desc, ok := detailed.Description["en"]; ok {
				updateStatus.ReleaseNotes = desc
			}
		} else {
			updateStatus.SeverityType = ""
		}

		pkglogger.Info("update available",
			"current", updateStatus.CurrentVersion,
			"available", targetVersion,
			"channel", currentChannel,
			"critical", manifest.IsCritical,
		)
	} else {
		updateStatus.UpdateAvailable = false
		updateStatus.IsCritical = false
		updateStatus.InstallerPath = ""
		updateStatus.Ready = false
		updateStatus.ReleaseNotes = ""
		updateStatus.SeverityType = ""
		pkglogger.Info("already on latest version",
			"current", updateStatus.CurrentVersion,
			"channel", currentChannel,
		)
	}

	updateStatus.Checking = false
	return updateStatus, nil
}

// loadUpdateManifest reads and parses version.json from the network share
func (a *App) loadUpdateManifest(updatePath string) (*UpdateManifest, error) {
	// Resolve path (handle UNC paths, file:// URLs, local paths)
	manifestPath, err := resolveUpdatePath(updatePath, "version.json")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve manifest path: %w", err)
	}

	pkglogger.Info("loading update manifest", "path", manifestPath)

	// Read manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Strip UTF-8 BOM if present (files saved by Windows editors may include it)
	data = bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF"))

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

// loadUpdateManifestFromAPI fetches the update manifest from the API endpoint.
// The path portion of apiBaseURL is stripped so the request always targets
// {scheme}://{host}/v2/updates/manifest regardless of the URL stored in config.
func (a *App) loadUpdateManifestFromAPI(apiBaseURL string) (*UpdateManifest, error) {
	u, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid API base URL: %w", err)
	}
	manifestURL := fmt.Sprintf("%s://%s/v2/updates/manifest", u.Scheme, u.Host)
	pkglogger.Info("fetching update manifest from API", "url", manifestURL)

	resp, err := a.httpClient.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest API returned HTTP %d", resp.StatusCode)
	}

	var manifest UpdateManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	if manifest.StableVersion == "" || manifest.StableDownload == "" {
		return nil, fmt.Errorf("invalid manifest: missing stable version or download")
	}

	return &manifest, nil
}

// resolveReleaseNotes picks the best available release note string for a given version.
// Priority: detailed notes in the configured language → detailed notes in English →
// simple release note string.
func resolveReleaseNotes(manifest *UpdateManifest, version, lang string) string {
	if detailed, ok := manifest.DetailedReleaseNotes[version]; ok {
		if lang != "" {
			if desc, ok := detailed.Description[lang]; ok && desc != "" {
				return desc
			}
		}
		if desc, ok := detailed.Description["en"]; ok && desc != "" {
			return desc
		}
	}
	if notes, ok := manifest.ReleaseNotes[version]; ok {
		return notes
	}
	return ""
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

	currentChannel := config.EMLy.GUIReleaseChannel

	// Use cached manifest if available; otherwise reload from the appropriate source
	manifest := cachedManifest
	isAPI := cachedManifestIsAPI
	if manifest == nil {
		var reloadErr error
		updatePath := strings.TrimSpace(config.EMLy.UpdatePath)
		apiBaseURL := strings.TrimSpace(config.EMLy.BugReportAPIURL)
		if apiBaseURL != "" {
			manifest, reloadErr = a.loadUpdateManifestFromAPI(apiBaseURL)
			if reloadErr == nil {
				isAPI = true
			}
		}
		if manifest == nil && updatePath != "" {
			manifest, reloadErr = a.loadUpdateManifest(updatePath)
		}
		if manifest == nil {
			updateStatus.ErrorMessage = "Failed to load manifest"
			return "", fmt.Errorf("failed to load manifest: %w", reloadErr)
		}
	}

	// Determine download URL/filename
	var downloadRef string
	if currentChannel == "beta" {
		downloadRef = manifest.BetaDownload
	} else {
		downloadRef = manifest.StableDownload
	}

	// Create temp directory for download
	tempDir := filepath.Join(os.TempDir(), "emly_update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		updateStatus.ErrorMessage = "Failed to create temp directory"
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	var destPath string

	isHTTP := strings.HasPrefix(downloadRef, "http://") || strings.HasPrefix(downloadRef, "https://")
	if isAPI || isHTTP {
		// HTTP download from API
		pkglogger.Info("downloading installer via HTTP", "url", downloadRef)
		parsedURL, parseErr := url.Parse(downloadRef)
		if parseErr != nil {
			updateStatus.ErrorMessage = "Invalid download URL"
			return "", fmt.Errorf("invalid download URL: %w", parseErr)
		}
		downloadFilename := filepath.Base(parsedURL.Path)
		if downloadFilename == "" || downloadFilename == "." {
			downloadFilename = fmt.Sprintf("EMLy_Installer_%s.exe", updateStatus.AvailableVersion)
		}
		destPath = filepath.Join(tempDir, downloadFilename)
		if err := a.downloadFileFromHTTP(downloadRef, destPath); err != nil {
			updateStatus.ErrorMessage = "Download failed"
			return "", fmt.Errorf("failed to download installer: %w", err)
		}
		// Verify checksum: API manifest keys checksums by version string
		checksumKey := updateStatus.AvailableVersion
		if checksum, ok := manifest.SHA256Checksums[checksumKey]; ok {
			pkglogger.Info("verifying checksum", "version", checksumKey)
			if err := verifyChecksum(destPath, checksum); err != nil {
				updateStatus.ErrorMessage = "Checksum verification failed"
				os.Remove(destPath)
				return "", fmt.Errorf("checksum verification failed: %w", err)
			}
			pkglogger.Info("checksum verified")
		} else {
			pkglogger.Warn("no checksum available", "version", checksumKey)
		}
	} else {
		// Network share copy
		updatePath := strings.TrimSpace(config.EMLy.UpdatePath)
		downloadFilename := downloadRef

		sourcePath, resolveErr := resolveUpdatePath(updatePath, downloadFilename)
		if resolveErr != nil {
			updateStatus.ErrorMessage = "Failed to resolve installer path"
			return "", fmt.Errorf("failed to resolve installer path: %w", resolveErr)
		}

		pkglogger.Info("downloading installer from share", "source", sourcePath)
		destPath = filepath.Join(tempDir, downloadFilename)

		if err := a.copyFileWithProgress(sourcePath, destPath); err != nil {
			updateStatus.ErrorMessage = "Download failed"
			return "", fmt.Errorf("failed to copy installer: %w", err)
		}

		// Network share manifest keys checksums by filename
		if checksum, ok := manifest.SHA256Checksums[downloadFilename]; ok {
			pkglogger.Info("verifying checksum", "file", downloadFilename)
			if err := verifyChecksum(destPath, checksum); err != nil {
				updateStatus.ErrorMessage = "Checksum verification failed"
				os.Remove(destPath)
				return "", fmt.Errorf("checksum verification failed: %w", err)
			}
			pkglogger.Info("checksum verified")
		} else {
			pkglogger.Warn("no checksum available", "file", downloadFilename)
		}
	}

	updateStatus.InstallerPath = destPath
	updateStatus.Ready = true
	updateStatus.DownloadProgress = 100
	pkglogger.Info("update downloaded", "path", destPath)

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

// downloadFileFromHTTP downloads a file via HTTP/HTTPS and emits progress events.
func (a *App) downloadFileFromHTTP(downloadURL, destPath string) error {
	resp, err := a.httpClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	buffer := make([]byte, 1024*1024) // 1MB buffer
	var copiedSize int64 = 0

	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := destFile.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			copiedSize += int64(n)

			if totalSize > 0 {
				progress := int((copiedSize * 100) / totalSize)
				if progress != updateStatus.DownloadProgress {
					updateStatus.DownloadProgress = progress
					runtime.EventsEmit(a.ctx, "update:status", updateStatus)
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}

// =============================================================================
// Install Methods
// =============================================================================

// InstallUpdate launches the downloaded installer silently as a detached process
// and quits the application if requested.
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

	pkglogger.Info("launching installer", "path", installerPath)

	logPath := filepath.Join(os.TempDir(), "emly_install.log")
	args := []string{
		"/VERYSILENT",
		"/SUPPRESSMSGBOXES",
		"/NORESTART",
		"/FORCEUPGRADE",
		fmt.Sprintf(`/LOG="%s"`, logPath),
	}

	var relaunchPath string
	if quitAfterLaunch {
		relaunchPath, _ = os.Executable()
	}

	if err := launchInstallerAndRelaunch(installerPath, args, relaunchPath); err != nil {
		updateStatus.ErrorMessage = fmt.Sprintf("Failed to launch installer: %v", err)
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	pkglogger.Info("installer launched")

	if quitAfterLaunch {
		time.Sleep(500 * time.Millisecond)
		runtime.Quit(a.ctx)
	}

	return nil
}

// launchDetachedInstaller launches the installer as a completely detached process
// using CreateProcess with DETACHED_PROCESS and CREATE_NEW_PROCESS_GROUP flags.
// This allows the installer to continue running and close EMLy without errors.
//
// Parameters:
//   - exePath: Full path to the installer executable
//   - args: Array of command-line arguments to pass to the installer
//
// Returns:
//   - error: Error if process creation fails
func launchDetachedInstaller(exePath string, args []string) error {
	// Build command line: executable path + arguments
	cmdLine := fmt.Sprintf(`"%s"`, exePath)
	if len(args) > 0 {
		cmdLine += " " + strings.Join(args, " ")
	}

	pkglogger.Info("launching detached installer", "cmd", cmdLine)

	// Convert to UTF16 for Windows API
	cmdLinePtr := syscall.StringToUTF16Ptr(cmdLine)

	// Setup process startup info
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation

	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = syscall.STARTF_USESHOWWINDOW
	si.ShowWindow = syscall.SW_HIDE // Hide installer window (silent mode)

	// Process creation flags:
	// CREATE_NEW_PROCESS_GROUP: Creates process in new process group
	// DETACHED_PROCESS: Process has no console, completely detached from parent
	const (
		CREATE_NEW_PROCESS_GROUP = 0x00000200
		DETACHED_PROCESS         = 0x00000008
	)
	flags := uint32(CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS)

	// Create the detached process
	err := syscall.CreateProcess(
		nil,        // Application name (nil = use command line)
		cmdLinePtr, // Command line
		nil,        // Process security attributes
		nil,        // Thread security attributes
		false,      // Inherit handles
		flags,      // Creation flags
		nil,        // Environment (nil = inherit)
		nil,        // Current directory (nil = inherit)
		&si,        // Startup info
		&pi,        // Process information (output)
	)

	if err != nil {
		pkglogger.Error("CreateProcess failed", "error", err.Error())
		return fmt.Errorf("failed to create detached process: %w", err)
	}

	// Close process and thread handles immediately
	// We don't need to wait for the process - it's fully detached
	syscall.CloseHandle(pi.Process)
	syscall.CloseHandle(pi.Thread)

	pkglogger.Info("detached installer process launched", "pid", pi.ProcessId)

	return nil
}

// launchInstallerAndRelaunch writes a temporary batch file that runs the installer
// and, if relaunchPath is non-empty, restarts the application afterwards.
// The batch is executed via a detached cmd.exe process so it survives EMLy exiting.
func launchInstallerAndRelaunch(installerPath string, args []string, relaunchPath string) error {
	installerCmd := fmt.Sprintf(`"%s" %s`, installerPath, strings.Join(args, " "))
	batch := "@echo off\r\n" + installerCmd + "\r\n"
	if relaunchPath != "" {
		batch += fmt.Sprintf("start \"\" \"%s\"\r\n", relaunchPath)
	}

	batchPath := filepath.Join(os.TempDir(), "emly_update.bat")
	if err := os.WriteFile(batchPath, []byte(batch), 0644); err != nil {
		return fmt.Errorf("failed to write update batch: %w", err)
	}

	pkglogger.Info("launching update batch", "path", batchPath)
	return launchDetachedInstaller("cmd.exe", []string{"/c", batchPath})
}

// InstallUpdateSilent downloads the update (if needed) then delegates to InstallUpdate.
//
// Returns:
//   - error: Error if download or launch fails
func (a *App) InstallUpdateSilent() error {
	pkglogger.Info("starting silent update installation")

	if !updateStatus.Ready || updateStatus.InstallerPath == "" {
		pkglogger.Info("installer not ready, downloading update first")
		if _, err := a.DownloadUpdate(); err != nil {
			pkglogger.Error("failed to download update", "error", err.Error())
			updateStatus.ErrorMessage = fmt.Sprintf("Failed to download update: %v", err)
			return fmt.Errorf("download failed: %w", err)
		}
	}

	return a.InstallUpdate(true)
}

// InstallUpdateSilentFromPath downloads an installer from a custom SMB/network path
// and launches it in silent mode with a detached process. Use this when you know the
// exact installer path (e.g., \\server\updates\EMLy_Installer.exe) without needing
// to check the version.json manifest.
//
// Parameters:
//   - smbPath: Full UNC path or local path to the installer (e.g., \\server\share\EMLy.exe)
//
// Returns:
//   - error: Error if download or launch fails
func (a *App) InstallUpdateSilentFromPath(smbPath string) error {
	pkglogger.Info("starting silent installation from custom path", "path", smbPath)

	// Verify source installer exists and is accessible
	if _, err := os.Stat(smbPath); os.IsNotExist(err) {
		pkglogger.Error("installer not found", "path", smbPath)
		return fmt.Errorf("installer not found at: %s", smbPath)
	}

	// Create temporary directory for installer
	tempDir := os.TempDir()
	installerFilename := filepath.Base(smbPath)
	tempInstallerPath := filepath.Join(tempDir, installerFilename)

	pkglogger.Info("copying installer to temp", "dest", tempInstallerPath)

	// Copy installer from SMB path to local temp
	sourceFile, err := os.Open(smbPath)
	if err != nil {
		pkglogger.Error("failed to open source installer", "error", err.Error())
		return fmt.Errorf("failed to open installer: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(tempInstallerPath)
	if err != nil {
		pkglogger.Error("failed to create temp installer", "error", err.Error())
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer destFile.Close()

	// Copy file
	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		pkglogger.Error("failed to copy installer", "error", err.Error())
		return fmt.Errorf("failed to copy installer: %w", err)
	}

	pkglogger.Info("installer copied", "bytes", bytesWritten)

	logPath := filepath.Join(os.TempDir(), "emly_install.log")
	args := []string{
		"/VERYSILENT",
		"/ALLUSERS",
		"/SUPPRESSMSGBOXES",
		"/NORESTART",
		"/FORCEUPGRADE",
		fmt.Sprintf(`/LOG="%s"`, logPath),
	}

	pkglogger.Info("launching installer", "args", args)

	relaunchPath, _ := os.Executable()
	if err := launchInstallerAndRelaunch(tempInstallerPath, args, relaunchPath); err != nil {
		pkglogger.Error("failed to launch installer", "error", err.Error())
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	pkglogger.Info("installer batch launched, quitting EMLy")

	time.Sleep(500 * time.Millisecond)
	runtime.Quit(a.ctx)

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
