package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

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

	err := wails.Run(appOptions)

	if err != nil {
		pkglogger.Error("application error", "error", err.Error())
	}
}
