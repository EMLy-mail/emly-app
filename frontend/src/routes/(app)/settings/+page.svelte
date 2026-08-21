<script lang="ts">
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { Label } from "$lib/components/ui/label";
    import { Separator } from "$lib/components/ui/separator";
    import { Switch } from "$lib/components/ui/switch";
    import { ChevronLeft, Flame, Sun, Moon } from "@lucide/svelte";
    import type {
        EMLy_GUI_Settings,
        PdfRendererEngine,
        ReleaseChannel,
    } from "$lib/types";
    import { toast } from "svelte-sonner";
    import { It, Us } from "svelte-flags";
    import * as RadioGroup from "$lib/components/ui/radio-group/index.js";
    import * as Select from "$lib/components/ui/select/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
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
    } from "$lib/stores/app";
    import { LogDebug, LogInfo } from "$lib/wailsjs/runtime/runtime";
    import {
        settingsStore,
        defaultSettings,
        releaseChannels,
        parseReleaseChannel,
    } from "$lib/stores/settings.svelte";
    import * as m from "$lib/paraglide/messages";
    import { setLocale } from "$lib/paraglide/runtime";
    import { mailState } from "$lib/stores/mail-state.svelte.js";
    import { dev } from "$app/environment";
    import {
        ReloadConfig,
        ShowOpenFolderDialog,
        SetExportAttachmentFolder,
        CheckFolderWritable,
        SetTrayIconEnabled,
        SetLogStartupTrace,
        SetOldAttachmentPreload,
        SetGUIReleaseChannel,
        OpenDevTools,
        RestartApp,
    } from "$lib/wailsjs/go/main/App";
    import SettingsSwitchLabel from "$lib/components/settings/SettingsSwitchLabel.svelte";
    import SafetyCheckDialog from "$lib/components/settings/SafetyCheckDialog.svelte";
    import { systemInfoStore } from '$lib/stores/system-info.svelte.js';

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
    let logStartupTraceEnabled = $state(false);
    let savingLogStartupTrace = $state(false);
    let oldAttachmentPreloadEnabled = $state(false);
    let savingOldAttachmentPreload = $state(false);

    $effect(() => {
        exportFolder = config?.ExportAttachmentFolder ?? "";
    });

    $effect(() => {
        trayIconEnabled = !(config?.DisableTrayIcon ?? false);
    });

    $effect(() => {
        logStartupTraceEnabled = config?.LogStartupTrace ?? false;
    });

    $effect(() => {
        oldAttachmentPreloadEnabled = config?.OldAttachmentPreload ?? false;
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

    async function toggleLogStartupTrace(checked: boolean) {
        savingLogStartupTrace = true;
        try {
            await SetLogStartupTrace(checked);
            logStartupTraceEnabled = checked;
            toast.success(m.settings_danger_startup_trace_saved());
        } catch (e) {
            console.error("Failed to set startup trace logging", e);
            toast.error(m.settings_danger_startup_trace_save_error());
        } finally {
            savingLogStartupTrace = false;
        }
    }

    async function toggleOldAttachmentPreload(checked: boolean) {
        savingOldAttachmentPreload = true;
        try {
            await SetOldAttachmentPreload(checked);
            oldAttachmentPreloadEnabled = checked;
            toast.success(m.settings_danger_attachment_preload_saved());
        } catch (e) {
            console.error("Failed to set old attachment preload", e);
            toast.error(m.settings_danger_attachment_preload_save_error());
        } finally {
            savingOldAttachmentPreload = false;
        }
    }

    async function selectExportFolder() {
        savingExportFolder = true;
        try {
            const folder = await ShowOpenFolderDialog();
            if (!folder) return;

            // The picker only returns folders that exist, so a failure here
            // means EMLy is not allowed to write into it. Fall back to the
            // default rather than leaving a previously configured custom
            // folder active while the UI shows the default.
            try {
                await CheckFolderWritable(folder);
            } catch (e) {
                console.error("Export folder is not writable", e);
                await SetExportAttachmentFolder("");
                exportFolder = "";
                toast.error(m.settings_export_folder_not_writable());
                return;
            }

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
        "png",
        "bmp",
        "gif",
        "webp",
        "tiff",
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
        // Declared in here on purpose: normalizeSettings runs during component
        // init, before any module-level const further down has been evaluated.
        const pdfRenderers: PdfRendererEngine[] = ["builtin", "native"];

        return {
            selectedLanguage:
                s.selectedLanguage || defaultSettings.selectedLanguage || "en",
            useBuiltinPreview: !!s.useBuiltinPreview,
            useBuiltinPDFViewer:
                s.useBuiltinPDFViewer ?? defaultSettings.useBuiltinPDFViewer ?? true,
            // Validated rather than defaulted: a value left over from an older
            // build would otherwise stick around and leave the radio group with
            // nothing selected.
            pdfRenderer: pdfRenderers.includes(s.pdfRenderer as PdfRendererEngine)
                ? (s.pdfRenderer as PdfRendererEngine)
                : (defaultSettings.pdfRenderer ?? "builtin"),
            useNativePdfToolbar:
                s.useNativePdfToolbar ?? defaultSettings.useNativePdfToolbar ?? false,
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
            showSidebar: s.showSidebar ?? defaultSettings.showSidebar ?? true,
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
            releaseChannel: parseReleaseChannel(s.releaseChannel),
            // Defaults to true when the stored/config.ini channel is already
            // "next" (e.g. hand-edited config.ini, or a build shipped on
            // that channel), so an already-next install doesn't get locked
            // out of its own selected channel.
            enableNextReleaseChannel:
                s.enableNextReleaseChannel ??
                defaultSettings.enableNextReleaseChannel ??
                parseReleaseChannel(s.releaseChannel) === "next",
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

        // config.ini is written before the store, and a failure aborts the
        // save entirely: the two must not drift apart, and on the next start
        // the layout re-reads the channel from the ini and would silently
        // undo a store-only change anyway.
        if (settings.releaseChannel !== lastSaved.releaseChannel) {
            try {
                await SetGUIReleaseChannel(settings.releaseChannel ?? "stable");
                config = (await ReloadConfig()).EMLy;
            } catch (e) {
                console.error("Failed to save release channel to config.ini", e);
                toast.error(m.settings_release_channel_save_error());
                return;
            }
        }

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
        // GUI_RELEASE_CHANNEL is deliberately left alone: it describes which
        // builds this installation was set up to follow, not a UI preference,
        // and the reload below pulls it back out of config.ini into the store.
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

    let safetyCheckDialogOpen = $state(false);

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

    let nextChannelWarningOpen = $state(false);

    /*
     * "next" can only be picked from the select once enableNextReleaseChannel
     * is on (its Select.Item is disabled otherwise), so by the time this runs
     * the warning has already been shown and confirmed via the switch below.
     */
    function pickReleaseChannel(value: string) {
        form.releaseChannel = parseReleaseChannel(value);
    }

    /*
     * The switch is bound to this local draft rather than directly to
     * form.enableNextReleaseChannel so the confirmation warning can still
     * intercept it: a two-way binding onto the form field would have already
     * written the "on" value before the dialog ever opened.
     */
    let nextChannelSwitchDraft = $state(false);
    let previousNextChannelSwitchDraft = false;

    $effect(() => {
        // Keep the draft in sync whenever the underlying form value changes
        // from outside a switch click (initial load, reset, revert).
        nextChannelSwitchDraft = form.enableNextReleaseChannel ?? false;
    });

    $effect(() => {
        if (nextChannelSwitchDraft && !previousNextChannelSwitchDraft) {
            // Turning on: only actually enabled once the warning is confirmed.
            if (!form.enableNextReleaseChannel) {
                nextChannelWarningOpen = true;
            }
        } else if (!nextChannelSwitchDraft && previousNextChannelSwitchDraft) {
            // Turning off is a safe operation, applied immediately without
            // a confirmation dialog.
            form.enableNextReleaseChannel = false;
            if (form.releaseChannel === "next") {
                form.releaseChannel = "stable";
            }
        }
        previousNextChannelSwitchDraft = nextChannelSwitchDraft;
    });

    function cancelNextChannel() {
        nextChannelSwitchDraft = false;
        nextChannelWarningOpen = false;
    }

    function confirmNextChannel() {
        form.enableNextReleaseChannel = true;
        nextChannelWarningOpen = false;
    }

    function releaseChannelLabel(channel: ReleaseChannel) {
        switch (channel) {
            case "beta":
                return m.settings_release_channel_beta();
            case "next":
                return m.settings_release_channel_next();
            default:
                return m.settings_release_channel_stable();
        }
    }

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
                        bind:featureBool={form.showSidebar}
                        labelText={m.settings_show_sidebar_label()}
                        hintText={m.settings_show_sidebar_hint()}
                        infoText={m.settings_show_sidebar_info()}
                    />

                    <Separator />

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
                                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
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

                <div class="space-y-3" class:opacity-50={!form.useBuiltinPDFViewer}>
                    <div class="rounded-lg border bg-card p-4 space-y-3">
                        <div>
                            <div class="font-medium">
                                {m.settings_preview_pdf_engine_label()}
                            </div>
                            <div class="text-sm text-muted-foreground">
                                {m.settings_preview_pdf_engine_hint()}
                            </div>
                        </div>
                        <RadioGroup.Root
                            bind:value={form.pdfRenderer}
                            disabled={!form.useBuiltinPDFViewer}
                            class="flex flex-col gap-3"
                        >
                            <div class="flex items-center space-x-2">
                                <RadioGroup.Item
                                    value="builtin"
                                    id="pdf-engine-builtin"
                                    class="cursor-pointer hover:cursor-pointer"
                                />
                                <Label
                                    for="pdf-engine-builtin"
                                    class="cursor-pointer hover:cursor-pointer"
                                >
                                    {m.settings_preview_pdf_engine_builtin()}
                                </Label>
                            </div>
                            <div class="flex items-center space-x-2">
                                <RadioGroup.Item
                                    value="native"
                                    id="pdf-engine-native"
                                    class="cursor-pointer hover:cursor-pointer"
                                />
                                <Label
                                    for="pdf-engine-native"
                                    class="cursor-pointer hover:cursor-pointer"
                                >
                                    {m.settings_preview_pdf_engine_native()}
                                </Label>
                            </div>
                        </RadioGroup.Root>
                    </div>
                    <p class="text-xs text-muted-foreground mt-2">
                        {m.settings_preview_pdf_engine_info()}
                    </p>
                </div>

                <Separator />

                <div class="space-y-3">
                    <SettingsSwitchLabel
                        bind:featureBool={form.useNativePdfToolbar}
                        labelText={m.settings_preview_pdf_native_toolbar_label()}
                        hintText={m.settings_preview_pdf_native_toolbar_hint()}
                        infoText={m.settings_preview_pdf_native_toolbar_info()}
                        disabled={!form.useBuiltinPDFViewer ||
                            form.pdfRenderer !== "native"}
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
                                >{m.settings_danger_startup_trace_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_startup_trace_hint()}
                            </div>
                            <div class="text-xs text-muted-foreground">
                                <strong>{m.settings_danger_startup_trace_info()}</strong>
                            </div>
                        </div>
                        <div class="flex items-center gap-2">
                            <Switch
                                checked={logStartupTraceEnabled}
                                class="cursor-pointer hover:cursor-pointer"
                                disabled={savingLogStartupTrace}
                                onCheckedChange={toggleLogStartupTrace}
                            />
                        </div>
                    </div>
                    <Separator />
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                    >
                        <div class="space-y-1">
                            <Label class="text-sm"
                                >{m.settings_danger_attachment_preload_label()}</Label
                            >
                            <div class="text-sm text-muted-foreground">
                                {m.settings_danger_attachment_preload_hint()}
                            </div>
                            <div class="text-xs text-muted-foreground">
                                <strong>{m.settings_danger_attachment_preload_info()}</strong>
                            </div>
                        </div>
                        <div class="flex items-center gap-2">
                            <Switch
                                checked={oldAttachmentPreloadEnabled}
                                class="cursor-pointer hover:cursor-pointer"
                                disabled={savingOldAttachmentPreload}
                                onCheckedChange={toggleOldAttachmentPreload}
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
                            onclick={() => (safetyCheckDialogOpen = true)}
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

                    <div class="space-y-3">
                        <div
                            class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
                        >
                            <div>
                                <Label class="text-sm">{m.settings_release_channel_title()}</Label>
                                <div class="text-sm text-muted-foreground">
                                    {m.settings_release_channel_hint()}
                                </div>
                            </div>
                            <Select.Root
                                type="single"
                                value={form.releaseChannel}
                                onValueChange={pickReleaseChannel}
                            >
                                <Select.Trigger
                                    class="w-40 cursor-pointer hover:cursor-pointer"
                                >
                                    {releaseChannelLabel(form.releaseChannel ?? "stable")}
                                </Select.Trigger>
                                <Select.Content>
                                    <Select.Group>
                                        {#each releaseChannels as channel (channel)}
                                            <Select.Item
                                                value={channel}
                                                label={releaseChannelLabel(channel)}
                                                disabled={channel === "next" &&
                                                    !form.enableNextReleaseChannel}
                                                class="cursor-pointer hover:cursor-pointer"
                                            />
                                        {/each}
                                    </Select.Group>
                                </Select.Content>
                            </Select.Root>
                        </div>
                        <div class="text-xs text-muted-foreground">
                            {m.settings_release_channel_info()}
                        </div>
                    </div>

                    <Separator />

                    <SettingsSwitchLabel
                        bind:featureBool={nextChannelSwitchDraft}
                        labelText={m.settings_release_channel_enable_next_label()}
                        hintText={m.settings_release_channel_enable_next_hint()}
                        infoText={m.settings_release_channel_enable_next_info()}
                        type="danger"
                        runningInDevMode={!runningInDevMode}
                    />

                    <Separator />

                    <div class="text-xs text-muted-foreground">
                        GUI: {config
                            ? `${config.GUISemver} (${config.GUIReleaseChannel})`
                            : m.settings_not_available()}
                        <br />
                        SDK: {config
                            ? `${config.SDKDecoderSemver} (${config.SDKDecoderReleaseChannel})`
                            : m.settings_not_available()}
                    </div>
                </Card.Content>
            </Card.Root>
        {/if}

        <SafetyCheckDialog bind:open={safetyCheckDialogOpen} />

        <AlertDialog.Root bind:open={nextChannelWarningOpen}>
            <AlertDialog.Content>
                <AlertDialog.Header>
                    <AlertDialog.Title
                        >{m.settings_release_channel_next_warning_title()}</AlertDialog.Title
                    >
                    <AlertDialog.Description>
                        {m.settings_release_channel_next_warning_rewrite()}
                        <br />
                        {m.settings_release_channel_next_warning_ui()}
                        <br />
                        {m.settings_release_channel_next_warning_parity()}
                    </AlertDialog.Description>
                </AlertDialog.Header>
                <AlertDialog.Footer>
                    <AlertDialog.Cancel
                        onclick={cancelNextChannel}
                        class="cursor-pointer hover:cursor-pointer"
                        >{m.settings_release_channel_next_warning_cancel()}</AlertDialog.Cancel
                    >
                    <AlertDialog.Action
                        onclick={confirmNextChannel}
                        class="cursor-pointer hover:cursor-pointer"
                        >{m.settings_release_channel_next_warning_confirm()}</AlertDialog.Action
                    >
                </AlertDialog.Footer>
            </AlertDialog.Content>
        </AlertDialog.Root>

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
