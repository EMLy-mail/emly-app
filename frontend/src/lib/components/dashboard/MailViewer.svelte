<script lang="ts">
  import { X, MailOpen, Image, FileText, File } from "@lucide/svelte";
  import { ShowOpenFileDialog, ReadEML, OpenPDF, OpenImageWindow } from "$lib/wailsjs/go/main/App";
  import type { internal } from "$lib/wailsjs/go/models";
  import { sidebarOpen } from "$lib/stores/app";
  import { onDestroy, onMount } from "svelte";
  import { toast } from "svelte-sonner";
  import { EventsOn, WindowShow, WindowUnminimise } from "$lib/wailsjs/runtime/runtime";
  import type { SupportedFileTypePreview } from "$lib/types";
  import { mailState } from "$lib/stores/mail-state.svelte";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import * as m from "$lib/paraglide/messages";

  let unregisterEvents = () => {};
  let isLoading = $state(false);

  function onClear() {
    mailState.clear();
  }

  $effect(() => {
    console.log("Current email changed:", mailState.currentEmail);
    if(mailState.currentEmail !== null) {
        sidebarOpen.set(false);
    }
  })

  onDestroy(() => {
    if (unregisterEvents) unregisterEvents();
  });

  onMount(async () => {
    // Listen for second instance args
    unregisterEvents = EventsOn("launchArgs", async (args: string[]) => {
      if (args && args.length > 0) {
        for (const arg of args) {
          if (arg.toLowerCase().endsWith(".eml")) {
            console.log("Loading EML from second instance:", arg);
            isLoading = true;
            try {
              const emlContent = await ReadEML(arg);
              mailState.setParams(emlContent);
              sidebarOpen.set(false);
              WindowUnminimise();
              WindowShow();
            } finally {
              isLoading = false;
            }
            break;
          }
        }
      }
    });
  });

  async function openPDFHandler(base64Data: string, filename: string) {
    try {
      await OpenPDF(base64Data, filename);
    } catch (error) {
      console.error("Failed to open PDF:", error);
        toast.error(m.mail_error_pdf());
    }
  }

  async function onOpenMail() {
    isLoading = true;
    const result = await ShowOpenFileDialog();
    if (result && result.length > 0) {
      // Handle opening the mail file
      try {
        const email: internal.EmailData = await ReadEML(result);
        mailState.setParams(email);
        sidebarOpen.set(false);

      } catch (error) {
        console.error("Failed to read EML file:", error);
      } finally {
        isLoading = false;
      }
    } else {
        isLoading = false;
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

  function getFileExtension(filename: string): string {
    return filename.slice((filename.lastIndexOf(".") - 1 >>> 0) + 2).toLowerCase();
  }

  function shouldPreview(filename: string): boolean {
    if (!settingsStore.settings.useBuiltinPreview) return false;
    const ext = getFileExtension(filename);
    const supported = settingsStore.settings.previewFileSupportedTypes || [];
    return supported.includes(ext as SupportedFileTypePreview);
  }
</script>

<div class="panel fill" aria-label="Events">
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
          </div>
        </div>

        <div class="email-attachments">
          <span class="att-section-label">{m.mail_attachments()}</span>
          <div class="att-list">
            {#if mailState.currentEmail.attachments && mailState.currentEmail.attachments.length > 0}
              {#each mailState.currentEmail.attachments as att}
                {#if att.contentType.startsWith("image/") && shouldPreview(att.filename)}
                  <button
                    class="att-btn image"
                    onclick={() => OpenImageWindow(arrayBufferToBase64(att.data), att.filename)}
                  >
                    <Image size="14" />
                    <span class="att-name">{att.filename}</span>
                  </button>
                {:else if att.contentType === "application/pdf" || att.filename.toLowerCase().endsWith(".pdf")}
                  <button
                    class="att-btn pdf"
                    onclick={() => openPDFHandler(arrayBufferToBase64(att.data), att.filename)}
                  >
                    <FileText size="15" />
                    <span class="att-name">{att.filename}</span>
                  </button>
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
            srcdoc={mailState.currentEmail.body +
              "<style>body{margin:0;padding:20px;font-family:sans-serif;} a{pointer-events:none!important;cursor:default!important;}</style>"}
            title="Email Body"
            class="email-iframe"
            sandbox="allow-same-origin"
          ></iframe>
        </div>
      </div>
    {/if}
    </div>  
</div>

<style>
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
</style>
