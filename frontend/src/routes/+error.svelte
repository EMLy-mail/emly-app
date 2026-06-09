<script lang="ts">
    import { page } from "$app/state";
    import * as m from "$lib/paraglide/messages.js";
    import { dev } from "$app/environment";
    import { onMount } from "svelte";
    import {
        WindowMinimise,
        WindowMaximise,
        WindowUnmaximise,
        WindowIsMaximised,
        Quit,
    } from "$lib/wailsjs/runtime/runtime";
    import { GetLogsDir, OpenFolderInExplorer } from "$lib/wailsjs/go/main/App";

    async function openLogs() {
        try {
            const dir = await GetLogsDir();
            await OpenFolderInExplorer(dir);
        } catch (e) {
            console.error("Failed to open logs folder", e);
        }
    }

    let isMaximized = $state(false);
    let windowFocused = $state(true);

    async function toggleMaximize() {
        if (isMaximized) {
            WindowUnmaximise();
        } else {
            WindowMaximise();
        }
        isMaximized = !isMaximized;
    }

    WindowIsMaximised().then((v) => (isMaximized = v));

    onMount(() => {
        window.addEventListener("focus", () => (windowFocused = true));
        window.addEventListener("blur", () => (windowFocused = false));
    });
</script>

<div class="app-layout">
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
        class="titlebar"
        ondblclick={toggleMaximize}
        style="--wails-draggable:drag"
    >
        <div class="title">EMLy</div>

        <div class="controls" style:opacity={windowFocused ? 1 : 0.4}>
            <button class="btn" onmousedown={() => WindowMinimise()}>─</button>
            <button class="btn" onmousedown={toggleMaximize}>
                {#if isMaximized}
                    ❐
                {:else}
                    ☐
                {/if}
            </button>
            <button class="btn close" onmousedown={() => Quit()}>✕</button>
        </div>
    </div>

    <div class="page">
        <div class="error-container">
            {#if page.status === 500}
                {#if dev}
                    <img class="error-img" src="/uhoh-proot.webp" alt="uh oh" />
                {/if}
                <p class="error-title">{m.error_500_title()}</p>
                <hr class="divider" />
                <p class="error-message">{m.error_500_message()}</p>
                <div class="btn-row">
                    <button class="quit-btn secondary" onclick={openLogs}>
                        {m.error_open_logs()}
                    </button>
                    <button class="quit-btn" onclick={() => Quit()}>
                        {m.error_close_app()}
                    </button>
                </div>
            {:else}
                <div class="error-icon">{page.status}</div>
                <p class="error-message">
                    {page.error?.message || m.error_unexpected()}
                </p>
            {/if}
        </div>
    </div>
</div>

<style>
    .app-layout {
        display: flex;
        flex-direction: column;
        height: 100vh;
        overflow: hidden;
    }

    .titlebar {
        height: 32px;
        background: var(--background);
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding-left: 12px;
        -webkit-app-region: drag;
        user-select: none;
        flex: 0 0 32px;
        z-index: 50;
        border-bottom: 1px solid var(--border);
    }

    .title {
        font-size: 13px;
        font-weight: 500;
        color: var(--muted-foreground);
    }

    .controls {
        display: flex;
        height: 100%;
    }

    .btn {
        width: 46px;
        height: 100%;
        border: none;
        background: transparent;
        color: var(--foreground);
        font-size: 14px;
        cursor: pointer;
        -webkit-app-region: no-drag;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .btn:hover {
        background: var(--accent);
    }

    .close:hover {
        background: #e81123;
        color: white;
    }

    .page {
        flex: 1;
        min-height: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 12px;
        box-sizing: border-box;
        overflow: hidden;
    }

    .error-container {
        padding: 40px;
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        max-width: 420px;
        width: 100%;
    }

    .error-icon {
        font-size: 64px;
        line-height: 1;
        margin: 0 0 12px 0;
        opacity: 0.85;
    }

    .error-img {
        width: 180px;
        height: auto;
        margin: 0 0 16px 0;
        border-radius: 8px;
    }


    .status-badge {
        display: inline-block;
        font-size: 11px;
        font-weight: 600;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: #e05555;
        background: rgba(220, 50, 50, 0.15);
        border: 1px solid rgba(220, 50, 50, 0.3);
        border-radius: 999px;
        padding: 2px 10px;
        margin: 0 0 14px 0;
    }

    .error-title {
        font-size: 20px;
        font-weight: 700;
        color: #e07070;
        margin: 0 0 12px 0;
        line-height: 1.3;
    }

    .divider {
        width: 48px;
        border: none;
        border-top: 1px solid rgba(220, 50, 50, 0.25);
        margin: 0 0 14px 0;
    }

    .error-message {
        font-size: 14px;
        opacity: 0.6;
        margin: 0 0 28px 0;
        line-height: 1.6;
    }

    .btn-row {
        display: flex;
        gap: 10px;
        flex-wrap: wrap;
        justify-content: center;
    }

    .quit-btn {
        padding: 8px 24px;
        border-radius: 6px;
        border: 1px solid rgba(220, 50, 50, 0.4);
        background: rgba(220, 50, 50, 0.15);
        color: #e07070;
        font-size: 14px;
        font-weight: 500;
        cursor: pointer;
        transition: background 0.15s, border-color 0.15s;
    }

    .quit-btn:hover {
        background: rgba(220, 50, 50, 0.28);
        border-color: rgba(220, 50, 50, 0.65);
    }

    .quit-btn.secondary {
        border-color: rgba(255, 255, 255, 0.15);
        background: rgba(255, 255, 255, 0.05);
        color: var(--muted-foreground, #888);
    }

    .quit-btn.secondary:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: rgba(255, 255, 255, 0.3);
    }
</style>
