// Package main provides the console startup banner shown when EMLy is
// launched from a terminal.
package main

import (
	"fmt"
	"os"

	pkglogger "emly/backend/logger"

	"github.com/mbndr/figlet4go"
	"golang.org/x/sys/windows"
)

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
