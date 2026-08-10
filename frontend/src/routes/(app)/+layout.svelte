<script lang="ts">
    import { browser } from "$app/environment";
    import { page, navigating } from "$app/state";
    import { beforeNavigate, goto } from "$app/navigation";
    import { locales, localizeHref } from "$lib/paraglide/runtime";
    import {
        unsavedChanges,
        sidebarOpen,
        bugReportDialogOpen,
        dangerZoneEnabled,
        runningInDebugMode,
        hostIntegrityFailed,
        hostIntegrityStanding,
    } from "$lib/stores/app";
    import { onMount } from "svelte";
    import * as m from "$lib/paraglide/messages.js";
    import type { utils } from "$lib/wailsjs/go/models";
    import { Toaster } from "$lib/components/ui/sonner/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
    import AppSidebar from "$lib/components/SidebarApp.svelte";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { dev } from "$app/environment";
    import {
        PanelRightClose,
        PanelRightOpen,
        Mail,
        Settings,
        Bug,
        Heart,
        Info,
        Music,
        TriangleAlert,
        ShieldAlert
    } from "@lucide/svelte";
    import { Separator } from "$lib/components/ui/separator/index.js";
    import { toast } from "svelte-sonner";
    import { buttonVariants } from "$lib/components/ui/button/index.js";
    import BugReportDialog from "$lib/components/BugReportDialog.svelte";
    import TitleBar from "$lib/components/TitleBar.svelte";

    import { RefreshCcwDot } from "@lucide/svelte";
    import {
        IsDebuggerRunning,
        QuitApp,
        IsAppInDebugMode,
    } from "$lib/wailsjs/go/main/App";
    import {
        settingsStore,
        parseReleaseChannel,
    } from "$lib/stores/settings.svelte.js";
  import { mailState } from "$lib/stores/mail-state.svelte.js";
    import { ensureHostIntegrityChecked } from "$lib/utils/hostIntegrityCheck";
    import { ensureExportFolderWritable } from "$lib/utils/export-folder-check";
    import { settingsResetStore } from "$lib/stores/settings-reset.svelte";
    import SettingsResetDialog from "$lib/components/SettingsResetDialog.svelte";

    let versionInfo: utils.Config | null = $state(null);
    let isDebugerOn: boolean = $state(false);
    let isDebbugerProtectionOn: boolean = $state(true);
    let configMissingDialogOpen = $state(false);
    let hostIntegrityDialogOpen = $state(false);
    let hostIntegrityChecked = false;
    let settingsResetDialogOpen = $derived(settingsResetStore.entries.length > 0);

    /**
     * Opt-out for the whole sidebar: when off it is never rendered and its
     * status bar toggle is hidden, so the only navigation left is the status
     * bar icons. Compared against `false` rather than coerced, so a settings
     * payload written by an older build (no `showSidebar` key at all) keeps the
     * sidebar instead of losing it.
     */
    let showSidebar = $derived(settingsStore.settings.showSidebar !== false);

    $effect(() => {
        $inspect("hostIntegrityDialogOpen", hostIntegrityDialogOpen);
    })

    // Opens the failed-integrity dialog whenever the gate flips true — not
    // just right after the initial check, since the fuller Updater/IPC
    // cross-check can upgrade the standing to "limited" later in the
    // background (see hostIntegrityCheck.ts).
    $effect(() => {
        if ($hostIntegrityFailed) {
            hostIntegrityDialogOpen = true;
        }
    });

    $effect(() => {
        if ($hostIntegrityStanding === "acceptable") {
            toast.warning(m.host_integrity_partial_toast());
        }
    });

    beforeNavigate(({ cancel, to }) => {
        const normalize = (p: string) => p.replace(/\/+$/, "") || "/";
        if (normalize(to?.url?.pathname ?? "") === normalize(page.url.pathname) && !to?.url.search.includes("?reload=true")) {
            cancel();
            return;
        }
        if ($unsavedChanges) {
            toast.warning(m.unsaved_changes_warning());
            cancel();
        }
    });

    onMount(async () => {
        // Log the entire local storage
        console.log("Local Storage Contents:");
        console.log(localStorage);
        if (dev) dangerZoneEnabled.set(true);
        if (browser && isDebbugerProtectionOn) {
            detectDebugging();
            setInterval(detectDebugging, 1000);
        }

        if (settingsStore.wasReset) {
            toast.warning(m.settings_toast_load_error());
        }

        versionInfo = data.data as utils.Config;
        if (!versionInfo) {
            configMissingDialogOpen = true;
        } else {
            // config.ini wins over localStorage for the release channel: it is
            // what the build was shipped with and what the updater reads, so a
            // stale value in the store (or an ini edited by hand between runs)
            // is corrected here, once per start.
            const channel = parseReleaseChannel(
                versionInfo.EMLy?.GUIReleaseChannel,
            );
            if (channel !== settingsStore.settings.releaseChannel) {
                settingsStore.update({ releaseChannel: channel });
            }
        }
        runningInDebugMode.set(await IsAppInDebugMode());

        // Independent of the integrity check below, and deliberately not
        // awaited: a folder on an unreachable network share can take seconds
        // to fail, and must not delay anything else. Started here, before
        // the integrity check's IPC round trips, so it genuinely runs
        // concurrently rather than only after integrity resolves.
        void ensureExportFolderWritable();

        if (!hostIntegrityChecked) {
            hostIntegrityChecked = true;
            // The failed/dialog-open reaction happens in the $effect above,
            // which also covers the background Updater/IPC cross-check
            // upgrading the standing to "limited" after this resolves.
            await ensureHostIntegrityChecked();
        }
    });

    async function detectDebugging() {
        if (!browser) return;
        if (isDebugerOn === true) return; // Prevent multiple detections
        isDebugerOn = await IsDebuggerRunning();
        if (isDebugerOn) {
            if (dev) toast.warning("Debugger is attached.");
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
        isDebbugerProtectionOn = settingsStore.settings
            .enableAttachedDebuggerProtection
            ? true
            : false;
        $inspect(isDebbugerProtectionOn, "isDebbugerProtectionOn");
        $inspect($runningInDebugMode, "isAppInDebugMode");

        applyTheme(stored === "light" ? "light" : "dark");

        if(page.url.pathname !== "/") {
            sidebarOpen.set(true);
        }
    });
