package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func buildZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, body := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractZip(t *testing.T) {
	data := buildZip(t, map[string]string{
		"a.txt":          "hello",
		"sub/":           "",
		"sub/deep/b.txt": "world",
	})
	dest := t.TempDir()
	if err := extractZip(data, dest); err != nil {
		t.Fatalf("extractZip: %v", err)
	}
	for path, want := range map[string]string{"a.txt": "hello", "sub/deep/b.txt": "world"} {
		got, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(path)))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q, %v; want %q", path, got, err, want)
		}
	}
}

func TestExtractZipRejectsZipSlip(t *testing.T) {
	for _, name := range []string{"../evil.txt", "sub/../../evil.txt", `..\evil.txt`, "/abs.txt", "C:/abs.txt", "a.txt:ads"} {
		data := buildZip(t, map[string]string{name: "x"})
		dest := t.TempDir()
		if err := extractZip(data, dest); err == nil || !strings.Contains(err.Error(), "unsafe path") {
			t.Errorf("%q: err = %v, want unsafe path", name, err)
		}
	}
}

func TestExtractZipRejectsEncrypted(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.CreateHeader(&zip.FileHeader{Name: "secret.txt", Flags: zipEncryptedFlag})
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("x"))
	w.Close()
	if err := extractZip(buf.Bytes(), t.TempDir()); !errors.Is(err, errZipEncrypted) {
		t.Errorf("err = %v, want errZipEncrypted", err)
	}
}

func TestExtractZipInvalid(t *testing.T) {
	if err := extractZip([]byte("not a zip"), t.TempDir()); err == nil {
		t.Error("want error for invalid archive")
	}
}

func TestExtractZipToTempCleansUpOnError(t *testing.T) {
	data := buildZip(t, map[string]string{"ok.txt": "x", "../evil.txt": "x"})
	before, _ := filepath.Glob(filepath.Join(os.TempDir(), zipTempFolderLabel+"_*_cleanup-test_*"))
	if _, err := extractZipToTemp(data, "cleanup-test.zip"); err == nil {
		t.Fatal("want error")
	}
	after, _ := filepath.Glob(filepath.Join(os.TempDir(), zipTempFolderLabel+"_*_cleanup-test_*"))
	if len(after) != len(before) {
		t.Errorf("temp folder left behind: %v", after)
	}
}

func TestZipTempFolderName(t *testing.T) {
	got := zipTempFolderName("20260101_000000", `Fatture: 2026?.zip`)
	if got != "emly_zip_20260101_000000_Fatture_ 2026_" {
		t.Errorf("got %q", got)
	}
	if got := zipTempFolderName("t", ".zip"); got != "emly_zip_t_archive" {
		t.Errorf("got %q", got)
	}
}
