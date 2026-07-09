// Package main provides detection of the EMLy Updater's installation and
// running state, shown in the settings Danger Zone.
package main

import (
	"os"
	"path/filepath"
	"syscall"

	pkglogger "emly/backend/logger"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// EMLyUpdaterStatus describes whether the EMLy Updater is installed on this
// machine and whether its service is currently running.
type EMLyUpdaterStatus struct {
	// Installed is true when the "EMLyUpdater" service is registered and
	// both its installation and config folders exist.
	Installed bool
	// Running is true when the "EMLyUpdater" service is currently started.
	Running bool
}

// GetEMLyUpdaterStatus reports the installation and running state of the
// EMLy Updater. Installed requires all three of the following:
//  1. The "EMLyUpdater" Windows service is registered
//  2. The installation folder exists (%ProgramFiles%\EMLyUpdater)
//  3. The config folder exists (%ProgramData%\EMLyUpdater)
//
// Every check reads state that a standard (non-administrator) account can
// already access, so no elevation is required.
func (a *App) GetEMLyUpdaterStatus() EMLyUpdaterStatus {
	registered := emlyUpdaterServiceRegistered()
	installDir := filepath.Join(programFilesDir(), "EMLyUpdater")
	configDir := filepath.Join(programDataDir(), "EMLyUpdater")
	installDirOk := dirExists(installDir)
	configDirOk := dirExists(configDir)
	running := emlyUpdaterServiceRunning()

	status := EMLyUpdaterStatus{
		Installed: registered && installDirOk && configDirOk,
		Running:   running,
	}
	pkglogger.Debug("EMLy Updater status check",
		"serviceRegistered", registered,
		"installDir", installDir,
		"installDirExists", installDirOk,
		"configDir", configDir,
		"configDirExists", configDirOk,
		"running", running,
		"result", status,
	)
	return status
}

// emlyUpdaterServiceRegistered checks the registry for the "EMLyUpdater"
// service. Reading HKLM\SYSTEM\CurrentControlSet\Services\<name> only
// requires KEY_READ, which standard accounts are granted by default -
// unlike opening a handle to the Service Control Manager to query status.
func emlyUpdaterServiceRegistered() bool {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\EMLyUpdater`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		pkglogger.Debug("EMLy Updater service registry key not readable", "error", err.Error())
		return false
	}
	defer k.Close()
	return true
}

// emlyUpdaterServiceRunning queries the Service Control Manager for the
// current run state of the "EMLyUpdater" service.
//
// This deliberately does NOT use golang.org/x/sys/windows/svc/mgr: that
// package's Connect() and OpenService() request SC_MANAGER_ALL_ACCESS /
// SERVICE_ALL_ACCESS, which the SCM denies to non-administrator accounts
// even for a read-only status query. Opening the manager and the service
// with only the QUERY-level access rights below works for standard users.
func emlyUpdaterServiceRunning() bool {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		pkglogger.Debug("EMLy Updater: OpenSCManager failed", "error", err.Error())
		return false
	}
	defer windows.CloseServiceHandle(scm)

	serviceName, err := syscall.UTF16PtrFromString("EMLyUpdater")
	if err != nil {
		pkglogger.Debug("EMLy Updater: failed to encode service name", "error", err.Error())
		return false
	}

	svcHandle, err := windows.OpenService(scm, serviceName, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		pkglogger.Debug("EMLy Updater: OpenService failed", "error", err.Error())
		return false
	}
	defer windows.CloseServiceHandle(svcHandle)

	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(svcHandle, &status); err != nil {
		pkglogger.Debug("EMLy Updater: QueryServiceStatus failed", "error", err.Error())
		return false
	}

	pkglogger.Debug("EMLy Updater: service status queried", "currentState", status.CurrentState)
	return status.CurrentState == windows.SERVICE_RUNNING
}

// dirExists reports whether path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// programFilesDir returns %ProgramFiles%, falling back to the standard
// default if the environment variable isn't set.
func programFilesDir() string {
	if dir := os.Getenv("ProgramFiles"); dir != "" {
		return dir
	}
	return `C:\Program Files`
}

// programDataDir returns %ProgramData%, falling back to the standard
// default if the environment variable isn't set.
func programDataDir() string {
	if dir := os.Getenv("ProgramData"); dir != "" {
		return dir
	}
	return `C:\ProgramData`
}
