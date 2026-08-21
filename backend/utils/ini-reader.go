package utils

import (
	"emly/backend/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/mod/semver"
	"gopkg.in/ini.v1"
)

// Config represents the structure of config.ini
type Config struct {
	EMLy EMLyConfig `ini:"EMLy"`
}

type EMLyConfig struct {
	SDKDecoderSemver         string `ini:"SDK_DECODER_SEMVER"`
	SDKDecoderReleaseChannel string `ini:"SDK_DECODER_RELEASE_CHANNEL"`
	GUISemver                string `ini:"GUI_SEMVER"`
	GUIReleaseChannel        string `ini:"GUI_RELEASE_CHANNEL"`
	Language                 string `ini:"LANGUAGE"`
	BugReportAPIURL          string `ini:"BUGREPORT_API_URL"`
	BugReportAPIKey          string `ini:"BUGREPORT_API_KEY"`
	LogLevel                 string `ini:"LOG_LEVEL"`
	ExportAttachmentFolder   string `ini:"EXPORT_ATTACHMENT_FOLDER"`
	DisableTrayIcon          bool   `ini:"DISABLE_TRAY_ICON"`

	// LogStartupTrace enables the plain-text startup-trace.log (see
	// backend/logger/trace.go) - off by default, since it's a diagnostic
	// tool for slow-mail-open investigations, not something every user
	// needs writing to disk on every launch. Settings → Danger Zone → "Log
	// startup trace". Takes effect on the next restart (read once in
	// main(), see InitStartupTrace).
	LogStartupTrace bool `ini:"LOG_STARTUP_TRACE"`

	// OldAttachmentPreload reverts ReadEML/ReadMSG/ReadPEC/ReadAuto to the
	// pre-fix behaviour of sending every attachment's full bytes in the
	// initial parse response, instead of the lazy default (see
	// stripAttachmentData/GetAttachmentData in app_mail.go). Off by
	// default - kept only as an escape hatch for experiments/regression
	// testing, since it reintroduces the slow-open problem the fix
	// solved. Settings → Danger Zone → "Old Pre-loading of attachments".
	// Takes effect immediately (read fresh on every call), no restart
	// needed.
	OldAttachmentPreload bool `ini:"OLD_ATTACHMENT_PRELOAD"`
}

// checkSemver reports whether version — the raw GUI_SEMVER value
// from config.ini, normally stored without a "v" prefix (e.g. "2.0.1") —
// is a valid semantic version per golang.org/x/mod/semver, which requires
// the "v" prefix. A value that already starts with "v" is used as-is so
// this stays a no-op for callers that pass an already-prefixed string.
func checkSemver(version string) bool {
	v := version
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	v = semver.Canonical(v)
	return semver.IsValid(v)
}

// LoadConfig reads the config.ini file at the given path and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	logger.Log("LoadConfig path:", path)
	cfg, err := ini.Load(path)
	if err != nil {
		logger.Log("Fail to read file:", err)
		return nil, err
	}

	config := new(Config)
	if err := cfg.MapTo(config); err != nil {
		logger.Log("Fail to map config:", err)
		return nil, err
	}

	if !checkSemver(config.EMLy.SDKDecoderSemver) {
		logger.Log("Invalid SDK_DECODER_SEMVER in config:", config.EMLy.SDKDecoderSemver)
		return nil, fmt.Errorf("invalid SDK_DECODER_SEMVER in config: %q", config.EMLy.SDKDecoderSemver)
	}

	if !checkSemver(config.EMLy.GUISemver) {
		logger.Log("Invalid GUI_SEMVER in config:", config.EMLy.GUISemver)
		fatalInvalidGUIVersion(config.EMLy.GUISemver)
	}

	return config, nil
}

// fatalInvalidGUIVersion shows a native Win32 error messagebox and
// terminates the process. LoadConfig is called both before Wails starts
// (main.go, before wails.Run) and from every runtime reload/binding call
// (Settings, tray, heartbeat), so a Wails context.Context isn't always
// available — hence a native dialog instead of runtime.MessageDialog.
// Reuses the user32.dll handle already opened in screenshot_windows.go.
func fatalInvalidGUIVersion(version string) {
	const (
		mbOK        = 0x00000000
		mbIconError = 0x00000010
	)

	title := "EMLy - Errore di configurazione"
	message := fmt.Sprintf(
		"La versione configurata (%q) non è un formato semver valido.\nControllare il campo GUI_SEMVER in config.ini.\n\nEMLy verrà chiuso.",
		version,
	)

	titlePtr, errTitle := syscall.UTF16PtrFromString(title)
	messagePtr, errMessage := syscall.UTF16PtrFromString(message)
	if errTitle == nil && errMessage == nil {
		messageBoxW := user32.NewProc("MessageBoxW")
		messageBoxW.Call(0, uintptr(unsafe.Pointer(messagePtr)), uintptr(unsafe.Pointer(titlePtr)), uintptr(mbOK|mbIconError))
	}

	os.Exit(1)
}

func SaveConfig(path string, config *Config) error {
	logger.Log("SaveConfig path:", path)
	cfg := ini.Empty()
	if err := cfg.ReflectFrom(config); err != nil {
		logger.Log("Fail to reflect config:", err)
		return err
	}
	if err := cfg.SaveTo(path); err != nil {
		logger.Log("Fail to save config file:", err)
		return err
	}
	return nil
}

func DefaultConfigPath() string {
	configName := "config.ini"
	if isDebugBuild {
		logger.Log("Debug build: using config.debug.ini")
		configName = "config.debug.ini"
	}

	// Prefer the config file next to the executable (packaged app), fallback to CWD (dev).
	exe, err := os.Executable()
	if err == nil {
		p := filepath.Join(filepath.Dir(exe), configName)
		if _, statErr := os.Stat(p); statErr == nil {
			return p
		}
	}
	return configName
}
