package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testEML is a minimal multipart mail with one base64 attachment
// ("hello attachment" as application/octet-stream).
const testEML = "From: a@example.com\r\n" +
	"To: b@example.com\r\n" +
	"Subject: test\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Type: multipart/mixed; boundary=\"BOUNDARY\"\r\n" +
	"\r\n" +
	"--BOUNDARY\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"\r\n" +
	"body\r\n" +
	"--BOUNDARY\r\n" +
	"Content-Type: application/octet-stream; name=\"note.bin\"\r\n" +
	"Content-Disposition: attachment; filename=\"note.bin\"\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"\r\n" +
	"aGVsbG8gYXR0YWNobWVudA==\r\n" +
	"--BOUNDARY--\r\n"

func writeTestEML(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "test.eml")
	if err := os.WriteFile(p, []byte(testEML), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func attachmentRequest(path, index string) *http.Request {
	q := url.Values{"path": {path}, "index": {index}}
	return httptest.NewRequest(http.MethodGet, attachmentEndpoint+"?"+q.Encode(), nil)
}

func TestAttachmentMiddleware(t *testing.T) {
	mailPath := writeTestEML(t)
	app := NewApp("test")

	passthrough := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := app.attachmentMiddleware(passthrough)

	serve := func(r *http.Request) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}

	t.Run("other paths pass through", func(t *testing.T) {
		rec := serve(httptest.NewRequest(http.MethodGet, "/index.html", nil))
		if rec.Code != http.StatusTeapot {
			t.Fatalf("got %d, want passthrough", rec.Code)
		}
	})

	t.Run("unknown mail is forbidden", func(t *testing.T) {
		rec := serve(attachmentRequest(mailPath, "0"))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("got %d, want 403", rec.Code)
		}
	})

	if _, err := app.ReadAuto(mailPath); err != nil {
		t.Fatalf("ReadAuto: %v", err)
	}

	t.Run("serves raw bytes", func(t *testing.T) {
		// Different case must still match the allowlist (Windows paths).
		rec := serve(attachmentRequest(strings.ToUpper(mailPath), "0"))
		if rec.Code != http.StatusOK {
			t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
		}
		if got := rec.Body.Bytes(); !bytes.Equal(got, []byte("hello attachment")) {
			t.Fatalf("body = %q", got)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/octet-stream" {
			t.Fatalf("Content-Type = %q", ct)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
			t.Fatalf("Cache-Control = %q", cc)
		}
	})

	t.Run("index out of range", func(t *testing.T) {
		if rec := serve(attachmentRequest(mailPath, "5")); rec.Code != http.StatusNotFound {
			t.Fatalf("got %d, want 404", rec.Code)
		}
	})

	t.Run("bad index", func(t *testing.T) {
		if rec := serve(attachmentRequest(mailPath, "x")); rec.Code != http.StatusBadRequest {
			t.Fatalf("got %d, want 400", rec.Code)
		}
	})

	t.Run("non-GET rejected", func(t *testing.T) {
		r := attachmentRequest(mailPath, "0")
		r.Method = http.MethodPost
		if rec := serve(r); rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("got %d, want 405", rec.Code)
		}
	})
}

func TestSaveMailAttachmentRequiresKnownMail(t *testing.T) {
	app := NewApp("test")
	if _, err := app.SaveMailAttachment(writeTestEML(t), 0, "note.bin"); err == nil {
		t.Fatal("expected error for a mail that was never loaded")
	}
}
