//go:build debug

package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	runtimepprof "runtime/pprof"
	"time"
)

// pprofAddr is the loopback-only address of the live pprof endpoint.
// Bound to 127.0.0.1 so it is never reachable from the network.
const pprofAddr = "127.0.0.1:6060"

// Profile files written in the working directory by debug builds: the CPU
// profile covers the whole run, the heap profile is written on exit (see
// App.stopProfiling). Both are gitignored (*.prof).
const (
	cpuProfileFile  = "cpu.prof"
	heapProfileFile = "mem.prof"
)

// liveCPUProfileEnv, set to 1, skips the cpu.prof recording: Go allows a
// single CPU profile at a time, so while cpu.prof is being recorded
// /debug/pprof/profile fails with "cpu profiling already in use". The other
// live endpoints (heap, goroutine, ...) work either way.
const liveCPUProfileEnv = "EMLY_PPROF_LIVE_CPU"

// startFileProfiling starts recording cpu.prof (unless EMLY_PPROF_LIVE_CPU=1)
// and returns the open file plus the mem.prof path for App.stopProfiling.
// Compiled only into debug builds (-tags debug).
func startFileProfiling() (cpuProfile *os.File, heapProfilePath string) {
	heapProfilePath = heapProfileFile

	if os.Getenv(liveCPUProfileEnv) == "1" {
		log.Println("cpu.prof disabled (" + liveCPUProfileEnv + "=1): CPU profile via http://" + pprofAddr + "/debug/pprof/profile")
		return nil, heapProfilePath
	}

	f, err := os.Create(cpuProfileFile)
	if err != nil {
		log.Println("Error creating CPU profile file:", err)
		return nil, heapProfilePath
	}
	if err := runtimepprof.StartCPUProfile(f); err != nil {
		log.Println("Error starting CPU profiling:", err)
		f.Close()
		return nil, heapProfilePath
	}
	return f, heapProfilePath
}

// startPprofServer exposes the net/http/pprof handlers on a dedicated mux
// (not http.DefaultServeMux) at http://127.0.0.1:6060/debug/pprof/.
// Compiled only into debug builds (-tags debug).
func startPprofServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              pprofAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Println("pprof endpoint listening on http://" + pprofAddr + "/debug/pprof/")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("pprof endpoint error:", err)
		}
	}()
}
