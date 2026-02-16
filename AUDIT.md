# EMLy Security Audit

**Date:** 2026-02-16

**Scope:** Main EMLy desktop application (Go backend + SvelteKit frontend). Server directory excluded.

---

## Critical (2)

### CRIT-1: API Key Committed to Repository
**File:** `config.ini:11`

`BUGREPORT_API_KEY` is in a tracked file and distributed with the binary. It is also returned to the frontend via `GetConfig()` and included in every bug report's `configData` field. Anyone who inspects the installed application directory, the repository, or the binary can extract this key.

**Risk:** Unauthorized access to the bug report API; potential abuse of any API endpoints authenticated by this key.

**Recommendation:** Rotate the key immediately. Stop distributing it in `config.ini`. Source it from an encrypted credential store or per-user environment variable. Strip it from the `GetConfig()` response to the frontend.

---

### CRIT-2: Path Traversal via Attachment Filename
**Files:** `app_viewer.go:83,153,223,285,321`

Email attachment filenames are used unsanitized in temp file paths. A malicious email could craft filenames like `../../malicious.exe` or absolute paths. `OpenPDFWindow` (line 223) is the worst offender: `filepath.Join(tempDir, filename)` with no timestamp prefix at all.

```go
// OpenPDFWindow — bare filename, no prefix
tempFile := filepath.Join(tempDir, filename)

// OpenImageWindow — timestamp prefix but still unsanitized
tempFile := filepath.Join(tempDir, fmt.Sprintf("%s_%s", timestamp, filename))
```

**Risk:** Overwriting arbitrary temp files; potential privilege escalation if a writable autorun target path can be hit.

**Recommendation:** Sanitize attachment filenames with `filepath.Base()` + a character allowlist `[a-zA-Z0-9._-]` before using them in temp paths.

---

## High (5)

### HIGH-1: Command Injection in `OpenURLInBrowser`
**File:** `app_system.go:156-159`

```go
func (a *App) OpenURLInBrowser(url string) error {
    cmd := exec.Command("cmd", "/c", "start", "", url)
    return cmd.Start()
}
```

Passes unsanitized URL to `cmd /c start`. A `file:///` URL or shell metacharacters (`&`, `|`) can execute arbitrary commands.

**Risk:** Arbitrary local file execution; command injection via crafted URL.

**Recommendation:** Validate that the URL uses `https://` scheme before passing it. Consider using `rundll32.exe url.dll,FileProtocolHandler` instead of `cmd /c start`.

---

### HIGH-2: Unsafe Path in `OpenFolderInExplorer`
**File:** `app_system.go:143-146`

```go
func (a *App) OpenFolderInExplorer(folderPath string) error {
    cmd := exec.Command("explorer", folderPath)
    return cmd.Start()
}
```

Raw frontend string passed to `explorer.exe` with no validation. This is a public Wails method callable from any frontend code.

**Risk:** Unexpected explorer behavior with crafted paths or UNC paths.

**Recommendation:** Validate that `folderPath` is a local directory path and exists before passing to explorer.

---

### HIGH-3: Iframe Sandbox Escape — Email Body XSS
**File:** `frontend/src/lib/components/MailViewer.svelte:387`

```svelte
<iframe
  srcdoc={mailState.currentEmail.body + iframeUtilHtml}
  sandbox="allow-same-origin allow-scripts"
/>
```

`allow-scripts` + `allow-same-origin` together allow embedded email scripts to remove the sandbox attribute entirely and access the parent Wails window + all Go backend bindings. MDN explicitly warns against this combination.

**Risk:** Full XSS in the Wails WebView context; arbitrary Go backend method invocation from a malicious email.

**Recommendation:** Remove `allow-same-origin` from the iframe sandbox. Replace `iframeUtilHtml` script injection with a `postMessage`-based approach from the parent so `allow-scripts` can also be removed entirely.

---

### HIGH-4: Arbitrary Code Execution via `InstallUpdateSilentFromPath`
**File:** `app_update.go:573-643`

