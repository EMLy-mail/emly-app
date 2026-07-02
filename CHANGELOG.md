# Changelog EMLy

# 1.8.1 (2026-07-02)
1) Backportata da Wails v3 la funzione per aprire programmaticamente la finestra DevTools di WebView2 (`App.OpenDevTools()`), applicata a Wails v2.12.0 tramite patch su `vendor/` (non fork) per evitare di dover ripullare e ripatchare manualmente i sorgenti ad ogni bump di versione.
2) Aggiunta una patch Bun per correggere un warning di Vite dev-server causato da un modulo che punta a un file `.map` non pubblicato.
3) Aggiornato il README: nuova sezione Configuration allineata alle chiavi reali di `config.ini`, rimossa la vecchia sezione del sistema di aggiornamento integrato (ora gestito da EMLy-Updater esterno), documentate le patch Wails/Bun e lo Standalone Image Reader.
4) Aggiornata la CI/CD (GitHub Actions) per eseguire `scripts/vendor-wails-patch.ps1` prima della build, così il binario rilasciato include sempre il backport DevTools di Wails v3.

# 1.8.0 (2026-06-01)
1) Aggiunto il supporto a EMLy di aprire direttamente vari formati di file immagine come lettore di default di Windows.
2) Aggiornate dipendenze lato Frontend.
3) Aggiornate dipendenze lato Backend (Go) e Wails v2.12.0.
4) Fixati vari bug

# 1.7.9 (2026-06-23)
1) Fixato un bug dove la finestra non veniva portata in primo piano quando si apriva una mail con EMLy già in esecuzione.

# 1.7.8 (2026-06-23)
1) Fixato un bug dove la data non veniva estratta sui file .EML non-PEC
2) Aggiornato alcune dipendenze lato Frontend.

# 1.7.7 (2026-06-12)
1) Aggiunto un error handling per errori 404.
2) Cambiato algortimo di HWID con offuscamento per privacy.
3) Cambiamenti i colori dei pulsanti degli allegati.
4) Aggiunta la possibilità di resettare a default la path di download degli allegati.
5) Aggiunto lo scrolling tra le tab (se in modalità tab).
6) Fixato un bug dove le icone degli allegati erano più piccole più il nome del file era più lungo.
7) Fixato lo scrolling nel Bug Report Dialog.
8) Cambiate le default delle impostazioni per attivare i link cliccabili e la correzione automatica del contrasto testo.
9) Bump SDK Decoder da Beta a Stable.

## 1.7.6 (2026-06-12)
1) Implementato un nuovo sistema di download nativo degli allegati che non passa più dal download manager di Edge/WebView2: i file vengono salvati direttamente su disco dal backend Go. Risolve il bug critico dove, scaricando un allegato non supportato e cliccando "Mantieni" sul prompt di sicurezza di Edge, una parte della WebView rimaneva freezata in modo permanente.
2) Aggiunta nelle Impostazioni (sezione Anteprima) la possibilità di scegliere la cartella di destinazione per gli allegati scaricati (default: cartella Download). Dopo ogni salvataggio viene mostrata una toast con il pulsante "Apri cartella" che apre Esplora Risorse con il file selezionato.
3) Fixato un bug critico dove gli allegati di email Yahoo (e mailer simili) venivano salvati con il Content-Id come nome file (es. `c181c7cd-...@yahoo.com` invece di `Dichiarazione.docx`), con conseguente estensione errata. Il parser MIME ora dà priorità al filename dichiarato in `Content-Disposition`/`Content-Type` rispetto all'header `Content-Id`, e i file embedded senza nome ricevono sempre l'estensione corretta derivata dal MIME type. Aggiunti test automatici per la classificazione degli allegati.
4) Aggiunto il supporto agli allegati .DOC e .DOCX: ora vengono aperti direttamente con l'applicazione predefinita di Windows (Word, LibreOffice, ecc.) invece di mostrare la toast per allegati non supportati.
5) Rimosso completamente il sistema di aggiornamento integrato (controllo manifest, download e installazione), ora sostituito da un updater esterno dedicato. Rimosse le relative opzioni dalle Impostazioni (canale di rilascio, sorgente aggiornamenti, percorso UNC) e le chiavi `UPDATE_*` da config.ini.
6) Aggiunta la possibilità di aprire gli allegati PDF e immagine come schede (tab) invece che in finestre separate, attivabile dalle Impostazioni quando la modalità tab è abilitata. Aggiunto un menu contestuale sulle schede.
7) L'opzione "Apri allegati come tab" viene ora disabilitata automaticamente quando la modalità tab viene disattivata.
8) Migliorata la pagina di errore: nuova immagine di errore personalizzata, messaggi aggiuntivi, e suggerimento mostrato quando il caricamento rimane bloccato troppo a lungo (possibile crash interno).
9) Aggiunti pulsanti di test del crash handler (errori 500/404) nella Danger Zone delle Impostazioni.
10) Aggiunti nuovi ringraziamenti nella pagina Credits.

