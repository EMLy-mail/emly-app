package utils

import "syscall"

// IsDebugged reports whether a debugger is attached (Windows).
func IsDebugged() bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")
	ret, _, _ := isDebuggerPresent.Call()
	return ret != 0
}
