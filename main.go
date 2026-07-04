package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/binary"
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
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

//go:embed all:frontend/build
var assets embed.FS

//go:embed build/windows/icon.ico
var trayIconICO []byte

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	var secondInstanceArgs []string
	secondInstanceArgs = secondInstanceData.Args

	pkglogger.Info("second instance launched",
		"args", strings.Join(secondInstanceData.Args, ","),
		"working_dir", secondInstanceData.WorkingDirectory,
	)
	a.bringToForeground()

	go runtime.EventsEmit(a.ctx, "launchArgs", secondInstanceArgs)
}

// bringToForeground restores and focuses the main window, used both when a
// second instance is launched and when the user picks "Mostra EMLy" from the
// tray menu.
func (a *App) bringToForeground() {
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
}

// trayIconBase64 extracts the ICO entry closest to preferredSize from the
// embedded icon and returns it base64-encoded, in the form Wails' Windows
// tray implementation expects when Image isn't a file path. Passing a file
// path instead doesn't work: it's loaded via the Win32 LoadIcon function,
// which (with a null module handle) only resolves predefined system icon
// IDs, not arbitrary file paths - it silently fails for any real .ico file.
func trayIconBase64(ico []byte, preferredSize int) (string, error) {
	if len(ico) < 6 || binary.LittleEndian.Uint16(ico[2:4]) != 1 {
		return "", fmt.Errorf("not a valid .ico file")
	}
	count := int(binary.LittleEndian.Uint16(ico[4:6]))

	var bestOffset, bestSize uint32
	bestDiff := int(^uint(0) >> 1)
	off := 6
	for i := 0; i < count && off+16 <= len(ico); i++ {
		width := int(ico[off])
		if width == 0 {
			width = 256
		}
		byteSize := binary.LittleEndian.Uint32(ico[off+8 : off+12])
		imageOffset := binary.LittleEndian.Uint32(ico[off+12 : off+16])

		absDiff := width - preferredSize
		if absDiff < 0 {
			absDiff = -absDiff
		}
		if absDiff < bestDiff {
			bestDiff = absDiff
			bestOffset = imageOffset
			bestSize = byteSize
		}
		off += 16
	}

	end := bestOffset + bestSize
	if bestSize == 0 || end > uint32(len(ico)) {
		return "", fmt.Errorf("no usable image found in .ico data")
	}
	return base64.StdEncoding.EncodeToString(ico[bestOffset:end]), nil
}

// newTrayMenu builds the system tray icon menu. Left-clicking the tray icon
// already restores the window (handled natively by Wails); the menu is only
// shown on right-click.
func newTrayMenu(app *App, iconBase64 string) *menu.TrayMenu {
	trayMenu := menu.NewMenu()
	trayMenu.AddText("Mostra EMLy", nil, func(_ *menu.CallbackData) {
		app.bringToForeground()
	})
	trayMenu.AddSeparator()
	trayMenu.AddText("Esci", nil, func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	return &menu.TrayMenu{
		Label:   "EMLy",
		Tooltip: "EMLy - EML Viewer for 3gIT",
		Image:   iconBase64,
		Menu:    trayMenu,
	}
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
	// viewer windows close normally when the user closes them.
	if isMainWindow {
		if iconBase64, err := trayIconBase64(trayIconICO, 32); err != nil {
			pkglogger.Error("failed to prepare tray icon", "error", err.Error())
		} else {
			appOptions.Tray = newTrayMenu(app, iconBase64)
			appOptions.HideWindowOnClose = true
		}
	}

	err := wails.Run(appOptions)

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
