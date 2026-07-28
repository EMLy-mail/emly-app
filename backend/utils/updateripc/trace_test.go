package updateripc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"testing"

	"emly/backend/utils/updateripc/ipcpb"
)

// traceTestNonce is the fixed auth nonce the fake servers in this file send,
// so a leak test can look for its exact bytes in the recorded trail.
var traceTestNonce = func() []byte {
	nonce := make([]byte, authNonceSize)
	for i := range nonce {
		nonce[i] = byte(0xA0 + i%16)
	}
	return nonce
}()

// serveV2HappyPath plays a well-behaved EMLyUpdater through one full v2
// exchange on conn, answering a SystemInfo request.
func serveV2HappyPath(conn net.Conn) error {
	tag, body, err := readFrame(conn)
	if err != nil {
		return err
	}
	if uErr := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, &ipcpb.ClientHello{}); uErr != nil {
		return uErr
	}
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, &ipcpb.ServerAnswHello{
		ProtocolVersion: protocolVersionV2, ServerVersion: MinCompatibleUpdaterVersionV2,
	}); err != nil {
		return err
	}

	tag, body, err = readFrame(conn)
	if err != nil {
		return err
	}
	if uErr := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND, &ipcpb.ClientSemverSend{}); uErr != nil {
		return uErr
	}
	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_OK, &ipcpb.ServerSemverOk{}); err != nil {
		return err
	}

	if err := writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_SERVER_REQUEST_AUTH_CHALLENGE, &ipcpb.ServerRequestAuthChallenge{Nonce: traceTestNonce}); err != nil {
		return err
	}
	tag, body, err = readFrame(conn)
	if err != nil {
		return err
	}
	authResp := &ipcpb.ClientAuthResponse{}
	if uErr := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_AUTH_RESPONSE, authResp); uErr != nil {
		return uErr
	}
	mac := hmac.New(sha256.New, sharedSecret)
	mac.Write(traceTestNonce)
	if !hmac.Equal(mac.Sum(nil), authResp.GetHmac()) {
		return fmt.Errorf("client sent an incorrect HMAC")
	}

	if _, _, err := readFrame(conn); err != nil {
		return err
	}
	return writeFrame(conn, ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{
		Hostname: "WKS01", Hwid: "hwid-1",
	})
}

func TestDoRequestV2RecordsEveryHandshakeStep(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	serverErrCh := make(chan error, 1)
	go func() { serverErrCh <- serveV2HappyPath(c2) }()

	meta, err := doRequestV2(c1, IPCMeta{},
		ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{})
	if err != nil {
		t.Fatalf("doRequestV2: %v", err)
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("fake server: %v", err)
	}

	want := []struct {
		phase     string
		direction string
		frame     string
	}{
		{PhaseHello, DirectionSend, "CLIENT_HELLO"},
		{PhaseHello, DirectionRecv, "SERVER_ANSW_HELLO"},
		{PhaseVersion, DirectionSend, "CLIENT_SEMVER_SEND"},
		{PhaseVersion, DirectionRecv, "SERVER_SEMVER_OK"},
		{PhaseAuth, DirectionRecv, "SERVER_REQUEST_AUTH_CHALLENGE"},
		{PhaseAuth, DirectionSend, "CLIENT_AUTH_RESPONSE"},
		{PhasePayload, DirectionSend, "CLIENT_SYSTEM_INFO_REQUEST"},
		{PhasePayload, DirectionRecv, "SERVER_SYSTEM_INFO_RESPONSE"},
	}
	if len(meta.Steps) != len(want) {
		t.Fatalf("recorded %d steps, want %d: %+v", len(meta.Steps), len(want), meta.Steps)
	}
	for i, w := range want {
		got := meta.Steps[i]
		if got.Seq != i+1 || got.Phase != w.phase || got.Direction != w.direction || got.Frame != w.frame {
			t.Errorf("step %d = %+v, want seq=%d phase=%s dir=%s frame=%s", i+1, got, i+1, w.phase, w.direction, w.frame)
		}
		if got.Error != "" {
			t.Errorf("step %d recorded an error on a successful exchange: %s", i+1, got.Error)
		}
	}
}

// The trail reaches both the app log and the DevTools console, so a step
// that echoed the shared secret, the auth nonce or the computed HMAC would
// hand an attacker with read access to either everything needed to forge a
// ClientAuthResponse.
func TestHandshakeTraceNeverLeaksAuthMaterial(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	serverErrCh := make(chan error, 1)
	go func() { serverErrCh <- serveV2HappyPath(c2) }()

	meta, err := doRequestV2(c1, IPCMeta{},
		ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{})
	if err != nil {
		t.Fatalf("doRequestV2: %v", err)
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("fake server: %v", err)
	}

	mac := hmac.New(sha256.New, sharedSecret)
	mac.Write(traceTestNonce)
	forbidden := map[string]string{
		"shared secret (raw)": string(sharedSecret),
		"shared secret (hex)": hex.EncodeToString(sharedSecret),
		"auth nonce (raw)":    string(traceTestNonce),
		"auth nonce (hex)":    hex.EncodeToString(traceTestNonce),
		"auth HMAC (raw)":     string(mac.Sum(nil)),
		"auth HMAC (hex)":     hex.EncodeToString(mac.Sum(nil)),
	}
	for _, step := range meta.Steps {
		text := step.Detail + " " + step.Error
		for name, secret := range forbidden {
			if strings.Contains(text, secret) {
				t.Errorf("step %d (%s) leaked the %s: %q", step.Seq, step.Frame, name, text)
			}
		}
	}
}

func TestHandshakeTraceTruncatesAtFailure(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- func() error {
			if _, _, err := readFrame(c2); err != nil {
				return err
			}
			// Answer the ClientHello with a protocol version this client
			// refuses, ending the exchange at step 2 of 8.
			return writeFrame(c2, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, &ipcpb.ServerAnswHello{
				ProtocolVersion: 99, ServerVersion: MinCompatibleUpdaterVersionV2,
			})
		}()
	}()

	meta, err := doRequestV2(c1, IPCMeta{},
		ipcpb.FrameType_FRAME_TYPE_CLIENT_SYSTEM_INFO_REQUEST, &ipcpb.SystemInfoRequest{},
		ipcpb.FrameType_FRAME_TYPE_SERVER_SYSTEM_INFO_RESPONSE, &ipcpb.SystemInfoResponse{})
	if err == nil {
		t.Fatal("doRequestV2 accepted an unsupported protocol version")
	}
	if err := <-serverErrCh; err != nil {
		t.Fatalf("fake server: %v", err)
	}

	if len(meta.Steps) != 2 {
		t.Fatalf("recorded %d steps, want the trail truncated at 2: %+v", len(meta.Steps), meta.Steps)
	}
	last := meta.Steps[1]
	if last.Frame != "SERVER_ANSW_HELLO" || last.Error == "" {
		t.Errorf("last step = %+v, want SERVER_ANSW_HELLO carrying the failure", last)
	}
	if !strings.Contains(last.Detail, "protocol_version=99") {
		t.Errorf("last step detail = %q, want it to name the rejected version", last.Detail)
	}
}
