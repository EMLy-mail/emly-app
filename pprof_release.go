//go:build !debug

package main

// livePprofEnabled is false in release builds: no live endpoint, so main
// records the CPU profile to cpu.prof instead.
const livePprofEnabled = false

// startPprofServer is a no-op in release builds: the live pprof endpoint
// is compiled only with -tags debug.
func startPprofServer() {}
