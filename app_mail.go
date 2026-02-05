// Package main provides email reading functionality for EMLy.
// This file contains methods for reading EML, MSG, and PEC email files.
package main

import (
	"emly/backend/utils/mail"
)

// =============================================================================
// Email Reading Methods
// =============================================================================

// ReadEML reads a standard .eml file and returns the parsed email data.
// EML files are MIME-formatted email messages commonly exported from email clients.
//
// Parameters:
//   - filePath: Absolute path to the .eml file
//
// Returns:
//   - *internal.EmailData: Parsed email with headers, body, and attachments
//   - error: Any parsing errors
func (a *App) ReadEML(filePath string) (*internal.EmailData, error) {
	return internal.ReadEmlFile(filePath)
}

// ReadPEC reads a PEC (Posta Elettronica Certificata) .eml file.
// PEC emails are Italian certified emails that contain an inner email message
// wrapped in a certification envelope with digital signatures.
//
// This method extracts and returns the inner original email, ignoring the
// certification wrapper (daticert.xml and signature files are available as attachments).
//
// Parameters:
//   - filePath: Absolute path to the PEC .eml file
//
// Returns:
//   - *internal.EmailData: The inner original email content
//   - error: Any parsing errors
func (a *App) ReadPEC(filePath string) (*internal.EmailData, error) {
	return internal.ReadPecInnerEml(filePath)
}

// ReadMSG reads a Microsoft Outlook .msg file and returns the email data.
// MSG files use the CFB (Compound File Binary) format, which is a proprietary
// format used by Microsoft Office applications.
//
// This method uses an external converter to properly parse the MSG format
// and extract headers, body, and attachments.
//
// Parameters:
//   - filePath: Absolute path to the .msg file
//   - useExternalConverter: Whether to use external conversion (currently always true)
//
// Returns:
//   - *internal.EmailData: Parsed email data
//   - error: Any parsing or conversion errors
func (a *App) ReadMSG(filePath string, useExternalConverter bool) (*internal.EmailData, error) {
	// The useExternalConverter parameter is kept for API compatibility
	// but the implementation always uses the internal MSG reader
	return internal.ReadMsgFile(filePath)
}

// ReadMSGOSS reads a .msg file using the open-source parser.
// This is an alternative entry point that explicitly uses the OSS implementation.
//
// Parameters:
//   - filePath: Absolute path to the .msg file
//
// Returns:
//   - *internal.EmailData: Parsed email data
//   - error: Any parsing errors
func (a *App) ReadMSGOSS(filePath string) (*internal.EmailData, error) {
	return internal.ReadMsgFile(filePath)
}

// ShowOpenFileDialog displays the system file picker dialog filtered for email files.
// This allows users to browse and select .eml or .msg files to open.
//
// The dialog is configured with filters for:
//   - EML files (*.eml)
//   - MSG files (*.msg)
//
// Returns:
//   - string: The selected file path, or empty string if cancelled
//   - error: Any dialog errors
func (a *App) ShowOpenFileDialog() (string, error) {
	return internal.ShowFileDialog(a.ctx)
}