## 1.7.5 (2026-06-05)
1) Aggiunto supporto al controllo aggiornamenti tramite API HTTP (`/v2/updates/manifest`), con download diretto dell'installer tramite URL completo restituito dal manifest.
2) Aggiunto campo `UPDATE_SOURCE` in `config.ini` (`api` o `unc`) per scegliere esplicitamente la sorgente degli aggiornamenti.
3) Aggiunto toggle nella pagina Impostazioni (Danger Zone) per commutare tra sorgente API e percorso di rete UNC, senza dover modificare il file di configurazione manualmente.
4) Il manifest API supporta ora il campo `isCritical` e le note di rilascio dettagliate (`detailedReleaseNotes`) con descrizione multilingua (IT/EN); le note vengono selezionate automaticamente in base alla lingua configurata.
5) Aggiornata la Action CI/CD per usare Node 24.

## 1.7.4 (2026-05-25)
1) Riorganizzata la pagina Impostazioni raggruppando meglio le opzioni per sezione, con relative traduzioni aggiornate.
2) Aggiornato il visualizzatore PDF per usare `@embedpdf/svelte-pdf-viewer`, con una nuova implementazione dedicata e una semplificazione della pagina PDF.
3) Aggiunta la visualizzazione della data dell'email nel Mail Viewer.
4) Rimossi gli asset statici OpenJPEG non piu' necessari dopo l'aggiornamento del visualizzatore PDF.
5) Aggiunta una nuova entry nei ringraziamenti per Amber, che ha fornito feedback preziosi per migliorare l'esperienza utente.
6) Aggiunto supporto a `detailedReleaseNotes` nel file `version.json`: ogni versione può ora specificare un `severityType` (`patch`, `feature`, `breaking`, `security`) e una descrizione localizzata (`en`/`it`). Il tipo di severità viene mostrato come badge nell'intestazione della sezione aggiornamenti nelle Impostazioni e cambia il colore del pannello di download di conseguenza. Se il tipo è `security`, viene mostrato automaticamente un AlertDialog che suggerisce l'aggiornamento immediato.

## 1.7.3 (2026-05-20)
1) Sistemato un critico bug di avvio, quando l'utente avviava EMLy tramite un .EML, andava in crash la WebView. Causato dall'uso di $effect invece che onMount.

## 1.7.2 (2026-05-11)
1) Aggiunta funzionalità di suggerimento abilitazione link: i link nelle email rimangono disabilitati per default, ma se l'utente tenta di cliccarci due o più volte viene mostrata una notifica toast con un pulsante "Abilita" per attivare il supporto ai link con conferma di sicurezza.

## 1.7.1 (2026-05-04)
1) Aggiunto controllo alla chiusura della finestra principale: se ci sono finestre di visualizzazione immagini o PDF ancora aperte, la chiusura viene bloccata e viene mostrata una finestra di avviso all'utente.
2) Aggiunto sistema di schede (tab) per la visualizzazione simultanea di più email: attivabile dalla Danger Zone nelle Impostazioni, permette di aprire ogni nuova email in una scheda separata e di chiuderle individualmente.
3) Fixato un bug dove le immagini inline (cid:) nelle email MSG non venivano visualizzate, perché il reader MSG non leggeva il campo PR_ATTACH_CONTENT_ID e non sostituiva i riferimenti cid: con data URI.
4) Fixato un bug dove le immagini inline (cid:) nelle email PEC annidate (es. email inoltrate con allegati immagine) non venivano visualizzate, perché il reader PEC non processava i file embedded dell'email interna.
5) Aggiunta opzione "Correzione automatica contrasto testo" nelle Impostazioni (sezione visualizzazione email): quando attiva, un algoritmo WCAG analizza ogni elemento del corpo email e inverte il colore del testo quando è troppo simile allo sfondo (rapporto di contrasto < 2:1), risolvendo il problema del testo nero su sfondo nero e viceversa.
6) Fixato un bug dove il reload tramite il pulsante "Ricarica" nella footerbar non funzionava correttamente, a causa di un confronto errato del pathname che non considerava i parametri di query (es. ?reload=true).

