# EMLy — UI Style Guide

Visual style reference for agents making UI changes to EMLy. Derived by running
`wails dev --tags "debug" --viteservertimeout 60` and screenshotting the app
(main view, settings, credits, image/PDF viewer windows, and several dialogs)
in both dark and light mode.

EMLy is a Wails v2 desktop app: Go backend + SvelteKit (Svelte 5) frontend,
styled with Tailwind CSS v4 and shadcn-svelte components (`baseColor: slate`,
see `frontend/components.json`). Design tokens live in
`frontend/src/routes/layout.css`.

## Overall aesthetic

Flat, minimal, "OS-native" utility tool. No gradients, no shadows-as-decoration,
no illustration/marketing style. Pure black/white neutral palette (OKLCH
`0 0 0` chroma) with color reserved almost entirely for semantic state
(destructive/red, success/green via toasts). Generous rounded corners
(`--radius: 0.625rem` ≈ 10px, cards often visually read as ~14px), thin
1px borders using a low-opacity foreground color rather than a distinct
gray, and a dense, information-forward layout (no card is purely
decorative — every panel groups a specific setting or piece of data).

Custom title bar (no native OS chrome) with app icon + name on the left,
version badge, a red "DEBUG BUILD" pill when applicable, and minimize/
maximize/close on the right. Title bar is drag-region (`-webkit-app-region:
drag`).

## Color tokens

Defined as CSS custom properties in `layout.css`, consumed via Tailwind
`@theme inline` (`bg-background`, `text-foreground`, `border-border`,
etc.) and by shadcn-svelte primitives. All colors are OKLCH.

Dark mode (`.dark`, the app's default — see `frontend/src/app.html`,
`localStorage["emly_theme"]`, defaults to `"dark"`):

| Token | Value | Use |
|---|---|---|
| `--background` | `oklch(0 0 0)` (pure black) | App background |
| `--foreground` | `oklch(0.985 0 0)` (near white) | Body text |
| `--card` | `oklch(0.205 0 0)` (dark gray) | Card / panel surfaces |
| `--card-foreground` | `oklch(0.985 0 0)` | Text on cards |
| `--popover` | `oklch(0.205 0 0)` | Dialog/dropdown surfaces |
| `--primary` | `oklch(0.922 0 0)` (light gray) | Primary buttons |
| `--primary-foreground` | `oklch(0.205 0 0)` | Text on primary buttons |
| `--secondary` / `--muted` / `--accent` | `oklch(0.269 0 0)` | Subtle fills, hover states |
| `--muted-foreground` | `oklch(0.708 0 0)` | Secondary/help text |
| `--destructive` | `oklch(0.704 0.191 22.216)` (red) | Danger buttons, error banners, delete actions |
| `--border` | `oklch(1 0 0 / 10%)` (white @ 10%) | Card/input borders — very subtle |
| `--input` | `oklch(1 0 0 / 15%)` | Input borders |
| `--ring` | `oklch(0.556 0 0)` | Focus ring |

Light mode (`:root`, opt-in via Settings → Aspetto → Chiaro): background
and card both pure white (`oklch(1 0 0)`), foreground near-black
(`oklch(0.145 0 0)`), border a light gray (`oklch(0.922 0 0)`) instead of
translucent white. Same structure, same component shapes — it's a direct
palette swap, not a different layout. Destructive red is slightly more
saturated/darker in light mode (`oklch(0.577 0.245 27.325)`).

Both themes are driven purely by toggling the `.dark` class on `<html>`;
never hardcode a light or dark color in a component — always use the
CSS variables / Tailwind semantic classes so components work in both.

## Typography

`font-family: system-ui, -apple-system, sans-serif` (see `app.html`) — no
custom webfont. Headings are bold and notably larger than body text (page
titles ~28px/`text-2xl` bold), body copy is small (~14px) and
muted-foreground for helper/description text under every setting. Monospace
is used for file paths / config values (e.g. `config.ini`, the downloads
folder path) via an inline `<code>`-style chip with `--muted` background.

## Layout shell

- Fixed-width left sidebar (~248px): app icon + "EMLy by 3gIT" wordmark at
  top, a "Menu" section label, then nav items with icon + label
  (Visualizza Mail / Impostazioni / Crediti). Active/hover items get a
  subtle background tint, not a colored accent bar.
- Sidebar is collapsible (toggle icon bottom-left status bar).
- Main content area (`main`) is the only scrollable region
  (`overflow: auto`); everything else uses `overflow: hidden` — long
  settings pages scroll internally while sidebar and status bar stay
  fixed.
- A thin status bar pinned to the bottom of the window: left side has
  small utility icons (sidebar toggle, mail, settings, info); right side
  has circular action buttons in destructive-red pill styling (update/
  reload icon, bug-report icon).
- Content pages follow a consistent header pattern: large bold `<h1>`
  title, one-line muted description underneath, and (on sub-pages) a
  "‹ Indietro" (back) link top-right.

## Cards / panels

Every settings group and content block is a `bg-card` panel with
`border border-border`, rounded corners, and internal padding — stacked
vertically with consistent gaps. Cards do not use drop shadows; separation
comes entirely from the subtle border + slightly lighter card background
against the pure-black page background. A row inside a card (label + control)
uses `flex items-center justify-between`, with an `<Info>`-style helper line
below in `muted-foreground`, prefixed with a bold **Info:** label.

