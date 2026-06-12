<script lang="ts">
  import {
    X,
    MailOpen,
    Image,
    FileText,
    File,
    ShieldCheck,
    Loader2,
    Download,
    Info,
    FolderOpen,
  } from '@lucide/svelte';
  import { dev } from '$app/environment';
  import { sidebarOpen, runningInDebugMode } from '$lib/stores/app';
  import { onDestroy, onMount } from 'svelte';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import { toast } from 'svelte-sonner';
  import { EventsOn, WindowShow, WindowUnminimise, BrowserOpenURL } from '$lib/wailsjs/runtime/runtime';
  import { mailState } from '$lib/stores/mail-state.svelte';
  import type { internal } from '$lib/wailsjs/go/models';
  import * as m from '$lib/paraglide/messages';
  import { showDefaultAttachmentToast, cancelCurrentToast } from '$lib/utils/open-default-attachment-toast';
  import { saveAttachmentNatively, saveAllAttachmentsNatively } from '$lib/utils/attachment-download';
  import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';

  import {
    DetectEmailFormat,
    OpenFolderInExplorer,
  } from '$lib/wailsjs/go/main/App';
  import { Button } from '$lib/components/ui/button/index.js';
  import {
    IFRAME_UTIL_HTML_DARK,
    IFRAME_UTIL_HTML_DARK_NO_LINKS,
    IFRAME_UTIL_HTML_LIGHT,
    IFRAME_UTIL_HTML_LIGHT_NO_LINKS,
    IFRAME_CONTRAST_FIX_JS,
    CONTENT_TYPES,
    PEC_FILES,
    arrayBufferToBase64,
    openPDFAttachment,
    openImageAttachment,
    openEMLAttachment,
    openDocAttachment,
    openAndLoadEmail,
    loadEmailFromPath,
    processEmailBody,
    isEmailFile,
  } from '$lib/utils/mail';
  import { settingsStore } from '$lib/stores/settings.svelte';

  // ============================================================================
  // Props
  // ============================================================================

  let {
    emailData = null,
    tabId = null,
    embedded = false,
  }: {
    emailData?: internal.EmailData | null;
    tabId?: string | null;
    embedded?: boolean;
  } = $props();

  // ============================================================================
  // State
  // ============================================================================

  let unregisterEvents = () => {};
  let isLoading = $state(false);
  let loadingText = $state('');
  let linkDialogOpen = $state(false);
  let pendingLinkUrl = $state('');
  let disabledLinkClickCount = $state(0);
  let debugModalOpen = $state(false);
  let debugFormat = $state('');
  let debugFormatLoading = $state(false);

  const LINK_HINT_TOAST_ID = 'emly-link-hint';

  // In tab mode, read from the specific tab in mailState.tabs.
  // In non-tab mode, read from mailState.currentEmail (which reads the active tab).
  let activeEmail = $derived<internal.EmailData | null>(
    tabId !== null
      ? (() => {
          const tab = mailState.tabs.find(t => t.id === tabId);
          return tab?.type === 'email' ? (tab.email ?? null) : null;
        })()
      : mailState.currentEmail
  );

  let activeFilePath = $derived<string | undefined>(
    tabId !== null
      ? (() => {
          const tab = mailState.tabs.find(t => t.id === tabId);
          return tab?.type === 'email' ? tab.filePath : undefined;
        })()
      : (() => {
          const tab = mailState.tabs.find(t => t.id === mailState.activeTabId);
          return tab?.type === 'email' ? tab.filePath : undefined;
        })()
  );

  let iframeUtilHtml = $derived(
    settingsStore.settings.useDarkEmailViewer !== false
      ? (settingsStore.settings.enableLinkClickConfirmation !== false ? IFRAME_UTIL_HTML_DARK : IFRAME_UTIL_HTML_DARK_NO_LINKS)
      : (settingsStore.settings.enableLinkClickConfirmation !== false ? IFRAME_UTIL_HTML_LIGHT : IFRAME_UTIL_HTML_LIGHT_NO_LINKS)
  );

  let contrastFixScript = $derived(
    settingsStore.settings.fixEmailTextContrast ? IFRAME_CONTRAST_FIX_JS : ''
  );

  // ============================================================================
  // Event Handlers
  // ============================================================================

  async function openDebugModal() {
    debugModalOpen = true;
    debugFormat = '';
    const fp = activeFilePath;
    if (fp) {
      debugFormatLoading = true;
      try {
        debugFormat = await DetectEmailFormat(fp) as string;
      } catch {
        debugFormat = 'unknown';
      }
      debugFormatLoading = false;
    }
  }

  function getDebugFolderPath(filePath: string): string {
    const lastSep = Math.max(filePath.lastIndexOf('/'), filePath.lastIndexOf('\\'));
    return lastSep >= 0 ? filePath.substring(0, lastSep) : filePath;
  }

  function getDebugFormatLabel(): string {
    if (debugFormatLoading) return '…';
    if (!activeFilePath && !activeEmail?.isPec) return activeEmail ? 'EML' : '—';
    const fmt = debugFormat.toLowerCase();
    if (fmt === 'msg') return 'MSG';
    if (fmt === 'eml' || fmt === '') return activeEmail?.isPec ? 'EML (PEC)' : 'EML';
    if (fmt === 'unknown') return m.debug_info_format_unknown();
    return fmt.toUpperCase();
  }

  function getBodyInfo(): string {
    const body = activeEmail?.body;
    if (!body) return m.debug_info_body_none();
    const trimmed = body.trimStart();
    const isHtmlBody = trimmed.startsWith('<') || /<!doctype html/i.test(trimmed) || /<html/i.test(trimmed);
    const kb = (body.length / 1024).toFixed(1);
    return `${isHtmlBody ? m.debug_info_body_html() : m.debug_info_body_text()}, ${kb} KB`;
  }

  function onClear() {
    cancelCurrentToast();
    if (tabId !== null) {
      mailState.removeTab(tabId);
    } else {
      mailState.clear();
      sidebarOpen.set(true);
    }
  }

  async function onDownloadAttachments() {
    if (!activeEmail || !activeEmail.attachments) return;

    await saveAllAttachmentsNatively(
      activeEmail.attachments.map((att) => ({
        filename: att.filename,
        base64: arrayBufferToBase64(att.data),
      }))
    );
  }

  async function onOpenMail() {
    isLoading = true;
    loadingText = m.layout_loading_text();

    const result = await openAndLoadEmail();

    if (result.cancelled) {
      isLoading = false;
      loadingText = '';
      return;
    }

    if (result.success && result.email) {
      if (tabId !== null) {
        mailState.addTab(result.email, result.filePath);
      } else {
        mailState.setParams(result.email, result.filePath);
        sidebarOpen.set(false);
      }
    } else if (result.error) {
      console.error('Failed to read email file:', result.error);
      toast.error(m.mail_error_opening());
    }

    isLoading = false;
    loadingText = '';
  }

  async function handleOpenPDF(base64Data: string, filename: string) {
    await openPDFAttachment(base64Data, filename);
  }

  async function handleOpenImage(base64Data: string, filename: string) {
    await openImageAttachment(base64Data, filename);
  }

  async function handleOpenEML(base64Data: string, filename: string) {
    await openEMLAttachment(base64Data, filename);
  }

  async function handleOpenDoc(base64Data: string, filename: string) {
    await openDocAttachment(base64Data, filename);
  }

  function handleWheel(event: WheelEvent) {
    if (event.ctrlKey) {
      event.preventDefault();
    }
  }

  function handleIframeMessage(event: MessageEvent) {
    if (event.data?.type === 'emly-link-disabled-click') {
      disabledLinkClickCount++;
      if (disabledLinkClickCount >= 2) {
        toast(m.mail_link_disabled_toast(), {
          id: LINK_HINT_TOAST_ID,
          duration: 10000,
          action: {
            label: m.mail_link_disabled_enable(),
            onClick: () => {
              settingsStore.update({ enableLinkClickConfirmation: true });
            },
          },
        });
      }
      return;
    }

    if (
      settingsStore.settings.enableLinkClickConfirmation !== false &&
      event.data?.type === 'emly-link-click' &&
      typeof event.data.url === 'string' &&
      event.data.url &&
      !linkDialogOpen
    ) {
      pendingLinkUrl = event.data.url;
      linkDialogOpen = true;
    }
  }

  function onConfirmOpenLink() {
    const url = pendingLinkUrl;
    linkDialogOpen = false;
    pendingLinkUrl = '';
    if (url) BrowserOpenURL(url);
  }

  function onCancelOpenLink() {
    linkDialogOpen = false;
    pendingLinkUrl = '';
  }

  // ============================================================================
  // Effects
  // ============================================================================

  $effect(() => {
    const processCurrentEmail = async () => {
      disabledLinkClickCount = 0;
      toast.dismiss(LINK_HINT_TOAST_ID);

      if (activeEmail?.body) {
        const processedBody = await processEmailBody(activeEmail.body);
        if (processedBody !== activeEmail.body) {
          activeEmail.body = processedBody;
        }
      }
      console.info('Current email changed:', activeEmail?.subject);
      // Log the email info: Format, N of attachments, and whether it has a body, but only if it's not too large (to avoid spamming the logs)
      if (activeEmail) {
        const bodyInfo = activeEmail.body ? `(body length: ${activeEmail.body.length})` : '(no body)';
        const attachmentsInfo = activeEmail.attachments ? `(${activeEmail.attachments.length} attachments)` : '(no attachments)';
        const isPecInfo = activeEmail.isPec ? '(PEC)' : '(not PEC)';
        console.info(`Email info: ${bodyInfo} ${attachmentsInfo} ${isPecInfo}`);
      }

      // Only close sidebar in non-tab mode (tab mode handled by the page)
      if (activeEmail !== null) {
        sidebarOpen.set(false);
      }
    };

    processCurrentEmail();
  });

  // ============================================================================
  // Lifecycle
  // ============================================================================

  onMount(async () => {
    window.addEventListener('message', handleIframeMessage);

    unregisterEvents = EventsOn('launchArgs', async (args: string[]) => {
      console.log('got event launchArgs:', args);

      // In tab mode: only the active tab handles this event
      if (tabId !== null && tabId !== mailState.activeTabId) return;

      if (!args || args.length === 0) return;

      for (const arg of args) {
        if (isEmailFile(arg)) {
          console.log('Loading file from second instance:', arg);
          isLoading = true;
          loadingText = m.layout_loading_text();

          if (arg.toLowerCase().endsWith('.msg')) {
            loadingText = m.mail_loading_msg_conversion();
          }

          const result = await loadEmailFromPath(arg);

          if (result.success && result.email) {
            if (tabId !== null) {
              // In tab mode: open in a new tab
              mailState.addTab(result.email, result.filePath);
              sidebarOpen.set(false);
            } else {
              mailState.setParams(result.email, result.filePath);
              sidebarOpen.set(false);
            }
            WindowUnminimise();
            WindowShow();
          } else if (result.error) {
            console.error('Failed to load email:', result.error);
            toast.error(m.mail_error_opening());
          }

          isLoading = false;
          loadingText = '';
          break;
        }
      }
    });
  });

  onDestroy(() => {
    cancelCurrentToast();
    if (unregisterEvents) {
      unregisterEvents();
    }
    window.removeEventListener('message', handleIframeMessage);
  });

  // ============================================================================
  // Helpers
  // ============================================================================

  function getAttachmentClass(att: { contentType: string; filename: string }): string {
    if (att.contentType.startsWith(CONTENT_TYPES.IMAGE)) return 'image';
    if (att.contentType === CONTENT_TYPES.PDF || att.filename.toLowerCase().endsWith('.pdf'))
      return 'pdf';
    if (att.filename.toLowerCase().endsWith('.eml')) return 'eml';
    return 'file';
  }

  function isPecSignature(filename: string, isPec: boolean): boolean {
    return isPec && filename.toLowerCase().endsWith(PEC_FILES.SIGNATURE);
  }

  function isPecCertificate(filename: string, isPec: boolean): boolean {
    return isPec && filename.toLowerCase() === PEC_FILES.CERTIFICATE;
  }
