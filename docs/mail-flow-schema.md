<!-- title: EMLy — Startup, Read/Parse, Attachments -->

# EMLy — Startup, Lettura, Parsing, Allegati

## 1. Avvio processo

```mermaid
flowchart TD
    A["main.go: main()"] --> B["InitLogger + LoadConfig (solo per GUISemver/User-Agent)"]
    B --> C["Scan os.Args #1: --view-image / --view-pdf?"]
    C -->|si| D["Modalita viewer: uniqueId per-file, no tray, isMainWindow=false"]
    C -->|no| E["Modalita main window"]
    D --> F["NewApp(userAgent)"]
    E --> F
    F --> G["Scan os.Args #2: arg con suffisso .eml/.msg?"]
    G -->|trovato| H["app.StartupFilePath = path"]
    G -->|no| I["StartupFilePath vuoto"]
    H --> J["options.App: AssetServer, Bind App, SingleInstanceLock, OnBeforeClose"]
    I --> J
    J --> K{"Main window?"}
    K -->|si| L["Setup tray (se non disabilitato via config)"]
    K -->|no| M["Skip tray"]
    L --> N["wails.Run(appOptions)"]
    M --> N
    N --> O{"Altra istanza gia attiva con stesso uniqueId?"}
    O -->|si| P["Handoff args a istanza esistente via OnSecondInstanceLaunch, processo corrente esce"]
    O -->|no| Q["App.startup(ctx)"]
    Q --> R["CurrentMailFilePath = StartupFilePath; applica LOG_LEVEL; registra EventsOn bringOnTop"]
    R --> S["Frontend boot: root +layout.svelte (host integrity, console hook, rimuove splash)"]
    S --> T["(app)/+layout.svelte: versionInfo, debugger-poll, export-folder check"]
    T --> U["(app)/+page.ts load(): GetViewerData()"]
    U -->|viewer mode| V["redirect a /image o /pdf"]
    U -->|normale| W["GetStartupFile()"]
    W -->|non vuoto| X["ReadEML/ReadMSG diretto per estensione (bypassa ReadAuto)"]
    W -->|vuoto| Y["Nessuna mail iniziale"]
    X --> Z["(app)/+page.svelte onMount: mailState.addTab o setParams"]
    Y --> Z
    Z --> AA["Registra EventsOn('launchArgs', ...) A LIVELLO PAGINA (non per-tab)"]
```

## 2. Apertura file — 4 percorsi che convergono sul parser

```mermaid
flowchart TD
    subgraph A["A. Dialogo file (toolbar Open)"]
        A1["openAndLoadEmail()"] --> A2["Go ShowOpenFileDialog()"]
        A2 --> A3["internal.ShowFileDialog: runtime.OpenFileDialog filtro *.eml;*.msg"]
    end

    subgraph B["B. Startup file / 'Apri con EMLy'"]
        B1["OS lancia EMLy.exe <path>"] --> B2["main.go arg-scan #2 -> StartupFilePath"]
        B2 --> B3["GetStartupFile() + ReadEML/ReadMSG diretto"]
    end

    subgraph C["C. Second-instance (EMLy gia aperto)"]
        C1["Nuova istanza rileva SingleInstanceLock attivo"] --> C2["onSecondInstanceLaunch (tray.go, processo ORIGINALE)"]
        C2 --> C3["bringToForeground() + EventsEmit 'launchArgs'"]
        C3 --> C4["Listener a livello pagina in (app)/+page.svelte riceve args"]
        C4 --> C5["loadEmailFromPath(primo path valido)"]
    end

    subgraph D["D. Allegato 'apri in nuova finestra' (image/pdf/eml)"]
        D1["App.openInViewerWindow: scrive allegato in temp file"] --> D2["exec.Command(exe, --view-image=... / --view-pdf=... / path)"]
        D2 --> D3["Rientra nell'INTERO avvio (sezione 1) come processo viewer"]
    end

    A3 --> P["loadEmailFromPath(filePath)"]
    C5 --> P
    P --> Q["Go ReadAuto(filePath)"]
    Q --> R["internal.DetectEmailFormat: legge 8 byte, magic OLE2/CFB?"]
    R -->|si| S["FormatMSG -> internal.ReadMsgFile"]
    R -->|no| T["Prova internal.ReadPecInnerEml (assume PEC firmata)"]
    T -->|fallisce| U["Fallback internal.ReadEmlFile"]
    S --> V["EmailData"]
    T -->|ok| V
    U --> V
    V --> W["DOMPurify.sanitize(body) + decode base64 se serve"]
    W --> X["mailState.addTab / setParams"]
    X --> Y["syncCurrentMailFilePath() -> Go SetCurrentMailFilePath"]
    B3 -.->|bypassa ReadAuto| V
```

## 3. Parsing EML — `backend/utils/mail/mailparser.go` + `eml_reader.go`

