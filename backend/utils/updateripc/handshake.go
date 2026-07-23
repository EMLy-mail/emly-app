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

	conn, err := dial(ctx)
	if err != nil {
		return meta, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	return doRequestV2(conn, meta, reqTag, reqMsg, wantRespTag, respMsg)
}

// doRequestV2 drives the v2 handshake and payload exchange over an
// already-dialed, already-verified conn, continuing from meta (whose
// Request* fields requestV2 has already populated). Split out from
// requestV2 so the handshake state machine can be exercised directly over
// net.Pipe() in tests, bypassing dial's real Windows named-pipe APIs
// (winio.DialPipeAccess, verifyServerIsSystem) the same way emly-updater's
// handleHandshake is tested independently of handleConn's real pipe Accept.
func doRequestV2(conn net.Conn, meta IPCMeta, reqTag ipcpb.FrameType, reqMsg proto.Message, wantRespTag ipcpb.FrameType, respMsg proto.Message) (IPCMeta, error) {
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, &ipcpb.ClientHello{ProtocolVersion: protocolVersionV2}); err != nil {
		return meta, err
	}

	tag, body, legacyErr, err := readFirstFrame(conn)
	if legacyErr != nil {
		meta.ErrorCode = legacyErr.GetCode().String()
		meta.ErrorMessage = legacyErr.GetMessage()
		return meta, fmt.Errorf("updateripc: updater rejected connection: %s (%s)", legacyErr.GetMessage(), legacyErr.GetCode())
	}
	hello := &ipcpb.ServerAnswHello{}
	if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, hello); uErr != nil {
		return meta, uErr
	}
	meta.ResponseProtocolVersion = hello.GetProtocolVersion()
	meta.ResponseSenderVersion = hello.GetServerVersion()
	if hello.GetProtocolVersion() != protocolVersionV2 {
		return meta, fmt.Errorf("updateripc: updater answered with unsupported protocol version %d", hello.GetProtocolVersion())
	}
	if err := checkPeerVersionV2(hello.GetServerVersion()); err != nil {
		return meta, err
	}

	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND, &ipcpb.ClientSemverSend{ClientVersion: version()}); err != nil {
		return meta, err
	}
	tag, body, err = readFrame(conn)
	if tag == ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_REJECT {
		reject := &ipcpb.ServerSemverReject{}
		if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_REJECT, reject); uErr != nil {
			return meta, uErr
		}
		meta.ErrorCode = ipcpb.ErrorCode_ERROR_CODE_UNSUPPORTED_VERSION.String()
		meta.ErrorMessage = reject.GetReason()
		return meta, fmt.Errorf("updateripc: updater rejected our version: %s", reject.GetReason())
	}
	if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_OK, &ipcpb.ServerSemverOk{}); uErr != nil {
		return meta, uErr
	}

	tag, body, err = readFrame(conn)
	challenge := &ipcpb.ServerRequestAuthChallenge{}
	if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_REQUEST_AUTH_CHALLENGE, challenge); uErr != nil {
		return meta, uErr
	}

	mac := hmac.New(sha256.New, sharedSecret)
	mac.Write(challenge.GetNonce())
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_CLIENT_AUTH_RESPONSE, &ipcpb.ClientAuthResponse{Hmac: mac.Sum(nil)}); err != nil {
		return meta, err
	}

	if err := writeFrame(conn, reqTag, reqMsg); err != nil {
		return meta, err
	}

	tag, body, err = readFrame(conn)
	if tag == ipcpb.FrameType_FRAME_TYPE_SERVER_ERROR {
		errResp := &ipcpb.ErrorResponse{}
		if uErr := unmarshalExpected(tag, body, err, ipcpb.FrameType_FRAME_TYPE_SERVER_ERROR, errResp); uErr != nil {
			return meta, uErr
		}
		meta.ErrorCode = errResp.GetCode().String()
		meta.ErrorMessage = errResp.GetMessage()
		return meta, fmt.Errorf("updateripc: updater rejected request: %s (%s)", errResp.GetMessage(), errResp.GetCode())
	}
	if uErr := unmarshalExpected(tag, body, err, wantRespTag, respMsg); uErr != nil {
		return meta, uErr
	}
	return meta, nil
}
