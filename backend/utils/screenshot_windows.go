package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")

	// user32 functions
	getForegroundWindow  = user32.NewProc("GetForegroundWindow")
	getWindowRect        = user32.NewProc("GetWindowRect")
	getClientRect        = user32.NewProc("GetClientRect")
	getDC                = user32.NewProc("GetDC")
	releaseDC            = user32.NewProc("ReleaseDC")
	findWindowW          = user32.NewProc("FindWindowW")
	getWindowDC          = user32.NewProc("GetWindowDC")
	printWindow          = user32.NewProc("PrintWindow")
	clientToScreen       = user32.NewProc("ClientToScreen")

	// gdi32 functions
	createCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	selectObject           = gdi32.NewProc("SelectObject")
	bitBlt                 = gdi32.NewProc("BitBlt")
	deleteDC               = gdi32.NewProc("DeleteDC")
	deleteObject           = gdi32.NewProc("DeleteObject")
	getDIBits              = gdi32.NewProc("GetDIBits")

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

// POINT structure for Windows API
type POINT struct {
	X int32
	Y int32
}

// BITMAPINFOHEADER structure
type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// BITMAPINFO structure
type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]uint32
}

const (
	SRCCOPY             = 0x00CC0020
	DIB_RGB_COLORS      = 0
	BI_RGB              = 0
	PW_CLIENTONLY       = 1
	PW_RENDERFULLCONTENT = 2
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

	// Get window DC
	hdcWindow, _, err := getWindowDC.Call(hwnd)
	if hdcWindow == 0 {
		return nil, fmt.Errorf("GetWindowDC failed: %v", err)
	}
	defer releaseDC.Call(hwnd, hdcWindow)

	// Create compatible DC
	hdcMem, _, err := createCompatibleDC.Call(hdcWindow)
	if hdcMem == 0 {
		return nil, fmt.Errorf("CreateCompatibleDC failed: %v", err)
	}
	defer deleteDC.Call(hdcMem)

	// Create compatible bitmap
	hBitmap, _, err := createCompatibleBitmap.Call(hdcWindow, uintptr(width), uintptr(height))
	if hBitmap == 0 {
		return nil, fmt.Errorf("CreateCompatibleBitmap failed: %v", err)
	}
	defer deleteObject.Call(hBitmap)

	// Select bitmap into DC
	oldBitmap, _, _ := selectObject.Call(hdcMem, hBitmap)
	defer selectObject.Call(hdcMem, oldBitmap)

	// Try PrintWindow first (works better with layered/composited windows)
	ret, _, _ = printWindow.Call(hwnd, hdcMem, PW_RENDERFULLCONTENT)
	if ret == 0 {
		// Fallback to BitBlt
		ret, _, err = bitBlt.Call(
			hdcMem, 0, 0, uintptr(width), uintptr(height),
			hdcWindow, 0, 0,
			SRCCOPY,
		)
		if ret == 0 {
			return nil, fmt.Errorf("BitBlt failed: %v", err)
		}
	}

	// Prepare BITMAPINFO
	bmi := BITMAPINFO{
		BmiHeader: BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			BiWidth:       int32(width),
			BiHeight:      -int32(height), // Negative for top-down DIB
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: BI_RGB,
		},
	}

	// Allocate buffer for pixel data
	pixelDataSize := width * height * 4
	pixelData := make([]byte, pixelDataSize)

	// Get the bitmap bits
	ret, _, err = getDIBits.Call(
		hdcMem,
		hBitmap,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixelData[0])),
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed: %v", err)
	}

	// Convert BGRA to RGBA
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < len(pixelData); i += 4 {
		img.Pix[i+0] = pixelData[i+2] // R <- B
		img.Pix[i+1] = pixelData[i+1] // G <- G
		img.Pix[i+2] = pixelData[i+0] // B <- R
		img.Pix[i+3] = pixelData[i+3] // A <- A
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
