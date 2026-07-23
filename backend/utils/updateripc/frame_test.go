package updateripc

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"google.golang.org/protobuf/proto"

	"emly/backend/utils/updateripc/ipcpb"
)

func TestFrameRoundTrip(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	want := &ipcpb.ClientHello{ProtocolVersion: protocolVersionV2}

	errCh := make(chan error, 1)
	go func() { errCh <- writeFrame(c1, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, want) }()

	tag, body, err := readFrame(c2)
	if err != nil {
		t.Fatalf("readFrame: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("writeFrame: %v", err)
	}
	if tag != ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO {
		t.Errorf("tag = %v, want FRAME_TYPE_CLIENT_HELLO", tag)
	}

	got := &ipcpb.ClientHello{}
	if err := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_CLIENT_HELLO, got); err != nil {
		t.Fatalf("unmarshalExpected: %v", err)
	}
	if got.GetProtocolVersion() != protocolVersionV2 {
		t.Errorf("round-tripped frame mismatch: %+v", got)
	}
}

func TestWriteFrameRejectsReservedTag(t *testing.T) {
	if err := writeFrame(new(bytes.Buffer), ipcpb.FrameType_FRAME_TYPE_UNSPECIFIED, &ipcpb.ClientHello{}); err == nil {
		t.Fatal("expected error writing FRAME_TYPE_UNSPECIFIED")
	}
}

func TestWriteFrameRejectsOversizedMessage(t *testing.T) {
	msg := &ipcpb.ClientSemverSend{ClientVersion: string(make([]byte, maxFrameSize+1))}
	if err := writeFrame(new(bytes.Buffer), ipcpb.FrameType_FRAME_TYPE_CLIENT_SEMVER_SEND, msg); err == nil {
		t.Fatal("expected error for a frame larger than maxFrameSize")
	}
}

func TestReadFrameRejectsOversizedFrame(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteByte(byte(ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO))
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], maxFrameSize+1)
	buf.Write(lenBuf[:])

	if _, _, err := readFrame(&buf); err == nil {
		t.Fatal("expected error for a frame larger than maxFrameSize")
	}
}

func TestReadFrameRejectsTruncatedFrame(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteByte(byte(ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO))
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], 10)
	buf.Write(lenBuf[:])
	buf.WriteString("short")

	if _, _, err := readFrame(&buf); err == nil {
		t.Fatal("expected error for a truncated frame body")
	}
}

func TestUnmarshalExpectedRejectsWrongTag(t *testing.T) {
	err := unmarshalExpected(ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_REJECT, []byte{}, nil,
		ipcpb.FrameType_FRAME_TYPE_SERVER_SEMVER_OK, &ipcpb.ServerSemverOk{})
	if err == nil {
		t.Fatal("expected error for a frame type mismatch")
	}
}

func TestReadFirstFrameDecodesLegacyRejection(t *testing.T) {
	var buf bytes.Buffer
	env := &ipcpb.Envelope{
		ProtocolVersion: 1,
		Body: &ipcpb.Envelope_Error{
			Error: &ipcpb.ErrorResponse{Code: ipcpb.ErrorCode_ERROR_CODE_UNAUTHORIZED, Message: "unauthorized"},
		},
	}
	b, err := proto.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(b)))
	buf.Write(lenBuf[:]) // first byte is 0x00, the legacy discriminator
	buf.Write(b)

	tag, body, legacyErr, err := readFirstFrame(&buf)
	if err != nil {
		t.Fatalf("readFirstFrame: %v", err)
	}
	if legacyErr == nil {
		t.Fatal("expected a decoded legacy ErrorResponse")
	}
	if legacyErr.Code != ipcpb.ErrorCode_ERROR_CODE_UNAUTHORIZED || legacyErr.Message != "unauthorized" {
		t.Errorf("unexpected legacy ErrorResponse: %+v", legacyErr)
	}
	if tag != 0 || body != nil {
		t.Errorf("expected zero tag/nil body alongside a legacy rejection, got tag=%v body=%v", tag, body)
	}
}

func TestReadFirstFrameReadsV2Frame(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- writeFrame(c1, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, &ipcpb.ServerAnswHello{ProtocolVersion: protocolVersionV2, ServerVersion: "1.3.0"})
	}()

	tag, body, legacyErr, err := readFirstFrame(c2)
	if err != nil {
		t.Fatalf("readFirstFrame: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("writeFrame: %v", err)
	}
	if legacyErr != nil {
		t.Fatalf("expected no legacy rejection, got %+v", legacyErr)
	}
	if tag != ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO {
		t.Fatalf("tag = %v, want FRAME_TYPE_SERVER_ANSW_HELLO", tag)
	}
	hello := &ipcpb.ServerAnswHello{}
	if err := unmarshalExpected(tag, body, nil, ipcpb.FrameType_FRAME_TYPE_SERVER_ANSW_HELLO, hello); err != nil {
		t.Fatalf("unmarshalExpected: %v", err)
	}
	if hello.GetServerVersion() != "1.3.0" {
		t.Errorf("unexpected ServerAnswHello: %+v", hello)
	}
}
