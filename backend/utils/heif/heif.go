// Package heif decodes HEIC/HEIF images in the backend via libheif.
//
// WebView2 cannot render HEIC, so EMLy converts it to JPEG before handing
// the bytes to the frontend. libheif (LGPL-3.0) is not linked into EMLy.exe:
// on Windows cgo builds (decode_windows.go) a small C shim loads
// libheif.dll at runtime, on first use, from EMLy's own folder only, and
// calls it through function pointers. Keeping libheif in a separate,
// user-replaceable DLL is what satisfies the LGPL.
//
// The DLL (with libde265 inside) is built by scripts/build-libheif.ps1 and
// shipped by the installer. If it is missing or fails to load, EMLy still
// starts: Available reports false, callers keep the original bytes and the
// frontend's libheif-js (WASM) fallback handles them. Builds without cgo or
// for other platforms use decode_stub.go, which never loads anything.
package heif

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

// ErrUnavailable is returned (wrapped) by Decode when libheif.dll cannot be
// loaded or this build cannot load it at all.
var ErrUnavailable = errors.New("heif: libheif unavailable")

// DefaultJPEGQuality matches the frontend decoder's canvas.toDataURL quality.
const DefaultJPEGQuality = 92

// heifBrands are the ftyp major brands of HEVC-coded HEIF files. mif1/msf1
// are generic HEIF brands; an AVIF file using them fails to decode (no AV1
// decoder is linked) and callers fall back to the original bytes.
var heifBrands = map[string]bool{
	"heic": true, "heix": true, "hevc": true, "hevx": true,
	"heim": true, "heis": true, "mif1": true, "msf1": true,
}

// IsHEIF reports whether data starts with an HEIF ftyp box (by content, not
// by filename).
func IsHEIF(data []byte) bool {
	if len(data) < 12 || string(data[4:8]) != "ftyp" {
		return false
	}
	return heifBrands[string(data[8:12])]
}

// Available reports whether libheif.dll is loaded, loading it on the first
// call. The result never changes afterwards.
func Available() bool { return load() == nil }

// LoadError is the reason libheif is unavailable, nil if it loaded.
func LoadError() error { return load() }

// Version is the loaded libheif version ("" if unavailable).
func Version() string {
	if load() != nil {
		return ""
	}
	return version
}

// Error is a libheif decoding error.
type Error struct {
	Code, Subcode int
	Message       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("heif: %s (code %d/%d)", e.Message, e.Code, e.Subcode)
}

// ToJPEG decodes the primary image of an HEIF file (rotation/mirroring
// applied) and re-encodes it as JPEG. Transparent areas are flattened onto
// white, as JPEG has no alpha.
func ToJPEG(data []byte, quality int) ([]byte, error) {
	img, err := Decode(data)
	if err != nil {
		return nil, err
	}

	var src image.Image = img
	if !img.Opaque() {
		flat := image.NewRGBA(img.Bounds())
		draw.Draw(flat, flat.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
		draw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, draw.Over)
		src = flat
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
