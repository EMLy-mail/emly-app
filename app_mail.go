// Package main provides email reading functionality for EMLy.
// This file contains methods for reading EML, MSG, and PEC email files.
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"emly/backend/utils"

	pkglogger "emly/backend/logger"
	"emly/backend/utils/mail"

	"github.com/ffois/mailfmt"
)

// =============================================================================
// Email Reading Methods
// =============================================================================

// ReadEML reads a standard .eml file and returns the parsed email data.
func (a *App) ReadEML(filePath string) (data *mailfmt.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadEML", start, err) }()

	logMailFileInfo("ReadEML", filePath)
	pkglogger.TraceStep("go_read_eml_start", filepath.Base(filePath), fileSizeDetail(filePath))
	data, err = mailfmt.ReadEmlFile(filePath)
	pkglogger.TraceStep("go_read_eml_done", fmt.Sprintf("took=%dms", time.Since(start).Milliseconds()), attachmentsDetail(data))
	if err == nil && data != nil {
		logParsedMailInfo("ReadEML", data)
		if !oldAttachmentPreloadEnabled() {
			stripAttachmentData(data)
		}
	}
	return data, err
}

// ReadPEC reads a PEC (Posta Elettronica Certificata) .eml file.
func (a *App) ReadPEC(filePath string) (data *mailfmt.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadPEC", start, err) }()

	logMailFileInfo("ReadPEC", filePath)
	pkglogger.TraceStep("go_read_pec_start", filepath.Base(filePath), fileSizeDetail(filePath))
	data, err = mailfmt.ReadPecInnerEml(filePath)
	pkglogger.TraceStep("go_read_pec_done", fmt.Sprintf("took=%dms", time.Since(start).Milliseconds()), attachmentsDetail(data))
	if err == nil && data != nil {
		logParsedMailInfo("ReadPEC", data)
		if !oldAttachmentPreloadEnabled() {
			stripAttachmentData(data)
		}
	}
	return data, err
}

// ReadMSG reads a Microsoft Outlook .msg file and returns the email data.
func (a *App) ReadMSG(filePath string) (data *mailfmt.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadMSG", start, err) }()

	logMailFileInfo("ReadMSG", filePath)
	pkglogger.TraceStep("go_read_msg_start", filepath.Base(filePath), fileSizeDetail(filePath))
	data, err = mailfmt.ReadMsgFile(filePath)
	pkglogger.TraceStep("go_read_msg_done", fmt.Sprintf("took=%dms", time.Since(start).Milliseconds()), attachmentsDetail(data))
	if err == nil && data != nil {
		logParsedMailInfo("ReadMSG", data)
		if !oldAttachmentPreloadEnabled() {
			stripAttachmentData(data)
		}
	}
	return data, err
}

