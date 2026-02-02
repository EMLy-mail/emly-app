package utils

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type FileMetadata struct {
	Name             string
	Size             int64
	LastModified     time.Time
	ProductVersion   string
	FileVersion      string
	OriginalFilename string
	ProductName      string
	FileDescription  string
	CompanyName      string
}

// GetFileMetadata returns metadata for the given file path.
// It retrieves basic file info and Windows-specific version info if available.
func GetFileMetadata(path string) (*FileMetadata, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	metadata := &FileMetadata{
		Name:         fileInfo.Name(),
		Size:         fileInfo.Size(),
		LastModified: fileInfo.ModTime(),
	}

	// Try to get version info
	versionInfo, err := getVersionInfo(path)
	if err == nil {
		metadata.ProductVersion = versionInfo["ProductVersion"]
		metadata.FileVersion = versionInfo["FileVersion"]
		metadata.OriginalFilename = versionInfo["OriginalFilename"]
		metadata.ProductName = versionInfo["ProductName"]
		metadata.FileDescription = versionInfo["FileDescription"]
		metadata.CompanyName = versionInfo["CompanyName"]
	}

	return metadata, nil
}

func getVersionInfo(path string) (map[string]string, error) {
	var version = syscall.NewLazyDLL("version.dll")
	var getFileVersionInfoSize = version.NewProc("GetFileVersionInfoSizeW")
	var getFileVersionInfo = version.NewProc("GetFileVersionInfoW")
	var verQueryValue = version.NewProc("VerQueryValueW")

	pathPtr, _ := syscall.UTF16PtrFromString(path)

	// Get size of version info
	size, _, err := getFileVersionInfoSize.Call(uintptr(unsafe.Pointer(pathPtr)), 0)
	if size == 0 {
		return nil, err
	}

	// Get version info
	info := make([]byte, size)
	ret, _, err := getFileVersionInfo.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		size,
		uintptr(unsafe.Pointer(&info[0])),
	)
	if ret == 0 {
		return nil, err
	}

	// Query language and codepage
	var langCodePagePtr *struct {
		Language uint16
		CodePage uint16
	}
	var length uint32
	subBlock := "\\VarFileInfo\\Translation"
	subBlockPtr, _ := syscall.UTF16PtrFromString(subBlock)

	ret, _, err = verQueryValue.Call(
		uintptr(unsafe.Pointer(&info[0])),
		uintptr(unsafe.Pointer(subBlockPtr)),
		uintptr(unsafe.Pointer(&langCodePagePtr)),
		uintptr(unsafe.Pointer(&length)),
	)
	if ret == 0 {
		return nil, err
	}

	// Helper to query string values
	queryValue := func(key string) string {
		query := fmt.Sprintf("\\StringFileInfo\\%04x%04x\\%s", langCodePagePtr.Language, langCodePagePtr.CodePage, key)
		queryPtr, _ := syscall.UTF16PtrFromString(query)
		var valPtr *uint16
		var valLen uint32

		ret, _, _ := verQueryValue.Call(
			uintptr(unsafe.Pointer(&info[0])),
			uintptr(unsafe.Pointer(queryPtr)),
			uintptr(unsafe.Pointer(&valPtr)),
			uintptr(unsafe.Pointer(&valLen)),
		)
		if ret != 0 && valLen > 0 {
			// valPtr points to a UTF-16 string, create a Go string from it
			// We need to iterate until null terminator because valLen includes it
			// but syscall.UTF16ToString expects a slice without the terminator if we want clean output,
			// or we can just use the pointer.
			// However, easier way with unsafe:
			return syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(valPtr))[:valLen])
		}
		return ""
	}

	results := make(map[string]string)
	keys := []string{"ProductVersion", "FileVersion", "OriginalFilename", "ProductName", "FileDescription", "CompanyName"}
	for _, key := range keys {
		results[key] = queryValue(key)
	}

	return results, nil
}
