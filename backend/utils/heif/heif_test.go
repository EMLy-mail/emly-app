package heif

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testdata/example.heic is libheif's examples/example.heic.
//
// Tests load the libheif.dll produced by scripts/build-libheif.ps1; without
// it the decoder tests skip (TestToJPEG then checks the fallback error).

func TestMain(m *testing.M) {
	path := "../../../third_party/libheif/bin/libheif.dll"
	// Set by the child processes of TestBrokenDLLIsUnavailable: the DLL is
	// loaded once per process, so each bad-DLL case needs its own process.
	if p := os.Getenv("EMLY_HEIF_TEST_DLL"); p != "" {
		path = p
	}
	SetLibraryPath(path)
	// CI sets EMLY_HEIF_REQUIRED=1: a libheif.dll that fails to load must
	// fail the run, not silently skip every decoder test.
	if os.Getenv("EMLY_HEIF_REQUIRED") == "1" && os.Getenv("EMLY_HEIF_TEST_CHILD") == "" && !Available() {
		fmt.Fprintf(os.Stderr, "EMLY_HEIF_REQUIRED=1 but libheif is unavailable: %v\n", LoadError())
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestIsHEIF(t *testing.T) {
	sample, err := os.ReadFile("testdata/example.heic")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		data []byte
		want bool
	}{
		{"sample", sample, true},
		{"jpeg", []byte{0xff, 0xd8, 0xff, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0}, false},
		{"mp4", append([]byte{0, 0, 0, 0x18}, []byte("ftypisom")...), false},
		{"avif", append([]byte{0, 0, 0, 0x18}, []byte("ftypavif")...), false},
		{"mif1", append([]byte{0, 0, 0, 0x18}, []byte("ftypmif1")...), true},
		{"short", []byte("ftyp"), false},
		{"empty", nil, false},
	}
	for _, c := range cases {
		if got := IsHEIF(c.data); got != c.want {
			t.Errorf("%s: IsHEIF = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestToJPEG(t *testing.T) {
	sample, err := os.ReadFile("testdata/example.heic")
	if err != nil {
		t.Fatal(err)
	}

	out, err := ToJPEG(sample, DefaultJPEGQuality)
	if !Available() {
		if !errors.Is(err, ErrUnavailable) {
			t.Fatalf("without libheif: err = %v, want ErrUnavailable", err)
		}
		t.Skipf("libheif unavailable: %v", LoadError())
	}
	t.Logf("libheif %s", Version())
	if err != nil {
		t.Fatalf("ToJPEG: %v", err)
	}

	img, err := jpeg.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("output is not a valid JPEG: %v", err)
	}
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		t.Fatalf("bad output size %v", b)
	}
	t.Logf("decoded %dx%d, %d bytes HEIC -> %d bytes JPEG", b.Dx(), b.Dy(), len(sample), len(out))
}

func TestDecodeRejectsGarbage(t *testing.T) {
	if !Available() {
		t.Skipf("libheif unavailable: %v", LoadError())
	}
	for _, data := range [][]byte{
		{},
		[]byte("not an image at all"),
		append([]byte{0, 0, 0, 0x18}, []byte("ftypheic\x00\x00\x00\x00heic")...),
	} {
		if _, err := Decode(data); err == nil {
			t.Errorf("Decode(%q) succeeded, want error", data)
		}
	}
}

// TestDecodeFuzzCorpus feeds libheif's own fuzzing corpus (malformed and
// edge-case files) through Decode: each must return an image or an error,
// never crash. Runs only when scripts/build-libheif.ps1 left the sources in
// third_party/.build.
func TestDecodeFuzzCorpus(t *testing.T) {
	if !Available() {
		t.Skipf("libheif unavailable: %v", LoadError())
	}
	files, _ := filepath.Glob("../../../third_party/.build/libheif-*/fuzzing/data/corpus/*")
	if len(files) == 0 {
		t.Skip("libheif fuzzing corpus not found")
	}
	var ok, failed int
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Decode(data); err != nil {
			failed++
		} else {
			ok++
		}
	}
	t.Logf("corpus: %d decoded, %d rejected", ok, failed)
}

// TestBrokenDLLIsUnavailable checks that a missing libheif.dll, or a DLL
// that is not libheif, leaves EMLy running with Available() == false
// instead of crashing.
func TestBrokenDLLIsUnavailable(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	notLibheif := filepath.Join(t.TempDir(), "libheif.dll")
	sys, err := os.ReadFile(filepath.Join(os.Getenv("SystemRoot"), "System32", "version.dll"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(notLibheif, sys, 0o644); err != nil {
		t.Fatal(err)
	}

	for name, dll := range map[string]string{
		"missing":     filepath.Join(t.TempDir(), "libheif.dll"),
		"not-libheif": notLibheif,
	} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestBrokenDLLChild$", "-test.v")
			cmd.Env = append(os.Environ(), "EMLY_HEIF_TEST_DLL="+dll, "EMLY_HEIF_TEST_CHILD=1")
			out, err := cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(out), "--- PASS: TestBrokenDLLChild") {
				t.Fatalf("child failed: %v\n%s", err, out)
			}
		})
	}
}

func TestBrokenDLLChild(t *testing.T) {
	if os.Getenv("EMLY_HEIF_TEST_CHILD") != "1" {
		t.Skip("run by TestBrokenDLLIsUnavailable")
	}
	if Available() {
		t.Fatal("Available() = true with a broken DLL")
	}
	if !errors.Is(LoadError(), ErrUnavailable) {
		t.Fatalf("LoadError = %v, want ErrUnavailable", LoadError())
	}
	t.Logf("load error: %v", LoadError())
	if _, err := ToJPEG([]byte("x"), DefaultJPEGQuality); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("ToJPEG err = %v, want ErrUnavailable", err)
	}
}
