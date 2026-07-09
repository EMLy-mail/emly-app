// Package updateripc is the EMLy-side client for the EMLyUpdater service's
// named-pipe IPC (see emly-updater's internal/ipc and proto/updateripc.proto,
// manually synced into ./proto/updateripc.proto in this repo). It fetches a
// SYSTEM-authoritative SystemInfo/ADStatus snapshot from the updater service,
// distinct from and not a replacement for backend/utils.GetExtendedMachineInfo
// (which runs unprivileged, in-process).
package updateripc

// SystemInfo is the EMLy-side view of the updater's SystemInfoResponse,
// decoupled from the generated protobuf type so callers don't depend on the
// wire schema directly.
type SystemInfo struct {
	Hostname   string
	HWID       string
	InternalIP string
	OSVersion  string
	// Meta carries the raw IPC exchange diagnostics for this request, even
	// on failure (partially populated — see IPCMeta's doc comment).
	Meta IPCMeta
}

// ADStatus is the EMLy-side view of the updater's ADStatusResponse.
type ADStatus struct {
	ADDomain     string
	DomainJoined bool
	// Meta carries the raw IPC exchange diagnostics for this request, even
	// on failure (partially populated — see IPCMeta's doc comment).
	Meta IPCMeta
}

// IPCMeta captures wire-level diagnostics from a single EMLyUpdater IPC
// exchange — the negotiated protocol_version and each side's own semver
// (sender_version), plus the raw ErrorResponse code/message when the
// updater rejected the request. Response* fields stay zero-valued when no
// response envelope was ever received (e.g. dial/read failure), since there
// is nothing to report in that case.
type IPCMeta struct {
	RequestProtocolVersion  uint32
	RequestSenderVersion    string
	ResponseProtocolVersion uint32
	ResponseSenderVersion   string
	ErrorCode               string
	ErrorMessage            string
}
