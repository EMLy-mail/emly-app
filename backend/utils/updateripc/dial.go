package updateripc

import (
	"context"
	"fmt"
	"net"

	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

// pipeName is the fixed, well-known named pipe the EMLyUpdater service
// listens on by default (config key [ipc] pipeName in that service). No
// per-machine configuration is needed client-side, matching the posture
// already used for the fixed EMLy install path C:\3gIT\EMLy shared between
// the two apps.
const pipeName = `\\.\pipe\EMLyUpdater`

// clientAccessMask is the exact desired-access value the server's pipe DACL
// grants to Authenticated Users (see emly-updater's internal/ipc/sddl.go,
// ClientAccessMask — 0x120083, i.e. READ_CONTROL | SYNCHRONIZE |
// FILE_READ_DATA | FILE_WRITE_DATA | FILE_READ_ATTRIBUTES).
//
// This must be requested explicitly: winio.DialPipe/DialPipeContext default
// to GENERIC_READ|GENERIC_WRITE, which the server's DACL does not grant
// (GENERIC_WRITE on a pipe object implicitly includes FILE_CREATE_PIPE_INSTANCE,
// which the server deliberately withholds from non-admin principals — see
// the server-side doc comment for why). A plain DialPipe call would be
// denied.
const clientAccessMask = windows.FILE_READ_DATA | windows.FILE_WRITE_DATA |
	windows.FILE_READ_ATTRIBUTES | windows.STANDARD_RIGHTS_READ | windows.SYNCHRONIZE

// dial connects to the updater's IPC pipe and verifies, before returning,
// that the pipe is owned by LocalSystem — defense in depth against a
// squatted pipe created by another process before the real service started
// (or after it crashed), independent of and in addition to the server's own
// DACL/client-authentication controls.
func dial(ctx context.Context) (net.Conn, error) {
	conn, err := winio.DialPipeAccess(ctx, pipeName, clientAccessMask)
	if err != nil {
		return nil, fmt.Errorf("updateripc: dialing %s: %w", pipeName, err)
	}
	if err := verifyServerIsSystem(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
