<script lang="ts">
  import { browser } from "$app/environment";
  import { page, navigating } from "$app/state";
  import { beforeNavigate, goto } from "$app/navigation";
  import { locales, localizeHref } from "$lib/paraglide/runtime";
  import { unsavedChanges, sidebarOpen } from "$lib/stores/app";
  import "../layout.css";
  import { onMount } from "svelte";
  import * as m from "$lib/paraglide/messages.js";
  import type { utils } from "$lib/wailsjs/go/models";
  import { Toaster } from "$lib/components/ui/sonner/index.js";
  import AppSidebar from "$lib/components/SidebarApp.svelte";
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import { dev } from "$app/environment";
  import {
    PanelRightClose,
    PanelRightOpen,
    House,
    Settings,
  } from "@lucide/svelte";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import { toast } from "svelte-sonner";
  import { buttonVariants } from "$lib/components/ui/button/index.js";

  import {
    WindowMinimise,
    WindowMaximise,
    WindowUnmaximise,
    WindowIsMaximised,
    Quit,
  } from "$lib/wailsjs/runtime/runtime";
  import { RefreshCcwDot } from "@lucide/svelte";
  import { IsDebuggerRunning, QuitApp } from "$lib/wailsjs/go/main/App";

  let versionInfo: utils.Config | null = $state(null);
  let isMaximized = $state(false);
  let isDebugerOn: boolean = $state(false);

  async function syncMaxState() {
    isMaximized = await WindowIsMaximised();
  }

  beforeNavigate(({ cancel }) => {
    if ($unsavedChanges) {
      toast.warning(m.unsaved_changes_warning());
      cancel();
    }
  });

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

  onMount(async () => {
    if (browser) {
      detectDebugging();
      setInterval(detectDebugging, 1000);
    }

    versionInfo = data.data as utils.Config;
  });

  function handleWheel(event: WheelEvent) {
    if (event.ctrlKey) {
      event.preventDefault();
    }
  }

  async function detectDebugging() {
    if (!browser) return;
    if (isDebugerOn === true) return; // Prevent multiple detections
    isDebugerOn = await IsDebuggerRunning();
    if (isDebugerOn) {
      if(dev) toast.warning("Debugger is attached.");
      await new Promise((resolve) => setTimeout(resolve, 5000));
      await QuitApp();
    }
  }

  let { data, children } = $props();

  const THEME_KEY = "emly_theme";
  let theme = $state<"dark" | "light">("dark");

  function applyTheme(next: "dark" | "light") {
    theme = next;
    if (!browser) return;
    document.documentElement.classList.toggle("dark", next === "dark");
    try {
      localStorage.setItem(THEME_KEY, next);
    } catch {
      // ignore
    }
  }

  $effect(() => {
    if (!browser) return;
    let stored: string | null = null;
    try {
      stored = localStorage.getItem(THEME_KEY);
    } catch {
      stored = null;
    }

    applyTheme(stored === "light" ? "light" : "dark");
  });

  syncMaxState();
</script>

