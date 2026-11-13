//go:build !windows || !cgo

package heif

import "image"

// DLLName is the file EMLy would load on Windows builds with cgo.
const DLLName = "libheif.dll"

var version string

// SetLibraryPath is a no-op: this build cannot load libheif.
func SetLibraryPath(string) {}

func load() error { return ErrUnavailable }

// Decode always fails: this build cannot load libheif (see package doc).
func Decode([]byte) (*image.NRGBA, error) {
	return nil, ErrUnavailable
}
