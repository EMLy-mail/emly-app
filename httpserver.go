// Package main provides HTTP asset-serving helpers for the Wails asset
// server: the SvelteKit SPA fallback and the outgoing User-Agent middleware.
package main

import (
	"bytes"
	"embed"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// spaFallbackHandler serves the embedded index.html for any GET request
// that doesn't match a real asset (e.g. a broken deep link or stale
// hashed-asset reference), so the SvelteKit app still boots and its own
// root +error.svelte renders the 404 — complete with draggable titlebar
// and window controls — instead of WebView2's native error page, which
// has none of that chrome.
func spaFallbackHandler(assets embed.FS) http.Handler {
	fallback, err := assets.ReadFile("frontend/build/index.html")
	if err == nil {
		// Force relative asset URLs (./_app/...) to resolve against the
		// site root regardless of how deep the unmatched path is.
		fallback = bytes.Replace(fallback, []byte("<head>"), []byte("<head><base href=\"/\">"), 1)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(fallback)
	})
}

// userAgentMiddleware returns an AssetServer middleware that sets the
// User-Agent header on every request to the given value.
func userAgentMiddleware(ua string) assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("User-Agent", ua)
			next.ServeHTTP(w, r)
		})
	}
}
