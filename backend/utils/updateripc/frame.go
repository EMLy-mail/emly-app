package updateripc

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"

	"emly/backend/utils/updateripc/ipcpb"
)

// maxFrameSize mirrors emly-updater's internal/ipc.MaxFrameSize. Independent
// definition (no shared module between the two repos) — keep in sync if
// either side changes it.
const maxFrameSize = 64 * 1024

// writeFrame writes a v2 frame to w as
// [1-byte FrameType tag][4-byte big-endian length][protobuf bytes], matching
// the framing implemented server-side in emly-updater's internal/ipc/frame.go.
func writeFrame(w io.Writer, tag ipcpb.FrameType, msg proto.Message) error {
	if tag == ipcpb.FrameType_FRAME_TYPE_UNSPECIFIED {
		return fmt.Errorf("updateripc: refusing to write reserved FrameType_FRAME_TYPE_UNSPECIFIED")
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("updateripc: marshal %T: %w", msg, err)
	}
	if len(b) > maxFrameSize {
		return fmt.Errorf("updateripc: frame too large (%d bytes)", len(b))
	}
	var head [5]byte
	head[0] = byte(tag)
	binary.BigEndian.PutUint32(head[1:], uint32(len(b)))
	if _, err := w.Write(head[:]); err != nil {
		return fmt.Errorf("updateripc: write frame header: %w", err)
	}
	if len(b) == 0 {
		// Several v2 messages (ServerSemverOk, SystemInfoRequest, ...)
		// marshal to zero bytes; skip the body write entirely — a
		// zero-length Write still requires a rendezvous with a matching
		// Read on some io.ReadWriter implementations (notably net.Pipe,
		// used by this package's tests), while a zero-length Read never
		// issues that Read — see the mirrored comment in emly-updater's
		// internal/ipc/frame.go.
		return nil
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("updateripc: write frame body: %w", err)
	}
	return nil
}

// readFrame reads one tag-prefixed v2 frame from r, returning its FrameType
// and raw (still-marshaled) body.
func readFrame(r io.Reader) (ipcpb.FrameType, []byte, error) {
	var tagBuf [1]byte
	if _, err := io.ReadFull(r, tagBuf[:]); err != nil {
		return 0, nil, fmt.Errorf("updateripc: read frame tag: %w", err)
	}
	return readFrameAfterTag(r, tagBuf[0])
}

// readFrameAfterTag reads the remainder of a v2 frame whose tag byte
// (tagByte) has already been read off the wire.
func readFrameAfterTag(r io.Reader, tagByte byte) (ipcpb.FrameType, []byte, error) {
	tag := ipcpb.FrameType(tagByte)
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return tag, nil, fmt.Errorf("updateripc: read frame length: %w", err)
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	if n > maxFrameSize {
		return tag, nil, fmt.Errorf("updateripc: frame too large (%d bytes)", n)
	}
	if n == 0 {
		return tag, []byte{}, nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return tag, nil, fmt.Errorf("updateripc: read frame body: %w", err)
	}
	return tag, buf, nil
}

// unmarshalExpected checks a (tag, body, err) triple as returned by
// readFrame/readFrameAfterTag against the FrameType expected at the current
// handshake step, then unmarshals body into out.
func unmarshalExpected(tag ipcpb.FrameType, body []byte, readErr error, want ipcpb.FrameType, out proto.Message) error {
	if readErr != nil {
		return readErr
	}
	if tag != want {
		return fmt.Errorf("updateripc: expected frame type %s, got %s", want, tag)
	}
	return proto.Unmarshal(body, out)
}

// readFirstFrame reads the very first frame of a v2 exchange, handling the
// one documented exception in this protocol: a pre-handshake auth rejection
// (the server's verifyClient failing before any wire read) is always sent
// in the frozen legacy Envelope/ErrorResponse shape — 0x00 as the first
// byte, then a 4-byte length prefix — since the server doesn't yet know
// which dialect the peer speaks. Every other read in this package uses
// plain readFrame; only the first read after dialing needs to check for
// this legacy shape.
func readFirstFrame(r io.Reader) (tag ipcpb.FrameType, body []byte, legacyErr *ipcpb.ErrorResponse, err error) {
	var first [1]byte
	if _, err := io.ReadFull(r, first[:]); err != nil {
		return 0, nil, nil, fmt.Errorf("updateripc: read first frame byte: %w", err)
	}
	if first[0] == 0 {
		env, err := readLegacyEnvelopeAfterFirstByte(r)
		if err != nil {
			return 0, nil, nil, err
		}
		errResp := env.GetError()
		if errResp == nil {
			return 0, nil, nil, fmt.Errorf("updateripc: legacy-shaped response was not an ErrorResponse (%T)", env.GetBody())
		}
		return 0, nil, errResp, nil
	}
	tag, body, err = readFrameAfterTag(r, first[0])
	return tag, body, nil, err
}

// readLegacyEnvelopeAfterFirstByte decodes a legacy (protocol_version 1)
// Envelope whose first length-prefix byte (always 0x00) has already been
// read — the only legacy wire shape this client ever needs to understand.
func readLegacyEnvelopeAfterFirstByte(r io.Reader) (*ipcpb.Envelope, error) {
	var lenBuf [4]byte // lenBuf[0] is already known to be 0
	if _, err := io.ReadFull(r, lenBuf[1:]); err != nil {
		return nil, fmt.Errorf("updateripc: read legacy length prefix: %w", err)
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	if n > maxFrameSize {
		return nil, fmt.Errorf("updateripc: legacy frame too large (%d bytes)", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("updateripc: read legacy envelope: %w", err)
	}
	env := &ipcpb.Envelope{}
	if err := proto.Unmarshal(buf, env); err != nil {
		return nil, fmt.Errorf("updateripc: unmarshal legacy envelope: %w", err)
	}
	return env, nil
}
