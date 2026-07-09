// Package main provides the Windows system tray integration for EMLy:
// showing/hiding the main window, the tray icon/menu, and second-instance
// activation (bringing an already-running EMLy to the foreground instead of
// launching a second copy).
package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	pkglogger "emly/backend/logger"
	"emly/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

	a.setTrayVisible(true)
}

// hideToTray hides the main window without quitting, mirroring what
// HideWindowOnClose already does when the user clicks the window's close
// button - used by the "Nascondi EMLy" tray menu item.
func (a *App) hideToTray() {
	runtime.WindowHide(a.ctx)
	a.setTrayVisible(false)
}

// setTrayVisible records whether the main window is currently shown and
// rebuilds the tray menu so its toggle item's label matches ("Nascondi
// EMLy" / "Mostra EMLy"). It's only called from our own show/hide code
// paths (tray menu clicks, second-instance activation) - if the window is
// ever hidden or shown by some other means, the label can lag until the
// next tray action, which then self-corrects it.
func (a *App) setTrayVisible(visible bool) {
	a.trayVisibleMux.Lock()
	a.trayVisible = visible
	a.trayVisibleMux.Unlock()

	if a.trayIconBase64 != "" {
		runtime.TraySetSystemTray(a.ctx, a.buildTrayMenu())
	}
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

// buildTrayMenu builds the system tray icon menu. Left-clicking the tray
// icon always restores the window (handled natively by Wails); the menu is
// only shown on right-click, and its "Mostra"/"Nascondi" item toggles based
// on the last known window visibility (see setTrayVisible).
func (a *App) buildTrayMenu() *menu.TrayMenu {
	trayMenu := menu.NewMenu()
	guiVersion := "unknown"
	if cfg, err := utils.LoadConfig(utils.DefaultConfigPath()); err == nil && cfg != nil {
		guiVersion = cfg.EMLy.GUISemver
	}
	trayMenu.AddText("Versione: "+guiVersion, nil, nil).Disable()
	trayMenu.AddSeparator()

	a.trayVisibleMux.Lock()
	visible := a.trayVisible
	a.trayVisibleMux.Unlock()

	if visible {
		trayMenu.AddText("Nascondi EMLy", nil, func(_ *menu.CallbackData) {
			a.hideToTray()
		})
	} else {
		trayMenu.AddText("Mostra EMLy", nil, func(_ *menu.CallbackData) {
			a.bringToForeground()
		})
	}
	trayMenu.AddSeparator()
	trayMenu.AddText("Esci", nil, func(_ *menu.CallbackData) {
		runtime.Quit(a.ctx)
	})

	return &menu.TrayMenu{
		Label:   "EMLy",
		Tooltip: "EMLy - EML Viewer for 3gIT",
		Image:   a.trayIconBase64,
		Menu:    trayMenu,
	}
}
