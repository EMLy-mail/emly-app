package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

//go:embed build/windows/icon.ico
var trayIconICO []byte

func main() {
	// Taken before anything else so the startup trace (see below) covers
	// the true process start, not just the point config/logger setup happens
	// to finish.
	processStart := time.Now()

	if err := InitLogger(); err != nil {
		log.Println("Error initializing logger:", err)
	}
	defer CloseLogger()

	// Build custom User-Agent from config version
	guiVersion := "unknown"
	if cfg, err := utils.LoadConfig(utils.DefaultConfigPath()); err == nil && cfg != nil {
		guiVersion = cfg.EMLy.GUISemver
	}

	printStartupBanner(guiVersion)

	// Check for custom args
	args := os.Args
	uniqueId := "emly-app-lock"
	windowTitle := "EMLy - EML Viewer for 3gIT"
	windowWidth := 1024
	windowHeight := 700
	frameless := true
	isMainWindow := true

	for _, arg := range args {
		if strings.Contains(arg, "--view-image") || isImageFilePath(arg) {
			uniqueId = "emly-viewer-" + arg // simplified uniqueness
			windowTitle = "EMLy Image Viewer"
			windowWidth = 800
			windowHeight = 600
			isMainWindow = false

		}
		if strings.Contains(arg, "--view-pdf") {
			uniqueId = "emly-pdf-viewer-" + strings.ReplaceAll(arg, "--view-pdf=", "")
			windowTitle = "EMLy PDF Viewer"
			windowWidth = 800
			windowHeight = 600
			frameless = true
			isMainWindow = false
		}
	}

	userAgent := fmt.Sprintf("EMLy/%s", guiVersion)

	// Startup tracing: a separate, plain-text timeline of launch + mail-load
	// timings (backend/logger/trace.go), truncated fresh on every run, so a
	// slow "Inizializzazione..." screen can be diagnosed step-by-step instead
	// of guessed at. Only for the main window - viewer windows (image/PDF)
	// are separate processes that would otherwise race to truncate the same
	// file, and their load path isn't what's being diagnosed here.
	//
	// Off by default (LOG_STARTUP_TRACE in config.ini / Settings → Danger
	// Zone → "Log startup trace") - it's a diagnostic tool for slow-open
	// investigations, not something every install needs writing to disk on
	// every launch. Takes effect on next restart, like the tray icon toggle
	// below.
	logStartupTrace := false
	if cfg, err := utils.LoadConfig(utils.DefaultConfigPath()); err == nil && cfg != nil {
		logStartupTrace = cfg.EMLy.LogStartupTrace
	}
	if isMainWindow && logStartupTrace {
		if configDir, err := os.UserConfigDir(); err == nil {
			tracePath := filepath.Join(configDir, "EMLy", "logs", "startup-trace.log")
			if err := pkglogger.InitStartupTrace(tracePath, processStart); err != nil {
				pkglogger.Error("failed to init startup trace", "error", err.Error())
			} else {
				defer pkglogger.CloseStartupTrace()
				pkglogger.TraceStep("process_start")
			}
		}
	}
	pkglogger.TraceStep("config_loaded", "gui_version="+guiVersion)

	// Create an instance of the app structure
	app := NewApp(userAgent)

	// Parse args again to set startup file on the app instance
	for _, arg := range args {
		if strings.HasSuffix(strings.ToLower(arg), ".eml") {
			app.StartupFilePath = arg
		}
		if strings.HasSuffix(strings.ToLower(arg), ".msg") {
			app.StartupFilePath = arg
		}
	}
	if app.StartupFilePath != "" {
		pkglogger.TraceStep("args_parsed", "startup_file="+filepath.Base(app.StartupFilePath))
	} else {
		pkglogger.TraceStep("args_parsed", "no_startup_file")
	}

	// Create application with options
	appOptions := &options.App{
		Title:  windowTitle,
		Width:  windowWidth,
		Height: windowHeight,
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Handler:    spaFallbackHandler(assets),
			Middleware: userAgentMiddleware(userAgent),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               uniqueId,
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		EnableDefaultContextMenu: true,
		MinWidth:                 964,
		MinHeight:                690,
		Frameless:                frameless,
		OnBeforeClose:            app.beforeClose,
	}

	// The tray icon only makes sense on the main window; standalone image/PDF
	// viewer windows close normally when the user closes them. It can also be
	// turned off entirely from the settings danger zone (DISABLE_TRAY_ICON in
	// config.ini), which takes effect on the next restart.
	trayIconEnabled := true
	if cfg, err := utils.LoadConfig(utils.DefaultConfigPath()); err == nil && cfg != nil {
		trayIconEnabled = !cfg.EMLy.DisableTrayIcon
	}

	if isMainWindow && trayIconEnabled {
		if iconBase64, err := trayIconBase64(trayIconICO, 32); err != nil {
			pkglogger.Error("failed to prepare tray icon", "error", err.Error())
		} else {
			app.trayIconBase64 = iconBase64
			app.trayVisible = true // the main window starts shown
			appOptions.Tray = app.buildTrayMenu()
			appOptions.HideWindowOnClose = true
		}
	}

	pkglogger.TraceStep("wails_run_starting")
	err := wails.Run(appOptions)
	pkglogger.TraceStep("wails_run_returned")

	if err != nil {
		pkglogger.Error("application error", "error", err.Error())
	}
}
