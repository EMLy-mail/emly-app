# Validazione semver di GUI_SEMVER in LoadConfig, con chiusura app su formato non valido

## Problema

`config.ini`/`config.debug.ini` contengono `GUI_SEMVER`, la versione corrente di EMLy (usata in `main.go` per lo User-Agent, in `backend/utils/updateripc/version.go` per i confronti di compatibilità col protocollo updater, e restituita al frontend via binding). Il valore non è mai validato: se un installer rotto, un edit manuale o un downgrade parziale scrivono un valore non-semver (es. stringa vuota, `"2.0"`, `"latest"`), il problema si propaga silenziosamente in tutti i consumer — User-Agent sporco, confronti di versione che falliscono in modi poco chiari, bug report con versione errata.

## Obiettivo e scope

Validare `EMLy.GUISemver` con [`golang.org/x/mod/semver`](https://pkg.go.dev/golang.org/x/mod/semver) (`semver.IsValid`) dentro `LoadConfig` (`backend/utils/ini-reader.go`), che è il punto centrale da cui passano tutte le letture di config — sia all'avvio (`main.go`, prima di `wails.Run`) sia ogni reload a runtime via i binding esposti al frontend (`app.go:GetConfig/ReloadEMLyConfig`, `app_settings.go:ReloadConfig`, `app_heartbeat.go`, `tray.go`, `backend/utils/machine-identifier.go`).

Se il valore non è un semver valido, l'app mostra una MessageBox nativa d'errore e chiude il processo. Scope esplicitamente confermato con l'utente: il check vale sia allo startup sia a ogni reload runtime/binding — non solo al primo avvio.

Fuori scope: validazione di `SDK_DECODER_SEMVER` (non è "la versione attuale" di EMLy, è la versione dell'SDK di parsing mail) e qualunque logica di auto-fix/migrazione del config.ini.

## Approccio scelto

1. **Dipendenza**: aggiungere `golang.org/x/mod` a `go.mod` (solo il subpackage `semver` verrà usato) e vendorizzarla (`go mod vendor`), seguendo il pattern vendor già in uso nel repo.
2. **Validazione**: `x/mod/semver.IsValid` richiede il prefisso `v` (`"2.0.1"` → non valido, `"v2.0.1"` → valido), mentre `GUI_SEMVER` in config.ini è scritto senza prefisso (vedi `config.ini:4`, `GUI_SEMVER = 2.0.1`). `LoadConfig` normalizza aggiungendo `v` solo per la validazione (il valore salvato in `Config.EMLy.GUISemver` resta invariato, senza prefisso, per non rompere gli altri consumer come `updateripc/version.go` che si aspettano il formato attuale).
3. **Messagebox nativa**: il primo check avviene in `main.go` prima di `wails.Run`, quindi senza `context.Context` Wails disponibile — `runtime.MessageDialog` non è utilizzabile in quel punto. Si usa una MessageBox Win32 nativa via `syscall`/`user32.dll` (stesso pattern già presente in `backend/utils/screenshot_windows.go`), così il comportamento è identico sia al primo avvio sia nei reload successivi (dove pure sarebbe disponibile `runtime.MessageDialog`, ma usare due meccanismi diversi a seconda del call site aggiungerebbe complessità senza benefici).
4. **Chiusura**: dopo la conferma della messagebox, `os.Exit(1)`. Nessun errore viene ritornato al chiamante di `LoadConfig` in questo caso — il processo termina direttamente, quindi tutti i call site (inclusi quelli che oggi ignorano l'errore di `LoadConfig`, es. `main.go:32`) chiudono l'app in modo uniforme senza bisogno di gestire un nuovo caso d'errore.

## File coinvolti

- `go.mod`, `go.sum` — nuova dipendenza diretta `golang.org/x/mod`.
- `vendor/golang.org/x/mod/semver/`, `vendor/modules.txt` — vendoring via `go mod vendor`.
- `backend/utils/ini-reader.go`:
  - `LoadConfig`: dopo `cfg.MapTo(config)`, valida `config.EMLy.GUISemver` con `semver.IsValid("v" + config.EMLy.GUISemver)`; se non valido, chiama il nuovo helper e poi `os.Exit(1)`.
  - Nuovo helper interno (es. `fatalInvalidGUIVersion(version string)`) che apre `user32.dll`, chiama `MessageBoxW` con flag `MB_OK | MB_ICONERROR` (icona di errore) con titolo `"EMLy - Errore di configurazione"` e messaggio in italiano che riporta il valore non valido e invita a correggere `GUI_SEMVER` in `config.ini`.

## Edge case

1. **Stringa vuota** (`GUI_SEMVER` mancante/vuoto) → `"v" + "" = "v"` → `semver.IsValid` ritorna `false` → chiude l'app. Coerente con la richiesta letterale ("se la versione attuale non è in formato semver").
2. **`ini.Load` fallisce prima** (file mancante, sintassi ini rotta) → invariato, `LoadConfig` ritorna l'errore esistente *prima* di arrivare alla validazione della versione; nessuna messagebox aggiuntiva in questo caso (il chiamante gestisce già questo `error`).
3. **Call site che oggi ignorano l'errore** (`main.go:32`, `main.go:112`, `tray.go:128`, usano `if cfg, err := ...; err == nil && cfg != nil`) → non cambia: con versione non valida il processo termina dentro `LoadConfig` stesso, prima che il controllo torni al chiamante.
4. **`machine-identifier.go:309`** chiama `LoadConfig` per un uso non legato alla versione (raccolta info macchina) — riceve comunque lo stesso guard centralizzato; accettato per design (singolo punto di validazione).
5. **Formato storico non-semver puro** (`updateripc/version.go` usa `"1.2.0b"` come costante per l'updater, non per `GUI_SEMVER`) → fuori scope, non è il valore validato qui.

## Verifica

`semver.IsValid` è puro e testabile senza toccare il path della messagebox/`os.Exit`. Aggiungere `backend/utils/ini-reader_test.go` con test tabellare sulla logica di normalizzazione+validazione (estratta in una funzione pura, es. `isValidSemver(version string) bool`, usata sia da `LoadConfig` sia dal test) coprendo: `"2.0.1"` valido, `"2.0.1-beta.1"` valido, `""` non valido, `"2.0"` non valido, `"latest"` non valido, `"v2.0.1"` (già con prefisso, non dovrebbe capitare da ini ma non deve rompere) valido.

Il path messagebox+`os.Exit` non è testabile via `go test` (termina il processo) — verifica manuale: impostare `GUI_SEMVER` a un valore non valido in `config.debug.ini`, avviare l'app in debug, confermare comparsa messagebox italiana e chiusura immediata; ripetere modificando il valore da Settings a runtime (se il flusso lo permette) per validare il reload path.
