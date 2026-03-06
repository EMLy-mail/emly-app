# EMLy Application Documentation

EMLy is a desktop email viewer application built for 3gIT, designed to open and display `.eml` and `.msg` email files on Windows. It provides a modern, user-friendly interface for viewing email content, attachments, and metadata.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [Backend (Go)](#backend-go)
5. [Frontend (SvelteKit)](#frontend-sveltekit)
6. [State Management](#state-management)
7. [Internationalization (i18n)](#internationalization-i18n)
8. [UI Components](#ui-components)
9. [Key Features](#key-features)
10. [Build & Development](#build--development)

---

## Architecture Overview

EMLy is built using the **Wails v2** framework, which combines a Go backend with a web-based frontend. This architecture allows:

- **Go Backend**: Handles file operations, Windows API calls, email parsing, and system interactions
- **Web Frontend**: Provides the user interface using SvelteKit with Svelte 5
- **Bridge**: Wails automatically generates TypeScript bindings for Go functions, enabling seamless communication

```
┌─────────────────────────────────────────────────────────┐
│                    EMLy Application                      │
├─────────────────────────────────────────────────────────┤
│  Frontend (SvelteKit + Svelte 5)                        │
│  ├── Routes & Pages                                     │
│  ├── UI Components (shadcn-svelte)                      │
│  ├── State Management (Svelte 5 Runes)                  │
│  └── i18n (ParaglideJS)                                 │
├─────────────────────────────────────────────────────────┤
│  Wails Bridge (Auto-generated TypeScript bindings)      │
├─────────────────────────────────────────────────────────┤
│  Backend (Go - Modular Architecture)                    │
│  ├── app.go          - Core struct & lifecycle          │
│  ├── app_mail.go     - Email parsing (EML/MSG/PEC)      │
│  ├── app_viewer.go   - Viewer window management         │
│  ├── app_screenshot.go - Window capture                 │
│  ├── app_bugreport.go  - Bug reporting system           │
│  ├── app_settings.go   - Settings import/export         │
│  ├── app_system.go     - Windows system utilities       │
│  ├── app_update.go     - Self-hosted update system      │
│  └── backend/utils/    - Shared utilities               │
└─────────────────────────────────────────────────────────┘
```

---

## Technology Stack

### Backend
- **Go 1.21+**: Primary backend language
- **Wails v2**: Desktop application framework
- **Windows Registry API**: For checking default file handlers
- **GDI/User32 APIs**: For screenshot functionality

### Frontend
- **SvelteKit**: Application framework
- **Svelte 5**: UI framework with Runes ($state, $derived, $effect, $props)
- **TypeScript**: Type-safe JavaScript
- **shadcn-svelte**: UI component library
- **Tailwind CSS**: Utility-first CSS framework
- **ParaglideJS**: Internationalization
- **Lucide Icons**: Icon library
- **Bun**: JavaScript runtime and package manager

---

## Project Structure

```
EMLy/
├── app.go                    # Core App struct, lifecycle, and configuration
├── app_mail.go               # Email reading methods (EML, MSG, PEC)
├── app_viewer.go             # Viewer window management (image, PDF, EML)
├── app_screenshot.go         # Screenshot capture functionality
├── app_bugreport.go          # Bug report creation and submission
├── app_heartbeat.go          # Bug report API heartbeat check
├── app_settings.go           # Settings import/export
├── app_system.go             # Windows system utilities (registry, encoding)
├── main.go                   # Application entry point
├── logger.go                 # Logging wrappers (canonical log lines, FrontendLog bridge)
├── wails.json                # Wails configuration
├── backend/
│   ├── logger/
│   │   └── logger.go             # Structured JSON logging (log/slog)
│   └── utils/
│       ├── mail/
│       │   ├── eml_reader.go     # EML file parsing
│       │   ├── msg_reader.go     # MSG file parsing
│       │   ├── mailparser.go     # MIME email parsing
│       │   └── file_dialog.go    # File dialog utilities
│       ├── screenshot_windows.go  # Windows screenshot capture
│       ├── debug_windows.go       # Debugger detection
│       ├── ini-reader.go          # Configuration file parsing
│       ├── machine-identifier.go  # System info collection
│       └── file-metadata.go       # File metadata utilities
├── frontend/
│   ├── src/
│   │   ├── routes/               # SvelteKit routes
│   │   │   ├── +layout.svelte    # Root layout
│   │   │   ├── +error.svelte     # Error page
│   │   │   ├── (app)/            # Main app route group
│   │   │   │   ├── +layout.svelte    # App layout with sidebar
│   │   │   │   ├── +page.svelte      # Mail viewer page
│   │   │   │   └── settings/
│   │   │   │       └── +page.svelte  # Settings page
│   │   │   ├── image/            # Image viewer route
│   │   │   │   ├── +layout.svelte
│   │   │   │   └── +page.svelte
│   │   │   └── pdf/              # PDF viewer route
│   │   │       ├── +layout.svelte
│   │   │       └── +page.svelte
│   │   ├── lib/
│   │   │   ├── components/       # Svelte components
│   │   │   │   ├── MailViewer.svelte
│   │   │   │   ├── SidebarApp.svelte
│   │   │   │   ├── UnsavedBar.svelte
│   │   │   │   └── ui/           # shadcn-svelte components
│   │   │   ├── stores/           # State management
│   │   │   │   ├── app.ts
│   │   │   │   ├── mail-state.svelte.ts
│   │   │   │   └── settings.svelte.ts
│   │   │   ├── paraglide/        # i18n runtime
│   │   │   ├── wailsjs/          # Auto-generated Go bindings
│   │   │   ├── types/            # TypeScript types
│   │   │   └── utils/            # Utility functions
│   │   │       └── mail/         # Email utilities (modular)
│   │   │           ├── index.ts          # Barrel export
│   │   │           ├── constants.ts      # IFRAME_UTIL_HTML, CONTENT_TYPES, etc.
│   │   │           ├── data-utils.ts     # arrayBufferToBase64, createDataUrl
│   │   │           ├── attachment-handlers.ts  # openPDFAttachment, openImageAttachment
│   │   │           └── email-loader.ts   # loadEmailFromPath, processEmailBody
│   │   └── messages/             # i18n translation files
│   │       ├── en.json
│   │       └── it.json
│   ├── static/                   # Static assets
│   └── package.json
└── config.ini                    # Application configuration
```

---

## Backend (Go)

### Entry Point (`main.go`)

The application starts in `main.go`, which:

1. Initializes the logger
2. Parses command-line arguments for:
   - `.eml` or `.msg` files to open on startup
   - `--view-image=<path>` for image viewer mode
   - `--view-pdf=<path>` for PDF viewer mode
3. Configures Wails with window options
4. Sets up single-instance lock to prevent multiple main windows
5. Binds the `App` struct for frontend access

```go
// Single instance with unique ID based on mode
uniqueId := "emly-app-lock"
if strings.Contains(arg, "--view-image") {
    uniqueId = "emly-viewer-" + arg
    windowTitle = "EMLy Image Viewer"
}
```

### Application Core (`app.go`)

The `App` struct is the main application controller, exposed to the frontend via Wails bindings. The code is organized into multiple files for maintainability.

#### Key Properties

```go
type App struct {
    ctx                 context.Context  // Wails application context
    StartupFilePath     string           // File opened via command line
    CurrentMailFilePath string           // Currently loaded mail file
    openImagesMux       sync.Mutex       // Mutex for image viewer tracking
    openImages          map[string]bool  // Track open image viewers
    openPDFsMux         sync.Mutex       // Mutex for PDF viewer tracking
    openPDFs            map[string]bool  // Track open PDF viewers
    openEMLsMux         sync.Mutex       // Mutex for EML viewer tracking
    openEMLs            map[string]bool  // Track open EML viewers
}
```

#### Backend File Organization

The Go backend is split into logical files:

| File | Purpose |
|------|---------|
| `app.go` | Core App struct, constructor, lifecycle methods (startup/shutdown), configuration |
| `app_mail.go` | Email reading: `ReadEML`, `ReadMSG`, `ReadPEC`, `ReadMSGOSS`, `ShowOpenFileDialog` |
| `app_viewer.go` | Viewer windows: `OpenImageWindow`, `OpenPDFWindow`, `OpenEMLWindow`, `OpenPDF`, `OpenImage`, `GetViewerData` |
| `app_screenshot.go` | Screenshots: `TakeScreenshot`, `SaveScreenshot`, `SaveScreenshotAs` |
| `app_bugreport.go` | Bug reports: `CreateBugReportFolder`, `SubmitBugReport`, `zipFolder` |
| `app_heartbeat.go` | API heartbeat: `CheckBugReportAPI` |
| `app_settings.go` | Settings I/O: `ExportSettings`, `ImportSettings` |
| `app_system.go` | System utilities: `CheckIsDefaultEMLHandler`, `OpenDefaultAppsSettings`, `ConvertToUTF8`, `OpenFolderInExplorer` |
| `app_update.go` | Update system: `CheckForUpdates`, `DownloadUpdate`, `InstallUpdate`, `GetUpdateStatus` |

#### Core Methods by Category

**Lifecycle & Configuration (`app.go`)**

| Method | Description |
|--------|-------------|
| `startup(ctx)` | Wails startup callback, saves context |
| `shutdown(ctx)` | Wails shutdown callback for cleanup |
| `QuitApp()` | Terminates the application |
| `GetConfig()` | Returns application configuration from `config.ini` |
| `SaveConfig(cfg)` | Saves configuration to `config.ini` |
| `GetStartupFile()` | Returns file path passed via command line |
| `SetCurrentMailFilePath()` | Updates the current mail file path |
| `GetMachineData()` | Returns system information |
| `IsDebuggerRunning()` | Checks if a debugger is attached |

**Email Reading (`app_mail.go`)**

| Method | Description |
|--------|-------------|
| `ReadEML(path)` | Parses a standard .eml file |
| `ReadMSG(path, useExternal)` | Parses a Microsoft .msg file |
| `ReadPEC(path)` | Parses PEC (Italian certified email) files |
| `ShowOpenFileDialog()` | Opens native file picker for EML/MSG files |

**Viewer Windows (`app_viewer.go`)**

| Method | Description |
|--------|-------------|
| `OpenImageWindow(data, filename)` | Opens image in built-in viewer |
| `OpenPDFWindow(data, filename)` | Opens PDF in built-in viewer |
| `OpenEMLWindow(data, filename)` | Opens EML attachment in new EMLy window |
| `OpenImage(data, filename)` | Opens image with system default app |
| `OpenPDF(data, filename)` | Opens PDF with system default app |
| `GetViewerData()` | Returns viewer data for viewer mode detection |

**Screenshots (`app_screenshot.go`)**

| Method | Description |
|--------|-------------|
| `TakeScreenshot()` | Captures window screenshot as base64 PNG |
| `SaveScreenshot()` | Saves screenshot to temp directory |
| `SaveScreenshotAs()` | Opens save dialog for screenshot |

**Bug Reports (`app_bugreport.go`)**

| Method | Description |
|--------|-------------|
| `CreateBugReportFolder()` | Creates folder with screenshot and mail file |
| `SubmitBugReport(input)` | Creates complete bug report with ZIP archive, attempts server upload |
| `UploadBugReport(folderPath, input)` | Uploads bug report files to configured API server via multipart POST |
| `CheckBugReportAPI()` | Checks if the bug report API is reachable via /health endpoint (3s timeout) |

**Settings (`app_settings.go`)**

| Method | Description |
|--------|-------------|
| `ExportSettings(json)` | Exports settings to JSON file |
| `ImportSettings()` | Imports settings from JSON file |

**System Utilities (`app_system.go`)**

| Method | Description |
|--------|-------------|
| `CheckIsDefaultEMLHandler()` | Checks if EMLy is default for .eml files |
| `OpenDefaultAppsSettings()` | Opens Windows default apps settings |
| `ConvertToUTF8(string)` | Converts string to valid UTF-8 |
| `OpenFolderInExplorer(path)` | Opens folder in Windows Explorer |

### Email Parsing (`backend/utils/mail/`)

#### EML Reader (`eml_reader.go`)
Reads standard `.eml` files using the `mailparser.go` MIME parser.

#### MSG Reader (`msg_reader.go`)
Handles Microsoft Outlook `.msg` files using external conversion.

#### Mail Parser (`mailparser.go`)
A comprehensive MIME email parser that handles:
- Multipart messages (mixed, alternative, related)
- Text and HTML bodies
- Attachments with proper content-type detection
- Embedded files (inline images)
- Various content transfer encodings (base64, quoted-printable, 7bit, 8bit)

The `EmailData` structure returned to the frontend:
```go
type EmailData struct {
    Subject     string
    From        string
    To          []string
    Cc          []string
    Bcc         []string
    Body        string          // HTML or text body
    Attachments []AttachmentData
    IsPec       bool            // Italian certified email
}
```

### Screenshot Utility (`backend/utils/screenshot_windows.go`)

Captures the application window using Windows GDI APIs:
- Uses `FindWindowW` to locate window by title
- Uses `DwmGetWindowAttribute` for DPI-aware window bounds
- Creates compatible DC and bitmap for capture
- Returns image as base64-encoded PNG

---

## Frontend (SvelteKit)

### Route Structure

SvelteKit uses file-based routing. The `(app)` folder is a route group that applies the main app layout.

```
routes/
├── +layout.svelte          # Root layout (minimal)
├── +error.svelte           # Global error page
├── (app)/                  # Main app group
│   ├── +layout.svelte      # App layout with titlebar, sidebar, footer
│   ├── +layout.ts          # Server data loader
│   ├── +page.svelte        # Main mail viewer page
│   └── settings/
│       ├── +page.svelte    # Settings page
│       └── +layout.ts      # Settings data loader
├── image/                  # Standalone image viewer
│   ├── +layout.svelte
│   └── +page.svelte
└── pdf/                    # Standalone PDF viewer
    ├── +layout.svelte
    └── +page.svelte
```

### Main App Layout (`(app)/+layout.svelte`)

The app layout provides:

1. **Custom Titlebar**: Windows-style titlebar with minimize/maximize/close buttons
   - Draggable for window movement
   - Double-click to maximize/restore

2. **Sidebar Provider**: Collapsible navigation sidebar

3. **Footer Bar**: Quick access icons for:
   - Toggle sidebar
   - Navigate to home
   - Navigate to settings
   - Open bug report dialog
   - Reload application

4. **Bug Report Dialog**: Complete bug reporting system with:
   - Screenshot capture on dialog open
   - Name, email, description fields
   - System info collection
   - Creates ZIP archive with all data

5. **Toast Notifications**: Using svelte-sonner

6. **Debugger Protection**: Detects attached debuggers and can quit if detected

### Mail Viewer (`(app)/+page.svelte` + `MailViewer.svelte`)

The mail viewer is split into two parts:
- `+page.svelte`: Page wrapper that initializes mail state from startup file
- `MailViewer.svelte`: Core email viewing component

#### MailViewer Features

- **Empty State**: Shows "Open EML/MSG File" button when no email loaded
- **Email Header**: Displays subject, from, to, cc, bcc fields
- **PEC Badge**: Shows green badge for Italian certified emails
- **Attachments Bar**: Horizontal scrollable list of attachments with type-specific icons
- **Email Body**: Rendered in sandboxed iframe for security
- **Loading Overlay**: Shows spinner during file loading

#### Attachment Handling

```typescript
// Different handlers based on file type
if (att.contentType.startsWith("image/")) {
    // Opens in built-in or external viewer based on settings
    await OpenImageWindow(base64Data, filename);
} else if (att.filename.toLowerCase().endsWith(".pdf")) {
    // Opens in built-in or external PDF viewer
    await OpenPDFWindow(base64Data, filename);
} else if (att.filename.toLowerCase().endsWith(".eml")) {
    // Opens in new EMLy instance
    await OpenEMLWindow(base64Data, filename);
} else {
    // Download as file
    <a href={dataUrl} download={filename}>...</a>
}
```

### Frontend Mail Utilities (`lib/utils/mail/`)

The frontend email handling code is organized into modular utility files:

| File | Purpose |
|------|---------|
| `index.ts` | Barrel export for all mail utilities |
| `constants.ts` | Constants: `IFRAME_UTIL_HTML`, `CONTENT_TYPES`, `PEC_FILES`, `EMAIL_EXTENSIONS` |
| `data-utils.ts` | Data conversion: `arrayBufferToBase64`, `createDataUrl`, `looksLikeBase64`, `tryDecodeBase64` |
| `attachment-handlers.ts` | Attachment opening: `openPDFAttachment`, `openImageAttachment`, `openEMLAttachment` |
| `email-loader.ts` | Email loading: `loadEmailFromPath`, `openAndLoadEmail`, `processEmailBody`, `isEmailFile` |

#### Key Functions

**Data Utilities** (`data-utils.ts`)
```typescript
// Convert ArrayBuffer to base64 string
function arrayBufferToBase64(buffer: unknown): string;

// Create data URL for file downloads
function createDataUrl(contentType: string, base64Data: string): string;

// Check if string looks like base64 encoded content
function looksLikeBase64(content: string): boolean;

// Attempt to decode base64, returns null on failure
function tryDecodeBase64(content: string): string | null;
```

**Attachment Handlers** (`attachment-handlers.ts`)
```typescript
// Open PDF using built-in or external viewer based on settings
async function openPDFAttachment(base64Data: string, filename: string): Promise<AttachmentHandlerResult>;

// Open image using built-in or external viewer based on settings
async function openImageAttachment(base64Data: string, filename: string): Promise<AttachmentHandlerResult>;

// Open EML attachment in new EMLy window
async function openEMLAttachment(base64Data: string, filename: string): Promise<AttachmentHandlerResult>;
```

**Email Loader** (`email-loader.ts`)
```typescript
// Load email from file path, handles EML/MSG/PEC detection
async function loadEmailFromPath(filePath: string): Promise<LoadEmailResult>;

// Open file dialog and load selected email
async function openAndLoadEmail(): Promise<LoadEmailResult>;

// Process email body (decode base64, fix encoding)
async function processEmailBody(body: string): Promise<string>;

// Check if file path is a valid email file
function isEmailFile(filePath: string): boolean;
```

### Settings Page (`(app)/settings/+page.svelte`)

Organized into cards with various configuration options:

1. **Language Settings**
   - Radio buttons for English/Italian
   - Triggers full page reload on change

2. **Export/Import Settings**
   - Export current settings to JSON file
   - Import settings from JSON file

3. **Preview Page Settings**
   - Supported image types (JPG, JPEG, PNG)
   - Toggle built-in image viewer
   - Toggle built-in PDF viewer

4. **Danger Zone** (Hidden by default, revealed by clicking settings 10 times rapidly)
   - Open DevTools hint
   - Reload application
   - Reset to defaults
   - Debugger protection toggle (disabled in production)
   - Version information display

#### Unsaved Changes Detection

The settings page tracks changes and shows a persistent toast when there are unsaved modifications:

```typescript
$effect(() => {
    const dirty = !isSameSettings(normalizeSettings(form), lastSaved);
    unsavedChanges.set(dirty);
    if (dirty) {
        showUnsavedChangesToast({
            onSave: saveToStorage,
            onReset: resetToLastSaved,
        });
    }
});
```

---

## State Management

EMLy uses a combination of Svelte 5 Runes and traditional Svelte stores.

### Mail State (`stores/mail-state.svelte.ts`)

Uses Svelte 5's `$state` rune for reactive email data:

```typescript
class MailState {
    currentEmail = $state<internal.EmailData | null>(null);

    setParams(email: internal.EmailData | null) {
        this.currentEmail = email;
    }

    clear() {
        this.currentEmail = null;
    }
}

export const mailState = new MailState();
```

### Settings Store (`stores/settings.svelte.ts`)

Manages application settings with localStorage persistence:

```typescript
class SettingsStore {
    settings = $state<EMLy_GUI_Settings>({ ...defaults });
    hasHydrated = $state(false);

    load() { /* Load from localStorage */ }
    save() { /* Save to localStorage */ }
    update(newSettings: Partial<EMLy_GUI_Settings>) { /* Merge and save */ }
    reset() { /* Reset to defaults */ }
}
```

Settings schema:
```typescript
interface EMLy_GUI_Settings {
    selectedLanguage: "en" | "it";
    useBuiltinPreview: boolean;
    useBuiltinPDFViewer: boolean;
    previewFileSupportedTypes: string[];
    enableAttachedDebuggerProtection: boolean;
}
```

### App Store (`stores/app.ts`)

Traditional Svelte writable stores for UI state:

```typescript
export const dangerZoneEnabled = writable<boolean>(false);
export const unsavedChanges = writable<boolean>(false);
export const sidebarOpen = writable<boolean>(true);
export const bugReportDialogOpen = writable<boolean>(false);
export const events = writable<AppEvent[]>([]);
```

---

## Internationalization (i18n)

EMLy uses **ParaglideJS** for compile-time type-safe translations.

### Translation Files

Located in `frontend/messages/`:
- `en.json` - English translations
- `it.json` - Italian translations

### Message Format

```json
{
    "$schema": "https://inlang.com/schema/inlang-message-format",
    "mail_no_email_selected": "No email selected",
    "mail_open_eml_btn": "Open EML/MSG File",
    "settings_title": "Settings",
    "settings_language_english": "English",
    "settings_language_italian": "Italiano"
}
```

### Usage in Components

```typescript
import * as m from "$lib/paraglide/messages";

// In template
<h1>{m.settings_title()}</h1>
<button>{m.mail_open_eml_btn()}</button>
```

### Changing Language

```typescript
import { setLocale } from "$lib/paraglide/runtime";

await setLocale("it", { reload: false });
location.reload(); // Page reload required for full update
```

---

## UI Components

EMLy uses **shadcn-svelte**, a port of shadcn/ui for Svelte. Components are located in `frontend/src/lib/components/ui/`.

### Available Components

| Component | Usage |
|-----------|-------|
| `Button` | Primary buttons with variants (default, destructive, outline, ghost) |
| `Card` | Container with header, content, footer sections |
| `Dialog` | Modal dialogs for bug reports, confirmations |
| `AlertDialog` | Confirmation dialogs with cancel/continue actions |
| `Switch` | Toggle switches for boolean settings |
| `Checkbox` | Multi-select checkboxes |
| `RadioGroup` | Single-select options (language selection) |
| `Label` | Form labels |
| `Input` | Text input fields |
| `Textarea` | Multi-line text input |
| `Separator` | Visual dividers |
| `Sidebar` | Collapsible navigation sidebar |
| `Tooltip` | Hover tooltips |
| `Sonner` | Toast notifications |
| `Badge` | Status badges |
| `Skeleton` | Loading placeholders |

### Custom Components

| Component | Purpose |
|-----------|---------|
| `MailViewer.svelte` | Email display with header, attachments, body |
| `SidebarApp.svelte` | Navigation sidebar with menu items |
| `UnsavedBar.svelte` | Unsaved changes notification bar |

---

## Key Features

### 1. Email Viewing

- Parse and display EML and MSG files
- Show email metadata (from, to, cc, bcc, subject)
- Render HTML email bodies in sandboxed iframe
- List and handle attachments with type-specific actions

### 2. Attachment Handling

- **Images**: Open in built-in viewer or external app
- **PDFs**: Open in built-in viewer or external app
- **EML files**: Open in new EMLy window
- **Other files**: Download directly

### 3. Multi-Window Support

The app can spawn separate viewer windows:
- Image viewer (`--view-image=<path>`)
- PDF viewer (`--view-pdf=<path>`)
- EML viewer (new app instance with file path)

Each viewer has a unique instance ID to allow multiple concurrent viewers.

### 4. Bug Reporting

Complete bug reporting system:
1. Captures screenshot when dialog opens
2. Collects user input (name, email, description)
3. Includes current mail file if loaded
4. Gathers system information
5. Creates ZIP archive in temp folder
6. Checks if the bug report API is online via heartbeat (`CheckBugReportAPI`)
7. If online, attempts to upload to the bug report API server
8. Falls back to local ZIP if server is offline or upload fails
9. Shows server confirmation with report ID, or local path with upload warning

#### Heartbeat Check (`app_heartbeat.go`)

Before uploading a bug report, the app sends a GET request to `{BUGREPORT_API_URL}/health` with a 3-second timeout. If the API doesn't respond with status 200, the upload is skipped entirely and only the local ZIP is created. The `CheckBugReportAPI()` method is also exposed to the frontend for UI status checks.

#### Bug Report API Server

A separate API server (`server/` directory) receives bug reports:
- **Stack**: Bun.js + ElysiaJS + MySQL 8
- **Deployment**: Docker Compose (`docker compose up -d` from `server/`)
- **Auth**: Static API key for clients (`X-API-Key`), static admin key (`X-Admin-Key`)
- **Rate limiting**: HWID-based, configurable (default 5 reports per 24h)
- **Logging**: Structured file logging to `logs/api.log` with format `[date] - [time] - [source] - message`
- **Endpoints**: `POST /api/bug-reports` (client), `GET/DELETE /api/admin/bug-reports` (admin)

#### Bug Report Dashboard

A web dashboard (`dashboard/` directory) for browsing, triaging, and downloading bug reports:
- **Stack**: SvelteKit (Svelte 5) + TailwindCSS v4 + Drizzle ORM + Bun.js
- **Deployment**: Docker service in `server/docker-compose.yml`, port 3001
- **Database**: Connects directly to the same MySQL database via Drizzle ORM (read/write)
- **Features**:
  - Paginated reports list with status filter and search (hostname, user, name, email)
  - Report detail view with metadata, description, system info (collapsible JSON), and file list
  - Status management (new → in_review → resolved → closed)
  - Inline screenshot preview for attached screenshots
  - Individual file download and bulk ZIP download (all files + report metadata)
  - Report deletion with confirmation dialog
  - Dark mode UI matching EMLy's aesthetic
- **Authentication**: Session-based auth with Lucia v3 + Drizzle ORM adapter
  - Default admin account: username `admin`, password `admin` (seeded on first migration)
  - Password hashing with argon2 via `@node-rs/argon2`
  - Session cookies with automatic refresh
  - Role-based access: `admin` and `user` roles
- **User Management**: Admin-only `/users` page for creating/deleting dashboard users
- **Development**: `cd dashboard && bun install && bun dev` (localhost:3001)

#### Configuration (config.ini)

```ini
[EMLy]
BUGREPORT_API_URL="https://your-server.example.com"
BUGREPORT_API_KEY="your-api-key"
```

### 5. Settings Management

- Language selection (English/Italian)
- Built-in viewer preferences
- Supported file type configuration
- Export/import settings as JSON
- Reset to defaults

### 6. Security Features

- Debugger detection and protection
- Sandboxed iframe for email body
- Single-instance lock for main window
- Disabled link clicking in email body

### 7. PEC Support

Special handling for Italian Posta Elettronica Certificata (PEC):
- Detects PEC emails
- Shows signed mail badge
- Handles P7S signature files
- Processes daticert.xml metadata

### 8. Self-Hosted Update System

**Corporate Network Update Management** - No third-party services required:

- **Network Share Integration**: Check for updates from corporate file shares (UNC paths like `\\server\emly-updates`)
- **Version Manifest**: JSON-based version.json controls what versions are available
- **Dual Channel Support**: Separate stable and beta release channels
- **Manual or Automatic**: Users can manually check, or app auto-checks on startup
- **Download & Verify**: Downloads installers from network share with SHA256 checksum verification
- **One-Click Install**: Auto-launches installer with UAC elevation, optionally quits app
- **UI Integration**: Full update UI in Settings page with progress indicators
- **Event-Driven**: Real-time status updates via Wails events

#### Configuration (config.ini)

```ini
[EMLy]
UPDATE_CHECK_ENABLED="true"      # Enable/disable update checking
UPDATE_PATH="\\server\updates"   # Network share or file:// path
UPDATE_AUTO_CHECK="true"         # Check on startup
```

#### Network Share Structure

```
\\server\emly-updates\
├── version.json                      # Update manifest
├── EMLy_Installer_1.5.0.exe          # Stable release installer
└── EMLy_Installer_1.5.1-beta.exe     # Beta release installer
```

#### version.json Format

```json
{
  "stableVersion": "1.5.0",
  "betaVersion": "1.5.1-beta",
  "stableDownload": "EMLy_Installer_1.5.0.exe",
  "betaDownload": "EMLy_Installer_1.5.1-beta.exe",
  "sha256Checksums": {
    "EMLy_Installer_1.5.0.exe": "abc123...",
    "EMLy_Installer_1.5.1-beta.exe": "def456..."
  },
  "releaseNotes": {
    "1.5.0": "Bug fixes and performance improvements",
    "1.5.1-beta": "New feature preview"
  }
}
```

#### Update Flow

1. **Check**: App reads `version.json` from configured network path
2. **Compare**: Compares current version with available version for active channel (stable/beta)
3. **Notify**: If update available, shows toast notification with action button
4. **Download**: User clicks download, installer copied from network share to temp folder
5. **Verify**: SHA256 checksum validated against manifest
6. **Install**: User clicks install, app launches installer with UAC, optionally quits

#### Backend Methods (app_update.go)

| Method | Description |
|--------|-------------|
| `CheckForUpdates()` | Reads manifest from network share, compares versions |
| `DownloadUpdate()` | Copies installer to temp folder, verifies checksum |
| `InstallUpdate(quit)` | Launches installer with UAC elevation |
| `GetUpdateStatus()` | Returns current update system state |
| `loadUpdateManifest(path)` | Parses version.json from network share |
| `compareSemanticVersions(v1, v2)` | Semantic version comparison |
| `verifyChecksum(file, hash)` | SHA256 integrity verification |
| `resolveUpdatePath(base, file)` | Handles UNC paths and file:// URLs |

#### Deployment Workflow for IT Admins

1. **Build new version**: `wails build --upx`
2. **Create installer**: Run Inno Setup with `installer/installer.iss`
3. **Generate checksum**: `certutil -hashfile EMLy_Installer_1.5.0.exe SHA256`
4. **Update manifest**: Edit `version.json` with new version and checksum
5. **Deploy to share**: Copy installer and manifest to `\\server\emly-updates\`
6. **Users notified**: Apps auto-check within 5 seconds of startup (if enabled)

---

## Build & Development

### Prerequisites

- Go 1.21+
- Node.js 18+ or Bun
- Wails CLI v2

### Development

```bash
# Install frontend dependencies
cd frontend && bun install

# Run in development mode
wails dev
```

### Building

```bash
# Build for Windows
wails build -platform windows/amd64

# Output: build/bin/EMLy.exe
```

### Configuration

`wails.json` configures the build:
```json
{
    "name": "EMLy",
    "frontend:install": "bun install",
    "frontend:build": "bun run build",
    "frontend:dev:watcher": "bun run dev",
    "fileAssociations": [
        {
            "ext": "eml",
            "name": "Email Message",
            "description": "EML File"
        }
    ]
}
```

### File Association

The app registers as a handler for `.eml` files via Windows file associations, configured in `wails.json`.

---

## Wails Bindings

Wails automatically generates TypeScript bindings for Go functions. These are located in `frontend/src/lib/wailsjs/`.

### Generated Files

- `go/main/App.ts` - TypeScript functions calling Go methods
- `go/models.ts` - TypeScript types for Go structs
- `runtime/runtime.ts` - Wails runtime functions

### Usage Example

```typescript
import { ReadEML, ShowOpenFileDialog } from "$lib/wailsjs/go/main/App";
import type { internal } from "$lib/wailsjs/go/models";

// Open file dialog
const filePath = await ShowOpenFileDialog();

// Parse email
const email: internal.EmailData = await ReadEML(filePath);
```

### Runtime Events

Wails provides event system for Go-to-JS communication:

```typescript
import { EventsOn } from "$lib/wailsjs/runtime/runtime";

// Listen for second instance launch
EventsOn("launchArgs", (args: string[]) => {
    // Handle file opened from second instance
});
```

---

## Logging

EMLy uses structured JSON logging based on Go's `log/slog` standard library.

### Backend (`backend/logger/`)

- **Output**: Simultaneous JSON output to `stdout` and `%APPDATA%/EMLy/logs/app.log`
- **Log Levels**: Configurable via `LOG_LEVEL` in `config.ini` (DEBUG, INFO, WARN, ERROR)
- **Structured Fields**: All log entries include timestamp, level, source file, and key-value attributes

```go
import pkglogger "emly/backend/logger"

pkglogger.Info("email loaded", "path", filePath, "format", "eml")
pkglogger.Error("parse failed", "error", err.Error(), "path", filePath)
pkglogger.Debug("attachment details", "count", len(attachments))
```

#### Canonical Log Lines

Every Wails-bound function emits a canonical log line at completion with function name, duration, and status:

```go
func (a *App) ReadEML(filePath string) (data *internal.EmailData, err error) {
    start := time.Now()
    defer func() { canonicalLog("ReadEML", start, err) }()
    return internal.ReadEmlFile(filePath)
}
// Output: {"level":"INFO","msg":"canonical_line","function":"ReadEML","duration_ms":42,"status":"success"}
```

#### Sensitive Data Redaction

Use `logger.Redacted` to mask passwords, API keys, and tokens in log output:

```go
slog.Any("api_key", pkglogger.Redacted(apiKey)) // logs "[REDACTED]"
pkglogger.RedactStruct(configMap)                // redacts known sensitive keys
```

### Frontend (`lib/utils/logger.ts`)

Structured logger service that sends logs to the Go backend via the `FrontendLog` Wails binding. Each entry includes browser context (URL, user agent).

```typescript
import { logger } from '$lib/utils/logger';

logger.info('email loaded', { filePath: '/tmp/test.eml' });
logger.error('failed to parse', { error: err.message });
```

### Console Hook (`lib/utils/logger-hook.ts`)

Intercepts `console.log/warn/error/info` and forwards them to the backend for unified logging. Called once at app startup via `setupConsoleLogger()`.

---

## Error Handling

### Frontend

Toast notifications for user-facing errors:
```typescript
import { toast } from "svelte-sonner";
import * as m from "$lib/paraglide/messages";

try {
    await ReadEML(filePath);
} catch (error) {
    toast.error(m.mail_error_opening());
}
```

### Backend

Errors are returned to frontend and logged:
```go
func (a *App) ReadEML(filePath string) (data *internal.EmailData, err error) {
    start := time.Now()
    defer func() { canonicalLog("ReadEML", start, err) }()
    logMailFileInfo("ReadEML", filePath)
    data, err = internal.ReadEmlFile(filePath)
    if err == nil && data != nil {
        logParsedMailInfo("ReadEML", data)
    }
    return data, err
}
```

When `LOG_LEVEL=DEBUG`, every mail loading call emits two debug entries:
1. **`loading mail file`** — file name, extension, size in bytes
2. **`mail parsed successfully`** — subject (truncated), from, to/cc count, body type (html/text/none), body length, attachment count, unique attachment MIME types, PEC flag, inner email flag

---

## Debugging

### Development Mode

In dev mode (`wails dev`):
- Hot reload enabled
- Debug logs visible
- DevTools accessible via Ctrl+Shift+F12
- Danger Zone always visible in settings

### Production

- Debugger protection can terminate app if debugger detected
- Danger Zone hidden by default
- Access Danger Zone by clicking settings link 10 times within 4 seconds

---

## Dashboard Features

### ZIP File Upload

The dashboard supports uploading `.zip` files created by EMLy's `SubmitBugReport` feature when the API upload fails. Accessible via the "Upload ZIP" button on the reports list page, it parses `report.txt` (name, email, description), `system_info.txt` (hostname, OS, HWID, IP), and imports all attached files (screenshots, mail files, localStorage, config) into the database as a new bug report.

**API Endpoint**: `POST /api/reports/upload` - Accepts multipart form data with a `.zip` file.

### User Enable/Disable

Admins can temporarily disable user accounts without deleting them. Disabled users cannot log in and active sessions are invalidated. The `user` table has an `enabled` BOOLEAN column (default TRUE). Toggle is available in the Users management page. Restrictions: admins cannot disable themselves or other admin users.

### Active Users / Presence Tracking

Real-time presence tracking using Server-Sent Events (SSE). Connected users are tracked in-memory with heartbeat updates every 15 seconds. The layout header shows avatar indicators for other active users with tooltips showing what they're viewing. The report detail page shows who else is currently viewing the same report.

**Endpoints**:
- `GET /api/presence` - SSE stream for real-time presence updates
- `POST /api/presence/heartbeat` - Client heartbeat with current page/report info

**Client Store**: `$lib/stores/presence.svelte.ts` - Svelte 5 reactive store managing SSE connection and heartbeats.

---

## License & Credits

EMLy is developed by FOISX @ 3gIT.