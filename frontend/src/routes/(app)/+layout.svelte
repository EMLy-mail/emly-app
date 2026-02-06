<script lang="ts">
  import { browser } from "$app/environment";
  import { page, navigating } from "$app/state";
  import { beforeNavigate, goto } from "$app/navigation";
  import { locales, localizeHref } from "$lib/paraglide/runtime";
  import { unsavedChanges, sidebarOpen, bugReportDialogOpen } from "$lib/stores/app";
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
    Bug,
    Loader2,
    Copy,
    FolderOpen,
    CheckCircle,
    Camera,
    Heart,
  } from "@lucide/svelte";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import { toast } from "svelte-sonner";
  import { Button, buttonVariants } from "$lib/components/ui/button/index.js";
  import * as Dialog from "$lib/components/ui/dialog/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import { Textarea } from "$lib/components/ui/textarea/index.js";

  import {
    WindowMinimise,
    WindowMaximise,
    WindowUnmaximise,
    WindowIsMaximised,
    Quit,
    EventsOn,
    EventsOff,
  } from "$lib/wailsjs/runtime/runtime";
  import { RefreshCcwDot } from "@lucide/svelte";
  import { IsDebuggerRunning, QuitApp, TakeScreenshot, SubmitBugReport, OpenFolderInExplorer } from "$lib/wailsjs/go/main/App";
  import { settingsStore } from "$lib/stores/settings.svelte.js";

  let versionInfo: utils.Config | null = $state(null);
  let isMaximized = $state(false);
  let isDebugerOn: boolean = $state(false);
  let isDebbugerProtectionOn: boolean = $state(true);

  // Bug report form state
  let userName = $state("");
  let userEmail = $state("");
  let bugDescription = $state("");

  // Bug report screenshot state
  let screenshotData = $state("");
  let isCapturing = $state(false);

  // Bug report UI state
  let isSubmitting = $state(false);
  let isSuccess = $state(false);
  let resultZipPath = $state("");

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
    if (browser && isDebbugerProtectionOn) {
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
    isDebbugerProtectionOn = settingsStore.settings.enableAttachedDebuggerProtection ? true : false;
    $inspect(isDebbugerProtectionOn, "isDebbugerProtectionOn");

    applyTheme(stored === "light" ? "light" : "dark");
  });

  // Bug report dialog effects
  $effect(() => {
    if ($bugReportDialogOpen) {
      // Capture screenshot immediately when dialog opens
      captureScreenshot();
    } else {
      // Reset form when dialog closes
      resetBugReportForm();
    }
  });

  // Listen for automatic update notifications
  $effect(() => {
    if (!browser) return;

    EventsOn("update:available", (status: any) => {
      toast.info(`Update ${status.availableVersion} is available!`, {
        description: "Go to Settings to download and install",
        duration: 10000,
        action: {
          label: "Open Settings",
          onClick: () => goto("/settings"),
        },
      });
    });

    return () => {
      EventsOff("update:available");
    };
  });

  async function captureScreenshot() {
    isCapturing = true;
    try {
      const result = await TakeScreenshot();
      screenshotData = result.data;
      console.log("Screenshot captured:", result.width, "x", result.height);
    } catch (err) {
      console.error("Failed to capture screenshot:", err);
    } finally {
      isCapturing = false;
    }
  }

  function resetBugReportForm() {
    userName = "";
    userEmail = "";
    bugDescription = "";
    screenshotData = "";
    isCapturing = false;
    isSubmitting = false;
    isSuccess = false;
    resultZipPath = "";
  }

  async function handleBugReportSubmit(event: Event) {
    event.preventDefault();

    if (!bugDescription.trim()) {
      toast.error("Please provide a bug description.");
      return;
    }

    isSubmitting = true;

    try {
      const result = await SubmitBugReport({
        name: userName,
        email: userEmail,
        description: bugDescription,
        screenshotData: screenshotData
      });

      resultZipPath = result.zipPath;
      isSuccess = true;
      console.log("Bug report created:", result.zipPath);
    } catch (err) {
      console.error("Failed to create bug report:", err);
      toast.error(m.bugreport_error());
    } finally {
      isSubmitting = false;
    }
  }

  async function copyBugReportPath() {
    try {
      await navigator.clipboard.writeText(resultZipPath);
      toast.success(m.bugreport_copied());
    } catch (err) {
      console.error("Failed to copy path:", err);
    }
  }

  async function openBugReportFolder() {
    try {
      const folderPath = resultZipPath.replace(/\.zip$/, "");
      await OpenFolderInExplorer(folderPath);
    } catch (err) {
      console.error("Failed to open folder:", err);
    }
  }

  function closeBugReportDialog() {
    $bugReportDialogOpen = false;
  }

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
          {:else if versionInfo?.EMLy.GUIReleaseChannel !== "stable"}
            v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy.GUIReleaseChannel}
          {:else}
            v{versionInfo?.EMLy.GUISemver}
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
    <Heart
      size="16"
      onclick={() => {
        if (page.url.pathname !== "/credits" && page.url.pathname !== "/credits/")
          goto("/credits");
      }}
      style="cursor: pointer; opacity: 0.7;"
      class="hover:opacity-100 transition-opacity"
    />

    <Separator orientation="vertical" />
    <Bug 
      size="16"
      onclick={() => {
        $bugReportDialogOpen = !$bugReportDialogOpen;
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

  <!-- Bug Report Dialog -->
  <Dialog.Root bind:open={$bugReportDialogOpen}>
    <Dialog.Content class="sm:max-w-[500px] w-full max-h-[80vh] overflow-y-auto custom-scrollbar">
      {#if isSuccess}
        <!-- Success State -->
        <Dialog.Header>
          <Dialog.Title class="flex items-center gap-2">
            <CheckCircle class="h-5 w-5 text-green-500" />
            {m.bugreport_success_title()}
          </Dialog.Title>
          <Dialog.Description>
            {m.bugreport_success_message()}
          </Dialog.Description>
        </Dialog.Header>

        <div class="grid gap-4 py-4">
          <div class="bg-muted rounded-md p-3">
            <code class="text-xs break-all select-all">{resultZipPath}</code>
          </div>

          <div class="flex gap-2">
            <Button variant="outline" class="flex-1" onclick={copyBugReportPath}>
              <Copy class="h-4 w-4 mr-2" />
              {m.bugreport_copy_path()}
            </Button>
            <Button variant="outline" class="flex-1" onclick={openBugReportFolder}>
              <FolderOpen class="h-4 w-4 mr-2" />
              {m.bugreport_open_folder()}
            </Button>
          </div>
        </div>

        <Dialog.Footer>
          <Button onclick={closeBugReportDialog}>
            {m.bugreport_close()}
          </Button>
        </Dialog.Footer>
      {:else}
        <!-- Form State -->
        <form onsubmit={handleBugReportSubmit}>
          <Dialog.Header>
            <Dialog.Title>{m.bugreport_title()}</Dialog.Title>
            <Dialog.Description>
              {m.bugreport_description()}
            </Dialog.Description>
          </Dialog.Header>

          <div class="grid gap-4 py-4">
            <div class="grid gap-2">
              <Label for="bug-name">{m.bugreport_name_label()}</Label>
              <Input
                id="bug-name"
                placeholder={m.bugreport_name_placeholder()}
                bind:value={userName}
                disabled={isSubmitting}
              />
            </div>

            <div class="grid gap-2">
              <Label for="bug-email">{m.bugreport_email_label()}</Label>
              <Input
                id="bug-email"
                type="email"
                placeholder={m.bugreport_email_placeholder()}
                bind:value={userEmail}
                disabled={isSubmitting}
              />
            </div>

            <div class="grid gap-2">
              <Label for="bug-description">{m.bugreport_text_label()}</Label>
              <Textarea
                id="bug-description"
                placeholder={m.bugreport_text_placeholder()}
                bind:value={bugDescription}
                disabled={isSubmitting}
                class="min-h-[120px]"
              />
            </div>

            <!-- Screenshot Preview -->
            <div class="grid gap-2">
              <Label class="flex items-center gap-2">
                <Camera class="h-4 w-4" />
                {m.bugreport_screenshot_label()}
              </Label>
              {#if isCapturing}
                <div class="flex items-center gap-2 text-muted-foreground text-sm">
                  <Loader2 class="h-4 w-4 animate-spin" />
                  Capturing...
                </div>
              {:else if screenshotData}
                <div class="border rounded-md overflow-hidden">
                  <img
                    src="data:image/png;base64,{screenshotData}"
                    alt="Screenshot preview"
                    class="w-full h-32 object-cover object-top opacity-80 hover:opacity-100 transition-opacity cursor-pointer"
                  />
                </div>
              {:else}
                <div class="text-muted-foreground text-sm">
                  No screenshot available
                </div>
              {/if}
            </div>

            <p class="text-muted-foreground text-sm">
              {m.bugreport_info()}
            </p>
          </div>

          <Dialog.Footer>
            <button type="button" class={buttonVariants({ variant: "outline" })} disabled={isSubmitting} onclick={closeBugReportDialog}>
              {m.bugreport_cancel()}
            </button>
            <Button type="submit" disabled={isSubmitting || isCapturing}>
              {#if isSubmitting}
                <Loader2 class="h-4 w-4 mr-2 animate-spin" />
                {m.bugreport_submitting()}
              {:else}
                {m.bugreport_submit()}
              {/if}
            </Button>
          </Dialog.Footer>
        </form>
      {/if}
    </Dialog.Content>
  </Dialog.Root>
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

  :global(.custom-scrollbar::-webkit-scrollbar) {
    width: 6px;
    height: 6px;
  }

  :global(.custom-scrollbar::-webkit-scrollbar-track) {
    background: transparent;
  }

  :global(.custom-scrollbar::-webkit-scrollbar-thumb) {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 6px;
  }

  :global(.custom-scrollbar::-webkit-scrollbar-thumb:hover) {
    background: rgba(255, 255, 255, 0.2);
  }

  :global(.custom-scrollbar::-webkit-scrollbar-corner) {
    background: transparent;
  }
</style>
