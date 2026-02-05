<script lang="ts">
  import { X, MailOpen, Image, FileText, File, ShieldCheck, Shield, Signature, FileCode, Loader2 } from "@lucide/svelte";
  import { ShowOpenFileDialog, ReadEML, OpenPDF, OpenImageWindow, OpenPDFWindow, OpenImage, ReadMSG, ReadPEC, OpenEMLWindow, ConvertToUTF8, SetCurrentMailFilePath } from "$lib/wailsjs/go/main/App";
  import type { internal } from "$lib/wailsjs/go/models";
  import { sidebarOpen } from "$lib/stores/app";
  import { onDestroy, onMount } from "svelte";
  import { toast } from "svelte-sonner";
  import { EventsOn, WindowShow, WindowUnminimise } from "$lib/wailsjs/runtime/runtime";
  import { mailState } from "$lib/stores/mail-state.svelte";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import * as m from "$lib/paraglide/messages";
  import { dev } from "$app/environment";
  import { isBase64, isHtml } from "$lib/utils";

  let unregisterEvents = () => {};
  let isLoading = $state(false);
  let loadingText = $state("");


  let iFrameUtilHTML = "<style>body{margin:0;padding:20px;font-family:sans-serif;} a{pointer-events:none!important;cursor:default!important;}</style><script>function handleWheel(event){if(event.ctrlKey){event.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});<\/script>";

  function onClear() {
    mailState.clear();
  }

  $effect(() => {
    const process = async () => {
      if (mailState.currentEmail?.body) {
        let content = mailState.currentEmail.body;
        // 1. Try to decode if not HTML
        if (!isHtml(content)) {
          const clean = content.replace(/[\s\r\n]+/g, '');
          if (isBase64(clean)) {
            try {
              const decoded = window.atob(clean);
              content = decoded;
            } catch (e) {
              console.warn("Failed to decode base64 body:", e);
            }
          }
        }

        // 2. If it is HTML (original or decoded), try to fix encoding
        if (isHtml(content)) {
            try {
              const fixed = await ConvertToUTF8(content);
              if (fixed) {
                content = fixed;
              }
            } catch (e) { 
                console.warn("Failed to fix encoding:", e);
            }
        }
        
        // 3. Update if changed
        if (content !== mailState.currentEmail.body) {
           mailState.currentEmail.body = content;
        }
      }

      if(dev) {
        console.debug("emailObj:", mailState.currentEmail)
      }
      console.info("Current email changed:", mailState.currentEmail?.subject);
      if(mailState.currentEmail !== null) {
        sidebarOpen.set(false);
      }
    };
    process();
  })

  onDestroy(() => {
    if (unregisterEvents) unregisterEvents();
  });

  onMount(async () => {
    // Listen for second instance args
    unregisterEvents = EventsOn("launchArgs", async (args: string[]) => {
      console.log("got event launchArgs:", args);
      if (args && args.length > 0) {
        for (const arg of args) {
          const lowerArg = arg.toLowerCase();
          if (lowerArg.endsWith(".eml") || lowerArg.endsWith(".msg")) {
            console.log("Loading file from second instance:", arg);
            isLoading = true;
            loadingText = m.layout_loading_text();
            
            try {
              let emlContent;
              
              if (lowerArg.endsWith(".msg")) {
                  loadingText = m.mail_loading_msg_conversion();
                  emlContent = await ReadMSG(arg, true);
              } else {
                  // EML handling
                  try {
                    emlContent = await ReadPEC(arg);
                  } catch (e) {
                    console.warn("ReadPEC failed, trying ReadEML:", e);
                    emlContent = await ReadEML(arg);
                  }

                  if (emlContent && emlContent.body) {
                    const trimmed = emlContent.body.trim();
                    const clean = trimmed.replace(/[\s\r\n]+/g, '');
                    if (clean.length > 0 && clean.length % 4 === 0 && /^[A-Za-z0-9+/]+=*$/.test(clean)) {
                      try {
                        emlContent.body = window.atob(clean);
                      } catch (e) { }
                    }
                  }
              }

              mailState.setParams(emlContent);
              sidebarOpen.set(false);
              WindowUnminimise();
              WindowShow();
            } catch (error) {
              console.error("Failed to load email:", error);
              toast.error("Failed to load email file");
            } finally {
              isLoading = false;
              loadingText = "";
            }
            break;
          }
        }
      }
    });
  });

  async function openPDFHandler(base64Data: string, filename: string) {
    try {
      if (settingsStore.settings.useBuiltinPDFViewer) {
         await OpenPDFWindow(base64Data, filename);
      } else {
         await OpenPDF(base64Data, filename);
      }
    } catch (error: string | any) {
      if(error.includes(filename) && error.includes("already open")) {
        toast.error(m.mail_pdf_already_open());
        return;
      }
      console.error("Failed to open PDF:", error);
      toast.error(m.mail_error_pdf());
    }
  }

  async function openImageHandler(base64Data: string, filename: string) {
    try {
      if (settingsStore.settings.useBuiltinPreview) {
         await OpenImageWindow(base64Data, filename);
      } else {
         await OpenImage(base64Data, filename);
      }
    } catch (error) {
      console.error("Failed to open image:", error);
        toast.error(m.mail_error_image());
    }
  }

  async function openEMLHandler(base64Data: string, filename: string) {
    try {
      await OpenEMLWindow(base64Data, filename);
    } catch (error) {
       console.error("Failed to open EML:", error);
       toast.error("Failed to open EML attachment");
    }
  }

  async function onOpenMail() {
    isLoading = true;
    loadingText = m.layout_loading_text();
    const result = await ShowOpenFileDialog();
    if (result && result.length > 0) {
      // Handle opening the mail file
      try {
        // If the file is .eml, otherwise if is .msg, read accordingly
        let email: internal.EmailData;
        if(result.toLowerCase().endsWith(".msg")) {
          loadingText = m.mail_loading_msg_conversion();
          email = await ReadMSG(result, true);
        } else {
          email = await ReadEML(result);
        }
        // Track the current mail file path for bug reports
        await SetCurrentMailFilePath(result);
        mailState.setParams(email);
        sidebarOpen.set(false);

      } catch (error) {
        console.error("Failed to read EML file:", error);
        toast.error(m.mail_error_opening());
      } finally {
        isLoading = false;
        loadingText = "";
      }
    } else {
        isLoading = false;
        loadingText = "";
    }
  }

  function arrayBufferToBase64(buffer: any): string {
    if (typeof buffer === "string") return buffer; // Already base64 string
    if (Array.isArray(buffer)) {
      let binary = "";
      const bytes = new Uint8Array(buffer);
      const len = bytes.byteLength;
      for (let i = 0; i < len; i++) {
        binary += String.fromCharCode(bytes[i]);
      }
      return window.btoa(binary);
    }
    return "";
  }

  function handleWheel(event: WheelEvent) {
    if (event.ctrlKey) {
      event.preventDefault();
    }
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
      <div class="email-view">
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
                <MailOpen size="15" ></MailOpen>
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

          <div class="email-meta-grid">
            <span class="label">{m.mail_from()}</span>
            <span class="value">{mailState.currentEmail.from}</span>

            {#if mailState.currentEmail.to && mailState.currentEmail.to.length > 0}
              <span class="label">{m.mail_to()}</span>
              <span class="value">{mailState.currentEmail.to.join(", ")}</span>
            {/if}

            {#if mailState.currentEmail.cc && mailState.currentEmail.cc.length > 0}
              <span class="label">{m.mail_cc()}</span>
              <span class="value">{mailState.currentEmail.cc.join(", ")}</span>
            {/if}

            {#if mailState.currentEmail.bcc && mailState.currentEmail.bcc.length > 0}
              <span class="label">{m.mail_bcc()}</span>
              <span class="value">{mailState.currentEmail.bcc.join(", ")}</span>
            {/if}

            {#if mailState.currentEmail.isPec}
              <span class="label">{m.mail_sign_label()}</span>
              <span class="value"><span class="pec-badge" title="Posta Elettronica Certificata">
                   <ShieldCheck size="14" />
                   PEC
                </span></span>
            {/if}
          </div>
        </div>

        <div class="email-attachments">
          <span class="att-section-label">{m.mail_attachments()}</span>
          <div class="att-list">
            {#if mailState.currentEmail.attachments && mailState.currentEmail.attachments.length > 0}
              {#each mailState.currentEmail.attachments as att}
                {#if att.contentType.startsWith("image/")}
                  <button
                    class="att-btn image"
                    onclick={() => openImageHandler(arrayBufferToBase64(att.data), att.filename)}
                  >
                    <Image size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if att.contentType === "application/pdf" || att.filename.toLowerCase().endsWith(".pdf")}
                  <button
                    class="att-btn pdf"
                    onclick={() => openPDFHandler(arrayBufferToBase64(att.data), att.filename)}
                  >
                    <FileText />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if att.filename.toLowerCase().endsWith(".eml")}
                   <button
                    class="att-btn eml"
                    onclick={() => openEMLHandler(arrayBufferToBase64(att.data), att.filename)}
                  >
                    <MailOpen size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if mailState.currentEmail.isPec && att.filename.toLowerCase().endsWith(".p7s")}
                  <a
                    class="att-btn file"
                    href={`data:${att.contentType};base64,${arrayBufferToBase64(att.data)}`}
                    download={att.filename}
                  >
                    <Signature size="14" />
                    <span class="att-name">{att.filename}</span>
                  </a>
                {:else if mailState.currentEmail.isPec && att.filename.toLowerCase() === "daticert.xml"}
                  <a
                    class="att-btn file"
                    href={`data:${att.contentType};base64,${arrayBufferToBase64(att.data)}`}
                    download={att.filename}
                  >
                    <FileCode size="14" />
                    <span class="att-name">{att.filename}</span>
                  </a>
                {:else}
                  <a
                    class="att-btn file"
                    href={`data:${att.contentType};base64,${arrayBufferToBase64(att.data)}`}
                    download={att.filename}
                  >
                   {#if att.contentType.startsWith("image/")}
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

        <div class="email-body-wrapper">
          <iframe
            srcdoc={mailState.currentEmail.body + iFrameUtilHTML}
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

  /* Make sure internal loader spins if not using class-based animation library like Tailwind */
  :global(.spinner) {
      animation: spin 1s linear infinite;
  }

  @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
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
  
  .att-btn.image { color: #60a5fa; border-color: rgba(96, 165, 250, 0.3); } 
  .att-btn.image:hover { color: #93c5fd; }

  .att-btn.pdf { color: #f87171; border-color: rgba(248, 113, 113, 0.3); }
  .att-btn.pdf:hover { color: #fca5a5; }

  .att-btn.eml { color: hsl(49, 80%, 49%); border-color: rgba(224, 206, 39, 0.3); }
  .att-btn.eml:hover { color: hsl(49, 80%, 65%); }

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

  .browse-btn:disabled, .btn:disabled {
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
