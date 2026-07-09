package updateripc

import (
	"fmt"
	"net"

	"golang.org/x/sys/windows"
)

// fder matches the promoted Fd() method exposed by go-winio's concrete pipe
// connection types. This mirrors emly-updater's internal/ipc/pipeconn.go —
// see that file's doc comment for why this is an implementation detail
// worth pinning the go-winio version over.
type fder interface {
	Fd() uintptr
}

// verifyServerIsSystem confirms the named pipe conn is connected to is owned
// by LocalSystem or BUILTIN\Administrators, independent of the server's own
// DACL and per-connection client authentication. This guards against a
// window where another process squatted the pipe name before the real
// service started, or after it crashed and before it was restarted — EMLy
// should not trust SystemInfo/ADStatus data from a pipe it did not itself
// confirm is admin-owned.
//
// BUILTIN\Administrators is accepted alongside LocalSystem because the
// stated threat model is "tampering requires Administrator", not
// "must literally run as SYSTEM": a manually-elevated (UAC) process's
// default object owner is BUILTIN\Administrators, not SYSTEM, which is the
// case for the installed Windows service too (it always runs as
// LocalSystem) but also for ad-hoc elevated testing/debugging runs.
func verifyServerIsSystem(conn net.Conn) error {
	f, ok := conn.(fder)
	if !ok {
		return fmt.Errorf("updateripc: connection type %T does not expose a raw handle", conn)
	}
	handle := windows.Handle(f.Fd())

	sd, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return fmt.Errorf("updateripc: reading pipe security descriptor: %w", err)
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return fmt.Errorf("updateripc: reading pipe owner: %w", err)
	}
	if !owner.IsWellKnown(windows.WinLocalSystemSid) && !owner.IsWellKnown(windows.WinBuiltinAdministratorsSid) {
		return fmt.Errorf("updateripc: pipe is not owned by LocalSystem or Administrators (owner=%s) — refusing to trust it", owner.String())
	}
	return nil
}
