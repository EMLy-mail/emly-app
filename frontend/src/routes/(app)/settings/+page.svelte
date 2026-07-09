<script lang="ts">
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { Label } from "$lib/components/ui/label";
    import { Separator } from "$lib/components/ui/separator";
    import { Switch } from "$lib/components/ui/switch";
    import {
        ChevronLeft,
        Flame,
        Sun,
        Moon,
        CircleCheck,
        CircleX,
        Loader2,
        OctagonX,
        TriangleAlert,
    } from "@lucide/svelte";
    import type { EMLy_GUI_Settings } from "$lib/types";
    import { toast } from "svelte-sonner";
    import { It, Us } from "svelte-flags";
    import * as RadioGroup from "$lib/components/ui/radio-group/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import { Badge } from "$lib/components/ui/badge/index.js";
    import { buttonVariants } from "$lib/components/ui/button/index.js";
    import { Checkbox } from "$lib/components/ui/checkbox/index.js";
    import {
        dismissUnsavedChangesToast,
        showUnsavedChangesToast,
    } from "$lib/utils/unsaved-changes-toast";
    import {
        dangerZoneEnabled,
        unsavedChanges,
        hostIntegrityFailed,
        runningInDebugMode,
    } from "$lib/stores/app";
    import { LogDebug, LogInfo } from "$lib/wailsjs/runtime/runtime";
    import { settingsStore, defaultSettings } from "$lib/stores/settings.svelte";
    import * as m from "$lib/paraglide/messages";
    import { setLocale } from "$lib/paraglide/runtime";
    import { mailState } from "$lib/stores/mail-state.svelte.js";
    import { dev } from "$app/environment";
    import {
        ReloadConfig,
        ShowOpenFolderDialog,
        SetExportAttachmentFolder,
        SetTrayIconEnabled,
        OpenDevTools,
        RestartApp,
        GetUpdaterADStatus,
        GetUpdaterSystemInfo,
    } from "$lib/wailsjs/go/main/App";
    import SettingsSwitchLabel from "$lib/components/settings/SettingsSwitchLabel.svelte";
    import { systemInfoStore } from '$lib/stores/system-info.svelte.js';
    import { updaterStatusStore } from '$lib/stores/updater-status.svelte.js';
    import { isInsideADDomain, isInsideTREGCCADDomain, evaluateHostname, isDevMachine } from '$lib/utils/hostIntegrity';
    import { logIPCRequest, logIPCResponse, logIPCError } from '$lib/utils/ipcLog';

    let { data } = $props();
    let config = $derived(data.config);
    let previousTheme = $state<string | undefined>(undefined);
    let runningInDevMode: boolean = dev;
    let machineData = $derived(systemInfoStore.data);
    
    // Clone store state for form editing
    // Use normalizeSettings to ensure new fields (like useBuiltinPDFViewer) are populated with defaults
    let form = $state<EMLy_GUI_Settings>(
        normalizeSettings(settingsStore.settings),
    );
    let lastSaved = $state<EMLy_GUI_Settings>(
        normalizeSettings(settingsStore.settings),
    );
    let dangerWarningOpen = $state(false);
    let reloadingConfig = $state(false);
    let exportFolder = $state("");
    let savingExportFolder = $state(false);
    let trayIconEnabled = $state(true);
    let savingTrayIcon = $state(false);

    $effect(() => {
        exportFolder = config?.ExportAttachmentFolder ?? "";
    });

    $effect(() => {
        trayIconEnabled = !(config?.DisableTrayIcon ?? false);
    });

    $effect(() => {
        $inspect("data", data)
    })

    async function toggleTrayIcon(checked: boolean) {
        savingTrayIcon = true;
        try {
            await SetTrayIconEnabled(checked);
            trayIconEnabled = checked;
            toast.success(m.settings_danger_tray_icon_saved());
        } catch (e) {
            console.error("Failed to set tray icon enabled", e);
            toast.error(m.settings_danger_tray_icon_save_error());
        } finally {
            savingTrayIcon = false;
        }
    }

    async function selectExportFolder() {
        savingExportFolder = true;
        try {
            const folder = await ShowOpenFolderDialog();
            if (!folder) return;
            await SetExportAttachmentFolder(folder);
            exportFolder = folder;
            toast.success(m.settings_export_folder_saved());
        } catch (e) {
            console.error("Failed to set export folder", e);
            toast.error(m.settings_export_folder_save_error());
        } finally {
            savingExportFolder = false;
        }
    }

    async function resetExportFolder() {
        savingExportFolder = true;
        try {
            await SetExportAttachmentFolder("");
            exportFolder = "";
            toast.success(m.settings_export_folder_saved());
        } catch (e) {
            console.error("Failed to reset export folder", e);
            toast.error(m.settings_export_folder_save_error());
        } finally {
            savingExportFolder = false;
        }
    }

    $effect(() => {
        if (!config) {
            toast.error(m.settings_config_load_error());
        }
    });

    const availablePreviewFileSupportedTypes: EMLy_GUI_Settings["previewFileSupportedTypes"] = [
        "jpg",
        "jpeg",
        "png"
    ];

    async function setLanguage(
        lang: EMLy_GUI_Settings["selectedLanguage"] | null,
    ) {
        if (!browser) return;
        try {
            await setLocale(lang || "en", { reload: false });
            toast.success(m.settings_toast_language_changed());
        } catch {
            toast.error(m.settings_toast_language_change_failed());
        }
    }

    function normalizeSettings(s: EMLy_GUI_Settings): EMLy_GUI_Settings {
        return {
            selectedLanguage:
                s.selectedLanguage || defaultSettings.selectedLanguage || "en",
            useBuiltinPreview: !!s.useBuiltinPreview,
            useBuiltinPDFViewer:
                s.useBuiltinPDFViewer ?? defaultSettings.useBuiltinPDFViewer ?? true,
            previewFileSupportedTypes:
                s.previewFileSupportedTypes ||
                defaultSettings.previewFileSupportedTypes ||
                [],
            enableAttachedDebuggerProtection:
                s.enableAttachedDebuggerProtection ??
                defaultSettings.enableAttachedDebuggerProtection ??
                true,
            enableHostIntegrityCheck:
                s.enableHostIntegrityCheck ??
                defaultSettings.enableHostIntegrityCheck ??
                true,
            useDarkEmailViewer:
                s.useDarkEmailViewer ?? defaultSettings.useDarkEmailViewer ?? true,
            reduceMotion: s.reduceMotion ?? defaultSettings.reduceMotion ?? false,
            theme: s.theme || defaultSettings.theme || "light",
            enableLinkClickConfirmation:
                s.enableLinkClickConfirmation ??
                defaultSettings.enableLinkClickConfirmation ??
                false,
            enableTabMode:
                s.enableTabMode ??
                defaultSettings.enableTabMode ??
                false,
            openAttachmentsAsTab:
                s.openAttachmentsAsTab ??
                defaultSettings.openAttachmentsAsTab ??
                false,
            fixEmailTextContrast:
                s.fixEmailTextContrast ??
                defaultSettings.fixEmailTextContrast ??
                false,
        };
    }

    function isSameSettings(a: EMLy_GUI_Settings, b: EMLy_GUI_Settings) {
        return JSON.stringify(normalizeSettings(a)) === JSON.stringify(normalizeSettings(b))
    }

    function resetToLastSaved() {
        form = { ...lastSaved };
        toast.info(m.settings_toast_reverted());
    }

    async function saveToStorage() {
        if (!browser) return;
        const settings = normalizeSettings(form);
        const languageChanged =
            settings.selectedLanguage !== lastSaved.selectedLanguage;
        const hostIntegrityCheckChanged =
            settings.enableHostIntegrityCheck !==
            lastSaved.enableHostIntegrityCheck;
        try {
            settingsStore.update(settings);
        } catch {
            toast.error(m.settings_toast_save_failed());
            return;
        }

        lastSaved = settings;
        form = settings;

        if (languageChanged) {
            await setLanguage(settings.selectedLanguage);
            location.reload();
        } else if (hostIntegrityCheckChanged) {
            await setLanguage(settings.selectedLanguage);
            location.reload();
        } else {
            toast.success(m.settings_toast_saved());
        }
    }

    async function resetToDefaults() {
        form = normalizeSettings(defaultSettings);
        lastSaved = normalizeSettings(defaultSettings);

        // Save to storage
        if (browser) {
            try {
                settingsStore.reset();
                settingsStore.update(form); // Ensure local form state is persisted
                sessionStorage.removeItem("debugWindowInSettings");
                if(!runningInDevMode) {
                    dangerZoneEnabled.set(false);
                    LogDebug("Reset danger zone setting to false.");
                }
                LogInfo("Settings reset to defaultSettings.");
            } catch {
                toast.error(m.settings_toast_reset_failed());
                return;
            }
        }
        try {
            await SetExportAttachmentFolder("");
            exportFolder = "";
        } catch (e) {
            console.error("Failed to reset export folder", e);
        }
        await setLanguage(form.selectedLanguage);
        mailState.clear();
        toast.info(m.settings_toast_reset_success());
        location.reload();
    }

    $effect(() => {
        if (!browser) return;

        const unsavedSettingsBool = !isSameSettings(normalizeSettings(form), lastSaved);
        unsavedChanges.set(unsavedSettingsBool);
        if (unsavedSettingsBool) {
            showUnsavedChangesToast({
                onSave: saveToStorage,
                onReset: resetToLastSaved,
            });
        } else {
            dismissUnsavedChangesToast();
        }
    });

    $effect(() => {
        // Sync initial state from store when hydrated
        // Ensure we don't update if the values are already practically identical to avoid loops
        if (
            settingsStore.hasHydrated &&
            isSameSettings(lastSaved, defaultSettings) &&
            !isSameSettings(lastSaved, settingsStore.settings)
        ) {
            // Ensure we normalize when syncing from store too
            form = normalizeSettings(settingsStore.settings);
            lastSaved = normalizeSettings(settingsStore.settings);
        }
    });

    let previousDangerZoneEnabled = $dangerZoneEnabled;

    $effect(() => {
        (async () => {
            if ($dangerZoneEnabled && !previousDangerZoneEnabled) {
                dangerWarningOpen = true;
                toast.info(m.settings_danger_here_be_dragons(), {
                    icon: Flame,
                });
            }
            previousDangerZoneEnabled = $dangerZoneEnabled;
        })();
    });

    // EMLy Updater's local (service) status and IPC pipe status are fetched
    // once, app-wide, from the root (app)/+layout.svelte onMount — read the
    // shared updaterStatusStore here rather than re-fetching.

    // "Run safety check" — a one-shot diagnostic combining host identity
    // (hostname/AD domain) with the EMLy Updater's local + IPC status, plus
    // a cross-check of AD domain/hostname as seen by the IPC-connected
    // SYSTEM-privileged service vs. this (unprivileged) process, as a
    // double failsafe against a spoofed/compromised local view.
    type SafetyCheckStanding = "perfect" | "usable" | "limited";

    type SafetyCheckResult = {
        hostnameOk: boolean;
        adDomainOk: boolean;
        updaterInstalled: boolean;
        updaterRunning: boolean;
        ipcInstalled: boolean;
        ipcResponding: boolean;
        crossCheckOk: boolean;
        hostIntegrityToggleOk: boolean;
        standing: SafetyCheckStanding;
    };

    let safetyCheckDialogOpen = $state(false);
    let runningSafetyCheck = $state(false);
    let safetyCheckResult = $state<SafetyCheckResult | null>(null);

    async function runSafetyCheck() {
        runningSafetyCheck = true;
        try {
            await updaterStatusStore.refreshUpdaterStatus();
            await updaterStatusStore.checkUpdaterIPCStatus();

            const hostname = machineData?.Hostname ?? "";
            const adDomain = machineData?.ADDomain ?? "";

            const hostnameOk = evaluateHostname(hostname);
            const adDomainOk = isInsideTREGCCADDomain(adDomain);

            const updaterInstalled = !!updaterStatusStore.installed;
            const updaterRunning = !!updaterStatusStore.running;
            const ipcInstalled = !!updaterStatusStore.ipcActive;
            const ipcResponding = !!updaterStatusStore.ipcValid;

            let crossCheckOk = false;
            if (ipcInstalled && ipcResponding) {
                try {
                    logIPCRequest("GetUpdaterADStatus");
                    logIPCRequest("GetUpdaterSystemInfo");
                    const [adStatus, sysInfo] = await Promise.all([
                        GetUpdaterADStatus(),
                        GetUpdaterSystemInfo(),
                    ]);
                    logIPCResponse("GetUpdaterADStatus", adStatus);
                    logIPCResponse("GetUpdaterSystemInfo", sysInfo);
                    crossCheckOk =
                        adStatus.ADDomain === adDomain &&
                        sysInfo.Hostname === hostname;
                } catch (err) {
                    logIPCError("GetUpdaterADStatus/GetUpdaterSystemInfo", err);
                    LogDebug(`Safety check IPC cross-check failed: ${err}`);
                    crossCheckOk = false;
                }
            }

            // The Host Integrity toggle must be on — the only exemption is
            // an actual DEBUG MODE build ($runningInDebugMode, the Go-side
            // build flag), never merely running under `wails dev`/vite
            // (`dev`) and never a recognized dev machine (which only
            // bypasses the hostname/AD checks above). Disabling it outside
            // of a debug build forces a Limited standing regardless of
            // every other result.
            const isDebugBuild = $runningInDebugMode;
            const hostIntegrityToggleOk =
                settingsStore.settings.enableHostIntegrityCheck ||
                isDebugBuild;

            let standing: SafetyCheckStanding;
            if (!hostIntegrityToggleOk) {
                standing = "limited";
            } else if (
                hostnameOk &&
                adDomainOk &&
                updaterInstalled &&
                updaterRunning &&
                ipcInstalled &&
                ipcResponding &&
                crossCheckOk
            ) {
                standing = "perfect";
            } else if (
                hostnameOk &&
                adDomainOk &&
                updaterInstalled &&
                !updaterRunning
            ) {
                standing = "usable";
            } else {
                standing = "limited";
            }

            safetyCheckResult = {
                hostnameOk,
                adDomainOk,
                updaterInstalled,
                updaterRunning,
                ipcInstalled,
                ipcResponding,
                crossCheckOk,
                hostIntegrityToggleOk,
                standing,
            };
        } finally {
            runningSafetyCheck = false;
        }
    }

    function openSafetyCheckDialog() {
        safetyCheckDialogOpen = true;
        runSafetyCheck();
    }

    // Sync theme with email viewer dark mode
    $effect(() => {
        if (!browser) return;
        if (previousTheme !== undefined && form.theme !== previousTheme) {
            form.useDarkEmailViewer = form.theme === "dark";
        }
        previousTheme = form.theme;
    });

    // Auto-disable openAttachmentsAsTab when tab mode is turned off
    $effect(() => {
        if (!form.enableTabMode) {
            form.openAttachmentsAsTab = false;
        }
    });

    async function reloadConfig() {
        reloadingConfig = true;
        try {
            const freshConfig = await ReloadConfig();
            config = freshConfig.EMLy;
            toast.success(m.settings_reload_config_success());
        } catch (err) {
            console.error("Failed to reload config:", err);
            toast.error(m.settings_reload_config_error());
        } finally {
            reloadingConfig = false;
        }
    }