</script>

<div class="app">
    <TitleBar>
        {#snippet titleContent()}
            <bold>EMLy</bold>
            <div class="version-wrapper">
                <version class="inline">
                    {#if dev && $runningInDebugMode}
                        v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy
                            .GUIReleaseChannel}
                        <debug><TriangleAlert size="16" /> DEBUG BUILD</debug>
                    {:else if dev}
                        v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy
                            .GUIReleaseChannel}
                        <dev><TriangleAlert size="16" /> DEV BUILD</dev>
                    {:else if $hostIntegrityFailed}
                        v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy
                                .GUIReleaseChannel}
                        <debug><ShieldAlert size="16" /> {m.version_host_integrity_check_failed_version_info()}</debug>
                    {:else if versionInfo?.EMLy.GUIReleaseChannel !== "stable"}
                        v{versionInfo?.EMLy.GUISemver}_{versionInfo?.EMLy
                            .GUIReleaseChannel}
                    {:else}
                        v{versionInfo?.EMLy.GUISemver}
                    {/if}
                </version>
                {#if versionInfo}
                    <div class="version-tooltip">
                        <div class="tooltip-item">
                            <span class="label">GUI:</span>
                            <span class="value"
                                >v{versionInfo.EMLy.GUISemver}</span
                            >
                            <span class="channel"
                                >({versionInfo.EMLy.GUIReleaseChannel})</span
                            >
                        </div>
                        <div class="tooltip-item">
                            <span class="label">SDK:</span>
                            <span class="value"
                                >v{versionInfo.EMLy.SDKDecoderSemver}</span
                            >
                            <span class="channel"
                                >({versionInfo.EMLy
                                    .SDKDecoderReleaseChannel})</span
                            >
                        </div>
                    </div>
                {/if}
            </div>
        {/snippet}
    </TitleBar>

    <div
        class="content"
        class:reduce-motion={settingsStore.settings.reduceMotion}
    >
        <Sidebar.Provider
            open={showSidebar && $sidebarOpen}
            onOpenChange={(v) => sidebarOpen.set(v)}
        >
            {#if showSidebar}
                <AppSidebar />
            {/if}
            <main>
                <Toaster />
                {#await navigating?.complete}
                    <div class="loading-overlay">
                        <div class="spinner"></div>
                        <span style="opacity: 0.5; font-size: 13px"
                            >{m.layout_loading_text()}</span
                        >
                    </div>
                {:then}
                    {#key page.url.pathname}
                        <div
                            class="page-enter"
                            class:no-motion={settingsStore.settings.reduceMotion}
                            style="height: 100%;"
                        >
                            {@render children()}
                        </div>
                    {/key}
                {/await}
            </main>
        </Sidebar.Provider>
    </div>

    <div class="footerbar">
        {#if showSidebar}
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
        {/if}

        <Mail
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
                const p = page.url.pathname as string;
                if (p !== "/settings" && p !== "/settings/") goto("/settings");
            }}
            style="cursor: pointer; opacity: 0.7;"
            class="hover:opacity-100 transition-opacity"
        />
        <Info
            size="16"
            onclick={() => {
                const p = page.url.pathname as string;
                if (p !== "/credits" && p !== "/credits/") goto("/credits");
            }}
            style="cursor: pointer; opacity: 0.7;"
            class="hover:opacity-100 transition-opacity"
        />

        <a
            data-sveltekit-reload
            href="/?reload=true"
            class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
            style="text-decoration: none; margin-left: auto; height: 24px; font-size: 12px; padding: 0 8px;"
            aria-label={m.settings_danger_reload_button_ui()}
            title={m.settings_danger_reload_button_ui()}
        >
            <RefreshCcwDot />
        </a>
        <!-- svelte-ignore a11y_invalid_attribute -->
        <a
            href="#"
            class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
            style="text-decoration: none; height: 24px; font-size: 12px; padding: 0 8px;"
            aria-label={m.layout_report_issue_label()}
            title={m.layout_report_issue_label()}
            onclick={() => {
                $bugReportDialogOpen = !$bugReportDialogOpen;
            }}
        >
            <Bug />
        </a>
    </div>

    <div style="display:none">
        {#each locales as locale}
            <a href={localizeHref(page.url.pathname, { locale })}>
                {locale}
            </a>
        {/each}
    </div>

    <BugReportDialog />

    <SettingsResetDialog
        open={settingsResetDialogOpen}
        entries={settingsResetStore.entries}
        onAcknowledge={() => settingsResetStore.clear()}
    />

    <AlertDialog.Root bind:open={configMissingDialogOpen}>
        <AlertDialog.Content>
            <AlertDialog.Header>
                <AlertDialog.Title
                    >{m.layout_config_missing_title()}</AlertDialog.Title
                >
                <AlertDialog.Description>
                    {m.layout_config_missing_description()}
                    <ul class="mt-2 list-disc pl-5 space-y-1">
                        <li>{m.layout_config_missing_item_settings()}</li>
                        <li>{m.layout_config_missing_item_bugreport()}</li>
                        <li>{m.layout_config_missing_item_features()}</li>
                    </ul>
                </AlertDialog.Description>
            </AlertDialog.Header>
            <AlertDialog.Footer>
                <AlertDialog.Action
                    onclick={() => {
                        configMissingDialogOpen = false;
                    }}
                    >{m.layout_config_missing_understood()}</AlertDialog.Action
                >
            </AlertDialog.Footer>
        </AlertDialog.Content>
    </AlertDialog.Root>

    <AlertDialog.Root bind:open={hostIntegrityDialogOpen}>
        <AlertDialog.Content>
            <AlertDialog.Header>
                <AlertDialog.Title
                    style="color: var(--destructive); opacity: 0.7;"
                    >
                    <ShieldAlert class="inline mr-1 -translate-y-0.5" />
                    {m.layout_host_integrity_title()}
                    </AlertDialog.Title
                >
                <AlertDialog.Description>
                    {m.layout_host_integrity_description()}
                </AlertDialog.Description>
            </AlertDialog.Header>
            <AlertDialog.Footer>
                <AlertDialog.Action
                    onclick={() => {
                        hostIntegrityDialogOpen = false;
                    }}
                    >{m.layout_host_integrity_understood()}</AlertDialog.Action
                >
            </AlertDialog.Footer>
        </AlertDialog.Content>
    </AlertDialog.Root>
</div>

<style>
    :global(body) {
        margin: 0;
        font-family: system-ui, sans-serif;
    }

    .app {
        display: flex;
        flex-direction: column;
        height: 100vh;
        overflow: hidden;
        background: var(--background);
        color: var(--foreground);
    }

    .footerbar {
        height: 32px;
        background: var(--background);
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 12px;
        padding: 0 12px;
        user-select: none;
        flex: 0 0 32px;
        border-top: 1px solid var(--border);
    }

    /* Rendered via TitleBar's titleContent snippet, defined (and thus
       scoped) here rather than in TitleBar.svelte itself. */
    bold {
        font-weight: 600;
        color: var(--foreground);
        opacity: 0.7;
    }

    version {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        color: var(--muted-foreground);
        opacity: 0.6;
    }

    version debug {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        color: var(--destructive);
        opacity: 1;
        font-weight: 600;
    }

    version dev {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        color: var(--color-yellow-400);
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
        background-color: var(--popover);
        color: var(--popover-foreground);
        border: 1px solid var(--border);
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
        color: var(--muted-foreground);
    }

    .tooltip-item .value {
        color: var(--foreground);
        font-family: monospace;
    }

    .tooltip-item .channel {
        color: var(--muted-foreground);
        font-size: 10px;
    }

    .content {
        flex: 1 1 auto;
        min-height: 0;
        display: flex;
        background: var(--background);
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

    /* Disable sidebar transitions when reduce-motion is active */
    :global(.content.reduce-motion [data-slot="sidebar-gap"]),
    :global(.content.reduce-motion [data-slot="sidebar-container"]) {
        transition-duration: 0s !important;
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

    .page-enter {
        animation: page-fade-in 150ms ease-out;
    }

    .page-enter.no-motion {
        animation: none;
    }

    @keyframes page-fade-in {
        from { opacity: 0; }
        to { opacity: 1; }
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
        background: var(--background);
    }

    .spinner {
        width: 32px;
        height: 32px;
        border: 2px solid var(--border);
        border-top-color: var(--primary);
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }
</style>
