// Package main provides system-level utilities for EMLy.
// This file contains methods for Windows registry access, character encoding
// conversion, and file system operations.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	pkglogger "emly/backend/logger"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// =============================================================================
// Windows Default App Handler
// =============================================================================

// CheckIsDefaultEMLHandler checks if EMLy is registered as the default handler
// for .eml files in Windows.
//
// This works by:
//  1. Getting the current executable path
//  2. Reading the UserChoice registry key for .eml files
//  3. Finding the command associated with the chosen ProgId
//  4. Comparing the command with our executable
//
// Returns:
//   - bool: True if EMLy is the default handler
//   - error: Error if registry access fails
func (a *App) CheckIsDefaultEMLHandler() (bool, error) {
	// Get current executable path for comparison
	exePath, err := os.Executable()
	if err != nil {
		return false, err
	}
	exePath = strings.ToLower(exePath)

	// Open the UserChoice key for .eml extension
	// This is where Windows stores the user's chosen default app
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\.eml\UserChoice`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		// Key doesn't exist - user hasn't made a specific choice
		// or system default is active (which is usually not us)
		return false, nil
	}
	defer k.Close()

	// Get the ProgId (program identifier) for the chosen app
	progId, _, err := k.GetStringValue("ProgId")
	if err != nil {
		return false, err
	}

	// Find the command associated with this ProgId
	classKeyPath := fmt.Sprintf(`%s\shell\open\command`, progId)
	classKey, err := registry.OpenKey(registry.CLASSES_ROOT, classKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("unable to find command for ProgId %s", progId)
	}
	defer classKey.Close()

	// Get the command string
	cmd, _, err := classKey.GetStringValue("")
	if err != nil {
		return false, err
	}

	// Compare command with our executable
	// Check if the command contains our executable name
	cmdLower := strings.ToLower(cmd)
	if strings.Contains(cmdLower, strings.ToLower(filepath.Base(exePath))) {
		return true, nil
	}

	return false, nil
}

// OpenDefaultAppsSettings opens the Windows Settings app to the Default Apps page.
// This allows users to easily set EMLy as the default handler for email files.
//
// Returns:
//   - error: Error if launching settings fails
func (a *App) OpenDefaultAppsSettings() error {
	cmd := exec.Command("cmd", "/c", "start", "ms-settings:defaultapps")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Start()
}

// =============================================================================
// Character Encoding
// =============================================================================

// ConvertToUTF8 attempts to convert a string to valid UTF-8.
// If the string is already valid UTF-8, it's returned as-is.
// Otherwise, it assumes Windows-1252 encoding (common for legacy emails)
// and attempts to decode it.
//
// This is particularly useful for email body content that may have been
// encoded with legacy Western European character sets.
//
// Parameters:
//   - s: The string to convert
//
// Returns:
//   - string: UTF-8 encoded string
func (a *App) ConvertToUTF8(s string) string {
	// If already valid UTF-8, return as-is
	if utf8.ValidString(s) {
		return s
	}

	// Assume Windows-1252 (superset of ISO-8859-1)
	// This is the most common encoding for legacy Western European text
	decoder := charmap.Windows1252.NewDecoder()
	decoded, _, err := transform.String(decoder, s)
	if err != nil {
		// Return original if decoding fails
		return s
	}
	return decoded
}

// =============================================================================
// File System Operations
// =============================================================================

// OpenFolderInExplorer opens the specified folder in Windows Explorer.
// This is used to show the user where bug report files are saved.
//
// Parameters:
//   - folderPath: The path to the folder to open
//
// Returns:
//   - error: Error if launching explorer fails
func (a *App) OpenFolderInExplorer(folderPath string) error {
	cmd := exec.Command("explorer", folderPath)
	return cmd.Start()
}

// =============================================================================
// EMLy Updater Detection
// =============================================================================

// EMLyUpdaterStatus describes whether the EMLy Updater is installed on this
// machine and whether its service is currently running.
type EMLyUpdaterStatus struct {
	// Installed is true when the "EMLyUpdater" service is registered and
	// both its installation and config folders exist.
	Installed bool
	// Running is true when the "EMLyUpdater" service is currently started.
	Running bool
}

// GetEMLyUpdaterStatus reports the installation and running state of the
// EMLy Updater. Installed requires all three of the following:
//  1. The "EMLyUpdater" Windows service is registered
//  2. The installation folder exists (%ProgramFiles%\EMLyUpdater)
//  3. The config folder exists (%ProgramData%\EMLyUpdater)
//
// Every check reads state that a standard (non-administrator) account can
// already access, so no elevation is required.
func (a *App) GetEMLyUpdaterStatus() EMLyUpdaterStatus {
	registered := emlyUpdaterServiceRegistered()
	installDir := filepath.Join(programFilesDir(), "EMLyUpdater")
	configDir := filepath.Join(programDataDir(), "EMLyUpdater")
	installDirOk := dirExists(installDir)
	configDirOk := dirExists(configDir)
	running := emlyUpdaterServiceRunning()

	status := EMLyUpdaterStatus{
		Installed: registered && installDirOk && configDirOk,
		Running:   running,
	}
	pkglogger.Debug("EMLy Updater status check",
		"serviceRegistered", registered,
		"installDir", installDir,
		"installDirExists", installDirOk,
		"configDir", configDir,
		"configDirExists", configDirOk,
		"running", running,
		"result", status,
	)
	return status
}

// emlyUpdaterServiceRegistered checks the registry for the "EMLyUpdater"
// service. Reading HKLM\SYSTEM\CurrentControlSet\Services\<name> only
// requires KEY_READ, which standard accounts are granted by default -
// unlike opening a handle to the Service Control Manager to query status.
func emlyUpdaterServiceRegistered() bool {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\EMLyUpdater`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		pkglogger.Debug("EMLy Updater service registry key not readable", "error", err.Error())
		return false
	}
	defer k.Close()
	return true
}

