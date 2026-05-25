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
        Download,
        RefreshCw,
        CheckCircle2,
        AlertCircle,
        Sun,
        Moon,
        Save,
    } from "@lucide/svelte";
    import { Input } from "$lib/components/ui/input";
    import type { EMLy_GUI_Settings } from "$lib/types";
    import { toast } from "svelte-sonner";
    import { It, Us } from "svelte-flags";
    import * as RadioGroup from "$lib/components/ui/radio-group/index.js";
    import * as Select from "$lib/components/ui/select/index.js";
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
    import { buttonVariants } from "$lib/components/ui/button/index.js";
    import { Badge } from "$lib/components/ui/badge/index.js";
    import { Checkbox } from "$lib/components/ui/checkbox/index.js";
    import {
        dismissUnsavedChangesToast,
        showUnsavedChangesToast,
    } from "$lib/utils/unsaved-changes-toast";
    import { dangerZoneEnabled, unsavedChanges } from "$lib/stores/app";
    import { LogDebug, LogInfo } from "$lib/wailsjs/runtime/runtime";
    import { settingsStore } from "$lib/stores/settings.svelte";
    import * as m from "$lib/paraglide/messages";
    import { setLocale } from "$lib/paraglide/runtime";
    import { mailState } from "$lib/stores/mail-state.svelte.js";
    import { dev } from "$app/environment";
    import {
        CheckForUpdates,
        DownloadUpdate,
        InstallUpdate,
        SetUpdateCheckerEnabled,
        ReloadConfig,
        SetUpdatePath,
        SetReleaseChannel,
    } from "$lib/wailsjs/go/main/App";
    import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";
    import { UPDATE_PATH_OPTIONS, type UpdatePathOption, UNC_PATH_RE, LOCAL_PATH_RE, UPDATE_PATH_LABELS } from "$lib/utils/settingsHelper";
    import SettingsSwitchLabel from "$lib/components/settings/SettingsSwitchLabel.svelte";

    let { data } = $props();
    let config = $derived(data.config);
    let previousTheme = $state<string | undefined>(undefined);
    let runningInDevMode: boolean = dev;
    
    // Clone store state for form editing
    // Use normalizeSettings to ensure new fields (like useBuiltinPDFViewer) are populated with defaults
    let form = $state<EMLy_GUI_Settings>(
        normalizeSettings(settingsStore.settings),
    );
    let lastSaved = $state<EMLy_GUI_Settings>(
        normalizeSettings(settingsStore.settings),
    );
    let dangerWarningOpen = $state(false);
    let updatePathSelection = $state<UpdatePathOption>("DC-RM2");
    let customUpdatePath = $state("");
    let savingUpdatePath = $state(false);
    let savingChannel = $state(false);
    let reloadingConfig = $state(false);
    let updatePathLabel = $derived(UPDATE_PATH_LABELS[updatePathSelection]);
    
    const customPathValid = $derived(
        updatePathSelection !== "Other" ||
            UNC_PATH_RE.test(customUpdatePath.trim()) ||
            (runningInDevMode && LOCAL_PATH_RE.test(customUpdatePath.trim())),
    );

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

    const defaults: EMLy_GUI_Settings = {
        selectedLanguage: "it",
        useBuiltinPreview: true,
        useBuiltinPDFViewer: true,
        previewFileSupportedTypes: ["jpg", "jpeg", "png"],
        enableAttachedDebuggerProtection: true,
        useDarkEmailViewer: true,
        enableUpdateChecker: true,
        reduceMotion: false,
        theme: "dark",
        enableLinkClickConfirmation: false,
        enableTabMode: false,
        fixEmailTextContrast: true,
    };

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
                s.selectedLanguage || defaults.selectedLanguage || "en",
            useBuiltinPreview: !!s.useBuiltinPreview,
            useBuiltinPDFViewer:
                s.useBuiltinPDFViewer ?? defaults.useBuiltinPDFViewer ?? true,
            previewFileSupportedTypes:
                s.previewFileSupportedTypes ||
                defaults.previewFileSupportedTypes ||
                [],
            enableAttachedDebuggerProtection:
                s.enableAttachedDebuggerProtection ??
                defaults.enableAttachedDebuggerProtection ??
                true,
            useDarkEmailViewer:
                s.useDarkEmailViewer ?? defaults.useDarkEmailViewer ?? true,
            enableUpdateChecker: runningInDevMode
                ? false
                : (s.enableUpdateChecker ??
                  defaults.enableUpdateChecker ??
                  true),
            reduceMotion: s.reduceMotion ?? defaults.reduceMotion ?? false,
            theme: s.theme || defaults.theme || "light",
            enableLinkClickConfirmation:
                s.enableLinkClickConfirmation ??
                defaults.enableLinkClickConfirmation ??
                false,
            enableTabMode:
                s.enableTabMode ??
                defaults.enableTabMode ??
                false,
            fixEmailTextContrast:
                s.fixEmailTextContrast ??
                defaults.fixEmailTextContrast ??
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
        } else {
            toast.success(m.settings_toast_saved());
        }
    }

    async function resetToDefaults() {
        form = normalizeSettings(defaults);
        lastSaved = normalizeSettings(defaults);

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
                LogInfo("Settings reset to defaults.");
            } catch {
                toast.error(m.settings_toast_reset_failed());
                return;
            }
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
            isSameSettings(lastSaved, defaults) &&
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

    // Sync update checker setting to backend config.ini
    let previousUpdateCheckerEnabled = $state<boolean | undefined>(undefined);
    $effect(() => {
        (async () => {
            if (!browser) return;
            if (previousUpdateCheckerEnabled === undefined) {
                previousUpdateCheckerEnabled = form.enableUpdateChecker;
                return;
            }
            if (form.enableUpdateChecker !== previousUpdateCheckerEnabled) {
                try {
                    await SetUpdateCheckerEnabled(
                        form.enableUpdateChecker ?? true,
                    );
                    LogDebug(
                        `Update checker ${form.enableUpdateChecker ? "enabled" : "disabled"}`,
                    );
                } catch (err) {
                    console.error(
                        "Failed to sync update checker setting:",
                        err,
                    );
                }
                previousUpdateCheckerEnabled = form.enableUpdateChecker;
            }
        })();
    });

    // Sync theme with email viewer dark mode
    $effect(() => {
        if (!browser) return;
        if (previousTheme !== undefined && form.theme !== previousTheme) {
            form.useDarkEmailViewer = form.theme === "dark";
        }
        previousTheme = form.theme;
    });

    async function reloadConfig() {
        reloadingConfig = true;
        try {
            const freshConfig = await ReloadConfig();
            config = freshConfig.EMLy;
            toast.success("Config ricaricato da config.ini");
        } catch (err) {
            console.error("Failed to reload config:", err);
            toast.error("Errore durante il ricaricamento del config");
        } finally {
            reloadingConfig = false;
        }
    }

    async function saveReleaseChannel(channel: string) {
        savingChannel = true;
        try {
            await SetReleaseChannel(channel);
            const freshConfig = await ReloadConfig();
            config = freshConfig.EMLy;
            // Reset update status: changing channel invalidates the previous check
            updateStatus = {
                currentVersion: updateStatus.currentVersion,
                availableVersion: "",
                updateAvailable: false,
                checking: false,
                downloading: false,
                downloadProgress: 0,
                ready: false,
                installerPath: "",
                errorMessage: "",
                releaseNotes: undefined,
                lastCheckTime: "",
                channel,
            };
            toast.success(m.settings_update_channel_saved({ channel }));
        } catch (err) {
            console.error("Failed to set release channel:", err);
            toast.error(m.settings_update_channel_error());
        } finally {
            savingChannel = false;
        }
    }

    async function saveUpdatePath() {
        const path =
            updatePathSelection === "Other"
                ? customUpdatePath.trim()
                : UPDATE_PATH_OPTIONS[updatePathSelection];
        if (!path) {
            toast.error("Inserire un percorso valido");
            return;
        }
        savingUpdatePath = true;
        try {
            await SetUpdatePath(path);
            const freshConfig = await ReloadConfig();
            config = freshConfig.EMLy;
            toast.success(`Percorso aggiornamento salvato: ${path}`);
        } catch (err) {
            console.error("Failed to set update path:", err);
            toast.error("Errore durante il salvataggio del percorso");
        } finally {
            savingUpdatePath = false;
        }
    }

    // Update System State
    type UpdateStatus = {
        currentVersion: string;
        availableVersion: string;
        updateAvailable: boolean;
        checking: boolean;
        downloading: boolean;
        downloadProgress: number;
        ready: boolean;
        installerPath: string;
        errorMessage: string;
        releaseNotes?: string;
        severityType?: string;
        lastCheckTime: string;
        channel?: string;
    };

    let updateStatus = $state<UpdateStatus>({
        currentVersion: "Unknown",
        availableVersion: "",
        updateAvailable: false,
        checking: false,
        downloading: false,
        downloadProgress: 0,
        ready: false,
        installerPath: "",
        errorMessage: "",
        lastCheckTime: "",
    });

    let showSecurityAlert = $state(false);
    let securityAlertShownForVersion = $state("");

    function getSeverityConfig(severityType: string | undefined) {
        switch (severityType) {
            case "security": return { border: "border-red-500/50",    bg: "bg-red-500/10",    badgeVariant: "destructive" as const, label: m.settings_updates_severity_security() };
            case "breaking": return { border: "border-amber-500/50",  bg: "bg-amber-500/10",  badgeVariant: "outline" as const,     label: m.settings_updates_severity_breaking() };
            case "feature":  return { border: "border-emerald-500/50",bg: "bg-emerald-500/10",badgeVariant: "secondary" as const,   label: m.settings_updates_severity_feature() };
            default:         return { border: "border-blue-500/30",   bg: "bg-blue-500/10",   badgeVariant: "outline" as const,     label: m.settings_updates_severity_patch() };
        }
    }

    const severityConfig = $derived(getSeverityConfig(updateStatus.severityType));

    // Sync current version from config
    $effect(() => {
        if (config?.GUISemver) {
            updateStatus.currentVersion = config.GUISemver;
        }
    });

    async function checkForUpdates() {
        try {
            const status = await CheckForUpdates();
            console.log("checkForUpdates status", status);
            updateStatus = status;

            if (status.updateAvailable && status.severityType === "security" && securityAlertShownForVersion !== status.availableVersion) {
                showSecurityAlert = true;
                securityAlertShownForVersion = status.availableVersion;
            }

            if (status.updateAvailable) {
                toast.success(
                    m.settings_toast_update_available({
                        version: status.availableVersion,
                    }),
                );
            } else if (!status.errorMessage) {
                toast.info(m.settings_toast_latest_version());
            } else {
                toast.error(status.errorMessage);
            }
        } catch (err) {
            console.error("Failed to check for updates:", err);
            updateStatus.checking = false;
            updateStatus.errorMessage = String(err);
            if(updateStatus.errorMessage.includes("failed to resolve manifest path: path not accessible: GetFileAttributesEx")) {
                updateStatus.errorMessage = m.settings_update_path_inaccessible();
            }
            updateStatus.lastCheckTime = new Date().toISOString();
            toast.error(m.settings_toast_check_failed());
        }
    }

    async function downloadUpdate() {
        try {
            await DownloadUpdate();
            toast.success(m.settings_toast_download_success());
        } catch (err) {
            console.error("Failed to download update:", err);
            toast.error(m.settings_toast_download_failed());
        }
    }

    async function installUpdate() {
        try {
            await InstallUpdate(true); // true = quit after launch
            // App will quit, so no toast needed
        } catch (err) {
            console.error("Failed to install update:", err);
            toast.error(m.settings_toast_install_failed());
        }
    }

    // Listen for update status events
    $effect(() => {
        if (!browser) return;

        EventsOn("update:status", (status: UpdateStatus) => {
            updateStatus = status;
            if (status.updateAvailable && status.severityType === "security" && securityAlertShownForVersion !== status.availableVersion) {
                showSecurityAlert = true;
                securityAlertShownForVersion = status.availableVersion;
            }
        });

        return () => {
            EventsOff("update:status");
        };
    });
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
            </Card.Content>
        </Card.Root>

        <!-- Update Section -->
        {#if form.enableUpdateChecker}
            <Card.Root>
                <Card.Header class="space-y-1">
                    <Card.Title class="flex items-center gap-2">
                        {m.settings_updates_title()}
                        {#if updateStatus.updateAvailable && updateStatus.severityType}
                            <Badge
                                variant={severityConfig.badgeVariant}
                                class={updateStatus.severityType === "breaking" ? "text-amber-600 dark:text-amber-400 border-amber-500/50" : updateStatus.severityType === "feature" ? "text-emerald-600 dark:text-emerald-400" : ""}
                            >
                                {severityConfig.label}
                            </Badge>
                        {/if}
                    </Card.Title>
                    <Card.Description
                        >{m.settings_updates_description()}</Card.Description
                    >
                </Card.Header>
                <Card.Content class="space-y-4">
                    <!-- Release Channel -->
                    <div class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4">
                        <div>
                            <div class="font-medium">{m.settings_update_channel_label()}</div>
                            <div class="text-sm text-muted-foreground">
                                {m.settings_update_channel_description()}
                            </div>
                        </div>
                        <RadioGroup.Root
                            value={config?.GUIReleaseChannel || "stable"}
                            onValueChange={saveReleaseChannel}
                            class="flex gap-4"
                            disabled={savingChannel}
                        >
                            <div class="flex items-center gap-2">
                                <RadioGroup.Item value="stable" id="ch-stable" class="cursor-pointer" />
                                <Label for="ch-stable" class="cursor-pointer">{m.settings_update_channel_stable()}</Label>
                            </div>
                            <div class="flex items-center gap-2">
                                <RadioGroup.Item value="beta" id="ch-beta" class="cursor-pointer" />
                                <Label for="ch-beta" class="cursor-pointer">{m.settings_update_channel_beta()}</Label>
                            </div>
                        </RadioGroup.Root>
                    </div>
                    <p class="text-xs text-muted-foreground">
                        {m.settings_update_channel_recheck_hint()}
                    </p>

                    <Separator />

                    <!-- Current Version -->
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
                    >
                        <div>
                            <div class="font-medium">
                                {m.settings_updates_current_version()}
                            </div>
                            <div class="text-sm text-muted-foreground">
                                {updateStatus.currentVersion} ({config?.GUIReleaseChannel ||
                                    "stable"})
                            </div>
                        </div>
                        {#if updateStatus.updateAvailable}
                            <div
                                class="flex items-center gap-2 text-sm font-medium text-green-600 dark:text-green-400"
                            >
                                <AlertCircle class="size-4" />
                                {m.settings_updates_available()}
                            </div>
                        {:else if updateStatus.errorMessage && updateStatus.lastCheckTime}
                            <div
                                class="flex items-center gap-2 text-sm text-destructive"
                            >
                                <AlertCircle class="size-4" />
                                {m.settings_updates_check_failed()}
                            </div>
                        {:else if updateStatus.lastCheckTime}
                            <div
                                class="flex items-center gap-2 text-sm text-muted-foreground"
                            >
                                <CheckCircle2 class="size-4" />
                                {m.settings_updates_no_updates()}
                            </div>
                        {/if}
                    </div>

                    <Separator />

                    <!-- Check for Updates -->
                    <div
                        class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
                    >
                        <div>
                            <div class="font-medium">
                                {m.settings_updates_check_label()}
                            </div>
                            <div class="text-sm text-muted-foreground">
                                {#if updateStatus.lastCheckTime}
                                    {m.settings_updates_last_checked({
                                        time: updateStatus.lastCheckTime,
                                    })}
                                {:else}
                                    {m.settings_updates_click_check()}
                                {/if}
                            </div>
                        </div>
                        <Button
                            variant="outline"
                            class="cursor-pointer hover:cursor-pointer"
                            onclick={checkForUpdates}
                            disabled={updateStatus.checking ||
                                updateStatus.downloading}
                        >
                            <RefreshCw
                                class="size-4 mr-2 {updateStatus.checking
                                    ? 'animate-spin'
                                    : ''}"
                            />
                            {updateStatus.checking
                                ? m.settings_updates_checking()
                                : m.settings_updates_check_now()}
                        </Button>
                    </div>

                    <!-- Download Update (shown when update available) -->
                    {#if updateStatus.updateAvailable && !updateStatus.ready}
                        <Separator />
                        <div
                            class="flex items-center justify-between gap-4 rounded-lg border {severityConfig.border} {severityConfig.bg} p-4"
                        >
                            <div>
                                <div class="font-medium">
                                    {m.settings_updates_version_available({
                                        version: updateStatus.availableVersion,
                                    })}
                                </div>
                                <div class="text-sm text-muted-foreground">
                                    {#if updateStatus.downloading}
                                        {m.settings_updates_downloading({
                                            progress:
                                                updateStatus.downloadProgress,
                                        })}
                                    {:else}
                                        {m.settings_updates_click_download()}
                                    {/if}
                                </div>
                                {#if updateStatus.releaseNotes}
                                    <div
                                        class="text-xs text-muted-foreground mt-2"
                                    >
                                        {updateStatus.releaseNotes}
                                    </div>
                                {/if}
                            </div>
                            <Button
                                variant="default"
                                class="cursor-pointer hover:cursor-pointer"
                                onclick={downloadUpdate}
                                disabled={updateStatus.downloading}
                            >
                                <Download class="size-4 mr-2" />
                                {updateStatus.downloading
                                    ? `${updateStatus.downloadProgress}%`
                                    : m.settings_updates_download_button()}
                            </Button>
                        </div>
                    {/if}

                    <!-- Install Update (shown when download ready) -->
                    {#if updateStatus.ready}
                        <Separator />
                        <div
                            class="flex items-center justify-between gap-4 rounded-lg border border-green-500/30 bg-green-500/10 p-4"
                        >
                            <div>
                                <div class="font-medium">
                                    {m.settings_updates_ready_title()}
                                </div>
                                <div class="text-sm text-muted-foreground">
                                    {m.settings_updates_ready_ref({
                                        version: updateStatus.availableVersion,
                                    })}
                                </div>
                            </div>
                            <Button
                                variant="default"
                                class="cursor-pointer hover:cursor-pointer bg-green-600 hover:bg-green-700"
                                onclick={installUpdate}
                            >
                                <CheckCircle2 class="size-4 mr-2" />
                                {m.settings_updates_install_button()}
                            </Button>
                        </div>
                    {/if}

                    <!-- Error Message -->
                    {#if updateStatus.errorMessage}
                        <div
                            class="rounded-lg border border-destructive/50 bg-destructive/10 p-3"
                        >
                            <div class="flex items-start gap-2">
                                <AlertCircle
                                    class="size-4 text-destructive mt-0.5"
                                />
                                <div class="text-sm text-destructive">
                                    {updateStatus.errorMessage}
                                </div>
                            </div>
                        </div>
                    {/if}

                    <!-- Info about update path -->
                    <div class="text-xs text-muted-foreground">
                        <strong>{m.settings_info_label()}</strong>
                        {m.settings_updates_info_message()}
                        {#if config?.UpdatePath}
                            {m.settings_updates_current_path()}
                            <code class="text-xs bg-muted px-1 py-0.5 rounded"
                                >{config?.UpdatePath}</code
                            >
                        {:else}
                            <span class="text-amber-600 dark:text-amber-400"
                                >{m.settings_updates_no_path()}</span
                            >
                        {/if}
                    </div>
                </Card.Content>
            </Card.Root>

            <!-- Security update alert dialog -->
            <AlertDialog.Root bind:open={showSecurityAlert}>
                <AlertDialog.Content>
                    <AlertDialog.Header>
                        <AlertDialog.Title style="color: var(--destructive); opacity: 0.7;">{m.settings_updates_security_alert_title()}</AlertDialog.Title>
                        <AlertDialog.Description>
                            {m.settings_updates_security_alert_description({ version: updateStatus.availableVersion })}
                        </AlertDialog.Description>
                    </AlertDialog.Header>
                    <AlertDialog.Footer>
                        <AlertDialog.Cancel>{m.settings_updates_security_alert_later()}</AlertDialog.Cancel>
                        <AlertDialog.Action onclick={() => (
                            downloadUpdate(),
                            showSecurityAlert = false
                        )}>
                            {m.settings_updates_security_alert_update_now()}
                        </AlertDialog.Action>
                    </AlertDialog.Footer>
                </AlertDialog.Content>
            </AlertDialog.Root>
        {/if}

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
                        </div>
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

                    {#if form.enableUpdateChecker}
                        <Separator />

                        <!-- Update location selector -->
                        <div
                            class="rounded-lg border border-destructive/30 bg-card p-4 space-y-3"
                        >
                            <div class="space-y-1">
                                <Label class="text-sm"
                                    >{m.settings_update_path_label()}</Label
                                >
                                <div class="text-sm text-muted-foreground">
                                    {m.settings_update_path_description()}
                                    <code
                                        class="text-xs bg-muted px-1 py-0.5 rounded"
                                        >config.ini</code
                                    >.
                                </div>
                            </div>
                            <div class="flex items-end gap-2">
                                <div class="flex flex-col gap-1.5 flex-1">
                                    <Label class="text-xs text-muted-foreground"
                                        >Server</Label
                                    >
                                    <Select.Root
                                        type="single"
                                        bind:value={updatePathSelection}
                                    >
                                        <Select.Trigger
                                            class="w-full cursor-pointer hover:cursor-pointer"
                                        >
                                            {updatePathLabel}
                                        </Select.Trigger>
                                        <Select.Content>
                                            <Select.Item
                                                value="DC-RM2"
                                                label={UPDATE_PATH_LABELS[
                                                    "DC-RM2"
                                                ]}
                                            />
                                            <Select.Item
                                                value="DC-CB"
                                                label={UPDATE_PATH_LABELS[
                                                    "DC-CB"
                                                ]}
                                            />
                                            <Select.Item
                                                value="Other"
                                                label={UPDATE_PATH_LABELS[
                                                    "Other"
                                                ]}
                                            />
                                        </Select.Content>
                                    </Select.Root>
                                </div>
                                <Button
                                    variant="destructive"
                                    class="cursor-pointer hover:cursor-pointer"
                                    onclick={saveUpdatePath}
                                    disabled={savingUpdatePath ||
                                        !customPathValid}
                                >
                                    <Save class="size-4 mr-2" />
                                    {m.settings_update_path_btn()}
                                </Button>
                            </div>
                            {#if updatePathSelection === "Other"}
                                <div class="flex flex-col gap-1.5">
                                    <Label class="text-xs text-muted-foreground"
                                        >{m.settings_update_unc_label()}</Label
                                    >
                                    <Input
                                        bind:value={customUpdatePath}
                                        placeholder={m.settings_update_unc_placeholder()}
                                        class="font-mono text-sm {customUpdatePath.trim() &&
                                        !customPathValid
                                            ? 'border-destructive focus-visible:ring-destructive'
                                            : ''}"
                                    />
                                    {#if customUpdatePath.trim() && !customPathValid}
                                        <p class="text-xs text-destructive">
                                            {m.settings_update_unc_invalid()}
                                        </p>
                                    {/if}
                                </div>
                            {/if}
                            {#if config?.UpdatePath}
                                <div class="text-xs text-muted-foreground">
                                    {m.settings_update_path_hint()}
                                    <code
                                        class="text-xs bg-muted px-1 py-0.5 rounded"
                                        >{config?.UpdatePath}</code
                                    >
                                </div>
                            {/if}
                        </div>
                    {/if}

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
                        bind:featureBool={form.enableUpdateChecker}
                        labelText={m.settings_danger_update_checker_label()}
                        hintText={m.settings_danger_update_checker_hint()}
                        infoText={m.settings_danger_update_checker_info()}
                        type="danger"
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
