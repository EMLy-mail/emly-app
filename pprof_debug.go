//go:build debug

package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/pprof"
	"time"
)

// pprofAddr is the loopback-only address of the live pprof endpoint.
// Bound to 127.0.0.1 so it is never reachable from the network.
const pprofAddr = "127.0.0.1:6060"

// livePprofEnabled reports whether the live pprof endpoint is compiled in.
// When true, main skips file-based CPU profiling to leave the (single)
// CPU profiler free for /debug/pprof/profile.
const livePprofEnabled = true

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