// emlyUpdaterServiceRunning queries the Service Control Manager for the
// current run state of the "EMLyUpdater" service.
//
// This deliberately does NOT use golang.org/x/sys/windows/svc/mgr: that
// package's Connect() and OpenService() request SC_MANAGER_ALL_ACCESS /
// SERVICE_ALL_ACCESS, which the SCM denies to non-administrator accounts
// even for a read-only status query. Opening the manager and the service
// with only the QUERY-level access rights below works for standard users.
func emlyUpdaterServiceRunning() bool {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		pkglogger.Debug("EMLy Updater: OpenSCManager failed", "error", err.Error())
		return false
	}
	defer windows.CloseServiceHandle(scm)

	serviceName, err := syscall.UTF16PtrFromString("EMLyUpdater")
	if err != nil {
		pkglogger.Debug("EMLy Updater: failed to encode service name", "error", err.Error())
		return false
	}

	svcHandle, err := windows.OpenService(scm, serviceName, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		pkglogger.Debug("EMLy Updater: OpenService failed", "error", err.Error())
		return false
	}
	defer windows.CloseServiceHandle(svcHandle)

	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(svcHandle, &status); err != nil {
		pkglogger.Debug("EMLy Updater: QueryServiceStatus failed", "error", err.Error())
		return false
	}

	pkglogger.Debug("EMLy Updater: service status queried", "currentState", status.CurrentState)
	return status.CurrentState == windows.SERVICE_RUNNING
}

// dirExists reports whether path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// programFilesDir returns %ProgramFiles%, falling back to the standard
// default if the environment variable isn't set.
func programFilesDir() string {
	if dir := os.Getenv("ProgramFiles"); dir != "" {
		return dir
	}
	return `C:\Program Files`
}

// programDataDir returns %ProgramData%, falling back to the standard
// default if the environment variable isn't set.
func programDataDir() string {
	if dir := os.Getenv("ProgramData"); dir != "" {
		return dir
	}
	return `C:\ProgramData`
}

// GetLogsDir returns the path to the EMLy logs directory.
func (a *App) GetLogsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("GetLogsDir: %w", err)
	}
	return filepath.Join(configDir, "EMLy", "logs"), nil
}

// OpenURLInBrowser opens the specified URL in the system's default web browser.
// Only http://, https://, and mailto: schemes are accepted to prevent
// command injection via cmd /c start.
//
// Parameters:
//   - url: The URL to open (must start with http://, https://, or mailto:)
//
// Returns:
//   - error: Error if the scheme is not allowed or launching the browser fails
func (a *App) OpenURLInBrowser(url string) error {
	lower := strings.ToLower(url)
	if !strings.HasPrefix(lower, "http://") &&
		!strings.HasPrefix(lower, "https://") &&
		!strings.HasPrefix(lower, "mailto:") {
		return fmt.Errorf("URL scheme not allowed: %s", url)
	}
	cmd := exec.Command("cmd", "/c", "start", "", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Start()
}