This exposed Wails method accepts an arbitrary UNC/local path from the frontend, copies the binary to temp, and executes it with UAC elevation (`/ALLUSERS`). There is no signature verification, no path allowlist, and no checksum validation.

**Risk:** Any attacker who can call this method (e.g., via XSS from HIGH-3) can execute any binary with administrator rights.

**Recommendation:** Restrict to validated inputs — check that installer paths match a known allowlist or have a valid code signature before execution.

---

### HIGH-5: Race Condition on `updateStatus`
**File:** `app_update.go:55-65`

```go
var updateStatus = UpdateStatus{ ... }
```

Global mutable variable accessed from multiple goroutines (startup check goroutine, frontend calls to `CheckForUpdates`, `DownloadUpdate`, `GetUpdateStatus`, `InstallUpdateSilent`) without any mutex protection. TOCTOU races possible on `Ready`/`InstallerPath` fields.

**Risk:** Installing from an empty path; checking stale ready status; data corruption.

**Recommendation:** Protect `updateStatus` with a `sync.RWMutex` or replace with an atomic struct/channel-based state machine.

---

## Medium (7)

### MED-1: API Key Leaked in Bug Reports
**Files:** `frontend/src/lib/components/BugReportDialog.svelte:92-101`, `logger.go:72-81`

`captureConfig()` calls `GetConfig()` and serializes the entire config including `BUGREPORT_API_KEY` into `configData`. This is sent to the remote API, written to `config.json` in the temp folder, and included in the zip. The `FrontendLog` function also logs all frontend output verbatim — any accidental `console.log(config)` would write the key to the log file.

**Recommendation:** Filter out `BUGREPORT_API_KEY` before serializing config data. Redact sensitive fields in `FrontendLog`.

---

### MED-2: No TLS Validation on API Requests
**Files:** `app_heartbeat.go:28-37`, `app_bugreport.go:376-384`

Both HTTP clients use the default transport with no certificate pinning and no enforcement of minimum TLS versions. The API URL from `config.ini` is not validated to be HTTPS before making requests. Bug report uploads contain PII (name, email, hostname, HWID, screenshot, email file) and the API key header.

**Recommendation:** Validate that `apiURL` starts with `https://`. Consider certificate pinning for the bug report API.

---

### MED-3: Raw Frontend String Written to Disk
**File:** `app_settings.go:31-63`

`ExportSettings` writes the raw `settingsJSON` string from the frontend to any user-chosen path with no content validation. A compromised frontend (e.g., via HIGH-3 XSS) could write arbitrary content.

**Recommendation:** Validate that `settingsJSON` is well-formed JSON matching the expected settings schema before writing.

---

### MED-4: Imported Settings Not Schema-Validated
**Files:** `app_settings.go:73-100`, `frontend/src/lib/stores/settings.svelte.ts:37`

Imported settings JSON is merged into the settings store via spread operator without schema validation. An attacker-supplied settings file could manipulate `enableAttachedDebuggerProtection` or inject unexpected values.

**Recommendation:** Validate imported JSON against the `EMLy_GUI_Settings` schema. Reject unknown keys.

---

### MED-5: `isEmailFile` Accepts Any String
**File:** `frontend/src/lib/utils/mail/email-loader.ts:42-44`

```typescript
export function isEmailFile(filePath: string): boolean {
  return filePath.trim().length > 0;
}
```

Any non-empty path passes validation and is sent to the Go backend for parsing, including paths to executables or sensitive files.

**Recommendation:** Check file extension against `EMAIL_EXTENSIONS` before passing to backend.

---

### MED-6: PATH Hijacking via `wmic` and `reg`
**File:** `backend/utils/machine-identifier.go:75-99`

`wmic` and `reg.exe` are resolved via PATH. If PATH is manipulated, a malicious binary could be executed instead. `wmic` is also deprecated since Windows 10 21H1.

**Recommendation:** Use full paths (`C:\Windows\System32\wbem\wmic.exe`, `C:\Windows\System32\reg.exe`) or replace with native Go syscalls/WMI COM interfaces.