"Danger zone" cards (irreversible actions — reset app data, safety check,
crash-handler test buttons) get a red-tinted treatment: `border-destructive/
30`, faint reddish card background, and destructive-variant buttons/badges,
clearly visually distinct from the neutral cards above them ("Zona
Pericolosa" section in Settings).

## Controls

- Buttons: shadcn-svelte `Button`, small (~24–36px tall), rounded, no
  all-caps, no heavy font-weight. Variants seen: default (light-gray fill
  in dark mode / near-black in light mode), `outline` (border only, used
  for the empty-state "Apri File EML/MSG" action), and `destructive`
  (solid red, white text) for anything dangerous or attention-grabbing
  (bug report, reload UI, reset data, safety check).
- Radio buttons (language, theme picker): plain circular radio + icon +
  label in a horizontal row, no card-per-option — just a vertical list of
  radio rows under one heading.
- Toggle switches (shadcn `Switch`): track is `muted`, thumb is white/
  foreground-colored, standard iOS-style sliding toggle, right-aligned in
  its row.
- Checkboxes (e.g. supported image types): square, small, filled/checked
  state uses foreground color with a check glyph.
- Text inputs (e.g. download folder path): full-width, `bg-input`-ish
  transparent fill, thin border, monospace-leaning content, paired with an
  adjacent outline button ("Seleziona cartella").

## Dialogs / modals

shadcn-svelte `Dialog`/`AlertDialog` primitives: centered overlay, dimmed
+ blurred backdrop, `bg-popover` panel with border and rounded corners,
bold title + one-line muted description at top, content in the middle,
action buttons bottom-right (secondary/cancel on the left, primary or
destructive on the right — e.g. "Annulla" / "Invia Segnalazione"). A small
`✕` close affordance sits top-right on dialogs that have one (not all do).

Observed dialogs:
- **Host integrity check failed** (`Controllo integrità dispositivo
  fallito`) — safety/warning dialog, red shield icon + red heading, single
  "Capito" acknowledge button. Can appear unprompted on startup/navigation
  when the device fails an integrity check.
- **Segnala un Bug** (BugReportDialog) — form dialog: Nome / Email text
  inputs, multi-line Descrizione textarea, an auto-attached screenshot
  thumbnail preview, "Annulla" (outline) + "Invia Segnalazione" (default)
  buttons. In debug mode the fields are pre-filled with test data.
- **Controllo di sicurezza** (SafetyCheckDialog, triggered from the Danger
  Zone) — shows an inline spinner + "Controllo in corso…" status while
  running, then "Esegui di nuovo" (outline) + "Chiudi" (destructive) once
  finished.

Not captured in this pass (need a loaded email to trigger): 
`MailDebugInfoDialog`, `LinkConfirmationDialog` (link-open safety
confirmation from within an email body).

## Toasts

`svelte-sonner`, bottom-right stacked, dark rounded pill/card with an icon
+ message, auto-dismiss. Success toasts use a check-circle icon
("Impostazioni salvate!"); there's also a whimsical easter-egg toast
("Qui ci sono i draghi!" — flame icon) that appears on the settings page.
Multiple toasts stack vertically without overlapping.

## Secondary windows

EMLy opens dedicated child windows for attachments instead of embedding
everything in the main window:
- **EMLy Viewer** (image viewer) — near-black (`#1e1e1e`-ish) toolbar with
  "Visualizzatore Immagini" label left, icon-only action buttons right
  (download, zoom in/out, rotate L/R, fit-to-window). Empty state is a
  centered red-outlined pill: "Nessun dato immagine fornito".
- **EMLy PDF Viewer** — pure black, minimal chrome, just a title-bar label
  and centered red status text when no PDF is loaded ("Nessun dato PDF
  fornito. Apri questa finestra dall'applicazione principale EMLy.").

Both reuse the same custom title-bar pattern as the main window (draggable,
native min/max/close on the right) and the same red-for-error-state
convention as the main app.

## Icons

`@lucide/svelte` throughout — consistent stroke-style line icons, no
filled/mixed icon sets. Sized small (16px inline, ~20–24px for toolbar
buttons).

## Localization

UI is fully localized via Paraglide (`$lib/paraglide/messages`); Italian
and English are supported (seen as `it`/`en` with flag icons in the
language picker). Screenshots in this pass were taken with the app in
Italian — don't be surprised by Italian copy in other design references.

## Practical notes for agents

- Don't invent new colors — extend the existing `--*` token set in
  `layout.css` (and its `.dark` override) if a new semantic color is
  genuinely needed, rather than hardcoding hex/oklch values in a
  component.
- Reuse `frontend/src/lib/components/ui/*` (shadcn-svelte primitives)
  before writing bespoke dialog/button/input markup.
- New settings should follow the existing card → row → helper-text
  pattern in `routes/(app)/settings/+page.svelte`, and dangerous actions
  belong in the red-tinted "Zona Pericolosa" section, not mixed in with
  normal settings.
- Preserve the flat/no-shadow, high-contrast, borderless-except-1px-hairline
  look — avoid adding drop shadows, gradients, or saturated accent colors
  outside of the destructive/error palette.
