# EMLy

EMLy is a lightweight, cross-platform EML (Email Message) viewer built with [Wails](https://wails.io/), Go, and Svelte. It allows users to open and view `.eml` files, including support for viewing attachments like images and PDFs.

## Requirements

Before developing or building EMLy, ensure you have the following installed:

*   **Go**: Version 1.24 or later (Project uses 1.24.4) - [Download](https://go.dev/dl/)
*   **Wails CLI**: Version v2 - [Installation Guide](https://wails.io/docs/gettingstarted/installation)
*   **Node.js**: Required for the frontend build tools.
*   **Bun**: The project uses `bun` as the frontend package manager. - [Install Bun](https://bun.sh/)
    *   Install via: `powershell -c "irm bun.sh/install.ps1 | iex"` (Windows) or `curl -fsSL https://bun.sh/install | bash` (macOS/Linux).

## Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/yourusername/EMLy.git
    cd EMLy
    ```

2.  Install frontend dependencies:
    ```bash
    cd frontend
    bun install
    cd ..
    ```

    *Note: The Wails build process is configured to handle `bun install` automatically if you run `wails build`, but running it manually ensures your IDE is happy.*

## Live Development

To run the application in development mode with hot-reload:

```bash
wails dev
```

*   This starts the Go backend and a Vite development server for the frontend.
*   The application window will open automatically.
*   You can also access the frontend in a browser at `http://localhost:34115` to debug with standard browser devtools (though backend calls will still route to the running Go process).

## Building

To build a redistributable, production-ready binary:

```bash
wails build
```

The output binary will be located in the `build/bin` directory.

### Cross-Compilation & Installers

The project contains configuration for building installers:
*   **Windows**: NSIS script located in `build/windows/installer/project.nsi`.
*   **MacOS**: Bundle configuration in `build/darwin/`.

To build with specific platform targets (if your environment supports it):

```bash
wails build -platform windows/amd64
```

## Testing

### Frontend Tests
Run the Svelte check to ensure type safety and valid Svelte code:

```bash
cd frontend
bun run check
```

### Backend Tests
To run Go unit tests (if any are added to `backend/`) or standard Go tests:

```bash
go test ./...
```

