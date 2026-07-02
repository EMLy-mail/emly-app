# EMLy

EMLy is a lightweight, cross-platform desktop email viewer built with [Wails v2](https://wails.io/), Go, and SvelteKit (Svelte 5). It lets you open and inspect `.eml`, `.msg`, and PEC files - including attachments like images and PDFs - without an email client.

## Features

- **Format support**: EML, MSG (Outlook), PEC (Italian certified email), TNEF/winmail.dat, nested email attachments
- **Attachment viewers**: Built-in image and PDF viewers, or open with the system default app
- **Standalone image reader**: EMLy can register itself as a handler for common raster image extensions (`.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`) and open them directly in a lightweight image-viewer window, the same way Windows' built-in Photos viewer does
- **Tab mode**: View multiple emails simultaneously in separate tabs (opt-in via Settings)
- **Bug reporting**: Capture a screenshot, attach system info and the current mail file, then submit to a self-hosted API or save a local ZIP
- **Light/Dark theme** with accessibility options (reduced motion, high contrast)
- **Internationalization**: English and Italian, switchable at runtime
- **Structured logging**: JSON logs to stdout and `%APPDATA%/EMLy/logs/app.log`, configurable level

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, Wails v2.12 |
| Frontend | SvelteKit, Svelte 5 (Runes), TypeScript |
| UI | shadcn-svelte, Tailwind CSS, Lucide Icons |
| i18n | ParaglideJS |
| Package manager | Bun |

## Requirements

- **Go** 1.26+ - [Download](https://go.dev/dl/)
- **Wails CLI** v2 - [Installation Guide](https://wails.io/docs/gettingstarted/installation)
- **Bun** - [Install](https://bun.sh/)
  - macOS/Linux: `curl -fsSL https://bun.sh/install | bash`
  - Windows: `powershell -c "irm bun.sh/install.ps1 | iex"`

## Getting Started

```bash
# 1. Clone
git clone https://github.com/EMLy-mail/emly-app.git
cd emly-app

# 2. Install frontend dependencies (also applies the Bun patches, see below)
cd frontend && bun install && cd ..

# 3. Vendor + patch the Go dependencies (see "Wails v2 Patches" below)
powershell -File scripts/vendor-wails-patch.ps1

# 4. Run in development mode (hot-reload)
wails dev -tags debug
```

The app window opens automatically. The frontend is also reachable at `http://localhost:34115` for browser DevTools.

## Building

```bash
# Production binary → build/bin/EMLy.exe
wails build -platform windows/amd64
```

Installer scripts are in `build/windows/installer/` (NSIS) and `installer/` (Inno Setup). macOS bundle config is in `build/darwin/`. File/image type associations (`.eml`, `.msg`, and the standalone image reader extensions) are registered via `[Registry]` entries in `installer/installer.iss`.

## Configuration

Copy `config.ini` (or `config.debug.ini` for debug builds, picked automatically) and adjust as needed:

```ini
[EMLy]
SDK_DECODER_SEMVER          = 1.6.0                 # version of the mail-parsing SDK, shown in Credits/bug reports
SDK_DECODER_RELEASE_CHANNEL = stable                 # stable | beta
GUI_SEMVER                  = 1.8.0                  # app version, used in the User-Agent and bug reports
GUI_RELEASE_CHANNEL         = pre-release             # stable | beta | pre-release
LANGUAGE                    = it                      # it | en, initial UI language
BUGREPORT_API_URL           = https://api.emly.ffois.it/v1
BUGREPORT_API_KEY           = your-api-key
LOG_LEVEL                   = INFO                    # DEBUG | INFO | WARN | ERROR
EXPORT_ATTACHMENT_FOLDER    =                          # default attachment download folder; empty = system Downloads
```

## Updates

EMLy no longer ships a built-in update checker/installer. Version rollout is handled entirely by **EMLy-Updater**, a separate external tool. There are no `UPDATE_*` keys in `config.ini` and no update UI in Settings.

## Wails v2 Patches

This project backports one feature from Wails v3 onto the pinned Wails v2.12.0: a `WindowOpenDevTools` runtime call (exposed on `App` as `OpenDevTools()`) that programmatically opens the WebView2 DevTools window, since v2 only exposes this via the F12 shortcut.

The change lives as a patch, not a fork, so it never needs to be manually re-applied against a re-pulled copy of Wails:

- `patches/wails-v2-opendevtools.patch` - unified diff against `vendor/github.com/wailsapp/wails/v2`
- `scripts/vendor-wails-patch.ps1` - runs `go mod vendor` and (re)applies the patch

`vendor/` itself is gitignored and regenerated locally; run the script above after cloning and after any Wails version bump (it will fail loudly if the patch no longer applies cleanly, meaning the upstream code moved and the patch needs a manual re-diff).

## Testing

```bash
# Frontend type checking
cd frontend && bun run check

# Go tests
go test ./...
```

## License & Credits

EMLy is developed by FOISX @ 3gIT.
