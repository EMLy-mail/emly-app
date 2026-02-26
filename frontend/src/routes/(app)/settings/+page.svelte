<script lang="ts">
  import { browser } from "$app/environment";
  import { goto } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Label } from "$lib/components/ui/label";
  import { Separator } from "$lib/components/ui/separator";
  import { Switch } from "$lib/components/ui/switch";
  import { ChevronLeft, Flame, Download, Upload, RefreshCw, CheckCircle2, AlertCircle, Sun, Moon, RefreshCcw } from "@lucide/svelte";
  import type { EMLy_GUI_Settings } from "$lib/types";
  import { toast } from "svelte-sonner";
  import { It, Us } from "svelte-flags";
  import * as RadioGroup from "$lib/components/ui/radio-group/index.js";
  import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
  import { buttonVariants } from "$lib/components/ui/button/index.js";
  import { Checkbox } from "$lib/components/ui/checkbox/index.js";
  import {
    dismissUnsavedChangesToast,
    showUnsavedChangesToast,
  } from "$lib/utils/unsaved-changes-toast";
  import { dangerZoneEnabled, unsavedChanges } from "$lib/stores/app";
  import { LogDebug } from "$lib/wailsjs/runtime/runtime";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import * as m from "$lib/paraglide/messages";
  import { setLocale } from "$lib/paraglide/runtime";
  import { mailState } from "$lib/stores/mail-state.svelte.js";
  import { dev } from '$app/environment';
  import { ExportSettings, ImportSettings, CheckForUpdates, DownloadUpdate, InstallUpdate, GetUpdateStatus, SetUpdateCheckerEnabled, RestartApp } from "$lib/wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "$lib/wailsjs/runtime/runtime";

  let { data } = $props();
  let config = $derived(data.config);

  let runningInDevMode: boolean = dev || false;

  const defaults: EMLy_GUI_Settings = {
    selectedLanguage: "it",
    useBuiltinPreview: true,
    useBuiltinPDFViewer: true,
    previewFileSupportedTypes: ["jpg", "jpeg", "png"],
    enableAttachedDebuggerProtection: true,
    useDarkEmailViewer: true,
    enableUpdateChecker: false,
    reduceMotion: false,
    theme: "dark",
    increaseWindowButtonsContrast: false,
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

  // Clone store state for form editing
  // Use normalizeSettings to ensure new fields (like useBuiltinPDFViewer) are populated with defaults
  let form = $state<EMLy_GUI_Settings>(normalizeSettings(settingsStore.settings));
  let lastSaved = $state<EMLy_GUI_Settings>(normalizeSettings(settingsStore.settings));
  let dangerWarningOpen = $state(false);

  function normalizeSettings(s: EMLy_GUI_Settings): EMLy_GUI_Settings {
    return {
      selectedLanguage: s.selectedLanguage || defaults.selectedLanguage || "en",
      useBuiltinPreview: !!s.useBuiltinPreview,
      useBuiltinPDFViewer:
        s.useBuiltinPDFViewer ?? defaults.useBuiltinPDFViewer ?? true,
      previewFileSupportedTypes:
        s.previewFileSupportedTypes || defaults.previewFileSupportedTypes || [],
      enableAttachedDebuggerProtection:
        s.enableAttachedDebuggerProtection ?? defaults.enableAttachedDebuggerProtection ?? true,
      useDarkEmailViewer:
        s.useDarkEmailViewer ?? defaults.useDarkEmailViewer ?? true,
      enableUpdateChecker: runningInDevMode
        ? false
        : (s.enableUpdateChecker ?? defaults.enableUpdateChecker ?? true),
      reduceMotion: s.reduceMotion ?? defaults.reduceMotion ?? false,
      theme: s.theme || defaults.theme || "light",
      increaseWindowButtonsContrast: s.increaseWindowButtonsContrast ?? defaults.increaseWindowButtonsContrast ?? false,
    };
  }

  function isSameSettings(a: EMLy_GUI_Settings, b: EMLy_GUI_Settings) {
    return (
      (a.selectedLanguage ?? "") === (b.selectedLanguage ?? "") &&
      !!a.useBuiltinPreview === !!b.useBuiltinPreview &&
      !!a.useBuiltinPDFViewer === !!b.useBuiltinPDFViewer &&
      !!a.enableAttachedDebuggerProtection === !!b.enableAttachedDebuggerProtection &&
      !!a.useDarkEmailViewer === !!b.useDarkEmailViewer &&
      !!a.enableUpdateChecker === !!b.enableUpdateChecker &&
      !!a.reduceMotion === !!b.reduceMotion &&
      (a.theme ?? "light") === (b.theme ?? "light") &&
      !!a.increaseWindowButtonsContrast === !!b.increaseWindowButtonsContrast &&
      JSON.stringify(a.previewFileSupportedTypes?.sort()) ===
        JSON.stringify(b.previewFileSupportedTypes?.sort())
    );
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
        dangerZoneEnabled.set(false);
        LogDebug("Reset danger zone setting to false.");
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

    const dirty = !isSameSettings(normalizeSettings(form), lastSaved);
    unsavedChanges.set(dirty);
    if (dirty) {
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
        toast.info(m.settings_danger_here_be_dragons(), { icon: Flame });
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
          await SetUpdateCheckerEnabled(form.enableUpdateChecker ?? true);
          LogDebug(`Update checker ${form.enableUpdateChecker ? 'enabled' : 'disabled'}`);
        } catch (err) {
          console.error('Failed to sync update checker setting:', err);
        }
        previousUpdateCheckerEnabled = form.enableUpdateChecker;
      }
    })();
  });

  // Sync theme with email viewer dark mode
  let previousTheme = $state<string | undefined>(undefined);
  $effect(() => {
    if (!browser) return;
    if (previousTheme !== undefined && form.theme !== previousTheme) {
      form.useDarkEmailViewer = form.theme === "dark";
    }
    previousTheme = form.theme;
  });

  async function exportSettings() {
    try {
      const settingsJSON = JSON.stringify(form, null, 2);
      const result = await ExportSettings(settingsJSON);
      if (result) {
        toast.success(m.settings_export_success());
      }
    } catch (err) {
      console.error("Failed to export settings:", err);
      toast.error(m.settings_export_error());
    }
  }

  async function importSettings() {
    try {
      const result = await ImportSettings();
      if (result) {
        try {
          const imported = JSON.parse(result) as EMLy_GUI_Settings;
          // Validate that it looks like a valid settings object
          if (typeof imported === 'object' && imported !== null) {
            form = normalizeSettings(imported);
            toast.success(m.settings_import_success());
          } else {
            toast.error(m.settings_import_invalid());
          }
        } catch {
          toast.error(m.settings_import_invalid());
        }
      }
    } catch (err) {
      console.error("Failed to import settings:", err);
      toast.error(m.settings_import_error());
    }
  }

  async function restartEntireApp() {
    try {
      await RestartApp();
    } catch(e) {
      toast.error(m.settings_danger_reload_app_error());
      console.error(e);
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
    lastCheckTime: string;
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

  // Sync current version from config
  $effect(() => {
    if (config?.GUISemver) {
      updateStatus.currentVersion = config.GUISemver;
    }
  });

  async function checkForUpdates() {
    try {
      const status = await CheckForUpdates();
      updateStatus = status;
      
      if (status.updateAvailable) {
        toast.success(m.settings_toast_update_available({ version: status.availableVersion }));
      } else if (!status.errorMessage) {
        toast.info(m.settings_toast_latest_version());
      } else {
        toast.error(status.errorMessage);
      }
    } catch (err) {
      console.error("Failed to check for updates:", err);
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
    });

    return () => {
      EventsOff("update:status");
    };
  });
