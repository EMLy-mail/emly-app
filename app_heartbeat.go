// Package main provides heartbeat checking for the bug report API.
package main

import (
	"fmt"
	"net/http"
	"time"

	"emly/backend/utils"
)

// CheckBugReportAPI sends a GET request to the bug report API's /health
// endpoint with a short timeout. Returns true if the API responds with
// status 200, false otherwise. This is exposed to the frontend.
func (a *App) CheckBugReportAPI() bool {
	cfgPath := utils.DefaultConfigPath()
	cfg, err := utils.LoadConfig(cfgPath)
	if err != nil {
		Log("Heartbeat: failed to load config:", err)
		return false
	}

	apiURL := cfg.EMLy.BugReportAPIURL
	if apiURL == "" {
		Log("Heartbeat: bug report API URL not configured")
		return false
	}

	endpoint := apiURL + "/health"
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(endpoint)
	if err != nil {
		Log("Heartbeat: API unreachable:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Log(fmt.Sprintf("Heartbeat: API returned status %d", resp.StatusCode))
		return false
	}

	return true
}
