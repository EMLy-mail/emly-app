<script lang="ts">
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Info, ShieldCheck, Loader2, FolderOpen } from "@lucide/svelte";
    import * as m from "$lib/paraglide/messages";
    import {
        DetectEmailFormat,
        OpenFolderInExplorer,
    } from "$lib/wailsjs/go/main/App";
    import type { mailfmt } from "$lib/wailsjs/go/models";

    let {
        open = $bindable(false),
        activeEmail,
        activeFilePath,
    }: {
        open: boolean;
        activeEmail: mailfmt.EmailData | null;
        activeFilePath?: string;
    } = $props();

    let debugFormat = $state("");
    let debugFormatLoading = $state(false);

    async function loadDebugFormat() {
        debugFormat = "";
        const fp = activeFilePath;
        if (fp) {
            debugFormatLoading = true;
            try {
                debugFormat = (await DetectEmailFormat(fp)) as string;
            } catch {
                debugFormat = "unknown";
            }
            debugFormatLoading = false;
        }
    }

    $effect(() => {
        if (open) {
            loadDebugFormat();
        }
    });

    function getDebugFolderPath(filePath: string): string {
        const lastSep = Math.max(
            filePath.lastIndexOf("/"),
            filePath.lastIndexOf("\\"),
        );
        return lastSep >= 0 ? filePath.substring(0, lastSep) : filePath;
    }

    function getDebugFormatLabel(): string {
        if (debugFormatLoading) return "…";
        if (!activeFilePath && !activeEmail?.isPec)
            return activeEmail ? "EML" : "—";
        const fmt = debugFormat.toLowerCase();
        if (fmt === "msg") return "MSG";
        if (fmt === "eml" || fmt === "")
            return activeEmail?.isPec ? "EML (PEC)" : "EML";
        if (fmt === "unknown") return m.debug_info_format_unknown();
        return fmt.toUpperCase();
    }

    function getBodyInfo(): string {
        const body = activeEmail?.body;
        if (!body) return m.debug_info_body_none();
        const trimmed = body.trimStart();
        const isHtmlBody =
            trimmed.startsWith("<") ||
            /<!doctype html/i.test(trimmed) ||
            /<html/i.test(trimmed);
        const kb = (body.length / 1024).toFixed(1);
        return `${isHtmlBody ? m.debug_info_body_html() : m.debug_info_body_text()}, ${kb} KB`;
    }
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="debug-dialog-content">
        <Dialog.Header>
            <Dialog.Title class="debug-dialog-title">
                <Info size="16" />
                {m.debug_info_title()}
            </Dialog.Title>
            <Dialog.Description
                >{m.debug_info_description()}</Dialog.Description
            >
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
                    <span class="pec-badge"
                        ><ShieldCheck size="11" /> PEC</span
                    >
                {:else}
                    {m.debug_info_no()}
                {/if}
            </span>

            <span class="debug-label">{m.debug_info_inner_email()}</span>
            <span class="debug-value"
                >{activeEmail?.hasInnerEmail
                    ? m.debug_info_yes()
                    : m.debug_info_no()}</span
            >

            <span class="debug-label">{m.debug_info_attachments()}</span>
            <span class="debug-value">
                {activeEmail?.attachments?.length ?? 0}
                {#if activeEmail?.attachments && activeEmail.attachments.length > 0}
                    <ul class="debug-att-list">
                        {#each activeEmail.attachments as att}
                            <li>
                                <span class="mono">{att.filename}</span>
                                <span class="debug-content-type"
                                    >{att.contentType}</span
                                >
                            </li>
                        {/each}
                    </ul>
                {/if}
            </span>

            <span class="debug-label">{m.debug_info_body()}</span>
            <span class="debug-value">{getBodyInfo()}</span>

            <span class="debug-label">{m.debug_info_date_raw()}</span>
            <span class="debug-value mono">{activeEmail?.date || "—"}</span>

            <span class="debug-label">{m.debug_info_file()}</span>
            <span class="debug-value mono debug-filepath"
                >{activeFilePath || "—"}</span
            >
        </div>

        <Dialog.Footer>
            {#if activeFilePath}
                <Button
                    variant="outline"
                    onclick={() =>
                        OpenFolderInExplorer(
                            getDebugFolderPath(activeFilePath!),
                        )}
                >
                    <FolderOpen size="14" />
                    {m.debug_info_show_in_explorer()}
                </Button>
            {/if}
            <Button onclick={() => (open = false)}
                >{m.debug_info_close()}</Button
            >
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>

<style>
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

    /* Duplicated from MailViewer.svelte's own <style>: that component still
       renders a .pec-badge in its header meta-grid, and Svelte's scoped CSS
       means this component needs its own copy for the badge rendered here. */
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
