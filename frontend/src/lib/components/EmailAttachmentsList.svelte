<script lang="ts">
    import { Image, FileText, MailOpen, File } from "@lucide/svelte";
    import { toast } from "svelte-sonner";
    import { hostIntegrityFailed } from "$lib/stores/app";
    import * as m from "$lib/paraglide/messages";
    import type { internal } from "$lib/wailsjs/go/models";
    import {
        CONTENT_TYPES,
        arrayBufferToBase64,
        openPDFAttachment,
        openImageAttachment,
        openEMLAttachment,
        openDocAttachment,
    } from "$lib/utils/mail";
    import { showDefaultAttachmentToast } from "$lib/utils/open-default-attachment-toast";
    import { saveAttachmentNatively } from "$lib/utils/attachment-download";

    let { attachments }: { attachments: internal.EmailAttachment[] | undefined } =
        $props();

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

    async function handleOpenPDF(base64Data: string, filename: string) {
        if (isAttachmentBlocked()) return;
        await openPDFAttachment(base64Data, filename);
    }

    async function handleOpenImage(base64Data: string, filename: string) {
        if (isAttachmentBlocked()) return;
        await openImageAttachment(base64Data, filename);
    }

    async function handleOpenEML(base64Data: string, filename: string) {
        if (isAttachmentBlocked()) return;
        await openEMLAttachment(base64Data, filename);
    }

    async function handleOpenDoc(base64Data: string, filename: string) {
        if (isAttachmentBlocked()) return;
        await openDocAttachment(base64Data, filename);
    }
</script>

<div class="email-attachments">
    <span class="att-section-label">{m.mail_attachments()}</span>
    <div class="att-list">
        {#if attachments && attachments.length > 0}
            {#each attachments as att}
                {@const base64 = arrayBufferToBase64(att.data)}
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

                {#if isImage}
                    <button
                        class="att-btn image"
                        class:integrity-blocked={$hostIntegrityFailed}
                        onclick={() =>
                            handleOpenImage(
                                base64,
                                att.filename,
                            )}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        <Image size="16" />
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else if isPdf}
                    <button
                        class="att-btn pdf"
                        class:integrity-blocked={$hostIntegrityFailed}
                        onclick={() =>
                            handleOpenPDF(base64, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        <FileText size="16" />
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else if isEml}
                    <button
                        class="att-btn eml"
                        class:integrity-blocked={$hostIntegrityFailed}
                        onclick={() =>
                            handleOpenEML(base64, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        <MailOpen size="16" />
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else if isDoc}
                    <button
                        class="att-btn doc"
                        class:integrity-blocked={$hostIntegrityFailed}
                        onclick={() =>
                            handleOpenDoc(base64, att.filename)}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        <FileText size="16" />
                        <span class="att-name"
                            >{att.filename}</span
                        >
                    </button>
                {:else}
                    <button
                        class="att-btn file"
                        class:integrity-blocked={$hostIntegrityFailed}
                        onclick={() => {
                            if (isAttachmentBlocked()) return;
                            showDefaultAttachmentToast({
                                onSave: () =>
                                    void saveAttachmentNatively(
                                        base64,
                                        att.filename,
                                    ),
                                onReset: () => {},
                            });
                        }}
                        title={$hostIntegrityFailed
                            ? m.mail_attachments_disabled_hint()
                            : undefined}
                    >
                        <File size="16" />
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

    .att-empty {
        font-size: 11px;
        color: var(--muted-foreground);
        font-style: italic;
    }
</style>
