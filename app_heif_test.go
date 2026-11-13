package main

import (
	"bytes"
	"os"
	"testing"

	"emly/backend/utils/heif"
)

func TestMain(m *testing.M) {
	// Tests run from a temp dir binary: point at the build-libheif.ps1 output.
	heif.SetLibraryPath("third_party/libheif/bin/libheif.dll")
	os.Exit(m.Run())
}

func TestHeifToDisplayJPEG(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n not really a png")
	if out, converted := heifToDisplayJPEG(png, "a.png"); converted || !bytes.Equal(out, png) {
		t.Fatal("non-HEIF bytes must pass through unchanged")
	}

	// A HEIF header with no valid content: decode fails, original returned.
	bogus := append([]byte{0, 0, 0, 0x18}, []byte("ftypheic\x00\x00\x00\x00heicmif1")...)
	if out, converted := heifToDisplayJPEG(bogus, "b.heic"); converted || !bytes.Equal(out, bogus) {
		t.Fatal("undecodable HEIF must pass through unchanged")
	}

	sample, err := os.ReadFile("backend/utils/heif/testdata/example.heic")
	if err != nil {
		t.Fatal(err)
	}
	out, converted := heifToDisplayJPEG(sample, "photo.heic")
	if converted != heif.Available() {
		t.Fatalf("converted = %v, want %v (libheif available)", converted, heif.Available())
	}
	if converted && !bytes.HasPrefix(out, []byte{0xff, 0xd8, 0xff}) {
		t.Fatal("converted output is not JPEG")
	}
	if !converted && !bytes.Equal(out, sample) {
		t.Fatal("without libheif the original bytes must be returned")
	}
}