```mermaid
flowchart TD
    A["ReadEmlFile / ReadPecInnerEml apre file"] --> B["Parse(io.Reader)"]
    B --> C["net/mail.ReadMessage: separa header/body"]
    C --> D["createEmailFromHeader: decodeMimeSentence (RFC 2047) su Subject/From/To/Cc/Bcc/Date/Message-ID/..."]
    D --> E["mime.ParseMediaType sul Content-Type root"]
    E --> F{"Tipo top-level"}
    F -->|multipart/mixed o signed| G["parseMultipartMixed"]
    F -->|multipart/alternative| H["parseMultipartAlternative"]
    F -->|multipart/related| I["parseMultipartRelated"]
    F -->|text/plain o text/html| J["readDecodedText diretto"]
    G --> K["Ricorsione su ogni parte"]
    H --> K
    I --> K
    K --> L{"Classificazione parte"}
    L -->|text/html o text/plain| M["Append a body"]
    L -->|message/rfc822| N["decodeRfc822Attachment -> allegato .eml, HasInnerEmail"]
    L -->|filename + Content-Disposition attachment| O["Attachment"]
    L -->|ha Content-Id| P["EmbeddedFile (inline)"]
    O --> Q["decodeContent: base64 / quoted-printable / 7bit/8bit/binary"]
    P --> Q
    Q --> R["ReadEmlFile post-processing"]
    R --> S["Per ogni EmbeddedFile: data:mime;base64 URI, sostituisce cid: nel body via regex, duplica in Attachments"]
    R --> T["Per ogni Attachment: check filename daticert.xml/smime.p7s (IsPec), .eml suffix (HasInnerEmail)"]
    S --> U["expandTNEFAttachments (vedi sezione 4)"]
    T --> U
    U --> V["ConvertToUTF8 su From/Subject/indirizzi/body (Windows-1252 se non valid UTF8)"]
    V --> W["body = HTMLBody, fallback TextBody"]
    W --> X["EmailData finale"]

    Y["ReadPecInnerEml"] -.-> B
    Y --> Z["Trova allegato postacert.eml nell'envelope esterno"]
    Z --> AA["Ri-parsa come Email interna, scarta envelope esterno"]
    AA --> X
```

## 3b. Parsing MSG — `msg_reader.go` + `rtf.go` (OLE2/CFB fatto a mano)

```mermaid
flowchart TD
    A["ReadMsgFile apre file"] --> B["Read(io.ReaderAt) -> newCFBReader"]
    B --> C["Legge header 512 byte, valida magic OLE2/CFB"]
    C --> D["Legge FAT (+ DIFAT per file grandi), directory entries"]
    D --> E["Costruisce albero dirNode (storage/stream, sibling+child)"]
    E --> F["Legge MiniFAT + ministream per stream piccoli <4096 byte"]
    F --> G["parseMessage: cerca stream __substg1.0_<propid><proptype>"]
    G --> H["props map (propID<<16 | propType) -> bytes"]
    H --> I["Subject: PidTagSubject 0x0037 (fallback ConversationTopic 0x0070)"]
    H --> J["Body: PidTagBodyHtml 0x1013"]
    J -->|assente| K["htmlFromCompressedRTF (PidTagRtfCompressed 0x1009)"]
    K --> L["decompressRTF: LZFu / MS-OXRTFCP, dizionario fisso 207 byte"]
    L --> M{"isEncapsulatedHTML? (marker \\fromhtml1)"}
    M -->|si| N["deEncapsulateHTML (MS-OXRTFEX): stack stato, htmlrtf/inHTML/skip/uc, decode \\'XX via codepage"]
    M -->|no| O["Fallback PidTagBody 0x1000 -> textToHTML (escape + newline->br)"]
    N --> P["Body HTML recuperato (cid: preservati)"]
    H --> Q["Sender: PidTagSenderName/EmailAddress (fallback SentRepresenting*)"]
    H --> R["Recipients: PidTagDisplayTo/Cc/Bcc, splitRecipients"]
    G --> S["Attachments: storage __attach_version1.0_#n"]
    S --> T["parseAttachment: AttachLongFilename/Filename, AttachMimeTag, AttachData, AttachContentId"]
    T --> U{"Content-Type multipart/*?"}
    U -->|si| V["extractMIMEAttachments (net/textproto + mime/multipart) -> HasInnerEmail"]
    U -->|no + ha ContentId| W["cidToDataURI map (inline)"]
    U -->|no, normale| X["Append a email.Attachments"]
    V --> X
    W --> X
    X --> Y["Sostituzione regex cid -> data URI in email.Body"]
    P --> Y
    O --> Y
    Y --> Z["Decode stringhe: PT_STRING8 (ANSI, ConvertToUTF8) o PT_UNICODE (UTF16LE)"]
    Z --> AA["EmailData finale"]
```

