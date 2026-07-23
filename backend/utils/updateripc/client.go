package updateripc

import (
	"context"
	"time"

	"emly/backend/utils/updateripc/ipcpb"
)

// requestTimeout bounds the whole dial+handshake+request+response exchange.
// The updater is a local SYSTEM service; a healthy one responds in
// milliseconds per round trip, so this is generous only to absorb transient
// load, not to wait out a hung or absent service. The v2 handshake is four
// sequential round trips instead of v1's one, hence the higher ceiling than
// the previous 3s.
const requestTimeout = 6 * time.Second

// GetSystemInfo asks the EMLyUpdater service for its SYSTEM-authoritative
// view of this machine's identity. This is a distinct source from
// backend/utils.GetExtendedMachineInfo (which runs unprivileged, in this
// process) — callers should not assume the two always agree, and should
// treat a failure here (e.g. the updater service is not installed or
// running) as expected/recoverable, not a hard error.
func GetSystemInfo(ctx context.Context) (SystemInfo, error) {
	resp := &ipcpb.SystemInfoResponse{}
	meta, err := requestV2(ctx,
		ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, resp)
	if err != nil {
		return SystemInfo{Meta: meta}, err
	}
	return SystemInfo{
		Hostname:   resp.GetHostname(),
		HWID:       resp.GetHwid(),
		InternalIP: resp.GetInternalIp(),
		OSVersion:  resp.GetOsVersion(),
		Meta:       meta,
	}, nil
}

// GetADStatus asks the EMLyUpdater service for its SYSTEM-authoritative AD
// domain status for this machine.
func GetADStatus(ctx context.Context) (ADStatus, error) {
	resp := &ipcpb.ADStatusResponse{}
	meta, err := requestV2(ctx,
		ipcpb.FrameType_FRAME_TYPE_CLIENT_AD_STATUS_REQUEST, &ipcpb.ADStatusRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_AD_STATUS_RESPONSE, resp)
	if err != nil {
		return ADStatus{Meta: meta}, err
	}
	return ADStatus{
		ADDomain:     resp.GetAdDomain(),
		DomainJoined: resp.GetDomainJoined(),
		Meta:         meta,
	}, nil
}
