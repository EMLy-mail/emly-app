// Package main provides by-reference attachment access for EMLy.
//
// Attachments are addressed as (mail file path, attachment index) instead of
// being shipped to the frontend as base64 strings. Before this, opening a
// 2.5MB attachment produced ~25-30MB of transient copies (re-parse, base64,
// several JSON marshals in the Wails IPC bridge, UTF-16 conversion for
// ExecuteScript), only for the frontend to send the same base64 straight
// back to Go to write a temp file. Now:
//   - actions that end in Go (viewer windows, external apps, save to disk)
//     load the bytes in Go and never touch the frontend;
//   - the frontend reads raw bytes, when it really needs them, from the
//     /emly-attachment AssetServer endpoint (binary HTTP, no base64/JSON).
package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pkglogger "emly/backend/logger"
	internal "emly/backend/utils/mail"

	"github.com/ffois/mailfmt"
)

// attachmentEndpoint is the AssetServer path that serves raw attachment
// bytes: GET /emly-attachment?path=<mail file>&index=<n>[&display=1]
// (display=1 converts HEIC/HEIF to JPEG, see heifToDisplayJPEG).
const attachmentEndpoint = "/emly-attachment"

// normalizeMailPath gives a stable allowlist key for a Windows path
// (cleaned, case-insensitive).
func normalizeMailPath(filePath string) string {
	return strings.ToLower(filepath.Clean(filePath))
}

// registerMailPath adds filePath to the allowlist of mail files whose
// attachments may be served. Called by the Read* methods after a
// successful parse.
func (a *App) registerMailPath(filePath string) {
	a.knownMailsMux.Lock()
	a.knownMails[normalizeMailPath(filePath)] = true
	a.knownMailsMux.Unlock()
}

// isKnownMailPath reports whether filePath was loaded via a Read* method.
func (a *App) isKnownMailPath(filePath string) bool {
	a.knownMailsMux.Lock()
	defer a.knownMailsMux.Unlock()
	return a.knownMails[normalizeMailPath(filePath)]
}

// loadAttachment returns the attachment at index of filePath. The
// detect-then-dispatch logic mirrors ReadAuto exactly, so indices stay
// consistent with the list the frontend received.
//
// MSG files use mailfmt.ReadMsgAttachmentFile, which reads only the requested
// attachment's bytes (the old full re-parse decoded every attachment and the
// body just to return one). EML/PEC still go through a full MIME parse.
func loadAttachment(filePath string, index int) (*mailfmt.EmailAttachment, error) {
	format, err := mailfmt.DetectEmailFormat(filePath)
	if err != nil {
		return nil, err
	}

	if format == mailfmt.FormatMSG {
		return mailfmt.ReadMsgAttachmentFile(filePath, index)
	}

	// FormatEML or FormatUnknown – try PEC first, fall back to plain EML
	data, err := mailfmt.ReadPecInnerEml(filePath)
	if err != nil {
		data, err = mailfmt.ReadEmlFile(filePath)
	}
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(data.Attachments) {
		return nil, fmt.Errorf("attachment index %d out of range (have %d)", index, len(data.Attachments))
	}
	return &data.Attachments[index], nil
}

// loadKnownAttachment is loadAttachment restricted to allowlisted mails.
func (a *App) loadKnownAttachment(filePath string, index int) (*mailfmt.EmailAttachment, error) {
	if !a.isKnownMailPath(filePath) {
		return nil, fmt.Errorf("mail file not loaded: %s", filepath.Base(filePath))
	}
	return loadAttachment(filePath, index)
}

// attachmentLoader adapts loadKnownAttachment to the viewer-window loader
// signature, so the bytes are only read after the duplicate-window check.
func (a *App) attachmentLoader(filePath string, index int) func() ([]byte, error) {
	return func() ([]byte, error) {
		att, err := a.loadKnownAttachment(filePath, index)
		if err != nil {
			return nil, err
		}
		return att.Data, nil
	}
}

// =============================================================================
// Bound methods (by reference: mail file path + attachment index)
// =============================================================================

