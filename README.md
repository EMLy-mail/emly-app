# EMLy

EMLy is a lightweight, cross-platform desktop email viewer built with [Wails v2](https://wails.io/), Go, and SvelteKit (Svelte 5). It lets you open and inspect `.eml`, `.msg`, and PEC files - including attachments like images and PDFs - without an email client.

## Features

- **Format support**: EML, MSG (Outlook), PEC (Italian certified email), TNEF/winmail.dat, nested email attachments
- **Attachment viewers**: Built-in image and PDF viewers, or open with the system default app
- **Tab mode**: View multiple emails simultaneously in separate tabs (opt-in via Settings)
- **Self-hosted updates**: Check for new versions from a corporate UNC share or an internal HTTP API
- **Bug reporting**: Capture a screenshot, attach system info and the current mail file, then submit to a self-hosted API or save a local ZIP
- **Light/Dark theme** with accessibility options (reduced motion, high contrast)
- **Internationalization**: English and Italian, switchable at runtime
- **Structured logging**: JSON logs to stdout and `%APPDATA%/EMLy/logs/app.log`, configurable level

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Wails v2.12 |
| Frontend | SvelteKit, Svelte 5 (Runes), TypeScript |
| UI | shadcn-svelte, Tailwind CSS, Lucide Icons |
| i18n | ParaglideJS |
| Package manager | Bun |

## Requirements

- **Go** 1.24+ - [Download](https://go.dev/dl/)
- **Wails CLI** v2 - [Installation Guide](https://wails.io/docs/gettingstarted/installation)
- **Bun** - [Install](https://bun.sh/)
  - macOS/Linux: `curl -fsSL https://bun.sh/install | bash`
  - Windows: `powershell -c "irm bun.sh/install.ps1 | iex"`

## Getting Started

```bash
# 1. Clone
git clone https://github.com/EMLy-mail/emly-app.git
cd emly-app

# 2. Install frontend dependencies
cd frontend && bun install && cd ..

# 3. Run in development mode (hot-reload)
wails dev -tags debug
```

The app window opens automatically. The frontend is also reachable at `http://localhost:34115` for browser DevTools.

## Building

```bash
# Production binary → build/bin/EMLy.exe
wails build -platform windows/amd64
```

Installer scripts are in `build/windows/installer/` (NSIS) and `installer/` (Inno Setup). macOS bundle config is in `build/darwin/`.

## Configuration

Copy `config.ini` and adjust as needed:

```ini
[EMLy]
LOG_LEVEL="INFO"                 # DEBUG | INFO | WARN | ERROR
UPDATE_CHECK_ENABLED="true"
UPDATE_SOURCE="unc"              # unc | api
UPDATE_PATH="\\server\emly-updates"
GUI_RELEASE_CHANNEL="stable"     # stable | beta
BUGREPORT_API_URL="https://your-server.example.com"
BUGREPORT_API_KEY="your-api-key"
```

## Update System

EMLy supports two update sources, switchable from Settings (Danger Zone):

- **UNC**: reads `version.json` from a corporate file share and downloads the installer directly.
- **API**: queries an HTTP endpoint (`/v2/updates/manifest`) that returns version info, download URL, and optional critical/security flags.

## Testing

```bash
# Frontend type checking
cd frontend && bun run check

# Go tests
go test ./...
```

## License & Credits

EMLy is developed by FOISX @ 3gIT.