---

### MED-7: Log File Grows Unboundedly
**File:** `logger.go:35`

The log file is opened in append mode with no size limit, rotation, or truncation. Frontend console output is forwarded to the logger, so verbose emails or a tight log loop can fill disk.

**Recommendation:** Implement log rotation (e.g., max 10MB, keep 3 rotated files) or use a library like `lumberjack`.

---

## Low (7)

### LOW-1: Temp Files Written with `0644` Permissions
**Files:** `app_bugreport.go`, `app_viewer.go`, `app_screenshot.go`

All temp files (screenshots, mail copies, attachments) are written with `0644`. Sensitive email content in predictable temp paths (`emly_bugreport_<timestamp>`) could be read by other processes.

**Recommendation:** Use `0600` for temp files containing sensitive content.

---

### LOW-2: Log Injection via `FrontendLog`
**File:** `logger.go:72-81`

`level` and `message` are user-supplied with no sanitization. Newlines in `message` can inject fake log entries. No rate limiting.

**Recommendation:** Strip newlines from `message`. Consider rate-limiting frontend log calls.

---

### LOW-3: `OpenPDFWindow` File Collision
**File:** `app_viewer.go:223`

Unlike other viewer methods, `OpenPDFWindow` uses the bare filename with no timestamp prefix. Two PDFs with the same name silently overwrite each other.

**Recommendation:** Add a timestamp prefix consistent with the other viewer methods.

---

### LOW-4: Single-Instance Lock Exposes File Path
**File:** `main.go:46-50`

Lock ID includes the full file path, which becomes a named mutex visible system-wide. Other processes can enumerate it to discover what files are being viewed.

**Recommendation:** Hash the file path before using it in the lock ID.

---

### LOW-5: External IP via Unauthenticated HTTP
**File:** `backend/utils/machine-identifier.go:134-147`

External IP fetched from `api.ipify.org` without certificate pinning. A MITM can spoof the IP. The request also reveals EMLy usage to the third-party service.

**Recommendation:** Consider making external IP lookup optional or using multiple sources.

---

### LOW-6: `GetConfig()` Exposes API Key to Frontend
**File:** `app.go:150-158`

Public Wails method returns the full `Config` struct including `BugReportAPIKey`. Any frontend JavaScript can retrieve it.

**Recommendation:** Create a `GetSafeConfig()` that omits sensitive fields, or strip the API key from the returned struct.

---

### LOW-7: Attachment Filenames Not Sanitized in Zip
**File:** `app_bugreport.go:422-465`

Email attachment filenames copied into the bug report folder retain their original names (possibly containing traversal sequences). These appear in the zip archive sent to the server.

**Recommendation:** Sanitize filenames with `filepath.Base()` before copying into the bug report folder.

---

## Info (4)

### INFO-1: `allow-same-origin` Could Be Removed from Iframe
**File:** `frontend/src/lib/components/MailViewer.svelte:387`

If `iframeUtilHtml` script injection were replaced with `postMessage`, both `allow-scripts` and `allow-same-origin` could be removed entirely.

### INFO-2: Unnecessary `cmd.exe` Shell Invocation
**File:** `app_system.go:92-94`

`ms-settings:` URIs can be launched via `rundll32.exe url.dll,FileProtocolHandler` without invoking `cmd.exe`, reducing shell attack surface.

### INFO-3: `unsafe.Pointer` Without Size Guards
**Files:** `backend/utils/file-metadata.go:115`, `backend/utils/screenshot_windows.go:94-213`

Cast to `[1 << 20]uint16` array and slicing by `valLen` is potentially out-of-bounds if the Windows API returns an unexpected length.

### INFO-4: No Memory Limits on Email Parsing
**Files:** `backend/utils/mail/mailparser.go`, `eml_reader.go`

All email parts read into memory via `io.ReadAll` with no size limit. A malicious `.eml` with a gigabyte-sized attachment would exhaust process memory. Consider `io.LimitReader`.
