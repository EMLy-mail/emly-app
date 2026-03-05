# Changelog EMLy

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
