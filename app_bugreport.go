// Package main provides bug reporting functionality for EMLy.
// This file contains methods for creating bug reports with screenshots,
// email files, and system information.
package main

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"emly/backend/utils"
)

// =============================================================================
// Bug Report Types
// =============================================================================

// BugReportResult contains paths to the generated bug report files.
type BugReportResult struct {
	// FolderPath is the path to the bug report folder in temp
	FolderPath string `json:"folderPath"`
	// ScreenshotPath is the path to the captured screenshot file
	ScreenshotPath string `json:"screenshotPath"`
	// MailFilePath is the path to the copied mail file (empty if no mail loaded)
	MailFilePath string `json:"mailFilePath"`
}

// BugReportInput contains the user-provided bug report details.
type BugReportInput struct {
	// Name is the user's name
	Name string `json:"name"`
	// Email is the user's email address for follow-up
	Email string `json:"email"`
	// Description is the detailed bug description
	Description string `json:"description"`
	// ScreenshotData is the base64-encoded PNG screenshot (captured before dialog opens)
	ScreenshotData string `json:"screenshotData"`
	// LocalStorageData is the JSON-encoded localStorage data
	LocalStorageData string `json:"localStorageData"`
	// ConfigData is the JSON-encoded config.ini data
	ConfigData string `json:"configData"`
}

// SubmitBugReportResult contains the result of submitting a bug report.
type SubmitBugReportResult struct {
	// ZipPath is the path to the created zip file
	ZipPath string `json:"zipPath"`
	// FolderPath is the path to the bug report folder
	FolderPath string `json:"folderPath"`
}

// =============================================================================
// Bug Report Methods
// =============================================================================

// CreateBugReportFolder creates a folder in temp with screenshot and optionally
// the current mail file. This is used for the legacy bug report flow.
//
// Returns:
//   - *BugReportResult: Paths to created files
//   - error: Error if folder creation or file operations fail
func (a *App) CreateBugReportFolder() (*BugReportResult, error) {
	// Create unique folder name with timestamp
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

	// Copy currently loaded mail file if one exists
	if a.CurrentMailFilePath != "" {
		mailData, err := os.ReadFile(a.CurrentMailFilePath)
		if err != nil {
			// Log but don't fail - screenshot is still valid
			Log("Failed to read mail file for bug report:", err)
		} else {
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

// SubmitBugReport creates a complete bug report with user input, saves all files,
// and creates a zip archive ready for submission.
//
// The bug report includes:
//   - User-provided description (report.txt)
//   - Screenshot (captured before dialog opens)
//   - Currently loaded mail file (if any)
//   - localStorage data (localStorage.json)
//   - Config.ini data (config.json)
//   - System information (hostname, OS version, hardware ID)
//
// Parameters:
//   - input: User-provided bug report details including pre-captured screenshot, localStorage, and config data
//
// Returns:
//   - *SubmitBugReportResult: Paths to the zip file and folder
//   - error: Error if any file operation fails
func (a *App) SubmitBugReport(input BugReportInput) (*SubmitBugReportResult, error) {
	// Create unique folder name with timestamp
	timestamp := time.Now().Format("20060102_150405")
	folderName := fmt.Sprintf("emly_bugreport_%s", timestamp)

	// Create folder in temp directory
	tempDir := os.TempDir()
	bugReportFolder := filepath.Join(tempDir, folderName)

	if err := os.MkdirAll(bugReportFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bug report folder: %w", err)
	}

	// Save the pre-captured screenshot (captured before dialog opened)
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

	// Save localStorage data if provided
	if input.LocalStorageData != "" {
		localStoragePath := filepath.Join(bugReportFolder, "localStorage.json")
		if err := os.WriteFile(localStoragePath, []byte(input.LocalStorageData), 0644); err != nil {
			Log("Failed to save localStorage data:", err)
		}
	}

	// Save config data if provided
	if input.ConfigData != "" {
		configPath := filepath.Join(bugReportFolder, "config.json")
		if err := os.WriteFile(configPath, []byte(input.ConfigData), 0644); err != nil {
			Log("Failed to save config data:", err)
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

	// Get and save machine/system information
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

	// Create zip archive of the folder
	zipPath := bugReportFolder + ".zip"
	if err := zipFolder(bugReportFolder, zipPath); err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}

	return &SubmitBugReportResult{
		ZipPath:    zipPath,
		FolderPath: bugReportFolder,
	}, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// zipFolder creates a zip archive containing all files from the source folder.
// Directories are traversed recursively but stored implicitly (no directory entries).
//
// Parameters:
//   - sourceFolder: Path to the folder to zip
//   - destZip: Path where the zip file should be created
//
// Returns:
//   - error: Error if any file operation fails
func zipFolder(sourceFolder, destZip string) error {
	// Create the zip file
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the folder and add all files
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

		// Skip directories (they're created implicitly)
		if info.IsDir() {
			return nil
		}

		// Create the file entry in the zip
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Read and write the file content
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = writer.Write(fileContent)
		return err
	})
}