// OpenAttachmentPDFWindow opens the PDF attachment at index of filePath in
// the built-in PDF viewer window. filename is used for the duplicate-window
// check and the temp file name.
func (a *App) OpenAttachmentPDFWindow(filePath string, index int, filename string) (err error) {
	start := time.Now()
	defer func() { canonicalLog("OpenAttachmentPDFWindow", start, err) }()
	return a.openPDFViewer(a.attachmentLoader(filePath, index), filename)
}

// OpenAttachmentImageWindow opens the image attachment at index of filePath
// in the built-in image viewer window.
func (a *App) OpenAttachmentImageWindow(filePath string, index int, filename string) (err error) {
	start := time.Now()
	defer func() { canonicalLog("OpenAttachmentImageWindow", start, err) }()
	return a.openImageViewer(a.attachmentLoader(filePath, index), filename)
}

// OpenAttachmentEMLWindow opens the EML attachment at index of filePath in a
// new EMLy window.
func (a *App) OpenAttachmentEMLWindow(filePath string, index int, filename string) (err error) {
	start := time.Now()
	defer func() { canonicalLog("OpenAttachmentEMLWindow", start, err) }()
	return a.openEMLViewer(a.attachmentLoader(filePath, index), filename)
}

// OpenAttachmentExternal writes the attachment at index of filePath to a
// temp file and opens it with the system's default application. Replaces
// OpenPDF/OpenImage/OpenDocument for by-reference attachments.
func (a *App) OpenAttachmentExternal(filePath string, index int, filename string) (err error) {
	start := time.Now()
	defer func() { canonicalLog("OpenAttachmentExternal", start, err) }()

	att, err := a.loadKnownAttachment(filePath, index)
	if err != nil {
		return err
	}
	return openBytesWithDefaultApp(att.Data, filename)
}

// SaveMailAttachment saves the attachment at index of filePath into the
// export folder (see SaveAttachment) and returns the saved path.
func (a *App) SaveMailAttachment(filePath string, index int, filename string) (savedPath string, err error) {
	start := time.Now()
	defer func() { canonicalLog("SaveMailAttachment", start, err) }()

	att, err := a.loadKnownAttachment(filePath, index)
	if err != nil {
		return "", err
	}
	savedPath, err = internal.SaveAttachmentBytesToFolder(filename, att.Data, a.GetExportAttachmentFolder())
	if err != nil {
		return "", err
	}

	pkglogger.Debug("attachment saved",
		"function", "SaveMailAttachment",
		"file", filepath.Base(savedPath),
		"folder", filepath.Dir(savedPath),
	)
	return savedPath, nil
}

// =============================================================================
// AssetServer endpoint
// =============================================================================

// attachmentMiddleware serves GET /emly-attachment?path=...&index=... with
// the raw attachment bytes and passes every other request through. Only
// allowlisted mail files are served (403 otherwise).
func (a *App) attachmentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != attachmentEndpoint {
			next.ServeHTTP(w, r)
			return
		}
		a.serveAttachment(w, r)
	})
}

func (a *App) serveAttachment(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var err error
	defer func() { canonicalLog("serveAttachment", start, err) }()

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	filePath := q.Get("path")
	index, convErr := strconv.Atoi(q.Get("index"))
	if filePath == "" || convErr != nil {
		err = fmt.Errorf("bad attachment request: %q", r.URL.RawQuery)
		http.Error(w, "missing or invalid path/index", http.StatusBadRequest)
		return
	}
	if !a.isKnownMailPath(filePath) {
		err = fmt.Errorf("attachment request for unknown mail: %s", filepath.Base(filePath))
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	att, err := loadAttachment(filePath, index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body := att.Data
	contentType := att.ContentType
	// display=1: the bytes are for an in-WebView viewer, not a save, so
	// HEIC/HEIF may be swapped for a JPEG the WebView can render.
	if q.Get("display") == "1" {
		if jpg, converted := heifToDisplayJPEG(body, att.Filename); converted {
			body, contentType = jpg, "image/jpeg"
		}
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	h := w.Header()
	h.Set("Content-Type", contentType)
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
}
