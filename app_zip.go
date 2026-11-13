// Package main provides ZIP attachment support for EMLy.
// A ZIP attachment is extracted into a fresh folder under the system temp
// directory, which is then opened in Windows Explorer.
package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkglogger "emly/backend/logger"
	internal "emly/backend/utils/mail"
)

// Extraction limits, a guard against zip bombs: the archive is an untrusted
// email attachment and its declared sizes cannot be trusted either.
const (
	zipMaxEntries      = 10000
	zipMaxTotalBytes   = 2 << 30 // 2 GiB uncompressed, all entries together
	zipEncryptedFlag   = 0x1     // general purpose bit 0: entry is encrypted
	zipTempFolderLabel = "emly_zip"
)

// errZipEncrypted is returned for password-protected archives, which
// archive/zip cannot decrypt.
var errZipEncrypted = errors.New("zip archive is password protected")

// zipTempFolderName builds the temp folder name for an archive: a prefix, a
// timestamp for uniqueness and the archive name without its extension.
func zipTempFolderName(timestamp, filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	base = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) || r < 32 {
			return '_'
		}
		return r
	}, base)
	base = strings.TrimRight(base, ". ")
	if base == "" {
		base = "archive"
	}
	return fmt.Sprintf("%s_%s_%s", zipTempFolderLabel, timestamp, base)
}

// extractZipToTemp extracts data (a ZIP archive) into a new folder under the
// system temp directory and returns that folder's path. On error the
// partially extracted folder is removed.
func extractZipToTemp(data []byte, filename string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	dest, err := os.MkdirTemp(os.TempDir(), zipTempFolderName(timestamp, filename)+"_")
	if err != nil {
		return "", fmt.Errorf("failed to create temp folder: %w", err)
	}
	if err := extractZip(data, dest); err != nil {
		_ = os.RemoveAll(dest)
		return "", err
	}
	return dest, nil
}

// extractZip extracts the ZIP archive in data into the existing folder dest.
// Entries whose path would land outside dest (zip slip) are rejected, and the
// total uncompressed size and entry count are capped.
func extractZip(data []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("invalid zip archive: %w", err)
	}
	if len(r.File) > zipMaxEntries {
		return fmt.Errorf("zip archive has too many entries (%d, max %d)", len(r.File), zipMaxEntries)
	}

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	var remaining int64 = zipMaxTotalBytes
	for _, f := range r.File {
		if f.Flags&zipEncryptedFlag != 0 {
			return errZipEncrypted
		}

		target, err := zipEntryTarget(destAbs, f.Name)
		if err != nil {
			return err
		}
		if target == "" {
			continue // the archive root itself
		}

		mode := f.Mode()
		switch {
		case mode.IsDir():
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("failed to create folder: %w", err)
			}
		case mode.IsRegular():
			n, err := extractZipFile(f, target, remaining)
			if err != nil {
				return err
			}
			remaining -= n
		default:
			// Symlinks and other special entries are skipped: a symlink
			// could point outside dest.
			continue
		}
	}
	return nil
}

// zipEntryTarget maps an entry name to its absolute path under destAbs, or
// fails if it would escape destAbs. Returns "" for the root entry.
func zipEntryTarget(destAbs, name string) (string, error) {
	name = strings.ReplaceAll(name, `\`, "/")
	if strings.HasPrefix(name, "/") || filepath.VolumeName(name) != "" || strings.Contains(name, ":") {
		return "", fmt.Errorf("unsafe path in zip archive: %q", name)
	}
	target := filepath.Join(destAbs, filepath.FromSlash(name))
	if target == destAbs {
		return "", nil
	}
	if !strings.HasPrefix(target, destAbs+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path in zip archive: %q", name)
	}
	return target, nil
}

// extractZipFile writes one regular entry to target, reading at most limit
// bytes, and returns the bytes written.
func extractZipFile(f *zip.File, target string, limit int64) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return 0, fmt.Errorf("failed to create folder: %w", err)
	}
	rc, err := f.Open()
	if err != nil {
		return 0, fmt.Errorf("failed to read %q from zip: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, fmt.Errorf("failed to create %q: %w", filepath.Base(target), err)
	}
	// Read one byte past the limit to tell "exactly at limit" from "over".
	n, err := io.Copy(out, io.LimitReader(rc, limit+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return n, fmt.Errorf("failed to extract %q: %w", f.Name, err)
	}
	if n > limit {
		return n, fmt.Errorf("zip archive too large once extracted (max %d bytes)", int64(zipMaxTotalBytes))
	}
	return n, nil
}

// openZipInExplorer extracts data to a temp folder and opens it in Explorer.
func openZipInExplorer(data []byte, filename string) error {
	dir, err := extractZipToTemp(data, filename)
	if err != nil {
		return err
	}
	pkglogger.Debug("zip attachment extracted",
		"function", "openZipInExplorer",
		"file", filename,
		"folder", dir,
	)
	return internal.OpenFileExplorer(dir)
}

// OpenAttachmentZip extracts the ZIP attachment at index of filePath into a
// temp folder and opens that folder in Windows Explorer.
func (a *App) OpenAttachmentZip(filePath string, index int, filename string) (err error) {
	start := time.Now()
	defer func() { canonicalLog("OpenAttachmentZip", start, err) }()

	att, err := a.loadKnownAttachment(filePath, index)
	if err != nil {
		return err
	}
	return openZipInExplorer(att.Data, filename)
}

// OpenZip is OpenAttachmentZip for base64 data already held by the frontend
// (attachment pre-loading enabled).
func (a *App) OpenZip(base64Data string, filename string) (err error) {
	start := time.Now()
	defer func() { canonicalLog("OpenZip", start, err) }()

	data, err := base64Loader(base64Data)()
	if err != nil {
		return err
	}
	return openZipInExplorer(data, filename)
}