## 1.7.0 (2026-04-20)
1) Selettore canale di aggiornamento (Stabile / Beta) spostato in cima alla sezione Aggiornamenti, ora sempre visibile prima dei controlli di versione.
2) Il cambio di canale azzera immediatamente lo stato della ricerca aggiornamenti, richiedendo una nuova verifica esplicita con il canale selezionato.
3) Aggiunto il campo `channel` nella risposta di `CheckForUpdates`, così il frontend sa sempre su quale canale è stato effettuato l'ultimo controllo.
4) Rimossi i cast `as any` superflui sull'oggetto `config` nella pagina Impostazioni, sostituiti con accesso tipizzato corretto.

## 1.6.7 (2026-03-23)
1) Aggiunto selettore del canale di rilascio (Stabile / Beta) nella sezione Aggiornamenti delle Impostazioni, che permette di scegliere da quale canale ricevere gli aggiornamenti e salva immediatamente la scelta in config.ini.
2) Aggiunta la toast per l'apertura degli allegati non supportati, con opzione per scaricare il file o meno.
3) I link presenti nel corpo delle email sono ora cliccabili: al click viene mostrato un avviso di sicurezza con l'URL di destinazione, richiedendo conferma prima di aprire il link nel browser.
4) Installer: le chiavi di registro per le associazioni file (.eml, .msg) vengono ora scritte in HKLM e i collegamenti in posizioni All Users; entrambi persistono tra profili AD temporanei e sono visibili a tutti gli utenti della macchina.

## 1.6.6 (2026-03-19)
1) Aggiunta navigazione tra le pagine nel visualizzatore PDF: pulsanti pagina precedente/successiva e contatore pagina corrente/totale nella toolbar.
2) Vari bug fix

## 1.6.5 (2026-03-10)
1) Rimosso il recupero dell'IP esterno (api.ipify.org) dalla raccolta informazioni di sistema nel report bug.
2) Rimosso il recupero delle informazioni GPU dalla raccolta informazioni di sistema nel report bug.
3) Rimosso dead code `_configCache` dal dialog bug report; rinominata `captureEnvironmentData` in `captureConfigData` per coerenza con la variabile di stato.
4) Reso interno (unexported) il metodo `uploadBugReport` in Go, rimuovendolo dai binding Wails; ottimizzato il recupero di apiURL/apiKey usando i dati già presenti in `machineInfo` invece di chiamare `GetConfig()` ridondante.
5) Aggiunte le transizioni CSS nello switching delle pagine.

## 1.6.4 (2026-03-06)
1) Fixato un bug nel visualizzatore PDF dove due operazioni di rendering concorrenti sullo stesso canvas causavano un errore all'apertura del file.
2) Aggiunto il supporto al codec OpenJPEG (JPEG 2000 / JPX) nel visualizzatore PDF tramite il modulo WASM incluso in pdfjs-dist, necessario per decodificare correttamente immagini JPX nei documenti PDF.

## 1.6.4 (2026-03-06)
1) Aggiunto logging di debug dettagliato durante il caricamento delle email: estensione file, dimensione, formato rilevato, tipo di body (HTML/testo), numero allegati, tipi MIME degli allegati, stato PEC, e presenza di email annidate.

## 1.6.4 (2026-03-05)
1) Implementato un sistema di logging strutturato in JSON basato su `log/slog`, con output simultaneo su stdout e file di log.
2) Aggiunta la "Canonical Log Line" per ogni funzione esposta al frontend (nome funzione, durata, stato success/error).
3) Aggiunto il mascheramento automatico dei dati sensibili nei log (password, API key, token).
4) Aggiunto il livello di log configurabile tramite `LOG_LEVEL` nel file `config.ini` (DEBUG, INFO, WARN, ERROR).
5) Creato il servizio di logging frontend (`logger.ts`) che invia log strutturati al backend con contesto del browser (URL, user agent).
6) Aggiornato il bridge `FrontendLog` per supportare il contesto JSON dal frontend.

