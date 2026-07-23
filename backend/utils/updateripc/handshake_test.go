package updateripc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"testing"

	"google.golang.org/protobuf/proto"

	"emly/backend/utils/updateripc/ipcpb"
)

func TestDoRequestV2HappyPath(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	nonce := make([]byte, authNonceSize)
	for i := range nonce {
		nonce[i] = byte(i)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- func() error {
			tag, body, err := readFrame(c2)
			if err != nil {
				return err
			}
			hello := &ipcpb.ClientHello{}
			if uErr := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, hello); uErr != nil {
				return uErr
			}
			if hello.GetProtocolVersion() != protocolVersionV2 {
				return fmt.Errorf("ClientHello.protocol_version = %d, want %d", hello.GetProtocolVersion(), protocolVersionV2)
			}

			if err := writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, &ipcpb.ServerAnswHello{
				ProtocolVersion: protocolVersionV2, ServerVersion: MinCompatibleUpdaterVersionV2,
			}); err != nil {
				return err
			}

			tag, body, err = readFrame(c2)
			if err != nil {
				return err
			}
			if uErr := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND, &ipcpb.ClientSemverSend{}); uErr != nil {
				return uErr
			}
			if err := writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_OK, &ipcpb.ServerSemverOk{}); err != nil {
				return err
			}

			if err := writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_REQUEST_AUTH_CHALLENGE, &ipcpb.ServerRequestAuthChallenge{Nonce: nonce}); err != nil {
				return err
			}

			tag, body, err = readFrame(c2)
			if err != nil {
				return err
			}
			authResp := &ipcpb.ClientAuthResponse{}
			if uErr := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_AUTH_RESPONSE, authResp); uErr != nil {
				return uErr
			}
			mac := hmac.New(sha256.New, sharedSecret)
			mac.Write(nonce)
			if !hmac.Equal(mac.Sum(nil), authResp.GetHmac()) {
				return fmt.Errorf("client sent an incorrect HMAC")
			}

			tag, body, err = readFrame(c2)
			if err != nil {
				return err
			}
			if tag != ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST {
				return fmt.Errorf("payload request tag = %v, want FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST", tag)
			}
			return writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{
				Hostname: "WKS01", Hwid: "hwid-1", InternalIp: "10.0.0.5", OsVersion: "Windows 11 Pro",
			})
		}()
	}()

	resp := &ipcpb.SystemInfoResponse{}
	meta, err := doRequestV2(c1, IPCMeta{RequestProtocolVersion: protocolVersionV2, RequestSenderVersion: "2.1.0"},
		ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, resp)
	if err != nil {
		t.Fatalf("doRequestV2: %v", err)
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("fake server: %v", err)
	}
	if resp.Hostname != "WKS01" || resp.Hwid != "hwid-1" {
		t.Errorf("unexpected SystemInfoResponse: %+v", resp)
	}
	if meta.ResponseProtocolVersion != protocolVersionV2 || meta.ResponseSenderVersion != MinCompatibleUpdaterVersionV2 {
		t.Errorf("unexpected meta: %+v", meta)
	}
}

func TestDoRequestV2HandlesSemverReject(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- func() error {
			if _, _, err := readFrame(c2); err != nil {
				return err
			}
			if err := writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, &ipcpb.ServerAnswHello{
				ProtocolVersion: protocolVersionV2, ServerVersion: MinCompatibleUpdaterVersionV2,
			}); err != nil {
				return err
			}
			if _, _, err := readFrame(c2); err != nil {
				return err
			}
			return writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_REJECT, &ipcpb.ServerSemverReject{Reason: "too old"})
		}()
	}()

	_, err := doRequestV2(c1, IPCMeta{}, ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{})
	if err == nil {
		t.Fatal("expected an error for a semver rejection")
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("fake server: %v", err)
	}
}

func TestDoRequestV2DecodesLegacyRejection(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- func() error {
			// Drain the ClientHello frame. A real server's pre-handshake
			// rejection happens before it ever reads from the client, but
			// net.Pipe requires a matching Read to unblock the client's
			// paired Write regardless of what a real named pipe would do —
			// see writeFrame's doc comment.
			if _, _, err := readFrame(c2); err != nil {
				return err
			}
			env := &ipcpb.Envelope{
				ProtocolVersion: 1,
				Body: &ipcpb.Envelope_Error{
					Error: &ipcpb.ErrorResponse{Code: ipcpb.ErrorCode_ERROR_CODE_UNAUTHORIZED, Message: "unauthorized"},
				},
			}
			b, err := proto.Marshal(env)
			if err != nil {
				return err
			}
			var lenBuf [4]byte
			binary.BigEndian.PutUint32(lenBuf[:], uint32(len(b)))
			if _, err := c2.Write(lenBuf[:]); err != nil {
				return err
			}
			_, err = c2.Write(b)
			return err
		}()
	}()

	_, err := doRequestV2(c1, IPCMeta{}, ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{})
	if err == nil {
		t.Fatal("expected an error for a legacy pre-handshake rejection")
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("fake server: %v", err)
	}
}
