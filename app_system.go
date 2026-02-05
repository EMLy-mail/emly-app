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
	"unicode/utf8"

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
