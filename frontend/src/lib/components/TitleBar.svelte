<script lang="ts">
    import type { Snippet } from "svelte";
    import { onMount, onDestroy } from "svelte";
    import {
        WindowMinimise,
        WindowMaximise,
        WindowUnmaximise,
        WindowIsMaximised,
        Quit,
        WindowToggleMaximise
    } from "$lib/wailsjs/runtime/runtime";

    let {
        title = "EMLy",
        titleContent,
    }: {
        /** Plain title text, used when titleContent isn't provided. */
        title?: string;
        /**
         * Replaces the plain title text with custom markup (e.g. the main
         * window's bold "EMLy" + version badge/tooltip). Defined - and thus
         * styled - by the consuming page, not by this component.
         */
        titleContent?: Snippet;
    } = $props();

    let isMaximized = $state(false);
    let windowFocused = $state(true);

    let interval: NodeJS.Timeout

    async function syncMaxState() {
        isMaximized = await WindowIsMaximised();
    }

    async function toggleMaximize() {
        WindowToggleMaximise();
        isMaximized = !isMaximized;
    }

    function minimize() {
        WindowMinimise();
    }

    function closeWindow() {
        Quit();
    }

    function onTitlebarDblClick() {
        toggleMaximize();
    }

    onMount(async () => {
        window.addEventListener("focus", () => (windowFocused = true));
        window.addEventListener("blur", () => (windowFocused = false));

        interval = setInterval(syncMaxState, 300);
    });

    onDestroy(() => {
        clearInterval(interval);
    });

    syncMaxState();
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
    class="titlebar"
    ondblclick={onTitlebarDblClick}
    style="--wails-draggable:drag"
>
    <div class="title">
        {#if titleContent}
            {@render titleContent()}
        {:else}
            {title}
        {/if}
    </div>

    <div class="controls" style:opacity={windowFocused ? 1 : 0.4}>
        <button class="btn" onmousedown={minimize}>─</button>
        <button class="btn" onmousedown={toggleMaximize} onclick={toggleMaximize}>
            {#if isMaximized}
                ❐
            {:else}
                ☐
            {/if}
        </button>
        <button class="btn close" onmousedown={closeWindow}>✕</button>
    </div>
</div>

<style>
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
        position: relative;
        border-bottom: 1px solid var(--border);
    }

    .title {
        font-size: 13px;
        font-weight: 500;
        opacity: 0.9;
        color: var(--muted-foreground);
    }

    .controls {
        display: flex;
        height: 100%;
        opacity: 0.5;
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

    .btn:disabled {
        opacity: 0.5;
        cursor: not-allowed;
        background: var(--muted);
    }

    .close:hover {
        background: #e81123;
        color: white;
    }
</style>
