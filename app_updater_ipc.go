// Package main provides EMLy's client for the EMLyUpdater service's IPC
// pipe. This is separate from app_system.go's GetEMLyUpdaterStatus (a local,
// unprivileged pre-check of whether the service is installed/running) —
// these methods perform the actual pipe round trip.
package main

import (
	"context"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils/updateripc"
)

// GetUpdaterSystemInfo queries the EMLyUpdater service over its local named
// pipe for its SYSTEM-authoritative view of this machine's identity. This is
// a distinct source from GetExtendedMachineData (which runs unprivileged, in
// this process) — it does not replace it, and failure here (e.g. the
// updater service is not installed or not running) is an expected,
// recoverable condition, not a hard error for the caller to surface loudly.
func (a *App) GetUpdaterSystemInfo() (info updateripc.SystemInfo, err error) {
	start := time.Now()
	defer func() { canonicalLog("GetUpdaterSystemInfo", start, err) }()

	info, err = updateripc.GetSystemInfo(context.Background())
	if err != nil {
		pkglogger.Warn("GetUpdaterSystemInfo: updater IPC request failed", "error", err.Error())
		return updateripc.SystemInfo{}, err
	}
	return info, nil
}

// GetUpdaterADStatus queries the EMLyUpdater service over its local named
// pipe for its SYSTEM-authoritative AD domain status for this machine.
func (a *App) GetUpdaterADStatus() (status updateripc.ADStatus, err error) {
	start := time.Now()
	defer func() { canonicalLog("GetUpdaterADStatus", start, err) }()

	status, err = updateripc.GetADStatus(context.Background())
	if err != nil {
		pkglogger.Warn("GetUpdaterADStatus: updater IPC request failed", "error", err.Error())
		return updateripc.ADStatus{}, err
	}
	return status, nil
}

// UpdaterIPCStatus describes the health of the EMLyUpdater service's
// named-pipe IPC, as observed by actually round-tripping a request over
// it. This is distinct from GetEMLyUpdaterStatus (a local, unprivileged
// pre-check of service registration/running state that never touches the
// pipe) — Active here means the pipe accepted a connection and answered,
// and Valid means the answer itself looks well-formed, not just present.
type UpdaterIPCStatus struct {
	// Active is true when the dial + request + response round trip over
	// the pipe succeeded.
	Active bool
	// Valid is true when Active is true and the response's fields are
	// non-empty, i.e. it's a genuine SystemInfo snapshot rather than an
	// empty/zero-value message.
	Valid bool
	// Meta carries the raw IPC exchange diagnostics (protocol/sender
	// versions, error code/message) for the SystemInfo request this check
	// issues under the hood.
	Meta updateripc.IPCMeta
}

// GetUpdaterIPCStatus checks whether the EMLyUpdater service's IPC pipe is
// active and returns a valid response, by issuing a real SystemInfo
// request and inspecting the result. A failed round trip (service not
// installed/running, pipe not SYSTEM-owned, timeout, ...) is an expected,
// recoverable condition — it is reported via Active=false rather than as
// an error, matching the posture of GetEMLyUpdaterStatus.
func (a *App) GetUpdaterIPCStatus() (status UpdaterIPCStatus, err error) {
	start := time.Now()
	defer func() { canonicalLog("GetUpdaterIPCStatus", start, err) }()

	info, ipcErr := updateripc.GetSystemInfo(context.Background())
	if ipcErr != nil {
		pkglogger.Debug("GetUpdaterIPCStatus: IPC request failed", "error", ipcErr.Error())
		return UpdaterIPCStatus{Active: false, Valid: false, Meta: info.Meta}, nil
	}

	valid := info.Hostname != "" && info.HWID != ""
	if !valid {
		pkglogger.Warn("GetUpdaterIPCStatus: IPC responded but SystemInfo looks malformed", "info", info)
	}
	return UpdaterIPCStatus{Active: true, Valid: valid, Meta: info.Meta}, nil
}
