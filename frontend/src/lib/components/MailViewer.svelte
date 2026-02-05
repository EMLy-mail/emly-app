<script lang="ts">
  import {
    X,
    MailOpen,
    Image,
    FileText,
    File,
    ShieldCheck,
    Signature,
    FileCode,
    Loader2,
  } from '@lucide/svelte';
  import { sidebarOpen } from '$lib/stores/app';
  import { onDestroy, onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { EventsOn, WindowShow, WindowUnminimise } from '$lib/wailsjs/runtime/runtime';
  import { mailState } from '$lib/stores/mail-state.svelte';
  import * as m from '$lib/paraglide/messages';
  import { dev } from '$app/environment';

  // Import refactored utilities
  import {
    IFRAME_UTIL_HTML,
    CONTENT_TYPES,
    PEC_FILES,
    arrayBufferToBase64,
    createDataUrl,
    openPDFAttachment,
    openImageAttachment,
    openEMLAttachment,
    openAndLoadEmail,
    loadEmailFromPath,
    processEmailBody,
    isEmailFile,
  } from '$lib/utils/mail';

  // ============================================================================
  // State
  // ============================================================================

  let unregisterEvents = () => {};
  let isLoading = $state(false);
  let loadingText = $state('');

  // ============================================================================
  // Event Handlers
  // ============================================================================

  function onClear() {
    mailState.clear();
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
      mailState.setParams(result.email);
      sidebarOpen.set(false);
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

  function handleWheel(event: WheelEvent) {
    if (event.ctrlKey) {
      event.preventDefault();
    }
  }

  // ============================================================================
  // Effects
  // ============================================================================

  // Process email body when current email changes
  $effect(() => {
    const processCurrentEmail = async () => {
      if (mailState.currentEmail?.body) {
        const processedBody = await processEmailBody(mailState.currentEmail.body);

        if (processedBody !== mailState.currentEmail.body) {
          mailState.currentEmail.body = processedBody;
        }
      }

      if (dev) {
        console.debug('emailObj:', mailState.currentEmail);
      }
      console.info('Current email changed:', mailState.currentEmail?.subject);

      if (mailState.currentEmail !== null) {
        sidebarOpen.set(false);
      }
    };

    processCurrentEmail();
  });

  // ============================================================================
  // Lifecycle
  // ============================================================================

  onMount(async () => {
    // Listen for second instance args (when another file is opened while app is running)
    unregisterEvents = EventsOn('launchArgs', async (args: string[]) => {
      console.log('got event launchArgs:', args);

      if (!args || args.length === 0) return;

      for (const arg of args) {
        if (isEmailFile(arg)) {
          console.log('Loading file from second instance:', arg);
          isLoading = true;
          loadingText = m.layout_loading_text();

          // Check if MSG file for special loading text
          if (arg.toLowerCase().endsWith('.msg')) {
            loadingText = m.mail_loading_msg_conversion();
          }

          const result = await loadEmailFromPath(arg);

          if (result.success && result.email) {
            mailState.setParams(result.email);
            sidebarOpen.set(false);
            WindowUnminimise();
            WindowShow();
          } else if (result.error) {
            console.error('Failed to load email:', result.error);
            toast.error('Failed to load email file');
          }

          isLoading = false;
          loadingText = '';
          break;
        }
      }
    });
  });

  onDestroy(() => {
    if (unregisterEvents) {
      unregisterEvents();
    }
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

<div class="panel fill" aria-label="Events">
  {#if isLoading}
    <div class="loading-overlay">
      <Loader2 class="spinner" size="48" />
      <div class="loading-text">{loadingText}</div>
    </div>
  {/if}

  <div class="events" role="log" aria-live="polite">
    {#if mailState.currentEmail === null}
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
            <div class="email-subject">
              {mailState.currentEmail.subject || m.mail_subject_no_subject()}
            </div>
            <div class="controls">
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
            <span class="value">{mailState.currentEmail.from}</span>

            {#if mailState.currentEmail.to && mailState.currentEmail.to.length > 0}
              <span class="label">{m.mail_to()}</span>
              <span class="value">{mailState.currentEmail.to.join(', ')}</span>
            {/if}

            {#if mailState.currentEmail.cc && mailState.currentEmail.cc.length > 0}
              <span class="label">{m.mail_cc()}</span>
              <span class="value">{mailState.currentEmail.cc.join(', ')}</span>
            {/if}

            {#if mailState.currentEmail.bcc && mailState.currentEmail.bcc.length > 0}
              <span class="label">{m.mail_bcc()}</span>
              <span class="value">{mailState.currentEmail.bcc.join(', ')}</span>
            {/if}

            {#if mailState.currentEmail.isPec}
              <span class="label">{m.mail_sign_label()}</span>
              <span class="value">
                <span class="pec-badge" title="Posta Elettronica Certificata">
                  <ShieldCheck size="14" />
                  PEC
                </span>
              </span>
            {/if}
          </div>
        </div>

        <!-- Attachments -->
        <div class="email-attachments">
          <span class="att-section-label">{m.mail_attachments()}</span>
          <div class="att-list">
            {#if mailState.currentEmail.attachments && mailState.currentEmail.attachments.length > 0}
              {#each mailState.currentEmail.attachments as att}
                {@const base64 = arrayBufferToBase64(att.data)}
                {@const isImage = att.contentType.startsWith(CONTENT_TYPES.IMAGE)}
                {@const isPdf =
                  att.contentType === CONTENT_TYPES.PDF ||
                  att.filename.toLowerCase().endsWith('.pdf')}
                {@const isEml = att.filename.toLowerCase().endsWith('.eml')}
                {@const isPecSig = isPecSignature(att.filename, mailState.currentEmail.isPec)}
                {@const isPecCert = isPecCertificate(att.filename, mailState.currentEmail.isPec)}

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
                {:else if isPecSig}
                  <a
                    class="att-btn file"
                    href={createDataUrl(att.contentType, base64)}
                    download={att.filename}
                  >
                    <Signature size="14" />
                    <span class="att-name">{att.filename}</span>
                  </a>
                {:else if isPecCert}
                  <a
                    class="att-btn file"
                    href={createDataUrl(att.contentType, base64)}
                    download={att.filename}
                  >
                    <FileCode size="14" />
                    <span class="att-name">{att.filename}</span>
                  </a>
                {:else}
                  <a
                    class="att-btn file"
                    href={createDataUrl(att.contentType, base64)}
                    download={att.filename}
                  >
                    {#if isImage}
                      <Image size="14" />
                    {:else}
                      <File size="14" />
                    {/if}
                    <span class="att-name">{att.filename}</span>
                  </a>
                {/if}
              {/each}
            {:else}
              <span class="att-empty">{m.mail_no_attachments()}</span>
            {/if}
          </div>
        </div>

        <!-- Email Body -->
        <div class="email-body-wrapper">
          <iframe
            srcdoc={mailState.currentEmail.body + IFRAME_UTIL_HTML}
            title="Email Body"
            class="email-iframe"
            sandbox="allow-same-origin allow-scripts"
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
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 14px;
    overflow: hidden;
  }

  .panel.fill {
    flex: 1 1 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 34px;
    padding: 0 12px;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.06);
    color: inherit;
    cursor: pointer;
    user-select: none;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: rgba(255, 255, 255, 0.5);
  }

  .btn:hover {
    background: rgba(255, 255, 255, 0.09);
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
    background: rgba(255, 255, 255, 0.05);
    padding: 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .email-subject {
    font-size: 18px;
    font-weight: 600;
    line-height: 1.25;
    color: inherit;
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
    color: rgba(255, 255, 255, 0.5);
    margin-right: 8px;
    font-weight: 500;
  }

  .email-meta-grid .value {
    color: rgba(255, 255, 255, 0.9);
    word-break: break-all;
    font-weight: 500;
  }

  .email-attachments {
    padding: 10px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(255, 255, 255, 0.03);
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
    color: rgba(255, 255, 255, 0.5);
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
    border: 1px solid rgba(255, 255, 255, 0.15);
    background: transparent;
    color: rgba(255, 255, 255, 0.8);
    font-size: 12px;
    cursor: pointer;
    text-decoration: none;
    max-width: 200px;
  }

  .att-btn:hover {
    background: rgba(255, 255, 255, 0.05);
    color: #fff;
  }

  .att-btn.image {
    color: #60a5fa;
    border-color: rgba(96, 165, 250, 0.3);
  }
  .att-btn.image:hover {
    color: #93c5fd;
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

  .att-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .email-body-wrapper {
    flex: 1;
    background: white;
    position: relative;
    min-height: 200px;
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
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    color: white;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .browse-btn:hover {
    background: rgba(255, 255, 255, 0.15);
    border-color: rgba(255, 255, 255, 0.25);
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
    background: rgba(255, 255, 255, 0.1);
    border-radius: 6px;
  }

  ::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  ::-webkit-scrollbar-corner {
    background: transparent;
  }

  .att-empty {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.4);
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
</style>
