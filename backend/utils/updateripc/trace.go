package updateripc

import (
	"strings"
	"time"

	pkglogger "emly/backend/logger"

	"emly/backend/utils/updateripc/ipcpb"
)

// Handshake phases, as reported in HandshakeStep.Phase. They group the v2
// exchange's frames into the four logical stages of handshake.go's state
// machine, so a reader can tell at a glance which stage an exchange died in.
const (
	PhaseDial    = "dial"
	PhaseHello   = "hello"
	PhaseVersion = "version"
	PhaseAuth    = "auth"
	PhasePayload = "payload"
)

// Frame directions, as reported in HandshakeStep.Direction.
const (
	DirectionSend = "->" // EMLy to EMLyUpdater
	DirectionRecv = "<-" // EMLyUpdater to EMLy
)

// HandshakeStep is one frame of a single v2 IPC exchange, recorded so the
// handshake's individual round trips (hello, version check, auth challenge,
// payload) are observable from the app log and the frontend DevTools
// console instead of collapsing into one opaque success/failure.
//
// Detail is deliberately a short, non-sensitive summary — it never carries
// sharedSecret, the raw auth nonce, or the computed HMAC, only their sizes.
type HandshakeStep struct {
	// Seq is the 1-based position of this frame within the exchange.
	Seq int
	// Phase is one of the Phase* constants.
	Phase string
	// Direction is one of the Direction* constants.
	Direction string
	// Frame is the FrameType name with its FRAME_TYPE_ prefix stripped.
	Frame string
	// Detail summarizes the frame's contents (versions, sizes, reject
	// reasons). Empty for frames with no meaningful payload.
	Detail string
	// ElapsedMs is milliseconds since the start of this exchange, not since
	// the previous step.
	ElapsedMs int64
	// Error is the failure that ended the exchange at this step, empty on
	// success.
	Error string
}

// tracer accumulates the HandshakeStep trail of one exchange and mirrors
// every step to the app log as it happens, so a hung exchange still leaves
// a partial trail behind even though its IPCMeta.Steps never reaches a
// caller.
type tracer struct {
	start time.Time
	steps []HandshakeStep
}

func newTracer() *tracer { return &tracer{start: time.Now()} }

// record appends one step and logs it. A non-nil err marks the step as the
// one that failed, and is logged at Warn rather than Debug.
func (t *tracer) record(phase, direction string, frame ipcpb.FrameType, detail string, err error) {
	step := HandshakeStep{
		Seq:       len(t.steps) + 1,
		Phase:     phase,
		Direction: direction,
		Frame:     frameName(frame),
		Detail:    detail,
		ElapsedMs: time.Since(t.start).Milliseconds(),
	}
	if err != nil {
		step.Error = err.Error()
	}
	t.steps = append(t.steps, step)

	args := []any{
		"seq", step.Seq,
		"phase", step.Phase,
		"dir", step.Direction,
		"frame", step.Frame,
		"elapsed_ms", step.ElapsedMs,
	}
	if step.Detail != "" {
		args = append(args, "detail", step.Detail)
	}
	if step.Error != "" {
		pkglogger.Warn("updateripc: IPC step failed", append(args, "error", step.Error)...)
		return
	}
	// Info, not Debug: a stock config.ini ships LOG_LEVEL=INFO, and the
	// whole point of the trail is to be readable in a production log
	// without asking the user to re-run under a debug build. The volume
	// stays low — an exchange is ~9 steps and a session runs a handful.
	pkglogger.Info("updateripc: IPC step", args...)
}

// frameName renders a FrameType for humans, dropping the generated
// FRAME_TYPE_ prefix every value carries. UNSPECIFIED is never a real frame
// on the wire (writeFrame refuses it) — it marks the steps that aren't a v2
// frame at all: the dial, and a legacy-shaped pre-handshake rejection.
func frameName(frame ipcpb.FrameType) string {
	if frame == ipcpb.FrameType_FRAME_TYPE_UNSPECIFIED {
		return "-"
	}
	return strings.TrimPrefix(frame.String(), "FRAME_TYPE_")
}
