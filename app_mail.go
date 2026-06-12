// Package main provides email reading functionality for EMLy.
// This file contains methods for reading EML, MSG, and PEC email files.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils/mail"
)

// =============================================================================
// Email Reading Methods
// =============================================================================

// ReadEML reads a standard .eml file and returns the parsed email data.
func (a *App) ReadEML(filePath string) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadEML", start, err) }()

	logMailFileInfo("ReadEML", filePath)
	data, err = internal.ReadEmlFile(filePath)
	if err == nil && data != nil {
		logParsedMailInfo("ReadEML", data)
	}
	return data, err
}

// ReadPEC reads a PEC (Posta Elettronica Certificata) .eml file.
func (a *App) ReadPEC(filePath string) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadPEC", start, err) }()

	logMailFileInfo("ReadPEC", filePath)
	data, err = internal.ReadPecInnerEml(filePath)
	if err == nil && data != nil {
		logParsedMailInfo("ReadPEC", data)
	}
	return data, err
}

// ReadMSG reads a Microsoft Outlook .msg file and returns the email data.
func (a *App) ReadMSG(filePath string) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadMSG", start, err) }()

	logMailFileInfo("ReadMSG", filePath)
	data, err = internal.ReadMsgFile(filePath)
	if err == nil && data != nil {
		logParsedMailInfo("ReadMSG", data)
	}
	return data, err
}

// DetectEmailFormat inspects the file's binary content to determine its format.
func (a *App) DetectEmailFormat(filePath string) (string, error) {
	start := time.Now()
	format, err := internal.DetectEmailFormat(filePath)
	canonicalLog("DetectEmailFormat", start, err)

	pkglogger.Debug("email format detected",
		"function", "DetectEmailFormat",
		"file", filepath.Base(filePath),
		"extension", strings.ToLower(filepath.Ext(filePath)),
		"detected_format", string(format),
	)
	return string(format), err
}

// ReadAuto automatically detects the email file format and delegates to the
// appropriate reader.
func (a *App) ReadAuto(filePath string) (result *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadAuto", start, err) }()

	logMailFileInfo("ReadAuto", filePath)

	format, err := internal.DetectEmailFormat(filePath)
	if err != nil {
		return nil, err
	}

	pkglogger.Debug("auto-detect chose format",
		"function", "ReadAuto",
		"file", filepath.Base(filePath),
		"detected_format", string(format),
	)

	switch format {
	case internal.FormatMSG:
		result, err = internal.ReadMsgFile(filePath)
	default: // FormatEML or FormatUnknown – try PEC first, fall back to plain EML
		result, err = internal.ReadPecInnerEml(filePath)
		if err != nil {
			pkglogger.Debug("PEC parse failed, falling back to plain EML",
				"function", "ReadAuto",
				"pec_error", err.Error(),
			)
			result, err = internal.ReadEmlFile(filePath)
		}
	}

	if err == nil && result != nil {
		logParsedMailInfo("ReadAuto", result)
	}
	return result, err
}

// ShowOpenFileDialog displays the system file picker dialog filtered for email files.
func (a *App) ShowOpenFileDialog() (string, error) {
	return internal.ShowFileDialog(a.ctx)
}

// ShowOpenFolderDialog displays the system directory picker dialog.
// Returns the selected folder path, or an empty string if cancelled.
func (a *App) ShowOpenFolderDialog() (string, error) {
	return internal.ShowFolderDialog(a.ctx)
}

// SaveAttachment saves a base64-encoded attachment to disk without going
// through the WebView2 download manager. The target folder is the
// EXPORT_ATTACHMENT_FOLDER from config.ini if set, otherwise the user's
// Downloads folder. Existing files are never overwritten.
//
// Parameters:
//   - filename: The name to save the file as
//   - base64Data: The base64-encoded attachment data
//
// Returns:
//   - string: The full path where the file was saved
//   - error: Any decoding or file system errors
func (a *App) SaveAttachment(filename string, base64Data string) (savedPath string, err error) {
	start := time.Now()
	defer func() { canonicalLog("SaveAttachment", start, err) }()

	savedPath, err = internal.SaveAttachmentToFolder(filename, base64Data, a.GetExportAttachmentFolder())
	if err != nil {
		return "", err
	}

	pkglogger.Debug("attachment saved",
		"function", "SaveAttachment",
		"file", filepath.Base(savedPath),
		"folder", filepath.Dir(savedPath),
	)
	return savedPath, nil
}

// OpenExplorerForPath opens Windows Explorer showing the specified file
// (selected) or folder.
//
// Parameters:
//   - path: The full path to the file or folder to show in Explorer
//
// Returns:
//   - error: Any execution errors
func (a *App) OpenExplorerForPath(path string) error {
	return internal.OpenFileExplorer(path)
}

// =============================================================================
// Debug Logging Helpers
// =============================================================================

// logMailFileInfo logs file-level details before parsing begins.
func logMailFileInfo(fn, filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))
	args := []any{
		"function", fn,
		"file", filepath.Base(filePath),
		"extension", ext,
	}
	if info, err := os.Stat(filePath); err == nil {
		args = append(args, "size_bytes", info.Size())
	}
	pkglogger.Debug("loading mail file", args...)
}

// logParsedMailInfo logs details extracted after successfully parsing an email.
func logParsedMailInfo(fn string, data *internal.EmailData) {
	bodyType := "none"
	if strings.Contains(data.Body, "<html") || strings.Contains(data.Body, "<HTML") || strings.Contains(data.Body, "<div") {
		bodyType = "html"
	} else if data.Body != "" {
		bodyType = "text"
	}

	// Collect unique MIME types from attachments
	mimeTypes := make(map[string]bool)
	for _, att := range data.Attachments {
		if att.ContentType != "" {
			mimeTypes[att.ContentType] = true
		}
	}
	var mimeList []string
	for mt := range mimeTypes {
		mimeList = append(mimeList, mt)
	}

	// Truncate subject for logging
	subject := data.Subject
	if len(subject) > 80 {
		subject = subject[:80] + "..."
	}

	pkglogger.Debug("mail parsed successfully",
		"function", fn,
		"subject", subject,
		"from", data.From,
		"to_count", len(data.To),
		"cc_count", len(data.Cc),
		"body_type", bodyType,
		"body_length", len(data.Body),
		"attachment_count", len(data.Attachments),
		"attachment_mimes", strings.Join(mimeList, ", "),
		"is_pec", data.IsPec,
		"has_inner_email", data.HasInnerEmail,
	)
}
