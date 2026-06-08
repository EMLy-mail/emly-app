<script lang="ts">
    import { page } from "$app/state";
    import * as m from "$lib/paraglide/messages.js";
    import { onMount } from "svelte";
    import {
        WindowMinimise,
        WindowMaximise,
        WindowUnmaximise,
        WindowIsMaximised,
        Quit,
    } from "$lib/wailsjs/runtime/runtime";

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
            <h1 class="error-code">:/</h1>
            {#if page.status === 500}
                <p class="error-title">{m.error_500_title()} :(</p>
                <p class="error-message">{m.error_500_message()}</p>
            {:else}
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
        max-width: 400px;
        width: 100%;
    }

    .error-code {
        font-size: 64px;
        font-weight: 700;
        margin: 0 0 8px 0;
        opacity: 0.9;
        line-height: 1;
    }

    .error-title {
        font-size: 18px;
        font-weight: 600;
        opacity: 0.85;
        margin: 0 0 8px 0;
    }

    .error-message {
        font-size: 16px;
        opacity: 0.6;
        margin: 0 0 32px 0;
        line-height: 1.5;
    }
</style>
