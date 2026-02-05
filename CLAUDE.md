# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

EMLy is a Windows desktop email viewer for `.eml` and `.msg` files built with Wails v2 (Go backend + SvelteKit/Svelte 5 frontend). It supports viewing email content, attachments, and Italian PEC (Posta Elettronica Certificata) certified emails.

## Build Commands

```bash
# Development mode with hot-reload
wails dev

# Production build (outputs to build/bin/EMLy.exe)
wails build

# Build for specific platform
wails build -platform windows/amd64

# Frontend type checking
cd frontend && bun run check

# Go tests
go test ./...

# Install frontend dependencies (also runs automatically on build)
cd frontend && bun install
```

## Architecture

### Wails Framework
- Go backend methods are automatically exposed to frontend via generated TypeScript bindings
- Bindings are generated in `frontend/src/lib/wailsjs/go/main/App.ts`
- All App methods with exported names become callable from TypeScript
- Wails runtime provides window control, dialogs, and event system

### Backend Structure (Go)

The `App` struct is the main controller, split across files by domain:

| File | Responsibility |
|------|----------------|
| `app.go` | Core struct, lifecycle (startup/shutdown), configuration |
| `app_mail.go` | Email reading: ReadEML, ReadMSG, ReadPEC, ShowOpenFileDialog |
| `app_viewer.go` | Viewer windows: OpenImageWindow, OpenPDFWindow, OpenEMLWindow |
| `app_screenshot.go` | Window capture using Windows GDI APIs |
| `app_bugreport.go` | Bug report creation with screenshots and system info |
| `app_settings.go` | Settings import/export to JSON |
| `app_system.go` | Windows registry, encoding conversion, file explorer |

Email parsing lives in `backend/utils/mail/`:
- `eml_reader.go` - Standard EML parsing
- `msg_reader.go` - Microsoft MSG (CFB format) parsing
- `mailparser.go` - MIME multipart handling

### Frontend Structure (SvelteKit + Svelte 5)

**Routes** (file-based routing):
- `(app)/` - Main app with sidebar layout, titlebar, footer
- `(app)/settings/` - Settings page
- `image/` - Standalone image viewer (launched with `--view-image=<path>`)
- `pdf/` - Standalone PDF viewer (launched with `--view-pdf=<path>`)

**Key patterns**:
- Svelte 5 runes: `$state`, `$effect`, `$derived`, `$props` (NOT legacy stores in components)
- State classes in `stores/*.svelte.ts` using `$state` for reactive properties
- Traditional Svelte stores in `stores/app.ts` for simple global state
- shadcn-svelte components in `lib/components/ui/`

**Mail utilities** (`lib/utils/mail/`):
- Modular utilities for email loading, attachment handling, data conversion
- Barrel exported from `index.ts`

### Multi-Window Support

The app spawns separate viewer processes:
- Main app uses single-instance lock with ID `emly-app-lock`
- Image/PDF viewers get unique IDs per file, allowing multiple concurrent viewers
- CLI args: `--view-image=<path>` or `--view-pdf=<path>` trigger viewer mode

### Internationalization

Uses ParaglideJS with compile-time type-safe translations:
- Translation files: `frontend/messages/en.json`, `it.json`
- Import: `import * as m from "$lib/paraglide/messages"`
- Usage: `{m.mail_open_btn_text()}`

## Key Patterns

### Calling Go from Frontend
```typescript
import { ReadEML, ShowOpenFileDialog } from "$lib/wailsjs/go/main/App";
const filePath = await ShowOpenFileDialog();
const email = await ReadEML(filePath);
```

### Event Communication
```typescript
import { EventsOn } from "$lib/wailsjs/runtime/runtime";
EventsOn("launchArgs", (args: string[]) => { /* handle file from second instance */ });
```

### Email Body Rendering
Email bodies render in sandboxed iframes with links disabled for security.

## Important Conventions

- Windows-only: Uses Windows APIs (registry, GDI, user32.dll)
- Bun package manager for frontend (not npm/yarn)
- Frontend assets embedded in binary via `//go:embed all:frontend/build`
- Custom frameless window with manual titlebar implementation
