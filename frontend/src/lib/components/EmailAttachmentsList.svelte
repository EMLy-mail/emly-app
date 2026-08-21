<script lang="ts">
    import { Image, FileText, MailOpen, File, Loader2 } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { hostIntegrityFailed } from "$lib/stores/app";
    import * as m from "$lib/paraglide/messages";
    import type { mailfmt } from "$lib/wailsjs/go/models";
    import { GetAttachmentData } from "$lib/wailsjs/go/main/App";
    import {
        CONTENT_TYPES,
        arrayBufferToBase64,
        hasPreloadedAttachmentData,
        openPDFAttachment,
        openImageAttachment,
        openEMLAttachment,
        openDocAttachment,
    } from "$lib/utils/mail";
    import { showDefaultAttachmentToast } from "$lib/utils/open-default-attachment-toast";
    import { saveAttachmentNatively } from "$lib/utils/attachment-download";
    import { settingsStore } from "$lib/stores/settings.svelte";

    let {
        attachments,
        filePath,
    }: {
        attachments: mailfmt.EmailAttachment[] | undefined;
        filePath: string | undefined;
    } = $props();

    // ReadEML/ReadMSG/ReadAuto no longer send attachment bytes up front (see
    // backend/app_mail.go: stripAttachmentData) - a heavy mail used to ship
    // every attachment's bytes across the Wails IPC bridge before the user
    // had looked at anything. Instead, fetch one attachment's bytes only
    // when its button is actually clicked (open or save).
    let pendingIndex = $state<number | null>(null);

    async function fetchAttachmentBase64(index: number): Promise<string | null> {
        // "Old Pre-loading of attachments" (Settings → Danger Zone, off by
        // default) makes the backend send full bytes up front - use them
        // directly and skip the extra round trip entirely.
        const att = attachments?.[index];
        if (att && hasPreloadedAttachmentData(att.data)) {
            return arrayBufferToBase64(att.data);
        }

        if (!filePath) return null;
        pendingIndex = index;
        try {
            return await GetAttachmentData(filePath, index);
        } catch (error) {
            console.error("Failed to fetch attachment data:", error);
            toast.error(m.attachment_fetch_error());
            return null;
        } finally {
            pendingIndex = null;
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

    function isSupportedImageType(contentType: string): boolean {
            let supportedTypes = settingsStore.settings.previewFileSupportedTypes;
            if (!supportedTypes || !contentType) return false;

            let normalizedContentType = contentType.toLowerCase().split(";")[0].trim();

            for (let type of supportedTypes) {
                if (!type) continue;

                let normalizedType = type.toLowerCase().trim();

                // Allow shorthand entries like "jpg", "jpeg", "png" in settings
                if (!normalizedType.includes("/")) {
                    normalizedType = normalizedType === "jpg" ? "image/jpeg" : `image/${normalizedType}`;
                }

                if (normalizedContentType === normalizedType) {
                    return true;
                }
            }
            return false;
        }

    async function handleOpenPDF(index: number, filename: string) {
        if (isAttachmentBlocked()) return;
        const base64Data = await fetchAttachmentBase64(index);
        if (base64Data === null) return;
        await openPDFAttachment(base64Data, filename);
    }

    async function handleOpenImage(index: number, filename: string) {
        if (isAttachmentBlocked()) return;
        const base64Data = await fetchAttachmentBase64(index);
        if (base64Data === null) return;
        await openImageAttachment(base64Data, filename);
    }

    async function handleOpenEML(index: number, filename: string) {
        if (isAttachmentBlocked()) return;
        const base64Data = await fetchAttachmentBase64(index);
        if (base64Data === null) return;
        await openEMLAttachment(base64Data, filename);
    }

    async function handleOpenDoc(index: number, filename: string) {
        if (isAttachmentBlocked()) return;
        const base64Data = await fetchAttachmentBase64(index);
        if (base64Data === null) return;
        await openDocAttachment(base64Data, filename);
    }

    async function handleSaveDefault(index: number, filename: string) {
        if (isAttachmentBlocked()) return;
        showDefaultAttachmentToast({
            onSave: async () => {
                const base64Data = await fetchAttachmentBase64(index);
                if (base64Data === null) return;
                void saveAttachmentNatively(base64Data, filename);
            },
            onReset: () => {},
        });
    }
</script>

<div class="email-attachments">
    <span class="att-section-label">{m.mail_attachments()}</span>
    <div class="att-list">
        {#if attachments && attachments.length > 0}
            {#each attachments as att, index}
                {@const isImage = att.contentType.startsWith(
                    CONTENT_TYPES.IMAGE,
                )}
                {@const isPdf =
                    att.contentType === CONTENT_TYPES.PDF ||
                    att.filename.toLowerCase().endsWith(".pdf")}
                {@const isEml = att.filename
                    .toLowerCase()
                    .endsWith(".eml")}
                {@const isDoc =
                    att.contentType === CONTENT_TYPES.DOCX ||
                    att.contentType === CONTENT_TYPES.DOC ||
                    att.filename
                        .toLowerCase()
                        .endsWith(".docx") ||
                    att.filename.toLowerCase().endsWith(".doc")}
                {@const isPending = pendingIndex === index}

                {#if isImage && isSupportedImageType(att.contentType)}
                    <button
                        class="att-btn image"
                        class:integrity-blocked={$hostIntegrityFailed}
                        disabled={isPending}
                        onclick={() => handleOpenImage(index, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        {#if isPending}
                            <Loader2 size="16" class="att-spinner" />
                        {:else}
                            <Image size="16" />
                        {/if}
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else if isPdf}
                    <button
                        class="att-btn pdf"
                        class:integrity-blocked={$hostIntegrityFailed}
                        disabled={isPending}
                        onclick={() => handleOpenPDF(index, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        {#if isPending}
                            <Loader2 size="16" class="att-spinner" />
                        {:else}
                            <FileText size="16" />
                        {/if}
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else if isEml}
                    <button
                        class="att-btn eml"
                        class:integrity-blocked={$hostIntegrityFailed}
                        disabled={isPending}
                        onclick={() => handleOpenEML(index, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        {#if isPending}
                            <Loader2 size="16" class="att-spinner" />
                        {:else}
                            <MailOpen size="16" />
                        {/if}
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else if isDoc}
                    <button
                        class="att-btn doc"
                        class:integrity-blocked={$hostIntegrityFailed}
                        disabled={isPending}
                        onclick={() => handleOpenDoc(index, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        {#if isPending}
                            <Loader2 size="16" class="att-spinner" />
                        {:else}
                            <FileText size="16" />
                        {/if}
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else}
                    <button
                        class="att-btn file"
                        class:integrity-blocked={$hostIntegrityFailed}
                        disabled={isPending}
                        onclick={() => handleSaveDefault(index, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        {#if isPending}
                            <Loader2 size="16" class="att-spinner" />
                        {:else}
                            <File size="16" />
                        {/if}
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {/if}
            {/each}
        {:else}
            <span class="att-empty"
                >{m.mail_no_attachments()}</span
            >
        {/if}
    </div>
</div>

<style>
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

    /* Not `:disabled` on purpose: the button stays interactive so hover
       still shows the title tooltip and click still fires (to surface the
       "blocked by integrity check" toast) instead of being swallowed. */
    .att-btn.integrity-blocked {
        opacity: 0.4;
        cursor: not-allowed;
    }

    .att-btn.integrity-blocked:hover {
        background: transparent;
        color: var(--foreground);
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
        min-width: 0;
    }

    .att-btn :global(svg) {
        flex-shrink: 0;
    }

    .att-btn:disabled {
        cursor: default;
        opacity: 0.7;
    }

    :global(.att-spinner) {
        animation: att-spin 0.8s linear infinite;
    }

    @keyframes att-spin {
        from { transform: rotate(0deg); }
        to   { transform: rotate(360deg); }
    }

    .att-empty {
        font-size: 11px;
        color: var(--muted-foreground);
        font-style: italic;
    }
</style>