<div class="app" onwheel={handleWheel}>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="titlebar"
    ondblclick={onTitlebarDblClick}
    style="--wails-draggable:drag"
  >
    <div class="title">
      <bold>EMLy</bold>
      <div class="version-wrapper">
        <version>
          {#if dev}
            v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy.GUIReleaseChannel}
            <debug>(DEBUG BUILD)</debug>
          {:else}
            v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy.GUIReleaseChannel}
          {/if}
        </version>
        {#if versionInfo}
          <div class="version-tooltip">
            <div class="tooltip-item">
              <span class="label">GUI:</span>
              <span class="value">v{versionInfo.EMLy.GUISemver}</span>
              <span class="channel">({versionInfo.EMLy.GUIReleaseChannel})</span
              >
            </div>
            <div class="tooltip-item">
              <span class="label">SDK:</span>
              <span class="value">v{versionInfo.EMLy.SDKDecoderSemver}</span>
              <span class="channel"
                >({versionInfo.EMLy.SDKDecoderReleaseChannel})</span
              >
            </div>
          </div>
        {/if}
      </div>
    </div>

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

  <div class="content">
    <Sidebar.Provider>
      {#if $sidebarOpen}
        <AppSidebar />
      {/if}
      <main>
        <!-- <Sidebar.Trigger /> -->
        <Toaster />
        {#await navigating?.complete}
          <div class="loading-overlay">
            <div class="spinner"></div>
            <span style="opacity: 0.5; font-size: 13px"
              >{m.layout_loading_text()}</span
            >
          </div>
        {:then}
          {@render children()}
        {/await}
      </main>
    </Sidebar.Provider>
  </div>

  <div class="footerbar">
    {#if !$sidebarOpen}
      <PanelRightClose
        size="17"
        onclick={() => {
          $sidebarOpen = !$sidebarOpen;
        }}
        style="cursor: pointer;"
      />
    {:else}
      <PanelRightOpen
        size="17"
        onclick={() => {
          $sidebarOpen = !$sidebarOpen;
        }}
        style="cursor: pointer;"
      />
    {/if}

    <Separator orientation="vertical" />

    <House
      size="16"
      onclick={() => {
        if (page.url.pathname !== "/") goto("/");
      }}
      style="cursor: pointer; opacity: 0.7;"
      class="hover:opacity-100 transition-opacity"
    />
    <Settings
      size="16"
      onclick={() => {
        if (
          page.url.pathname !== "/settings" &&
          page.url.pathname !== "/settings/"
        )
          goto("/settings");
      }}
      style="cursor: pointer; opacity: 0.7;"
      class="hover:opacity-100 transition-opacity"
    />

    <a
      data-sveltekit-reload
      href="/"
      class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
      style="text-decoration: none; margin-left: auto; height: 24px; font-size: 12px; padding: 0 8px;"
      aria-label={m.settings_danger_reload_button()}
      title={m.settings_danger_reload_button() + " app"}
    >
      <RefreshCcwDot />
    </a>
  </div>

  <div style="display:none">
    {#each locales as locale}
      <a href={localizeHref(page.url.pathname, { locale })}>
        {locale}
      </a>
    {/each}
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
    background: oklch(0 0 0);
    color: #eaeaea;
    font-family: system-ui, sans-serif;
  }

  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
  }

  .titlebar {
    height: 32px;
    background: oklch(0 0 0);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-left: 12px;
    -webkit-app-region: drag;
    user-select: none;
    flex: 0 0 32px;
    z-index: 50;
    position: relative;
  }

  .footerbar {
    height: 32px;
    background: oklch(0 0 0);
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 12px;
    padding: 0 12px;
    user-select: none;
    flex: 0 0 32px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }

  .title {
    font-size: 13px;
    opacity: 0.9;
    color: gray;
  }

  .title bold {
    font-weight: 600;
    color: white;
    opacity: 0.7;
  }

  .title version {
    color: rgb(228, 221, 221);
    opacity: 0.4;
  }

  .title version debug {
    color: #e11d48;
    opacity: 1;
    font-weight: 600;
  }

  .version-wrapper {
    position: relative;
    display: inline-block;
    cursor: default;
  }

  .version-tooltip {
    visibility: hidden;
    opacity: 0;
    position: absolute;
    top: 100%;
    left: 0;
    background-color: #111;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 6px;
    padding: 8px 12px;
    z-index: 1000;
    margin-top: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    transition: all 0.2s ease-in-out;
    transform: translateY(-5px);
    pointer-events: none;
    -webkit-app-region: no-drag;
  }

  .version-wrapper:hover .version-tooltip {
    visibility: visible;
    opacity: 1;
    transform: translateY(0);
    pointer-events: auto;
  }

  .tooltip-item {
    display: grid;
    grid-template-columns: 40px auto auto;
    gap: 8px;
    font-size: 11px;
    line-height: 1.6;
    white-space: nowrap;
    align-items: center;
  }

  .tooltip-item .label {
    color: #9ca3af;
  }

  .tooltip-item .value {
    color: #f3f4f6;
    font-family: monospace;
  }

  .tooltip-item .channel {
    color: #6b7280;
    font-size: 10px;
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
    color: white;
    font-size: 14px;
    cursor: pointer;
    -webkit-app-region: no-drag;
  }

  .btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: rgba(255, 255, 255, 0.02);
  }

  .close:hover {
    background: #e81123;
  }

  .content {
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    background: oklch(0 0 0);
    overflow: hidden;
    position: relative;
  }

  main {
    flex: 1 1 auto;
    min-width: 0;
    min-height: 0;
    overflow: auto;
    position: relative;
  }

  /* Override Shadcn Sidebar defaults to fit in content area */
  :global(.content .group\/sidebar-wrapper) {
    min-height: 0 !important;
    height: 100% !important;
  }

  /* Target the fixed container of the sidebar */
  :global(.content [data-slot="sidebar-container"]) {
    position: absolute !important;
    height: 100% !important;
    /* Ensure it doesn't take viewport height */
    max-height: 100% !important;
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

  .loading-overlay {
    position: absolute;
    inset: 0;
    z-index: 50;
    display: flex;
    flex-direction: column;
    gap: 10px;
    align-items: center;
    justify-content: center;
    background: oklch(0 0 0);
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 2px solid rgba(255, 255, 255, 0.1);
    border-top-color: rgba(255, 255, 255, 0.8);
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
