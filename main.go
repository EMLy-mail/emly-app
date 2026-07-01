package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils"

	"github.com/mbndr/figlet4go"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

//go:embed all:frontend/build
var assets embed.FS

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	var secondInstanceArgs []string
	secondInstanceArgs = secondInstanceData.Args

	pkglogger.Info("second instance launched",
		"args", strings.Join(secondInstanceData.Args, ","),
		"working_dir", secondInstanceData.WorkingDirectory,
	)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)

	// Windows blocks SetForegroundWindow calls from background processes
	// (foreground lock), so WindowShow alone often fails to bring the
	// window above other apps. Toggling AlwaysOnTop forces a Z-order
	// change that isn't subject to that restriction.
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	go func() {
		time.Sleep(200 * time.Millisecond)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
	}()

	go runtime.EventsEmit(a.ctx, "launchArgs", secondInstanceArgs)
}

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

	for _, arg := range args {
		if strings.Contains(arg, "--view-image") || isImageFilePath(arg) {
			uniqueId = "emly-viewer-" + arg // simplified uniqueness
			windowTitle = "EMLy Image Viewer"
			windowWidth = 800
			windowHeight = 600

		}
		if strings.Contains(arg, "--view-pdf") {
			uniqueId = "emly-pdf-viewer-" + strings.ReplaceAll(arg, "--view-pdf=", "")
			windowTitle = "EMLy PDF Viewer"
			windowWidth = 800
			windowHeight = 600
			frameless = true
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
	err := wails.Run(&options.App{
		Title:  windowTitle,
		Width:  windowWidth,
		Height: windowHeight,
		AssetServer: &assetserver.Options{
			Assets:     assets,
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
	})

	if err != nil {
		pkglogger.Error("application error", "error", err.Error())
	}
}

// enableVTMode turns on ANSI/VT100 escape sequence processing on the
// attached console, which is off by default on Windows so 24-bit color
// codes would otherwise be printed as raw escape sequences.
func enableVTMode() {
	stdout := windows.Handle(os.Stdout.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(stdout, &mode); err != nil {
		return
	}

	_ = windows.SetConsoleMode(stdout, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

// printStartupBanner prints the EMLy ASCII art logo and version to the
// console at startup, with "EML" in gold-grey and "Y" in dark purple.
func printStartupBanner(version string) {
	enableVTMode()

	goldGrey, err := figlet4go.NewTrueColorFromHexString("B8A989")
	if err != nil {
		pkglogger.Error("failed to build startup banner color", "error", err.Error())
		return
	}
	darkPurple, err := figlet4go.NewTrueColorFromHexString("4A1942")
	if err != nil {
		pkglogger.Error("failed to build startup banner color", "error", err.Error())
		return
	}

	renderOptions := figlet4go.NewRenderOptions()
	renderOptions.FontColor = []figlet4go.Color{
		goldGrey, goldGrey, goldGrey, darkPurple,
	}

	ascii := figlet4go.NewAsciiRender()
	banner, err := ascii.RenderOpts("EMLy", renderOptions)
	if err != nil {
		pkglogger.Error("failed to render startup banner", "error", err.Error())
		return
	}

	fmt.Print(banner)
	fmt.Printf("  v%s\n\n", version)
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
