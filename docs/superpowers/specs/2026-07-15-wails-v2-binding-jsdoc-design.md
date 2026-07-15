# Nomi parametro reali + JSDoc nei binding Wails v2 generati

## Problema

Wails v2 genera i binding TS/JS (`frontend/src/lib/wailsjs/go/main/App.d.ts`, `App.js`) usando reflection runtime pura, in `vendor/github.com/wailsapp/wails/v2/internal/binding/`. Il risultato: i parametri delle funzioni esportate sono sempre `arg1, arg2, arg3, ...` e non c'è mai un commento/JSDoc. `reflect.Type` non porta né nomi parametro né doc comment — sono informazioni che esistono solo a livello di sorgente Go (AST), mai a runtime.

## Obiettivo e scope

Patchare il generatore vendorizzato per estrarre nomi parametro reali e doc comment del metodo dal sorgente Go (`go/parser`+`go/ast`, solo stdlib, nessuna nuova dipendenza) e iniettarli nei file `.js`/`.d.ts` generati.

Scope: solo le firme dei metodi (nomi parametro + doc comment del metodo). **Non** i commenti sui campi delle struct in `models.ts` — esplicitamente fuori scope.

## Approccio scelto

Patch al codice vendorizzato Wails v2, seguendo l'infrastruttura di patching già esistente nel progetto (`patches/*.patch` applicati da `scripts/vendor-wails-patch.ps1` dopo `go mod vendor`), invece di un post-processor esterno che riparserizza i file già generati.

Motivazione: il modello dati di Wails v2 ha già i campi necessari — `Parameter.Name` (`parameter.go`) e `BoundMethod.Comments` (`boundMethod.go`) esistono ma sono sempre valorizzati a vuoto (`newParameter("", input)`, `Comments: """`) in `reflect.go`. Il lavoro è riempirli, non aggiungerli. Un post-processor esterno dovrebbe invece riparserizzare l'output testuale generato con regex, duplicando logica già presente nel generatore e richiedendo un passo extra in ogni punto in cui viene invocata la generazione (`wails build`, `wails dev`, `wails generate module`).

## Scoperta critica che guida il design

`binding.NewBindings`/`getMethods` non viene chiamato solo in fase di codegen. Viene invocato incondizionatamente ad ogni avvio dell'app:

- `vendor/.../internal/app/app_dev.go:225` (`wails dev`)
- `vendor/.../internal/app/app_production.go:78` (binario di produzione, serve per instradare le chiamate JS→Go)
- `vendor/.../internal/app/app_bindings.go:62` (build tag `//go:build bindings`, usato solo da `wails generate module -tags bindings` / `wails build -tags bindings`)

Se il parsing AST vivesse incondizionatamente in `reflect.go`, ogni binario di produzione tenterebbe di aprire e parsare i file sorgente `.go` originali ad ogni avvio, usando path assoluti presi dalla macchina di build (`runtime.FuncForPC(...).FileLine()`) — path che non esistono sulla macchina dell'utente finale. Il fallback lo rende innocuo, ma resta I/O/CPU sprecato ad ogni avvio per una feature utile solo in fase di generazione.

**Soluzione**: split via build tag, esattamente come Wails fa già per `app_bindings.go`:

- `source.go` (nessun tag) — struct condivisa `methodSource{paramNames []string, comment string}`.
- `source_ast.go` (`//go:build bindings`) — implementazione reale con `go/parser`+`go/ast`, compilata solo quando si generano i binding.
- `source_stub.go` (`//go:build !bindings`) — no-op a costo zero, compilata in `wails dev`/produzione.

`reflect.go` chiama la stessa API (`newSourceCache()`/`.lookup(...)`) in entrambi i casi — il build tag decide quale implementazione viene linkata.

## File coinvolti

Tutti in `vendor/github.com/wailsapp/wails/v2/internal/binding/`:

- `source.go`, `source_ast.go`, `source_stub.go` (nuovi)
- `reflect.go` — `getMethods` usa `sourceCache` per popolare `Comments` e i nomi dei parametri **input** (non degli output — i tipi di ritorno non hanno bisogno di nomi)
- `generate.go` — nuovi helper `paramArgName` (nome reale se identificatore JS valido e non riservato, altrimenti fallback `arg<N>`) e `formatJSDoc` (converte un doc comment Go in blocco JSDoc, escaping `*/`), usati sia nella generazione `.js` che `.d.ts`
- `parameter.go`, `boundMethod.go` — nessuna modifica, i campi esistono già

La patch viene salvata in `patches/wails-v2-jsdoc-bindings.patch`, applicata da `scripts/vendor-wails-patch.ps1` come le due patch esistenti (`wails-v2-opendevtools.patch`, `wails-v2-tray-icon.patch`).

## Edge case

1. Parsing/lookup fallisce (file non trovato, parse error, `FuncDecl` non trovato — es. metodo promosso da embedding) → `methodSource{}` zero-value → fallback silenzioso ad `arg<N>`/nessun commento, identico al comportamento odierno. La generazione non può mai fallire per questo.
2. Blank identifier `_` come nome parametro Go → risolto in `reflect.go` prima che il nome arrivi a `generate.go`.
3. Commenti multi-riga / `*/` letterale nel testo → `ast.CommentGroup.Text()` + escaping in `formatJSDoc`.
4. I parametri di output non ricevono mai un nome (comportamento invariato — solo il tipo conta per costruire `Promise<Tipo>`).
5. Nomi parametro Go che sono parole riservate JS o identificatori non validi (es. ipotetico `func (a *App) Foo(new string)`) → fallback forzato ad `arg<N>` tramite `jsIdentifierRegex`/`jsReservedWords`, per non generare output JS/TS sintatticamente rotto.

## Verifica

Nessun test automatico in CI copre l'output generato (`.github/workflows/syntax-check.yml` esegue solo `vendor-wails-patch.ps1` + `go vet ./...`, che valida solo che il Go vendorizzato patchato compili). Verifica manuale:

1. `powershell -File scripts/vendor-wails-patch.ps1` da stato pulito → le tre patch si applicano senza conflitti.
2. `go vet ./...` (esercita `source_stub.go`) e `go vet -tags bindings ./...` (esercita `source_ast.go`).
3. `wails generate module -tags bindings` → rigenera solo `App.d.ts`/`App.js`.
4. Ispezione manuale su casi noti in `app.go`/`app_system.go`/`logger.go`:
   - `RestartApp` (`app.go:173-176`, doc comment 3 righe, zero param) → JSDoc presente, firma invariata.
   - `SaveConfig` (`app.go:227-229`, doc comment 2 righe, un param `cfg *utils.Config`) → JSDoc + `SaveConfig(cfg:utils.Config)` invece di `arg1`.
   - `CheckIsDefaultEMLHandler` (`app_system.go`, doc comment multi-paragrafo con lista numerata e sezione "Returns:") → blocco JSDoc renderizzato senza rompere `*/`.
   - `GetExtendedMachineData` (`app.go:271`, nessun doc comment) → nessun blocco `/** */` vuoto emesso.
   - `FrontendLog` (`logger.go:23-27`, oggi `arg1,arg2,arg3` in `App.d.ts`) → nomi reali `level, message, contextJSON`.
5. Confronto strutturale di `App.js`/`App.d.ts` con l'output pre-modifica (import, logica dei tipi di ritorno, import dei modelli) per assicurarsi che nient'altro sia cambiato.

Nessuna modifica al `CHANGELOG.md` — è una modifica di tooling di sviluppo, non un comportamento visibile all'utente finale dell'app.
