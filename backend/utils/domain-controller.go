package utils

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"emly/backend/logger"

	"golang.org/x/sys/windows"
)

var (
	modNetapi32       = windows.NewLazySystemDLL("netapi32.dll")
	procDsGetDcNameW  = modNetapi32.NewProc("DsGetDcNameW")
	procNetApiBufferFree = modNetapi32.NewProc("NetApiBufferFree")
)

// DS_RETURN_DNS_NAME asks DsGetDcNameW to return the domain controller's
// DNS name (rather than its NetBIOS name) in DomainControllerName.
const dsReturnDNSName = 0x40000000

// domainControllerInfo mirrors the Win32 DOMAIN_CONTROLLER_INFOW struct.
// Field order and sizes must match the native layout exactly.
type domainControllerInfo struct {
	DomainControllerName        *uint16
	DomainControllerAddress     *uint16
	DomainControllerAddressType uint32
	DomainGUID                  windows.GUID
	DomainName                  *uint16
	DNSForestName               *uint16
	Flags                       uint32
	DCSiteName                  *uint16
	ClientSiteName              *uint16
}

// DomainControllerInfo is the JSON-friendly result returned to callers
// (and, via the App wrapper, to the frontend).
type DomainControllerInfo struct {
	Name string `json:"Name"`
	Site string `json:"Site"`
}

// GetNearestDomainController resolves the domain controller nearest to this
// machine using the native DsGetDcNameW Win32 API (site-aware, no external
// process spawned). Pass an empty domain to resolve the DC for the machine's
// own AD domain.
//
// Returns the DC's DNS name and the AD site name it belongs to.
func GetNearestDomainController(domain string) (*DomainControllerInfo, error) {
	logger.Debug("GetNearestDomainController: resolving nearest DC", "domain", domain)

	var domainPtr *uint16
	if domain != "" {
		var err error
		domainPtr, err = syscall.UTF16PtrFromString(domain)
		if err != nil {
			return nil, fmt.Errorf("invalid domain name %q: %w", domain, err)
		}
	}

	// infoPtr is declared as a typed pointer (rather than uintptr) so the
	// Win32-written pointer value can be read back directly, without a
	// second unsafe.Pointer(uintptr(...)) conversion that `go vet` cannot
	// verify as safe.
	var infoPtr *domainControllerInfo
	ret, _, _ := procDsGetDcNameW.Call(
		0, // ComputerName (NULL = local machine)
		uintptr(unsafe.Pointer(domainPtr)),
		0, // DomainGuid
		0, // SiteName
		uintptr(dsReturnDNSName),
		uintptr(unsafe.Pointer(&infoPtr)),
	)
	if ret != 0 {
		err := syscall.Errno(ret)
		logger.Error("GetNearestDomainController: DsGetDcNameW failed", "error", err.Error(), "code", ret)
		return nil, fmt.Errorf("DsGetDcName failed: %w", err)
	}
	defer procNetApiBufferFree.Call(uintptr(unsafe.Pointer(infoPtr)))

	// DsGetDcName prefixes both names with "\\"; strip it for display/use.
	dcName := strings.TrimPrefix(windows.UTF16PtrToString(infoPtr.DomainControllerName), `\\`)
	siteName := windows.UTF16PtrToString(infoPtr.DCSiteName)

	logger.Debug("GetNearestDomainController: resolved", "dc", dcName, "site", siteName)
	return &DomainControllerInfo{Name: dcName, Site: siteName}, nil
}