## 1.6.4 (2026-03-04)
1) Fixato un bug dove email con `Content-Transfer-Encoding: 8Bit` (maiuscolo) non venivano parsate correttamente a causa di un confronto case-sensitive.
2) Fixato un bug dove email con struttura `multipart/alternative` contenente una parte `multipart/mixed` (come quelle inviate da Apple Mail) mostravano un allegato fantasma denominato `embedded_image.mixed` invece di mostrare correttamente gli allegati reali.
3) Aggiunta la toast di errore quando si verifica un errore durante il caricamento dell'email.

## 1.6.3 (2026-03-03)
1) Fixato un bug dove scaricando un singolo allegato PDF dal visualizzatore, il file scaricato era corrotto con dimensioni di 0 byte.
2) Aggiunta la possibilità di selezionare il percorso di aggiornamento (DC-RM2, DC-CB, o percorso personalizzato) direttamente dalle impostazioni.
3) Inserito disclaimer all'avvio se il file config.ini non è presente o non è accessibile.
4) Aggiunti più dati di diagnostica nel report di segnalazione bug (IP interno, dominio Active Directory, configurazione EMLy)
5) Fixato un bug dove se l'aggiornamento falliva, il pulsante di aggiornamento rimaneva bloccato.
6) Fixato un bug dove se l'aggiornamento falliva, il testo diceva che si era all'ultimo aggiornamento disponibile.

## 1.6.2 (2026-02-27)
1) Aggiunto il supporto al MIME "message/rfc822" per visualizzare correttamente le mail con allegati mail (mail annidate).
2) Aggiunto il supporto ai raw Quoted-Printable, per gestire correttamente le mail con codifica non standard.
3) Migliorato il sistema di segnalazione bug: il report ora include informazioni estese sulla macchina (IP interno, dominio Active Directory, configurazione EMLy).
4) Aggiunta la possibilità di ricaricare il file config.ini dal disco senza riavviare l'app (Danger Zone nelle impostazioni).
5) Aggiunta la selezione del percorso aggiornamenti (DC-RM2, DC-CB, o percorso personalizzato) direttamente dalle impostazioni.

## 1.6.1 (2026-02-26)
1) Sistemato un bug del sistema di aggiornamento, dove, in alcuni casi, non veniva scaricata la nuova versione, anche se disponibile. (Il sistema di aggiornamento è ancora in fase di test, ma questo fix dovrebbe migliorare l'affidabilità del processo)
2) Sistemate alcune traduzioni mancanti.
3) Cambiata la path di installazione predefinita.

## 1.6.0 (2026-02-17)
1) Implementazione in sviluppo del sistema di aggiornamento automatico e manuale, con supporto per canali di rilascio (stable/beta) e gestione delle versioni. (Ancora non attivo di default, in fase di test)

## 1.5.5 (2026-02-14)
1) Aggiunto il supporto al caricamento dei bug report su un server esterno, per facilitare la raccolta e gestione dei report da parte degli sviluppatori. (Con fallback locale in caso di errore)
2) Aggiunto il supporto alle mail con formato TNEF/winmail.dat, per estrarre gli allegati correttamente.

## 1.5.4 (2026-02-10)
1) Aggiunti i pulsanti "Download" al MailViewer, PDF e Image viewer, per scaricare il file invece di aprirlo direttamente.
2) Refactor del sistema di bug report.
3) Rimosso temporaneamente il fetching dei dati macchina all'apertura della pagine delle impostazioni, per evitare problemi di performance.
4) Fixato un bug dove, nel Bug Reporting, non si disattivaa il pulsante di invio, se tutti i campi erano compilati.
5) Aggiunto il supprto all'allegare i file di localStorage e config.ini al Bug Report, per investigare meglio i problemi legati all'ambiente dell'utente.



## 1.5.3 (2026-02-10)
1) Sistemato un bug dove, al primo avvio, il tema chiaro era applicato insieme all'opzioni del tema scuro sul contenuto mail, causando un contrasto eccessivo.



## 1.5.2 (2026-02-10)
1) Supporto tema chiaro/scuro.
2) Internazionalizzazione completa (Italiano/Inglese).
3) Opzioni di accessibilità (riduzione animazioni, contrasto).


## 1.5.1 (2026-02-09)
1) Sistemato un bug del primo avvio, con mismatch della lingua.
2) Aggiunto il supporto all'installazione sotto AppData/Local


## 1.5.0 (2026-02-08)
1) Sistema di aggiornamento automatico self-hosted (ancora non attivo di default).
2) Sistema di bug report integrato.


## 1.4.1 (2026-02-06)
1) Export/Import impostazioni.
2) Aggiornamento configurazione installer.