// DetectEmailFormat inspects the file's binary content to determine its format.
func (a *App) DetectEmailFormat(filePath string) (string, error) {
	start := time.Now()
	format, err := mailfmt.DetectEmailFormat(filePath)
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
func (a *App) ReadAuto(filePath string) (result *mailfmt.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadAuto", start, err) }()

	logMailFileInfo("ReadAuto", filePath)
	pkglogger.TraceStep("go_read_auto_start", filepath.Base(filePath), fileSizeDetail(filePath))

	format, err := mailfmt.DetectEmailFormat(filePath)
	if err != nil {
		pkglogger.TraceStep("go_read_auto_detect_failed", err.Error())
		return nil, err
	}
	pkglogger.TraceStep("go_read_auto_detected", string(format), fmt.Sprintf("took=%dms", time.Since(start).Milliseconds()))

	pkglogger.Debug("auto-detect chose format",
		"function", "ReadAuto",
		"file", filepath.Base(filePath),
		"detected_format", string(format),
	)

	switch format {
	case mailfmt.FormatMSG:
		result, err = mailfmt.ReadMsgFile(filePath)
	default: // FormatEML or FormatUnknown – try PEC first, fall back to plain EML
		result, err = mailfmt.ReadPecInnerEml(filePath)
		if err != nil {
			pkglogger.Debug("PEC parse failed, falling back to plain EML",
				"function", "ReadAuto",
				"pec_error", err.Error(),
			)
			pkglogger.TraceStep("go_read_auto_pec_fallback", err.Error())
			result, err = mailfmt.ReadEmlFile(filePath)
		}
	}
	pkglogger.TraceStep("go_read_auto_done", fmt.Sprintf("took=%dms", time.Since(start).Milliseconds()), attachmentsDetail(result))

	if err == nil && result != nil {
		logParsedMailInfo("ReadAuto", result)
		if !oldAttachmentPreloadEnabled() {
			stripAttachmentData(result)
		}
	}
	return result, err
}

// GetAttachmentData re-parses filePath and returns the base64-encoded bytes
// of the attachment at the given index - its position in the Attachments
// slice that ReadEML/ReadMSG/ReadPEC/ReadAuto returned, which no longer
// carry attachment bytes (see stripAttachmentData). Called on demand, only
// when the user actually opens or saves an attachment, instead of shipping
// every attachment's bytes across the Wails IPC bridge on every mail load -
// see startup-trace.log: for a 41MB .msg with 5 attachments, parsing took
// 196ms but transferring the old, byte-carrying response took over 3s.
//
// Re-parsing (instead of caching the first parse) keeps this stateless and
// always correct, and is cheap for the same reason: parsing was never the
// slow part. The detect-then-dispatch logic mirrors ReadAuto exactly, so
// attachment order/indices stay consistent with whatever the frontend
// originally listed (ReadAuto is used everywhere an email is loaded - see
// +page.ts).
func (a *App) GetAttachmentData(filePath string, index int) (b64 string, err error) {
	start := time.Now()
	defer func() { canonicalLog("GetAttachmentData", start, err) }()

	pkglogger.TraceStep("go_get_attachment_start", filepath.Base(filePath), fmt.Sprintf("index=%d", index))

	format, err := mailfmt.DetectEmailFormat(filePath)
	if err != nil {
		return "", err
	}

	var data *mailfmt.EmailData
	switch format {
	case mailfmt.FormatMSG:
		data, err = mailfmt.ReadMsgFile(filePath)
	default: // FormatEML or FormatUnknown – try PEC first, fall back to plain EML
		data, err = mailfmt.ReadPecInnerEml(filePath)
		if err != nil {
			data, err = mailfmt.ReadEmlFile(filePath)
		}
	}
	if err != nil {
		return "", err
	}

	if index < 0 || index >= len(data.Attachments) {
		err = fmt.Errorf("attachment index %d out of range (have %d)", index, len(data.Attachments))
		return "", err
	}

	b64 = base64.StdEncoding.EncodeToString(data.Attachments[index].Data)
	pkglogger.TraceStep("go_get_attachment_done",
		fmt.Sprintf("took=%dms attachment_bytes=%d", time.Since(start).Milliseconds(), len(data.Attachments[index].Data)))
	return b64, nil
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
// Startup Trace Helpers
// =============================================================================

// fileSizeDetail returns a "size_bytes=N" trace detail for filePath, or ""
// if the file can't be stat'd. Kept separate from logMailFileInfo (which
// logs the same thing to app.log) because the trace file needs it inline on
// the *_start line, before parsing begins - the whole point is comparing
// on-disk size against how long the parse that follows takes.
func fileSizeDetail(filePath string) string {
	if info, err := os.Stat(filePath); err == nil {
		return fmt.Sprintf("size_bytes=%d", info.Size())
	}
	return ""
}

// attachmentsDetail returns an "attachments=N" trace detail, or "" if data
// is nil (parse failed).
func attachmentsDetail(data *mailfmt.EmailData) string {
	if data == nil {
		return ""
	}
	return fmt.Sprintf("attachments=%d", len(data.Attachments))
}

// oldAttachmentPreloadEnabled reports whether OLD_ATTACHMENT_PRELOAD is set
// in config.ini (Settings → Danger Zone → "Old Pre-loading of
// attachments") - an opt-in escape hatch back to the pre-fix behaviour of
// shipping every attachment's full bytes in the initial parse response,
// kept only for experiments/regression testing. Off by default. Read fresh
// on every call (like a.GetConfig() elsewhere) rather than cached, so the
// toggle takes effect immediately, no restart needed.
func oldAttachmentPreloadEnabled() bool {
	cfg, err := utils.LoadConfig(utils.DefaultConfigPath())
	if err != nil || cfg == nil {
		return false
	}
	return cfg.EMLy.OldAttachmentPreload
}

// stripAttachmentData clears every attachment's binary payload in place,
// keeping only metadata (filename, content type). ReadEML/ReadMSG/ReadPEC/
// ReadAuto used to return full attachment bytes, so opening a heavy mail
// meant marshaling and shipping its entire attachment payload across the
// Wails IPC bridge before the user had looked at anything - see
// GetAttachmentData, which fetches one attachment's bytes on demand
// instead, only when the frontend actually needs them (open/save click).
func stripAttachmentData(data *mailfmt.EmailData) {
	if data == nil {
		return
	}
	for i := range data.Attachments {
		data.Attachments[i].Data = nil
	}
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
func logParsedMailInfo(fn string, data *mailfmt.EmailData) {
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