</script>

<AlertDialog.Root bind:open={linkDialogOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>{m.mail_link_security_title()}</AlertDialog.Title>
      <AlertDialog.Description>
        {m.mail_link_security_description()}
      </AlertDialog.Description>
    </AlertDialog.Header>
    <div class="link-url-box">
      <span class="link-url-label">{m.mail_link_security_url_label()}</span>
      <span class="link-url-value">{pendingLinkUrl}</span>
    </div>
    <AlertDialog.Footer>
      <AlertDialog.Cancel onclick={onCancelOpenLink}>{m.mail_link_security_cancel()}</AlertDialog.Cancel>
      <AlertDialog.Action onclick={onConfirmOpenLink}>{m.mail_link_security_open()}</AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

{#if dev || $runningInDebugMode}
  <Dialog.Root bind:open={debugModalOpen}>
    <Dialog.Content class="debug-dialog-content">
      <Dialog.Header>
        <Dialog.Title class="debug-dialog-title">
          <Info size="16" />
          {m.debug_info_title()}
        </Dialog.Title>
        <Dialog.Description>{m.debug_info_description()}</Dialog.Description>
      </Dialog.Header>

      <div class="debug-grid">
        <span class="debug-label">{m.debug_info_format()}</span>
        <span class="debug-value">
          {#if debugFormatLoading}
            <Loader2 size="12" class="spinner" />
          {:else}
            {getDebugFormatLabel()}
          {/if}
        </span>

        <span class="debug-label">{m.debug_info_pec()}</span>
        <span class="debug-value">
          {#if activeEmail?.isPec}
            <span class="pec-badge"><ShieldCheck size="11" /> PEC</span>
          {:else}
            {m.debug_info_no()}
          {/if}
        </span>

        <span class="debug-label">{m.debug_info_inner_email()}</span>
        <span class="debug-value">{activeEmail?.hasInnerEmail ? m.debug_info_yes() : m.debug_info_no()}</span>

        <span class="debug-label">{m.debug_info_attachments()}</span>
        <span class="debug-value">
          {activeEmail?.attachments?.length ?? 0}
          {#if activeEmail?.attachments && activeEmail.attachments.length > 0}
            <ul class="debug-att-list">
              {#each activeEmail.attachments as att}
                <li><span class="mono">{att.filename}</span> <span class="debug-content-type">{att.contentType}</span></li>
              {/each}
            </ul>
          {/if}
        </span>

        <span class="debug-label">{m.debug_info_body()}</span>
        <span class="debug-value">{getBodyInfo()}</span>

        <span class="debug-label">{m.debug_info_date_raw()}</span>
        <span class="debug-value mono">{activeEmail?.date || '—'}</span>

        <span class="debug-label">{m.debug_info_file()}</span>
        <span class="debug-value mono debug-filepath">{activeFilePath || '—'}</span>
      </div>

      <Dialog.Footer>
        {#if activeFilePath}
          <Button
            variant="outline"
            onclick={() => OpenFolderInExplorer(getDebugFolderPath(activeFilePath!))}
          >
            <FolderOpen size="14" />
            {m.debug_info_show_in_explorer()}
          </Button>
        {/if}
        <Button onclick={() => (debugModalOpen = false)}>{m.debug_info_close()}</Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>
{/if}

<div class="panel fill" class:embedded aria-label={m.mail_panel_label()}>
  {#if isLoading}
    <div class="loading-overlay">
      <Loader2 class="spinner" size="48" />
      <div class="loading-text">{loadingText}</div>
    </div>
  {/if}

  <div class="events" role="log" aria-live="polite">
    {#if activeEmail === null}
      <!-- Empty State -->
      <div class="empty-state">
        <div class="empty-icon">
          <MailOpen size="48" strokeWidth={1} />
        </div>
        <div class="empty-text">{m.mail_no_email_selected()}</div>
        <button class="browse-btn" onclick={onOpenMail} disabled={isLoading}>
          {m.mail_open_eml_btn()}
        </button>
      </div>
    {:else}
      <!-- Email View -->
      <div class="email-view">
        <!-- Header -->
        <div class="email-header-content">
          <div class="subject-row">
            <div class="subject-left">
              <div class="email-subject">
                {activeEmail.subject || m.mail_subject_no_subject()}
              </div>
              {#if dev || $runningInDebugMode}
                <button
                  class="debug-info-btn"
                  onclick={openDebugModal}
                  title="Debug Info"
                  aria-label="Mostra info di debug"
                >
                  <Info size="13" />
                </button>
              {/if}
            </div>
            <div class="controls">
              <button
                class="btn"
                onclick={onDownloadAttachments}
                aria-label={m.mail_download_btn_label()}
                title={m.mail_download_btn_title()}
                disabled={isLoading}
              >
                <Download size="15" />
                {m.mail_download_btn_text()}
              </button>
              <button
                class="btn"
                onclick={onOpenMail}
                aria-label={m.mail_open_btn_label()}
                title={m.mail_open_btn_title()}
                disabled={isLoading}
              >
                <MailOpen size="15" />
                {m.mail_open_btn_text()}
              </button>
              <button
                class="btn"
                onclick={onClear}
                aria-label={m.mail_close_btn_label()}
                title={m.mail_close_btn_title()}
                disabled={isLoading}
              >
                <X size="15" />
                {m.mail_close_btn_text()}
              </button>
            </div>
          </div>

          <!-- Meta Grid -->
          <div class="email-meta-grid">
            <span class="label">{m.mail_from()}</span>
            <span class="value">{activeEmail.from}</span>

            {#if activeEmail.to && activeEmail.to.length > 0}
              <span class="label">{m.mail_to()}</span>
              <span class="value">{activeEmail.to.join(', ')}</span>
            {/if}

            {#if activeEmail.cc && activeEmail.cc.length > 0}
              <span class="label">{m.mail_cc()}</span>
              <span class="value">{activeEmail.cc.join(', ')}</span>
            {/if}

            {#if activeEmail.bcc && activeEmail.bcc.length > 0}
              <span class="label">{m.mail_bcc()}</span>
              <span class="value">{activeEmail.bcc.join(', ')}</span>
            {/if}

            {#if activeEmail.isPec}
              <span class="label">{m.mail_sign_label()}</span>
              <span class="value">
                <span class="pec-badge" title="Posta Elettronica Certificata">
                  <ShieldCheck size="14" />
                  PEC
                </span>
              </span>
            {/if}

            {#if activeEmail.date}
              <span class="label">{m.mail_date()}</span>
              {#if settingsStore.settings.selectedLanguage === 'it'}
                <span class="value">{new Intl.DateTimeFormat('it-IT', { dateStyle: 'full', timeStyle: 'long' }).format(new Date(activeEmail.date))}</span>
              {:else}
              <span class="value">{new Intl.DateTimeFormat('en-GB', { dateStyle: 'full', timeStyle: 'long' }).format(new Date(activeEmail.date))}</span>
              {/if}
             {/if}
          </div>
        </div>

        <!-- Attachments -->
        <div class="email-attachments">
          <span class="att-section-label">{m.mail_attachments()}</span>
          <div class="att-list">
            {#if activeEmail.attachments && activeEmail.attachments.length > 0}
              {#each activeEmail.attachments as att}
                {@const base64 = arrayBufferToBase64(att.data)}
                {@const isImage = att.contentType.startsWith(CONTENT_TYPES.IMAGE)}
                {@const isPdf =
                  att.contentType === CONTENT_TYPES.PDF ||
                  att.filename.toLowerCase().endsWith('.pdf')}
                {@const isEml = att.filename.toLowerCase().endsWith('.eml')}
                {@const isDoc =
                  att.contentType === CONTENT_TYPES.DOCX ||
                  att.contentType === CONTENT_TYPES.DOC ||
                  att.filename.toLowerCase().endsWith('.docx') ||
                  att.filename.toLowerCase().endsWith('.doc')}
                {@const isPecSig = isPecSignature(att.filename, activeEmail.isPec)}
                {@const isPecCert = isPecCertificate(att.filename, activeEmail.isPec)}

                {#if isImage}
                  <button
                    class="att-btn image"
                    onclick={() => handleOpenImage(base64, att.filename)}
                  >
                    <Image size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if isPdf}
                  <button class="att-btn pdf" onclick={() => handleOpenPDF(base64, att.filename)}>
                    <FileText size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if isEml}
                  <button class="att-btn eml" onclick={() => handleOpenEML(base64, att.filename)}>
                    <MailOpen size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if isDoc}
                  <button class="att-btn doc" onclick={() => handleOpenDoc(base64, att.filename)}>
                    <FileText size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else}
                  <button
                    class="att-btn file"
                    onclick={() => showDefaultAttachmentToast({
                      onSave: () => void saveAttachmentNatively(base64, att.filename),
                      onReset: () => {},
                    })}
                  >
                    <File size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {/if}
              {/each}
            {:else}
              <span class="att-empty">{m.mail_no_attachments()}</span>
            {/if}
          </div>
        </div>

        <!-- Email Body -->
        <div class="email-body-wrapper" class:light-theme={settingsStore.settings.useDarkEmailViewer === false}>
          <iframe
            srcdoc={activeEmail.body + iframeUtilHtml + contrastFixScript}
            title={m.mail_email_body_title()}
            class="email-iframe"
            sandbox="allow-scripts"
            onwheel={handleWheel}
          ></iframe>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    z-index: 50;
    backdrop-filter: blur(4px);
    gap: 16px;
  }

  :global(.spinner) {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .loading-text {
    color: white;
    font-size: 16px;
    font-weight: 500;
  }

  .panel {
    background: var(--card);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
    position: relative;
  }

  .panel.fill {
    flex: 1 1 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  /* When embedded inside a tabbed container, the parent provides border/radius */
  .panel.embedded {
    border: none;
    border-radius: 0;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 34px;
    padding: 0 12px;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--muted);
    color: var(--muted-foreground);
    cursor: pointer;
    user-select: none;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .btn:hover {
    background: var(--accent);
    color: var(--accent-foreground);
  }

  .events {
    flex: 1 1 auto;
    min-height: 0;
    overflow: auto;
    padding: 0;
  }

  .email-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    gap: 0;
  }

  .email-header-content {
    background: var(--card);
    padding: 16px;
    border-bottom: 1px solid var(--border);
  }

  .email-subject {
    font-size: 18px;
    font-weight: 600;
    line-height: 1.25;
    color: var(--foreground);
    min-width: 0;
    overflow-wrap: break-word;
  }

  .subject-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 12px;
  }

  .subject-row .controls {
    display: flex;
    gap: 6px;
    flex-shrink: 0;
  }

  .subject-row .btn {
    height: 28px;
    padding: 0 8px;
  }

  .email-meta-grid {
    display: grid;
    grid-template-columns: 60px 1fr;
    gap: 4px;
    font-size: 13px;
  }

  .email-meta-grid .label {
    text-align: right;
    color: var(--muted-foreground);
    margin-right: 8px;
    font-weight: 500;
  }

  .email-meta-grid .value {
    color: var(--foreground);
    word-break: break-all;
    font-weight: 500;
  }

  .email-attachments {
    padding: 10px 16px;
    border-bottom: 1px solid var(--border);
    background: var(--muted);
    display: flex;
    align-items: center;
    gap: 12px;
    overflow-x: auto;
  }

  .att-section-label {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--muted-foreground);
    flex-shrink: 0;
  }

  .att-list {
    display: flex;
    gap: 8px;
  }

  .att-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 28px;
    padding: 0 10px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--foreground);
    font-size: 12px;
    cursor: pointer;
    text-decoration: none;
    max-width: 200px;
  }

  .att-btn:hover {
    background: var(--accent);
    color: var(--accent-foreground);
  }

  .att-btn.image {
    color: #4ade80;
    border-color: rgba(74, 222, 128, 0.3);
  }
  .att-btn.image:hover {
    color: #86efac;
  }

  .att-btn.pdf {
    color: #f87171;
    border-color: rgba(248, 113, 113, 0.3);
  }
  .att-btn.pdf:hover {
    color: #fca5a5;
  }

  .att-btn.eml {
    color: hsl(49, 80%, 49%);
    border-color: rgba(224, 206, 39, 0.3);
  }
  .att-btn.eml:hover {
    color: hsl(49, 80%, 65%);
  }

  .att-btn.doc {
    color: #60a5fa;
    border-color: rgba(96, 165, 250, 0.3);
  }
  .att-btn.doc:hover {
    color: #93c5fd;
  }

  .att-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .email-body-wrapper {
    flex: 1;
    background: #0d0d0d;
    position: relative;
    min-height: 200px;
    border-radius: 0 0 14px 14px;
    overflow: hidden;
  }

  .embedded .email-body-wrapper {
    border-radius: 0;
  }

  .email-body-wrapper.light-theme {
    background: #ffffff;
  }

  .email-iframe {
    width: 100%;
    height: 100%;
    border: none;
    display: block;
  }

  .empty-state {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    opacity: 0.6;
    padding: 20px;
  }

  .empty-icon {
    opacity: 0.5;
  }

  .empty-text {
    font-size: 14px;
    font-weight: 500;
  }

  .browse-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 36px;
    padding: 0 16px;
    background: var(--muted);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--foreground);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .browse-btn:hover {
    background: var(--accent);
    border-color: var(--accent-foreground);
  }

  .browse-btn:disabled,
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
  }

  ::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  ::-webkit-scrollbar-track {
    background: transparent;
  }

  ::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 6px;
  }

  ::-webkit-scrollbar-thumb:hover {
    background: var(--muted-foreground);
  }

  ::-webkit-scrollbar-corner {
    background: transparent;
  }

  .att-empty {
    font-size: 11px;
    color: var(--muted-foreground);
    font-style: italic;
  }

  .pec-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.3);
    padding: 2px 6px;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 700;
    vertical-align: middle;
    user-select: none;
    width: fit-content;
  }

  .subject-left {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    min-width: 0;
    flex: 1;
  }

  .debug-info-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted-foreground);
    cursor: pointer;
    flex-shrink: 0;
    margin-top: 3px;
    opacity: 0.6;
    transition: opacity 0.15s, background 0.15s, color 0.15s;
  }

  .debug-info-btn:hover {
    opacity: 1;
    background: var(--muted);
    color: var(--foreground);
  }

  :global(.debug-dialog-content) {
    max-width: 520px !important;
  }

  :global(.debug-dialog-title) {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .debug-grid {
    display: grid;
    grid-template-columns: 110px 1fr;
    gap: 6px 12px;
    font-size: 13px;
    padding: 4px 0;
  }

  .debug-label {
    color: var(--muted-foreground);
    font-weight: 500;
    text-align: right;
    padding-top: 1px;
  }

  .debug-value {
    color: var(--foreground);
    word-break: break-all;
  }

  .debug-filepath {
    word-break: break-all;
    user-select: all;
  }

  .mono {
    font-family: monospace;
    font-size: 12px;
  }

  .debug-att-list {
    list-style: none;
    padding: 0;
    margin: 4px 0 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .debug-att-list li {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .debug-content-type {
    font-size: 11px;
    color: var(--muted-foreground);
  }

  .link-url-box {
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: var(--muted);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 10px 12px;
    margin: 4px 0;
  }

  .link-url-label {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--muted-foreground);
  }

  .link-url-value {
    font-size: 12px;
    color: var(--foreground);
    word-break: break-all;
    font-family: monospace;
  }
</style>