## 4. Estrazione/gestione allegati

```mermaid
flowchart TD
    A["Parte MIME / stream MSG con filename"] --> B{"Tipo"}
    B -->|Content-Disposition attachment o filename senza Content-Id| C["Attachment normale"]
    B -->|ha Content-Id| D["EmbeddedFile / inline"]
    B -->|message/rfc822| E["Allegato .eml nidificato"]
    D --> F["data:mime;base64 URI sostituito nel body via cid:, duplicato in Attachments list"]
    E --> G["decodeRfc822Attachment / extractMIMEAttachments -> HasInnerEmail=true"]
    C --> H{"Filename contiene daticert.xml + smime.p7s?"}
    H -->|si| I["IsPec=true, ReadPecInnerEml cerca postacert.eml"]
    H -->|no| J["Attachment normale"]
    J --> K["expandTNEFAttachments: e' winmail.dat / vnd.ms-tnef / magic 78 9F 3E 22?"]
    K -->|si| L["github.com/teamwork/tnef decode"]
    L --> M["Scan ricorsivo (depth 10) attAttachment 0x9005 -> PR_ATTACH_DATA_OBJ 0x3701 per MAPI nidificati"]
    M --> N["Filtra placeholder Outlook 'allegato incorporato MAPI'"]
    N --> O["Sostituisce blob con file estratti in Attachments"]
    K -->|no| O

    O --> P["Frontend: EmailAttachmentsList.svelte classifica per content-type/estensione"]
    P --> Q{"Azione utente"}
    Q -->|Salva| R["Go SaveAttachment(filename, base64)"]
    R --> S["internal.SaveAttachmentToFolder: decode base64, folder da EXPORT_ATTACHMENT_FOLDER o Downloads"]
    S --> T["ExpandWindowsEnvVars + sanitizeFilename (no path traversal) + uniquePath (mai overwrite)"]
    T --> U["OpenExplorerForPath (opzionale, mostra file salvato)"]
    Q -->|Visualizza| V["attachment-handlers.ts sceglie in base a settings"]
    V -->|tab mode + open-as-tab| W["mailState.addImageTab/addPDFTab (solo frontend)"]
    V -->|default| X["OpenImageWindow/OpenPDFWindow/OpenEMLWindow -> relaunch EMLy.exe --view-*"]
    V -->|app esterna| Y["OpenImage/OpenPDF/OpenDocument -> cmd /c start tempfile"]
```

## 5. Tab state / sync / second-instance (sequence)

```mermaid
sequenceDiagram
    participant OS
    participant MainProc as EMLy (istanza principale)
    participant Page as (app)/+page.svelte
    participant MailState as mailState (Svelte store)
    participant GoApp as App (Go)

    OS->>MainProc: doppio click file2.eml (istanza gia attiva)
    MainProc->>MainProc: SingleInstanceLock rileva conflitto uniqueId
    MainProc->>MainProc: onSecondInstanceLaunch(args) [tray.go]
    MainProc->>MainProc: bringToForeground() (unminimize, AlwaysOnTop toggle)
    MainProc->>Page: EventsEmit("launchArgs", args)
    Note over Page: Listener registrato UNA VOLTA a livello pagina,<br/>non per singolo tab - funziona anche se<br/>un tab PDF/immagine ha il focus
    Page->>Page: trova primo path valido (isEmailFile)
    Page->>MailState: loadEmailFromPath(path) -> ReadAuto
    MailState->>GoApp: ReadAuto(path)
    GoApp-->>MailState: EmailData
    MailState->>MailState: addTab(email) o setParams
    MailState->>GoApp: syncCurrentMailFilePath() -> SetCurrentMailFilePath
    MailState->>Page: collassa sidebar
    Page->>GoApp: WindowUnminimise / WindowShow
    Page->>GoApp: EventsEmit("bringOnTop")
    GoApp->>GoApp: handler registrato in startup() forza foreground
```

## Note chiave

- Nessun drag-and-drop: apertura solo via dialogo file, associazione OS ("apri con"), o second-instance relaunch.
- `ReadAuto` fa auto-detect formato + tentativo PEC; il percorso B (startup file) lo bypassa e chiama `ReadEML`/`ReadMSG` diretto in base all'estensione.
- Nessuna cache lato Go: ogni `ReadEML`/`ReadMSG`/`ReadAuto` ri-parsa da disco. L'unica "cache" e l'array `mailState.tabs` in memoria nel frontend.
- Ogni finestra viewer (immagine/PDF/eml) e un processo `EMLy.exe` separato, rientra nell'intero avvio con flag `--view-image`/`--view-pdf`, `SingleInstanceLock` per-file.
- `beforeClose` blocca chiusura finestra principale finche ci sono viewer figli aperti (`openImages`/`openPDFs` map).
