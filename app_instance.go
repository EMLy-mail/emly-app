// Package main: second-instance activation - bringing an already-running
// EMLy to the foreground (and forwarding the new launch args to it) instead
// of launching a second copy.
package main

import (
	"strings"
	"time"

	pkglogger "emly/backend/logger"

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

// bringToForeground restores and focuses the main window, used when a
// second instance is launched.
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
