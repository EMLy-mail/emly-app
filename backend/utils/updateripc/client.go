package updateripc

import (
	"context"
	"fmt"
	"time"

	"emly/backend/utils/updateripc/ipcpb"
)

// protocolVersion must match emly-updater's internal/ipc.ProtocolVersion.
// Bump alongside it on any wire-incompatible proto change.
const protocolVersion = 1

// requestTimeout bounds the whole dial+request+response exchange. The
// updater is a local SYSTEM service; a healthy one responds in
// milliseconds, so this is generous only to absorb transient load, not to
// wait out a hung or absent service.
const requestTimeout = 3 * time.Second

// GetSystemInfo asks the EMLyUpdater service for its SYSTEM-authoritative
// view of this machine's identity. This is a distinct source from
// backend/utils.GetExtendedMachineInfo (which runs unprivileged, in this
// process) — callers should not assume the two always agree, and should
// treat a failure here (e.g. the updater service is not installed or
// running) as expected/recoverable, not a hard error.
func GetSystemInfo(ctx context.Context) (SystemInfo, error) {
	req := &ipcpb.Envelope{
		ProtocolVersion: protocolVersion,
		SenderVersion:   version(),
		Body:            &ipcpb.Envelope_SystemInfoRequest{SystemInfoRequest: &ipcpb.SystemInfoRequest{}},
	}
	resp, meta, err := request(ctx, req)
	if err != nil {
		return SystemInfo{Meta: meta}, err
	}
	info := resp.GetSystemInfoResponse()
	if info == nil {
		return SystemInfo{Meta: meta}, fmt.Errorf("updateripc: expected SystemInfoResponse, got %T", resp.GetBody())
	}
	return SystemInfo{
		Hostname:   info.GetHostname(),
		HWID:       info.GetHwid(),
		InternalIP: info.GetInternalIp(),
		OSVersion:  info.GetOsVersion(),
		Meta:       meta,
	}, nil
}

// GetADStatus asks the EMLyUpdater service for its SYSTEM-authoritative AD
// domain status for this machine.
func GetADStatus(ctx context.Context) (ADStatus, error) {
	req := &ipcpb.Envelope{
		ProtocolVersion: protocolVersion,
		SenderVersion:   version(),
		Body:            &ipcpb.Envelope_AdStatusRequest{AdStatusRequest: &ipcpb.ADStatusRequest{}},
	}
	resp, meta, err := request(ctx, req)
	if err != nil {
		return ADStatus{Meta: meta}, err
	}
	status := resp.GetAdStatusResponse()
	if status == nil {
		return ADStatus{Meta: meta}, fmt.Errorf("updateripc: expected ADStatusResponse, got %T", resp.GetBody())
	}
	return ADStatus{
		ADDomain:     status.GetAdDomain(),
		DomainJoined: status.GetDomainJoined(),
		Meta:         meta,
	}, nil
}

// request performs one dial → verify → write → read → close exchange,
// returning the raw response envelope alongside an IPCMeta snapshot of the
// exchange's wire-level diagnostics. meta is always returned (even on
// error), populated as far as the exchange got before failing.
func request(ctx context.Context, req *ipcpb.Envelope) (*ipcpb.Envelope, IPCMeta, error) {
	meta := IPCMeta{
		RequestProtocolVersion: req.GetProtocolVersion(),
		RequestSenderVersion:   req.GetSenderVersion(),
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	conn, err := dial(ctx)
	if err != nil {
		return nil, meta, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if err := writeEnvelope(conn, req); err != nil {
		return nil, meta, err
	}

	resp, err := readEnvelope(conn)
	if err != nil {
		return nil, meta, err
	}
	meta.ResponseProtocolVersion = resp.GetProtocolVersion()
	meta.ResponseSenderVersion = resp.GetSenderVersion()

	if errResp := resp.GetError(); errResp != nil {
		meta.ErrorCode = errResp.GetCode().String()
		meta.ErrorMessage = errResp.GetMessage()
		return resp, meta, fmt.Errorf("updateripc: updater rejected request: %s (%s)", errResp.GetMessage(), errResp.GetCode())
	}
	if err := checkPeerVersion(resp.GetSenderVersion()); err != nil {
		return resp, meta, err
	}
	return resp, meta, nil
}
