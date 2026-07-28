package updateripc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net"

	"google.golang.org/protobuf/proto"

	"emly/backend/utils/updateripc/ipcpb"
)

// protocolVersionV2 must match emly-updater's internal/ipc.ProtocolVersion.
// Bump alongside it on any wire-incompatible proto change.
const protocolVersionV2 = 2

// authNonceSize is the length, in bytes, of the random nonce the server
// sends in ServerRequestAuthChallenge, which this client HMACs (with
// sharedSecret) and returns in ClientAuthResponse.
const authNonceSize = 32

// requestV2 performs one full v2 exchange over a fresh connection:
//
//	ClientHello -> ServerAnswHello
//	ClientSemverSend -> ServerSemverOk | ServerSemverReject
//	ServerRequestAuthChallenge -> ClientAuthResponse
//	reqTag/reqMsg -> wantRespTag/respMsg
//
// It returns IPCMeta populated as far as the exchange got, even on failure,
// mirroring the pre-handshake request() helper's contract.
func requestV2(ctx context.Context, reqTag ipcpb.FrameType, reqMsg proto.Message, wantRespTag ipcpb.FrameType, respMsg proto.Message) (IPCMeta, error) {
	meta := IPCMeta{
		RequestProtocolVersion: protocolVersionV2,
		RequestSenderVersion:   version(),
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// One tracer spans dial and exchange, so the trail's ElapsedMs is
	// measured from the same origin the caller's requestTimeout is.
	tr := newTracer()

	conn, err := dial(ctx)
	if err != nil {
		tr.record(PhaseDial, DirectionSend, ipcpb.FrameType_FRAME_TYPE_UNSPECIFIED, pipeName, err)
		meta.Steps = tr.steps
		return meta, err
	}
	defer conn.Close()
	tr.record(PhaseDial, DirectionSend, ipcpb.FrameType_FRAME_TYPE_UNSPECIFIED,
		fmt.Sprintf("connected to %s, peer verified as SYSTEM", pipeName), nil)

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	meta, err = exchangeV2(conn, meta, tr, reqTag, reqMsg, wantRespTag, respMsg)
	meta.Steps = tr.steps
	return meta, err
}

// doRequestV2 drives the v2 handshake and payload exchange over an
// already-dialed, already-verified conn, continuing from meta (whose
// Request* fields requestV2 has already populated). Split out from
// requestV2 so the handshake state machine can be exercised directly over
// net.Pipe() in tests, bypassing dial's real Windows named-pipe APIs
// (winio.DialPipeAccess, verifyServerIsSystem) the same way emly-updater's
// handleHandshake is tested independently of handleConn's real pipe Accept.
func doRequestV2(conn net.Conn, meta IPCMeta, reqTag ipcpb.FrameType, reqMsg proto.Message, wantRespTag ipcpb.FrameType, respMsg proto.Message) (IPCMeta, error) {
	tr := newTracer()
	meta, err := exchangeV2(conn, meta, tr, reqTag, reqMsg, wantRespTag, respMsg)
	// Attached on every path, success or failure: a truncated trail is
	// exactly what makes a failed exchange diagnosable.
	meta.Steps = tr.steps
	return meta, err
}

// exchangeV2 is doRequestV2's state machine proper, recording every frame it
// sends or receives into tr. Split from doRequestV2 only so the trail can be
// attached to meta once, at a single return point, instead of at each of the
// dozen-odd early returns below.
func exchangeV2(conn net.Conn, meta IPCMeta, tr *tracer, reqTag ipcpb.FrameType, reqMsg proto.Message, wantRespTag ipcpb.FrameType, respMsg proto.Message) (IPCMeta, error) {
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, &ipcpb.ClientHello{ProtocolVersion: protocolVersionV2}); err != nil {
		tr.record(PhaseHello, DirectionSend, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, "", err)
		return meta, err
	}
	tr.record(PhaseHello, DirectionSend, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO,
		fmt.Sprintf("protocol_version=%d", protocolVersionV2), nil)

	tag, body, legacyErr, err := readFirstFrame(conn)
	if legacyErr != nil {
		meta.ErrorCode = legacyErr.GetCode().String()
		meta.ErrorMessage = legacyErr.GetMessage()
		rejErr := fmt.Errorf("updateripc: updater rejected connection: %s (%s)", legacyErr.GetMessage(), legacyErr.GetCode())
		tr.record(PhaseHello, DirectionRecv, ipcpb.FrameType_FRAME_TYPE_UNSPECIFIED,
			fmt.Sprintf("legacy ErrorResponse code=%s", legacyErr.GetCode()), rejErr)
		return meta, rejErr
	}
	hello := &ipcpb.ServerAnswHello{}
	if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, hello); uErr != nil {
		tr.record(PhaseHello, DirectionRecv, tag, "", uErr)
		return meta, uErr
	}
	meta.ResponseProtocolVersion = hello.GetProtocolVersion()
	meta.ResponseSenderVersion = hello.GetServerVersion()
	helloDetail := fmt.Sprintf("protocol_version=%d server_version=%s", hello.GetProtocolVersion(), hello.GetServerVersion())
	if hello.GetProtocolVersion() != protocolVersionV2 {
		verErr := fmt.Errorf("updateripc: updater answered with unsupported protocol version %d", hello.GetProtocolVersion())
		tr.record(PhaseHello, DirectionRecv, tag, helloDetail, verErr)
		return meta, verErr
	}
	if err := checkPeerVersionV2(hello.GetServerVersion()); err != nil {
		tr.record(PhaseHello, DirectionRecv, tag, helloDetail, err)
		return meta, err
	}
	tr.record(PhaseHello, DirectionRecv, tag, helloDetail, nil)

	clientVersion := version()
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND, &ipcpb.ClientSemverSend{ClientVersion: clientVersion}); err != nil {
		tr.record(PhaseVersion, DirectionSend, ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND, "", err)
		return meta, err
	}
	tr.record(PhaseVersion, DirectionSend, ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND,
		fmt.Sprintf("client_version=%s", clientVersion), nil)

	tag, body, err = readFrame(conn)
	if tag == ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_REJECT {
		reject := &ipcpb.ServerSemverReject{}
		if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_REJECT, reject); uErr != nil {
			tr.record(PhaseVersion, DirectionRecv, tag, "", uErr)
			return meta, uErr
		}
		meta.ErrorCode = ipcpb.ErrorCode_ERROR_CODE_UNSUPPORTED_VERSION.String()
		meta.ErrorMessage = reject.GetReason()
		rejErr := fmt.Errorf("updateripc: updater rejected our version: %s", reject.GetReason())
		tr.record(PhaseVersion, DirectionRecv, tag, fmt.Sprintf("reason=%s", reject.GetReason()), rejErr)
		return meta, rejErr
	}
	if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_OK, &ipcpb.ServerSemverOk{}); uErr != nil {
		tr.record(PhaseVersion, DirectionRecv, tag, "", uErr)
		return meta, uErr
	}
	tr.record(PhaseVersion, DirectionRecv, tag, "", nil)

	tag, body, err = readFrame(conn)
	challenge := &ipcpb.ServerRequestAuthChallenge{}
	if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_REQUEST_AUTH_CHALLENGE, challenge); uErr != nil {
		tr.record(PhaseAuth, DirectionRecv, tag, "", uErr)
		return meta, uErr
	}
	// Sizes only, never the nonce itself and never the HMAC — a trace that
	// leaks either turns the app log and the DevTools console into material
	// for forging a ClientAuthResponse.
	nonceDetail := fmt.Sprintf("nonce=%d bytes", len(challenge.GetNonce()))
	if len(challenge.GetNonce()) != authNonceSize {
		nonceDetail += fmt.Sprintf(" (expected %d)", authNonceSize)
	}
	tr.record(PhaseAuth, DirectionRecv, tag, nonceDetail, nil)

	mac := hmac.New(sha256.New, sharedSecret)
	mac.Write(challenge.GetNonce())
	sum := mac.Sum(nil)
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_CLIENT_AUTH_RESPONSE, &ipcpb.ClientAuthResponse{Hmac: sum}); err != nil {
		tr.record(PhaseAuth, DirectionSend, ipcpb.FrameType_FRAME_TYPE_CLIENT_AUTH_RESPONSE, "", err)
		return meta, err
	}
	tr.record(PhaseAuth, DirectionSend, ipcpb.FrameType_FRAME_TYPE_CLIENT_AUTH_RESPONSE,
		fmt.Sprintf("hmac-sha256=%d bytes", len(sum)), nil)

	if err := writeFrame(conn, reqTag, reqMsg); err != nil {
		tr.record(PhasePayload, DirectionSend, reqTag, "", err)
		return meta, err
	}
	tr.record(PhasePayload, DirectionSend, reqTag, "", nil)

	tag, body, err = readFrame(conn)
	if tag == ipcpb.FrameType_FRAME_TYPE_SERVER_ERROR {
		errResp := &ipcpb.ErrorResponse{}
		if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_ERROR, errResp); uErr != nil {
			tr.record(PhasePayload, DirectionRecv, tag, "", uErr)
			return meta, uErr
		}
		meta.ErrorCode = errResp.GetCode().String()
		meta.ErrorMessage = errResp.GetMessage()
		rejErr := fmt.Errorf("updateripc: updater rejected request: %s (%s)", errResp.GetMessage(), errResp.GetCode())
		tr.record(PhasePayload, DirectionRecv, tag, fmt.Sprintf("code=%s", errResp.GetCode()), rejErr)
		return meta, rejErr
	}
	if uErr := unmarshalExpected(tag, body, err, wantRespTag, respMsg); uErr != nil {
		tr.record(PhasePayload, DirectionRecv, tag, "", uErr)
		return meta, uErr
	}
	tr.record(PhasePayload, DirectionRecv, tag, fmt.Sprintf("body=%d bytes", len(body)), nil)
	return meta, nil
}
