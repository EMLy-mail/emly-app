//go:build !debug

package main

import "os"

// startFileProfiling is a no-op in release builds: no cpu.prof / mem.prof
// is ever written next to the user's working directory. Profiling is
// compiled only with -tags debug (pprof_debug.go).
func startFileProfiling() (cpuProfile *os.File, heapProfilePath string) {
	return nil, ""
}

// startPprofServer is a no-op in release builds: the live pprof endpoint
// is compiled only with -tags debug.
func startPprofServer() {}
