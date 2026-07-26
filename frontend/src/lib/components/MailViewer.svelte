<script lang="ts">
    import {
        X,
        MailOpen,
        ShieldCheck,
        ShieldAlert,
        Loader2,
        Download,
        Info,
    } from "@lucide/svelte";
    import { dev } from "$app/environment";
    import {
        sidebarOpen,
        runningInDebugMode,
        hostIntegrityFailed,
    } from "$lib/stores/app";
    import { onDestroy, onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import {
        BrowserOpenURL,
    } from "$lib/wailsjs/runtime/runtime";
    import { mailState } from "$lib/stores/mail-state.svelte";
    import type { mailfmt } from "$lib/wailsjs/go/models";
    import * as m from "$lib/paraglide/messages";
    import { cancelCurrentToast } from "$lib/utils/open-default-attachment-toast";
    import { saveAllAttachmentsNatively } from "$lib/utils/attachment-download";
    import LinkConfirmationDialog from "$lib/components/LinkConfirmationDialog.svelte";
    import MailDebugInfoDialog from "$lib/components/MailDebugInfoDialog.svelte";
    import EmailAttachmentsList from "$lib/components/EmailAttachmentsList.svelte";

    import {
        IFRAME_UTIL_HTML_DARK,
        IFRAME_UTIL_HTML_DARK_NO_LINKS,
        IFRAME_UTIL_HTML_LIGHT,
        IFRAME_UTIL_HTML_LIGHT_NO_LINKS,
        IFRAME_CONTRAST_FIX_JS,
        arrayBufferToBase64,
        openAndLoadEmail,
        processEmailBody,
        isPecOpenBlocked,
    } from "$lib/utils/mail";
    import { settingsStore } from "$lib/stores/settings.svelte";

    // ============================================================================
    // Props
    // ============================================================================

    let {
        emailData = null,
        tabId = null,
        embedded = false,
    }: {
        emailData?: mailfmt.EmailData | null;
        tabId?: string | null;
        embedded?: boolean;
    } = $props();

    // ============================================================================
    // State
    // ============================================================================

    let isLoading = $state(false);
    let loadingText = $state("");
    let linkDialogOpen = $state(false);
    let pendingLinkUrl = $state("");
    let disabledLinkClickCount = $state(0);
    let debugModalOpen = $state(false);

    const LINK_HINT_TOAST_ID = "emly-link-hint";

    // In tab mode, read from the specific tab in mailState.tabs.
    // In non-tab mode, read from mailState.currentEmail (which reads the active tab).
    let activeEmail = $derived<mailfmt.EmailData | null>(
        tabId !== null
            ? (() => {
                  const tab = mailState.tabs.find((t) => t.id === tabId);
                  return tab?.type === "email" ? (tab.email ?? null) : null;
              })()
            : mailState.currentEmail,
    );

    

    let activeFilePath = $derived<string | undefined>(
        tabId !== null
            ? (() => {
                  const tab = mailState.tabs.find((t) => t.id === tabId);
                  return tab?.type === "email" ? tab.filePath : undefined;
              })()
            : (() => {
                  const tab = mailState.tabs.find(
                      (t) => t.id === mailState.activeTabId,
                  );
                  return tab?.type === "email" ? tab.filePath : undefined;
              })(),
    );

    let iframeUtilHtml = $derived(
        settingsStore.settings.useDarkEmailViewer !== false
            ? settingsStore.settings.enableLinkClickConfirmation !== false
                ? IFRAME_UTIL_HTML_DARK
                : IFRAME_UTIL_HTML_DARK_NO_LINKS
            : settingsStore.settings.enableLinkClickConfirmation !== false
              ? IFRAME_UTIL_HTML_LIGHT
              : IFRAME_UTIL_HTML_LIGHT_NO_LINKS,
    );

    let contrastFixScript = $derived(
        settingsStore.settings.fixEmailTextContrast
            ? IFRAME_CONTRAST_FIX_JS
            : "",
    );

    // ============================================================================
    // Event Handlers
    // ============================================================================

    function onClear() {
        cancelCurrentToast();
        if (tabId !== null) {
            mailState.removeTab(tabId);
        } else {
            mailState.clear();
            sidebarOpen.set(true);
        }
    }

    // Returns true (and shows a toast) if attachment interactions should be
    // blocked because the host integrity check failed. Buttons stay natively
    // enabled (not `disabled`) so hover tooltips keep working; this guard is
    // what actually stops the action on click.
    function isAttachmentBlocked(): boolean {
        if ($hostIntegrityFailed) {
            toast.error(m.mail_attachments_disabled_hint());
            return true;
        }
        return false;
    }

    async function onDownloadAttachments() {
        if (isAttachmentBlocked()) return;
        if (!activeEmail || !activeEmail.attachments) return;

        await saveAllAttachmentsNatively(
            activeEmail.attachments.map((att) => ({
                filename: att.filename,
                base64: arrayBufferToBase64(att.data),
            })),
        );
    }

    async function onOpenMail() {
        isLoading = true;
        loadingText = m.layout_loading_text();

        const result = await openAndLoadEmail();

        if (result.cancelled) {
            isLoading = false;
            loadingText = "";
            return;
        }

        if (result.success && result.email) {
            if (isPecOpenBlocked(result.email)) {
                isLoading = false;
                loadingText = "";
                return;
            }
            if (tabId !== null) {
                mailState.addTab(result.email, result.filePath);
            } else {
                mailState.setParams(result.email, result.filePath);
                sidebarOpen.set(false);
            }
        } else if (result.error) {
            console.error("Failed to read email file:", result.error);
            toast.error(m.mail_error_opening());
        }

        isLoading = false;
        loadingText = "";
    }

    function handleWheel(event: WheelEvent) {
        if (event.ctrlKey) {
            event.preventDefault();
        }
    }

    function handleIframeMessage(event: MessageEvent) {
        // Only react to messages coming from this instance's own iframe.
        if (!iframeEl || event.source !== iframeEl.contentWindow) return;

        if (event.data?.type === "emly-link-disabled-click") {
            disabledLinkClickCount++;
            if (disabledLinkClickCount >= 2) {
                toast(m.mail_link_disabled_toast(), {
                    id: LINK_HINT_TOAST_ID,
                    duration: 10000,
                    action: {
                        label: m.mail_link_disabled_enable(),
                        onClick: () => {
                            settingsStore.update({
                                enableLinkClickConfirmation: true,
                            });
                        },
                    },
                });
            }
            return;
        }

        if (
            settingsStore.settings.enableLinkClickConfirmation !== false &&
            event.data?.type === "emly-link-click" &&
            typeof event.data.url === "string" &&
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
        pendingLinkUrl = "";
        if (url) BrowserOpenURL(url);
    }

    function onCancelOpenLink() {
        linkDialogOpen = false;
        pendingLinkUrl = "";
    }

    // ============================================================================
    // Effects
    // ============================================================================

    $effect(() => {
        const processCurrentEmail = async () => {
            disabledLinkClickCount = 0;
            toast.dismiss(LINK_HINT_TOAST_ID);

            if (activeEmail?.body) {
                let startParseTime = new Date();
                const processedBody = await processEmailBody(activeEmail.body);
                let finishParseTime = new Date();
                let parseTime = finishParseTime.getTime() - startParseTime.getTime();
                console.log("Parse -> B64 to HTML -> time took: " + parseTime + " ms")
                if (processedBody !== activeEmail.body) {
                    activeEmail.body = processedBody;
                }
            }
            console.info("Current email changed:", activeEmail?.subject);
            // Log the email info: Format, N of attachments, and whether it has a body, but only if it's not too large (to avoid spamming the logs)
            if (activeEmail) {
                const bodyInfo = activeEmail.body
                    ? `(body length: ${activeEmail.body.length})`
                    : "(no body)";
                const attachmentsInfo = activeEmail.attachments
                    ? `(${activeEmail.attachments.length} attachments)`
                    : "(no attachments)";
                const isPecInfo = activeEmail.isPec ? "(PEC)" : "(not PEC)";
                console.info(
                    `Email info: ${bodyInfo} ${attachmentsInfo} ${isPecInfo}`,
                );
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

    onMount(() => {
        window.addEventListener("message", handleIframeMessage);
    });

    onDestroy(() => {
        cancelCurrentToast();
        window.removeEventListener("message", handleIframeMessage);
    });

</script>

<LinkConfirmationDialog
    bind:open={linkDialogOpen}
    url={pendingLinkUrl}
    onConfirm={onConfirmOpenLink}
    onCancel={onCancelOpenLink}
/>

{#if dev || $runningInDebugMode}
    <MailDebugInfoDialog
        bind:open={debugModalOpen}
        {activeEmail}
        {activeFilePath}
    />
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
                <button
                    class="browse-btn"
                    onclick={onOpenMail}
                    disabled={isLoading}
                >
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
                                {activeEmail.subject ||
                                    m.mail_subject_no_subject()}
                            </div>
                            {#if dev || $runningInDebugMode}
                                <button
                                    class="debug-info-btn"
                                    onclick={() => (debugModalOpen = true)}
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
                                class:integrity-blocked={$hostIntegrityFailed}
                                onclick={onDownloadAttachments}
                                aria-label={m.mail_download_btn_label()}
                                title={$hostIntegrityFailed
                                    ? m.mail_attachments_disabled_hint()
                                    : m.mail_download_btn_title()}
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
                            <span class="value"
                                >{activeEmail.to.join(", ")}</span
                            >
                        {/if}

                        {#if activeEmail.cc && activeEmail.cc.length > 0}
                            <span class="label">{m.mail_cc()}</span>
                            <span class="value"
                                >{activeEmail.cc.join(", ")}</span
                            >
                        {/if}

                        {#if activeEmail.bcc && activeEmail.bcc.length > 0}
                            <span class="label">{m.mail_bcc()}</span>
                            <span class="value"
                                >{activeEmail.bcc.join(", ")}</span
                            >
                        {/if}

                        {#if activeEmail.isPec}
                            <span class="label">{m.mail_sign_label()}</span>
                            <span class="value">
                                <span
                                    class="pec-badge"
                                    title="Posta Elettronica Certificata"
                                >
                                    <ShieldCheck size="14" />
                                    PEC
                                </span>
                            </span>
                        {/if}

                        {#if activeEmail.date}
                            <span class="label">{m.mail_date()}</span>
                            {#if settingsStore.settings.selectedLanguage === "it"}
                                <span class="value"
                                    >{new Intl.DateTimeFormat("it-IT", {
                                        dateStyle: "full",
                                        timeStyle: "long",
                                    }).format(new Date(activeEmail.date))}</span
                                >
                            {:else}
                                <span class="value"
                                    >{new Intl.DateTimeFormat("en-GB", {
                                        dateStyle: "full",
                                        timeStyle: "long",
                                    }).format(new Date(activeEmail.date))}</span
                                >
                            {/if}
                        {/if}
                    </div>
                </div>

                <!-- Attachments -->
                <EmailAttachmentsList attachments={activeEmail.attachments} />

                <!-- Email Body -->
                <div
                    class="email-body-wrapper"
                    class:light-theme={settingsStore.settings
                        .useDarkEmailViewer === false}
                    class:integrity-blurred={$hostIntegrityFailed}
                >
                    <iframe
                        bind:this={iframeEl}
                        srcdoc={activeEmail.body +
                            iframeUtilHtml +
                            contrastFixScript}
                        title={m.mail_email_body_title()}
                        class="email-iframe"
                        sandbox="allow-scripts"
                        onwheel={handleWheel}
                    ></iframe>
                    {#if $hostIntegrityFailed}
                        <div class="integrity-blur-notice">
                            <ShieldAlert size="18" />
                            {m.mail_body_blurred_hint()}
                        </div>
                    {/if}
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

    .email-body-wrapper.integrity-blurred .email-iframe {
        filter: blur(14px);
        pointer-events: none;
        user-select: none;
    }

    .integrity-blur-notice {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        z-index: 5;
        display: flex;
        align-items: center;
        gap: 8px;
        max-width: 80%;
        padding: 10px 16px;
        border-radius: 10px;
        background: rgba(0, 0, 0, 0.75);
        color: #fff;
        border: 1px solid rgba(255, 255, 255, 0.15);
        font-size: 13px;
        font-weight: 500;
        text-align: center;
        pointer-events: none;
        user-select: none;
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

    /* Not `:disabled`: stays interactive so the title tooltip and the
       "blocked by integrity check" toast on click both keep working. */
    .btn.integrity-blocked {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .btn.integrity-blocked:hover {
        background: var(--muted);
        color: var(--muted-foreground);
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
        transition:
            opacity 0.15s,
            background 0.15s,
            color 0.15s;
    }

    .debug-info-btn:hover {
        opacity: 1;
        background: var(--muted);
        color: var(--foreground);
    }
</style>
