<script lang="ts">
  import {
    WindowMinimise,
    WindowMaximise,
    WindowUnmaximise,
    WindowIsMaximised,
    Quit,
  } from "$lib/wailsjs/runtime/runtime";
  import type { LayoutProps } from "./$types";

  let { data, children }: LayoutProps = $props();

  let isMaximized = $state(false);

  async function syncMaxState() {
    isMaximized = await WindowIsMaximised();
  }

  async function toggleMaximize() {
    if (isMaximized) {
      WindowUnmaximise();
    } else {
      WindowMaximise();
    }
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

  function handleWheel(event: WheelEvent) {
    if (event.ctrlKey) {
      event.preventDefault();
    }
  }

  syncMaxState();
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="app-layout" onwheel={handleWheel}>
  <!-- Titlebar -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="titlebar"
    ondblclick={onTitlebarDblClick}
    style="--wails-draggable:drag"
  >
    <div class="title">EMLy PDF Viewer</div>

    <div class="controls">
      <button class="btn" onclick={minimize}>─</button>
      <button class="btn" onclick={toggleMaximize}>
        {#if isMaximized}
          ❐
        {:else}
          ☐
        {/if}
      </button>
      <button class="btn close" onclick={closeWindow}>✕</button>
    </div>
  </div>

  <!-- Content -->
  <main class="content">
    {@render children()}
  </main>
</div>

<style>
  :global(body) {
    margin: 0;
    background: var(--background);
    color: var(--foreground);
    overflow: hidden;
  }

  .app-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
    background: var(--background);
    color: var(--foreground);
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
    opacity: 0.9;
    color: var(--muted-foreground);
    font-weight: 500;
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

  .content {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--background);
  }
</style>