</script>

<div
    class="min-h-[calc(100vh-1rem)] from-background to-muted/30"
>
    <div
        class="mx-auto flex max-w-3xl flex-col gap-4 px-4 py-6 sm:px-6 sm:py-10 opacity-80"
    >
        <header class="flex items-start justify-between gap-3">
            <div class="min-w-0">
                <h1
                    class="text-balance text-2xl font-semibold tracking-tight sm:text-3xl"
                >
                    {m.settings_title()}
                </h1>
                <p class="mt-2 text-sm text-muted-foreground">
                    {m.settings_description()}
                </p>
            </div>
            <Button
                class="cursor-pointer hover:cursor-pointer"
                variant="ghost"
                onclick={() => goto("/")}
            >
                <ChevronLeft class="size-4" /> {m.settings_back()}
            </Button>
        </header>

        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title>{m.settings_language_title()}</Card.Title>
                <Card.Description
                    >{m.settings_language_description()}</Card.Description
                >
            </Card.Header>
            <Card.Content>
                <RadioGroup.Root
                    bind:value={form.selectedLanguage}
                    class="flex flex-col gap-3"
                >
                    <div class="flex items-center space-x-2">
                        <RadioGroup.Item
                            value="en"
                            id="en"
                            class="cursor-pointer hover:cursor-pointer"
                        />
                        <Label
                            for="en"
                            class="flex items-center gap-2 cursor-pointer hover:cursor-pointer"
                        >
                            <Us class="size-4 rounded-sm shadow-sm" />
                            {m.settings_language_english()}
                        </Label>
                    </div>
                    <div class="flex items-center space-x-2">
                        <RadioGroup.Item
                            value="it"
                            id="it"
                            class="cursor-pointer hover:cursor-pointer"
                        />
                        <Label
                            for="it"
                            class="flex items-center gap-2 cursor-pointer hover:cursor-pointer"
                        >
                            <It class="size-4 rounded-sm shadow-sm" />
                            {m.settings_language_italian()}
                        </Label>
                    </div>
                </RadioGroup.Root>
                <div class="text-xs text-muted-foreground mt-4">
                    <strong>{m.settings_info_label()}</strong>
                    {m.settings_language_info()}
                </div>
            </Card.Content>
        </Card.Root>

        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title>{m.settings_appearance_title()}</Card.Title>
                <Card.Description
                    >{m.settings_appearance_description()}</Card.Description
                >
            </Card.Header>
            <Card.Content class="space-y-4">
                <RadioGroup.Root
                    bind:value={form.theme}
                    class="flex flex-col gap-3"
                >
                    <div class="flex items-center space-x-2">
                        <RadioGroup.Item
                            value="light"
                            id="theme-light"
                            class="cursor-pointer hover:cursor-pointer"
                        />
                        <Label
                            for="theme-light"
                            class="flex items-center gap-2 cursor-pointer hover:cursor-pointer"
                        >
                            <Sun class="size-4" />
                            {m.settings_theme_light()}
                        </Label>
                    </div>
                    <div class="flex items-center space-x-2">
                        <RadioGroup.Item
                            value="dark"
                            id="theme-dark"
                            class="cursor-pointer hover:cursor-pointer"
                        />
                        <Label
                            for="theme-dark"
                            class="flex items-center gap-2 cursor-pointer hover:cursor-pointer"
                        >
                            <Moon class="size-4" />
                            {m.settings_theme_dark()}
                        </Label>
                    </div>
                </RadioGroup.Root>
                <div class="text-xs text-muted-foreground mt-4">
                    <strong>{m.settings_info_label()}</strong>
                    {m.settings_theme_hint()}
                </div>

                <Separator />

                <div class="space-y-3">
                    <SettingsSwitchLabel
                        bind:featureBool={form.reduceMotion}
                        labelText={m.settings_reduce_motion_label()}
                        hintText={m.settings_reduce_motion_hint()}
                        infoText={m.settings_reduce_motion_info()}
                    />
                </div>
            </Card.Content>
        </Card.Root>

        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title>{m.settings_mailviewer_title()}</Card.Title>
                <Card.Description
                    >{m.settings_mailviewer_description()}</Card.Description
                >
            </Card.Header>
            <Card.Content class="space-y-4">
                <div class="space-y-3">
                    <SettingsSwitchLabel
                        bind:featureBool={form.useDarkEmailViewer}
                        labelText={m.settings_email_dark_viewer_label()}
                        hintText={m.settings_email_dark_viewer_hint()}
                        infoText={m.settings_email_dark_viewer_info()}
                    />

                    <Separator />

                    <SettingsSwitchLabel
                        bind:featureBool={form.enableLinkClickConfirmation}
                        labelText={m.settings_link_confirmation_label()}
                        hintText={m.settings_link_confirmation_hint()}
                        infoText={m.settings_link_confirmation_info()}
                    />

                    <Separator />

                    <SettingsSwitchLabel
                        bind:featureBool={form.fixEmailTextContrast}
                        labelText={m.settings_contrast_fix_label()}
                        hintText={m.settings_contrast_fix_hint()}
                        infoText={m.settings_contrast_fix_info()}
                    />

                    <Separator />

                    <SettingsSwitchLabel
                        bind:featureBool={form.enableTabMode}
                        labelText={m.settings_danger_tab_mode_label()}
                        hintText={m.settings_danger_tab_mode_hint()}
                        infoText={m.settings_danger_tab_mode_info()}
                    />
                </div>
            </Card.Content>
        </Card.Root>

        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title>{m.settings_preview_page_title()}</Card.Title>
                <Card.Description
                    >{m.settings_preview_page_description()}</Card.Description
                >
            </Card.Header>
            <Card.Content class="space-y-4">
                <div class="space-y-4">
                    <Label>{m.settings_preview_images_label()}</Label>
                    <div class="flex flex-col gap-3">
                    {#each availablePreviewFileSupportedTypes || [] as type}
                        <div class="flex items-center space-x-2">
                            <Checkbox
                                id={"preview-" + type}
                                checked={form.previewFileSupportedTypes?.includes(
                                    type,
                                )}
                                onCheckedChange={(checked) => {
                                    console.log(checked)
                                    const types = new Set(
                                        form.previewFileSupportedTypes || [],
                                    );
                                    console.log(types)
                                    if (checked) types.add(type);
                                    else types.delete(type);
                                    form.previewFileSupportedTypes = Array.from(
                                        types,
                                    ).sort() as any[];
                                }}
                            />
                            <Label
                                for={"preview-" + type}
                                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                            >
                                {type.toUpperCase()} (.{type})
                            </Label>
                        </div>
                    {/each}
                        
                        <p class="text-xs text-muted-foreground mt-2">
                            {m.settings_preview_images_hint()}
                        </p>
                        <Separator />
                    </div>
                </div>
                <div class="space-y-3">
                    <SettingsSwitchLabel
                        bind:featureBool={form.useBuiltinPreview}
                        labelText={m.settings_preview_builtin_label()}
                        hintText={m.settings_preview_builtin_hint()}
                        infoText={m.settings_preview_builtin_info()}
                    />
                </div>
                <Separator />

                <div class="space-y-3">

                    <SettingsSwitchLabel
                        bind:featureBool={form.useBuiltinPDFViewer}
                        labelText={m.settings_preview_pdf_builtin_label()}
                        hintText={m.settings_preview_pdf_builtin_hint()}
                        infoText={m.settings_preview_pdf_builtin_info()}
                    />
                </div>

                <Separator />

                <div class="space-y-3">
                    <SettingsSwitchLabel
                        bind:featureBool={form.openAttachmentsAsTab}
                        labelText={m.settings_attachments_tab_label()}
                        hintText={m.settings_attachments_tab_hint()}
                        infoText={m.settings_attachments_tab_info()}
                        disabled={!form.enableTabMode}
                    />
                </div>

                <Separator />

                <div class="rounded-lg border bg-card p-4 space-y-3">
                    <div class="space-y-1">
                        <Label class="text-sm"
                            >{m.settings_export_folder_label()}</Label
                        >
                        <div class="text-sm text-muted-foreground">
                            {m.settings_export_folder_hint()}
                        </div>
                    </div>
                    <div class="flex items-end gap-2">
                        <div class="flex flex-col gap-1.5 flex-1 min-w-0">
                            <Label class="text-xs text-muted-foreground"
                                >{m.settings_export_folder_label()}</Label
                            >
                            <div
                                class="flex h-9 w-full items-center rounded-md border border-input bg-muted px-3 py-1 text-sm shadow-sm font-mono truncate text-muted-foreground"
                                title={exportFolder || "%USERPROFILE%\\Downloads"}
                            >
                                {exportFolder || "%USERPROFILE%\\Downloads"}
                            </div>
                        </div>
                        <Button
                            variant="outline"
                            class="cursor-pointer hover:cursor-pointer shrink-0 h-9"
                            onclick={selectExportFolder}
                            disabled={savingExportFolder}
                        >
                            {m.settings_select_folder_button()}
                        </Button>
                    </div>
                    <div class="text-xs text-muted-foreground">
                        {m.settings_export_folder_info()}
                    </div>
                </div>
            </Card.Content>
        </Card.Root>

        {#if $dangerZoneEnabled || dev}
            <Card.Root class="border-destructive/50 bg-destructive/15">
                <Card.Header class="space-y-1">
                    <Card.Title class="text-destructive"
                        >{m.settings_danger_zone_title()}</Card.Title
                    >
                    <Card.Description
                        >{m.settings_danger_zone_description()}</Card.Description
                    >
                </Card.Header>
                <Card.Content class="space-y-3">
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_danger_devtools_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_devtools_hint()}
                            </div>
                            {#if $hostIntegrityFailed}
                                <div class="text-xs text-destructive">
                                    {m.settings_danger_devtools_blocked_hint()}
                                </div>
                            {/if}
                        </div>
                        <Button
                                variant="destructive"
                                class="cursor-pointer hover:cursor-pointer devtools-btn"
                                onclick={() => OpenDevTools()}
                                disabled={$hostIntegrityFailed}
                            >
                                {m.settings_danger_devtools_btn_label()}
                            </Button>
                    </div>
                    <Separator />
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_danger_tray_icon_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_tray_icon_hint()}
                            </div>
                            <div class="text-xs text-muted-foreground">
                                <strong>{m.settings_danger_tray_icon_info()}</strong>
                            </div>
                        </div>
                        <div class="flex items-center gap-2">
                            <Switch
                                checked={trayIconEnabled}
                                class="cursor-pointer hover:cursor-pointer"
                                disabled={savingTrayIcon}
                                onCheckedChange={toggleTrayIcon}
                            />
                        </div>
                    </div>
                    <Separator />
    
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_danger_safety_check_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_safety_check_hint()}
                            </div>
                        </div>
                        <Button
                            variant="destructive"
                            class="cursor-pointer hover:cursor-pointer"
                            onclick={openSafetyCheckDialog}
                        >
                            {m.settings_danger_safety_check_btn()}
                        </Button>
                    </div>
                    <Separator />
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_danger_reload_all_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_reload_hint()}
                            </div>
                        </div>

                        <div class="flex items-center gap-2">
                            <a
                                data-sveltekit-reload
                                href="/"
                                class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
                                style="text-decoration: none;"
                                id="reload-ui-btn"
                            >
                                {m.settings_danger_reload_button_ui()}
                            </a>
                            
                        </div>
                    </div>

                    <Separator />

                    <!-- Reload config from disk -->
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_reload_config_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_reload_config_description_p1()}
                                <code
                                    class="text-xs bg-muted px-1 py-0.5 rounded"
                                    >config.ini</code
                                >
                                {m.settings_reload_config_description_p2()}
                            </div>
                        </div>
                        <Button
                            variant="destructive"
                            class="cursor-pointer hover:cursor-pointer"
                            onclick={reloadConfig}
                            disabled={reloadingConfig}
                        >
                            {m.settings_reload_config_btn()}
                        </Button>
                    </div>

                    <Separator />

                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_danger_reset_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_reset_hint()}
                            </div>
                        </div>

                        <AlertDialog.Root>
                            <AlertDialog.Trigger
                                class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
                            >
                                {m.settings_danger_reset_button()}
                            </AlertDialog.Trigger>
                            <AlertDialog.Content>
                                <AlertDialog.Header>
                                    <AlertDialog.Title
                                        style="color: var(--destructive); opacity: 0.7;"
                                    >
                                        <u
                                            >{m.settings_danger_reset_dialog_title()}</u
                                        >
                                    </AlertDialog.Title>
                                    <AlertDialog.Description>
                                        {m.settings_danger_reset_dialog_description_part1()}
                                        <br />
                                        {m.settings_danger_reset_dialog_description_part2()}
                                    </AlertDialog.Description>
                                </AlertDialog.Header>
                                <AlertDialog.Footer>
                                    <AlertDialog.Cancel
                                        class="cursor-pointer hover:cursor-pointer"
                                        >{m.settings_danger_reset_dialog_cancel()}</AlertDialog.Cancel
                                    >
                                    <AlertDialog.Action
                                        onclick={() => {
                                            resetToDefaults();
                                            goto("/");
                                        }}
                                        class="cursor-pointer hover:cursor-pointer"
                                        >{m.settings_danger_reset_dialog_continue()}</AlertDialog.Action
                                    >
                                </AlertDialog.Footer>
                            </AlertDialog.Content>
                        </AlertDialog.Root>
                    </div>
                    <div class="text-xs text-muted-foreground">
                        <strong>{m.settings_danger_warning()}</strong>
                    </div>
                    <Separator />

                    <SettingsSwitchLabel
                        bind:featureBool={form.enableAttachedDebuggerProtection}
                        labelText={m.settings_danger_debugger_protection_label()}
                        hintText={m.settings_danger_debugger_protection_hint()}
                        infoText={m.settings_danger_debugger_protection_info()}
                        type="danger"
                        runningInDevMode={!runningInDevMode}
                    />

                    <Separator />

                    <SettingsSwitchLabel
                        bind:featureBool={form.enableHostIntegrityCheck}
                        labelText={m.settings_danger_host_integrity_label()}
                        hintText={m.settings_danger_host_integrity_hint()}
                        infoText={m.settings_danger_host_integrity_info()}
                        type="danger-bypass"
                        runningInDevMode={!runningInDevMode}
                        machineData={machineData}
                    />

                    <Separator />

                    <!-- Test crash handler buttons -->
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm">{m.settings_danger_test_crash_label()}</Label>
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_test_crash_hint()}
                            </div>
                        </div>
                        <div class="flex items-center gap-2">
                            <Button
                                variant="destructive"
                                class="cursor-pointer hover:cursor-pointer"
                                onclick={() => goto("/test-500")}
                            >
                                500
                            </Button>
                            <Button
                                variant="destructive"
                                class="cursor-pointer hover:cursor-pointer"
                                onclick={() => goto("/emly-test-404-nonexistent")}
                            >
                                404
                            </Button>
                        </div>
                    </div>

                    <Separator />

                    <div class="text-xs text-muted-foreground">
                        GUI: {config
                            ? `${config.GUISemver} (${config.GUIReleaseChannel})`
                            : m.settings_not_available()}
                        <br />
                        SDK: {config
                            ? `${config.SDKDecoderSemver} (${config.SDKDecoderReleaseChannel})`
                            : m.settings_not_available()}
                        <br />
                        <br />
                        <span class="inline-flex items-center gap-1">
                            Hostname:
                            {#if evaluateHostname(machineData?.Hostname ?? "")}
                                <CircleCheck class="size-4" />
                            {:else}
                                <CircleX class="size-4" />
                            {/if}
                        </span>
                        <br />
                        <span class="inline-flex items-center gap-1">
                            TREGCC AD:
                            {#if isInsideTREGCCADDomain(machineData?.ADDomain ?? "")}
                                <CircleCheck class="size-4" />
                            {:else}
                                <CircleX class="size-4" />
                            {/if}
                        </span>
                    </div>
                </Card.Content>
            </Card.Root>
        {/if}

        <Dialog.Root bind:open={safetyCheckDialogOpen}>
            <Dialog.Content class="sm:max-w-md">
                <Dialog.Header>
                    <Dialog.Title>{m.settings_safety_check_dialog_title()}</Dialog.Title>
                    <Dialog.Description>
                        {m.settings_safety_check_dialog_description()}
                    </Dialog.Description>
                </Dialog.Header>

                {#if runningSafetyCheck && !safetyCheckResult}
                    <div class="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground">
                        <Loader2 class="size-4 animate-spin" />
                        {m.settings_safety_check_running()}
                    </div>
                {:else if safetyCheckResult}
                    <div class="space-y-2">
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_hostname()}
                            {#if safetyCheckResult.hostnameOk}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_ad_domain()}
                            {#if safetyCheckResult.adDomainOk}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_updater_installed()}
                            {#if safetyCheckResult.updaterInstalled}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_updater_running()}
                            {#if safetyCheckResult.updaterRunning}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_ipc_installed()}
                            {#if safetyCheckResult.ipcInstalled}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_ipc_responding()}
                            {#if safetyCheckResult.ipcResponding}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_cross_check()}
                            {#if safetyCheckResult.crossCheckOk}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>
                        <div class="flex items-center justify-between gap-2 text-sm">
                            {m.settings_safety_check_item_host_integrity_toggle()}
                            {#if safetyCheckResult.hostIntegrityToggleOk}
                                <CircleCheck class="size-4 text-green-500" />
                            {:else}
                                <CircleX class="size-4 text-red-500" />
                            {/if}
                        </div>

                        <Separator />

                        <div class="flex items-center justify-between gap-2 pt-1">
                            <span class="text-sm font-medium">
                                {m.settings_safety_check_standing_label()}
                            </span>
                            {#if safetyCheckResult.standing === "perfect"}
                                <Badge class="bg-green-600 text-white border-transparent">
                                    <CircleCheck />
                                    {m.settings_safety_check_standing_perfect()}
                                </Badge>
                            {:else if safetyCheckResult.standing === "usable"}
                                <Badge class="bg-yellow-500 text-black border-transparent">
                                    <TriangleAlert />
                                    {m.settings_safety_check_standing_usable()}
                                </Badge>
                            {:else}
                                <Badge variant="destructive">
                                    <OctagonX /> 
                                    {m.settings_safety_check_standing_limited()}
                                </Badge>
                            {/if}
                        </div>
                        <div class="text-xs text-muted-foreground">
                            {#if safetyCheckResult.standing === "perfect"}
                                {m.settings_safety_check_standing_perfect_hint()}
                            {:else if safetyCheckResult.standing === "usable"}
                                {m.settings_safety_check_standing_usable_hint()}
                            {:else}
                                {m.settings_safety_check_standing_limited_hint()}
                            {/if}
                        </div>
                    </div>
                {/if}

                <Dialog.Footer>
                    <Button
                        variant="outline"
                        class="cursor-pointer hover:cursor-pointer"
                        onclick={runSafetyCheck}
                        disabled={runningSafetyCheck}
                    >
                        {#if runningSafetyCheck}
                            <Loader2 class="size-4 animate-spin" />
                        {/if}
                        {m.settings_safety_check_rerun_btn()}
                    </Button>
                    <Dialog.Close
                        class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
                    >
                        {m.settings_safety_check_close_btn()}
                    </Dialog.Close>
                </Dialog.Footer>
            </Dialog.Content>
        </Dialog.Root>

        {#if !runningInDevMode}
            <AlertDialog.Root bind:open={dangerWarningOpen}>
                <AlertDialog.Content>
                    <AlertDialog.Header>
                        <AlertDialog.Title
                            >{m.settings_danger_alert_title()}</AlertDialog.Title
                        >
                        <AlertDialog.Description>
                            {m.settings_danger_alert_description_part1()}
                            <br />
                            {m.settings_danger_alert_description_part2()}
                            <br />
                            {m.settings_danger_alert_description_part3()}
                        </AlertDialog.Description>
                    </AlertDialog.Header>
                    <AlertDialog.Footer>
                        <AlertDialog.Action
                            onclick={() => (dangerWarningOpen = false)}
                            >{m.settings_danger_alert_understood()}</AlertDialog.Action
                        >
                    </AlertDialog.Footer>
                </AlertDialog.Content>
            </AlertDialog.Root>
        {/if}
    </div>
</div>

<style>
    /* Overrides the button base's `disabled:pointer-events-none` so hovering
       a host-integrity-disabled button still shows a "not-allowed" cursor
       instead of falling through to whatever is behind it. Native `disabled`
       already blocks clicks/activation regardless of pointer-events. */
    :global(button.devtools-btn:disabled) {
        pointer-events: auto;
        cursor: not-allowed;
    }
</style>
