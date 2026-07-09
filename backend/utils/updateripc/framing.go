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

// writeEnvelope writes env to w as [4-byte big-endian length][protobuf bytes],
// matching the framing implemented server-side in emly-updater's
// internal/ipc/framing.go.
func writeEnvelope(w io.Writer, env *ipcpb.Envelope) error {
	b, err := proto.Marshal(env)
	if err != nil {
		return fmt.Errorf("updateripc: marshal envelope: %w", err)
	}
	if len(b) > maxFrameSize {
		return fmt.Errorf("updateripc: envelope too large (%d bytes)", len(b))
	}
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(b)))
	if _, err := w.Write(lenBuf[:]); err != nil {
		return fmt.Errorf("updateripc: write length prefix: %w", err)
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("updateripc: write envelope: %w", err)
	}
	return nil
}

// readEnvelope reads one length-prefixed Envelope from r.
func readEnvelope(r io.Reader) (*ipcpb.Envelope, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, fmt.Errorf("updateripc: read length prefix: %w", err)
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	if n > maxFrameSize {
		return nil, fmt.Errorf("updateripc: frame too large (%d bytes)", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("updateripc: read envelope: %w", err)
	}
	env := &ipcpb.Envelope{}
	if err := proto.Unmarshal(buf, env); err != nil {
		return nil, fmt.Errorf("updateripc: unmarshal envelope: %w", err)
	}
	return env, nil
}
