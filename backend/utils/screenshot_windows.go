package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"

	"github.com/kbinani/screenshot"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	dwmapi = syscall.NewLazyDLL("dwmapi.dll")

	// user32 functions
	getForegroundWindow = user32.NewProc("GetForegroundWindow")
	getWindowRect       = user32.NewProc("GetWindowRect")
	findWindowW         = user32.NewProc("FindWindowW")

	// dwmapi functions
	dwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
)

// RECT structure for Windows API
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

const (
	DWMWA_EXTENDED_FRAME_BOUNDS = 9
)

// CaptureWindowByHandle captures a screenshot of a specific window by its handle
func CaptureWindowByHandle(hwnd uintptr) (*image.RGBA, error) {
	if hwnd == 0 {
		return nil, fmt.Errorf("invalid window handle")
	}

	// Try to get the actual window bounds using DWM (handles DPI scaling better)
	var rect RECT
	ret, _, _ := dwmGetWindowAttribute.Call(
		hwnd,
		uintptr(DWMWA_EXTENDED_FRAME_BOUNDS),
		uintptr(unsafe.Pointer(&rect)),
		uintptr(unsafe.Sizeof(rect)),
	)

	// Fallback to GetWindowRect if DWM fails
	if ret != 0 {
		ret, _, err := getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
		if ret == 0 {
			return nil, fmt.Errorf("GetWindowRect failed: %v", err)
		}
	}

	width := int(rect.Right - rect.Left)
	height := int(rect.Bottom - rect.Top)

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid window dimensions: %dx%d", width, height)
	}

	// Using kbinani/screenshot to capture the rectangle on screen
	img, err := screenshot.CaptureRect(image.Rect(int(rect.Left), int(rect.Top), int(rect.Right), int(rect.Bottom)))
	if err != nil {
		return nil, fmt.Errorf("screenshot.CaptureRect failed: %v", err)
	}

	return img, nil
}

// CaptureForegroundWindow captures the currently focused window
func CaptureForegroundWindow() (*image.RGBA, error) {
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return nil, fmt.Errorf("no foreground window found")
	}
	return CaptureWindowByHandle(hwnd)
}

// CaptureWindowByTitle captures a window by its title
func CaptureWindowByTitle(title string) (*image.RGBA, error) {
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return nil, fmt.Errorf("failed to convert title: %v", err)
	}

	hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return nil, fmt.Errorf("window with title '%s' not found", title)
	}
	return CaptureWindowByHandle(hwnd)
}

// ScreenshotToBase64PNG captures a window and returns it as a base64-encoded PNG string
func ScreenshotToBase64PNG(img *image.RGBA) (string, error) {
	if img == nil {
		return "", fmt.Errorf("nil image provided")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %v", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// CaptureWindowToBase64 is a convenience function that captures a window and returns base64 PNG
func CaptureWindowToBase64(hwnd uintptr) (string, error) {
	img, err := CaptureWindowByHandle(hwnd)
	if err != nil {
		return "", err
	}
	return ScreenshotToBase64PNG(img)
}

// CaptureForegroundWindowToBase64 captures the foreground window and returns base64 PNG
func CaptureForegroundWindowToBase64() (string, error) {
	img, err := CaptureForegroundWindow()
	if err != nil {
		return "", err
	}
	return ScreenshotToBase64PNG(img)
}

// CaptureWindowByTitleToBase64 captures a window by title and returns base64 PNG
func CaptureWindowByTitleToBase64(title string) (string, error) {
	img, err := CaptureWindowByTitle(title)
	if err != nil {
		return "", err
	}
	return ScreenshotToBase64PNG(img)
}
