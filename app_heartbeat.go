// Package main provides heartbeat checking for the bug report API.
package main

import (
	"net/http"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils"
)

// CheckBugReportAPI sends a GET request to the bug report API's /health
// endpoint with a short timeout. Returns true if the API responds with
// status 200, false otherwise. This is exposed to the frontend.
func (a *App) CheckBugReportAPI() bool {
	cfgPath := utils.DefaultConfigPath()
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		pkglogger.Warn("heartbeat: failed to load config", "error", err.Error())
		return false
	}

	apiURL := cfg.EMLy.BugReportAPIURL
	if apiURL == "" {
		pkglogger.Warn("heartbeat: bug report API URL not configured")
		return false
	}

	endpoint := apiURL + "/health"
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(endpoint)
	if err != nil {
		pkglogger.Debug("heartbeat: API unreachable", "error", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		pkglogger.Warn("heartbeat: API returned non-200 status", "status", resp.StatusCode)
		return false
	}

	return true
}