</script>

<div class="min-h-[calc(100vh-1rem)] bg-linear-to-b from-background to-muted/30">
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
        ><ChevronLeft class="size-4" /> {m.settings_back()}</Button
      >
    </header>

    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title>{m.settings_language_title()}</Card.Title>
        <Card.Description>{m.settings_language_description()}</Card.Description>
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
        <Card.Description>{m.settings_appearance_description()}</Card.Description>
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
          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
          >
            <div>
              <div class="font-medium">
                {m.settings_reduce_motion_label()}
              </div>
              <div class="text-sm text-muted-foreground">
                {m.settings_reduce_motion_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.reduceMotion}
              class="cursor-pointer hover:cursor-pointer"
            />
          </div>
          <p class="text-xs text-muted-foreground mt-2">
            {m.settings_reduce_motion_info()}
          </p>

          <Separator />

          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
          >
            <div>
              <div class="font-medium">
                {m.settings_window_buttons_contrast_label()}
              </div>
              <div class="text-sm text-muted-foreground">
                {m.settings_window_buttons_contrast_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.increaseWindowButtonsContrast}
              class="cursor-pointer hover:cursor-pointer"
            />
          </div>

          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
          >
            <div>
              <div class="font-medium">
                {m.settings_email_dark_viewer_label()}
              </div>
              <div class="text-sm text-muted-foreground">
                {m.settings_email_dark_viewer_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.useDarkEmailViewer}
              class="cursor-pointer hover:cursor-pointer"
            />
          </div>
          <p class="text-xs text-muted-foreground mt-2">
            {m.settings_email_dark_viewer_info()}
          </p>
        </div>
      </Card.Content>
    </Card.Root>

    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title>{m.settings_export_import_title()}</Card.Title>
        <Card.Description>{m.settings_export_import_description()}</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div
          class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
        >
          <div>
            <div class="font-medium">{m.settings_export_button()}</div>
            <div class="text-sm text-muted-foreground">
              {m.settings_export_hint()}
            </div>
          </div>
          <Button
            variant="outline"
            class="cursor-pointer hover:cursor-pointer"
            onclick={exportSettings}
          >
            <Download class="size-4 mr-2" />
            {m.settings_export_button()}
          </Button>
        </div>

        <div
          class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
        >
          <div>
            <div class="font-medium">{m.settings_import_button()}</div>
            <div class="text-sm text-muted-foreground">
              {m.settings_import_hint()}
            </div>
          </div>
          <Button
            variant="outline"
            class="cursor-pointer hover:cursor-pointer"
            onclick={importSettings}
          >
            <Upload class="size-4 mr-2" />
            {m.settings_import_button()}
          </Button>
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
            <div class="flex items-center space-x-2">
              <Checkbox
                id="preview-jpg"
                checked={form.previewFileSupportedTypes?.includes("jpg")}
                onCheckedChange={(checked) => {
                  const types = new Set(form.previewFileSupportedTypes || []);
                  if (checked) types.add("jpg");
                  else types.delete("jpg");
                  form.previewFileSupportedTypes = Array.from(
                    types,
                  ).sort() as any[];
                }}
              />
              <Label
                for="preview-jpg"
                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                JPG (.jpg)
              </Label>
            </div>

            <div class="flex items-center space-x-2">
              <Checkbox
                id="preview-jpeg"
                checked={form.previewFileSupportedTypes?.includes("jpeg")}
                onCheckedChange={(checked) => {
                  const types = new Set(form.previewFileSupportedTypes || []);
                  if (checked) types.add("jpeg");
                  else types.delete("jpeg");
                  form.previewFileSupportedTypes = Array.from(
                    types,
                  ).sort() as any[];
                }}
              />
              <Label
                for="preview-jpeg"
                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                JPEG (.jpeg)
              </Label>
            </div>

            <div class="flex items-center space-x-2">
              <Checkbox
                id="preview-png"
                checked={form.previewFileSupportedTypes?.includes("png")}
                onCheckedChange={(checked) => {
                  const types = new Set(form.previewFileSupportedTypes || []);
                  if (checked) types.add("png");
                  else types.delete("png");
                  form.previewFileSupportedTypes = Array.from(
                    types,
                  ).sort() as any[];
                }}
              />
              <Label
                for="preview-png"
                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                PNG (.png)
              </Label>
            </div>
            <p class="text-xs text-muted-foreground mt-2">
              {m.settings_preview_images_hint()}
            </p>
            <Separator />
          </div>
        </div>
        <div class="space-y-3">
          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
          >
            <div>
              <div class="font-medium">
                {m.settings_preview_builtin_label()}
              </div>
              <div class="text-sm text-muted-foreground">
                {m.settings_preview_builtin_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.useBuiltinPreview}
              class="cursor-pointer hover:cursor-pointer"
            />
          </div>
          <p class="text-xs text-muted-foreground mt-2">
            {m.settings_preview_builtin_info()}
          </p>
        </div>
        <Separator />

        <div class="space-y-3">
          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
          >
            <div>
              <div class="font-medium">
                {m.settings_preview_pdf_builtin_label()}
              </div>
              <div class="text-sm text-muted-foreground">
                {m.settings_preview_pdf_builtin_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.useBuiltinPDFViewer}
              class="cursor-pointer hover:cursor-pointer"
            />
          </div>
          <p class="text-xs text-muted-foreground mt-2">
            {m.settings_preview_pdf_builtin_info()}
          </p>
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Update Section -->
    {#if form.enableUpdateChecker}
    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title>{m.settings_updates_title()}</Card.Title>
        <Card.Description>{m.settings_updates_description()}</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <!-- Current Version -->
        <div class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4">
          <div>
            <div class="font-medium">{m.settings_updates_current_version()}</div>
            <div class="text-sm text-muted-foreground">
              {updateStatus.currentVersion} ({config?.GUIReleaseChannel || "stable"})
            </div>
          </div>
          {#if updateStatus.updateAvailable}
            <div class="flex items-center gap-2 text-sm font-medium text-green-600 dark:text-green-400">
              <AlertCircle class="size-4" />
              {m.settings_updates_available()}
            </div>
          {:else if updateStatus.errorMessage && updateStatus.lastCheckTime}
            <div class="flex items-center gap-2 text-sm text-destructive">
              <AlertCircle class="size-4" />
              {m.settings_updates_check_failed()}
            </div>
          {:else if updateStatus.lastCheckTime}
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <CheckCircle2 class="size-4" />
              {m.settings_updates_no_updates()}
            </div>
          {/if}
        </div>

        <Separator />

        <!-- Check for Updates -->
        <div class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4">
          <div>
            <div class="font-medium">{m.settings_updates_check_label()}</div>
            <div class="text-sm text-muted-foreground">
              {#if updateStatus.lastCheckTime}
                {m.settings_updates_last_checked({ time: updateStatus.lastCheckTime })}
              {:else}
                {m.settings_updates_click_check()}
              {/if}
            </div>
          </div>
          <Button
            variant="outline"
            class="cursor-pointer hover:cursor-pointer"
            onclick={checkForUpdates}
            disabled={updateStatus.checking || updateStatus.downloading}
          >
            <RefreshCw class="size-4 mr-2 {updateStatus.checking ? 'animate-spin' : ''}" />
            {updateStatus.checking ? m.settings_updates_checking() : m.settings_updates_check_now()}
          </Button>
        </div>

        <!-- Download Update (shown when update available) -->
        {#if updateStatus.updateAvailable && !updateStatus.ready}
          <Separator />
          <div class="flex items-center justify-between gap-4 rounded-lg border border-blue-500/30 bg-blue-500/10 p-4">
            <div>
              <div class="font-medium">{m.settings_updates_version_available({ version: updateStatus.availableVersion })}</div>
              <div class="text-sm text-muted-foreground">
                {#if updateStatus.downloading}
                  {m.settings_updates_downloading({ progress: updateStatus.downloadProgress })}
                {:else}
                  {m.settings_updates_click_download()}
                {/if}
              </div>
              {#if updateStatus.releaseNotes}
                <div class="text-xs text-muted-foreground mt-2">
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
              {updateStatus.downloading ? `${updateStatus.downloadProgress}%` : m.settings_updates_download_button()}
            </Button>
          </div>
        {/if}

        <!-- Install Update (shown when download ready) -->
        {#if updateStatus.ready}
          <Separator />
          <div class="flex items-center justify-between gap-4 rounded-lg border border-green-500/30 bg-green-500/10 p-4">
            <div>
              <div class="font-medium">{m.settings_updates_ready_title()}</div>
              <div class="text-sm text-muted-foreground">
                {m.settings_updates_ready_ref({ version: updateStatus.availableVersion })}
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
          <div class="rounded-lg border border-destructive/50 bg-destructive/10 p-3">
            <div class="flex items-start gap-2">
              <AlertCircle class="size-4 text-destructive mt-0.5" />
              <div class="text-sm text-destructive">
                {updateStatus.errorMessage}
              </div>
            </div>
          </div>
        {/if}

        <!-- Info about update path -->
        <div class="text-xs text-muted-foreground">
          <strong>{m.settings_info_label()}</strong> {m.settings_updates_info_message()}
          {#if (config as any)?.UpdatePath}
            {m.settings_updates_current_path()} <code class="text-xs bg-muted px-1 py-0.5 rounded">{(config as any).UpdatePath}</code>
          {:else}
            <span class="text-amber-600 dark:text-amber-400">{m.settings_updates_no_path()}</span>
          {/if}
        </div>
      </Card.Content>
    </Card.Root>
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
              <Label class="text-sm">{m.settings_danger_devtools_label()}</Label
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
              <Label class="text-sm">{m.settings_danger_reload_all_label()}</Label>
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
              <Button
                class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer disabled:cursor-not-allowed disabled:opacity-50`}
                onclick={restartEntireApp}
                style="text-decoration: none;"
                disabled={runningInDevMode}
              >
                {m.settings_danger_reload_button_app()}
              </Button>
            </div>
          </div>
          
          <Separator />

          <div
            class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
          >
            <div class="space-y-1">
              <Label class="text-sm">{m.settings_danger_reset_label()}</Label>
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
                    <u>{m.settings_danger_reset_dialog_title()}</u>
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

          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4 border-destructive/30"
          >
            <div class="space-y-1">
              <Label class="text-sm">{m.settings_danger_debugger_protection_label()}</Label>
              <div class="text-sm text-muted-foreground">
                {m.settings_danger_debugger_protection_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.enableAttachedDebuggerProtection}
              class="cursor-pointer hover:cursor-pointer"
              disabled={!runningInDevMode}
            />
          </div>
          <div class="text-xs text-muted-foreground">
            <strong>{m.settings_danger_debugger_protection_info()}</strong>
          </div>
          <Separator />

          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4 border-destructive/30"
          >
            <div class="space-y-1">
              <Label class="text-sm">{m.settings_danger_update_checker_label()}</Label>
              <div class="text-sm text-muted-foreground">
                {m.settings_danger_update_checker_hint()}
              </div>
            </div>
            <Switch
              bind:checked={form.enableUpdateChecker}
              class="cursor-pointer hover:cursor-pointer"
              disabled={runningInDevMode}
            />
          </div>
          <div class="text-xs text-muted-foreground">
            {m.settings_danger_update_checker_info()}
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
          <AlertDialog.Action onclick={() => (dangerWarningOpen = false)}
            >{m.settings_danger_alert_understood()}</AlertDialog.Action
          >
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog.Root>
    {/if}
  </div>
</div>
