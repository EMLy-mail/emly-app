<script lang="ts">
  import { browser } from "$app/environment";
  import { goto } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Label } from "$lib/components/ui/label";
  import { Separator } from "$lib/components/ui/separator";
  import { Switch } from "$lib/components/ui/switch";
  import { ChevronLeft, Command, Option, Flame } from "@lucide/svelte";
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
  import { GetMachineData } from "$lib/wailsjs/go/main/App";
  import * as m from "$lib/paraglide/messages";
  import { setLocale } from "$lib/paraglide/runtime";
  import { mailState } from "$lib/stores/mail-state.svelte.js";

  let { data } = $props();
  let machineData = $derived(data.machineData);
  let config = $derived(data.config);

  const defaults: EMLy_GUI_Settings = {
    selectedLanguage: "it",
    useBuiltinPreview: true,
    previewFileSupportedTypes: ["jpg", "jpeg", "png"],
  };


  async function setLanguage(lang: EMLy_GUI_Settings["selectedLanguage"] | null) {
    if (!browser) return;
    try {
      await setLocale(lang || "en", {reload: false});
      toast.success(m.settings_toast_language_changed());
    } catch {
      toast.error(m.settings_toast_language_change_failed());
    }
  }

  // Clone store state for form editing
  let form = $state<EMLy_GUI_Settings>({ ...settingsStore.settings });
  let lastSaved = $state<EMLy_GUI_Settings>({
    ...settingsStore.settings,
  });
  let dangerWarningOpen = $state(false);

  function normalizeSettings(
    s: EMLy_GUI_Settings,
  ): EMLy_GUI_Settings {
    return {
      selectedLanguage: s.selectedLanguage || defaults.selectedLanguage || "en",
      useBuiltinPreview: !!s.useBuiltinPreview,
      previewFileSupportedTypes:
        s.previewFileSupportedTypes || defaults.previewFileSupportedTypes || [],
    };
  }

  function isSameSettings(
    a: EMLy_GUI_Settings,
    b: EMLy_GUI_Settings,
  ) {
    return (
      (a.selectedLanguage ?? "") === (b.selectedLanguage ?? "") &&
      !!a.useBuiltinPreview === !!b.useBuiltinPreview &&
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
      form = { ...settingsStore.settings };
      lastSaved = { ...settingsStore.settings };
    }
  });

  let previousDangerZoneEnabled = $dangerZoneEnabled;

  $effect(() => {
    (async () => {
      if ($dangerZoneEnabled && !previousDangerZoneEnabled) {
        if (!data.machineData || data.machineData === undefined) {
          data.machineData = await GetMachineData();
        }
        dangerWarningOpen = true;
        toast.info("Here be dragons!", { icon: Flame });
      }
      previousDangerZoneEnabled = $dangerZoneEnabled;
    })();
  });
</script>

<div class="min-h-[calc(100vh-1rem)] from-background to-muted/30">
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
        onclick={() => goto("/")}><ChevronLeft class="size-4" /> {m.settings_back()}</Button
      >
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
          <strong>Info:</strong> {m.settings_language_info()}
        </div>
      </Card.Content>
    </Card.Root>

    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title>{m.settings_preview_files_title()}</Card.Title>
        <Card.Description
          >{m.settings_preview_files_description()}</Card.Description
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
          </div>
          <p class="text-xs text-muted-foreground mt-2">
            {m.settings_preview_images_hint()}
          </p>
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
        <div class="space-y-3">
          <div
            class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
          >
            <div>
              <div class="font-medium">{m.settings_preview_builtin_label()}</div>
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

        
      </Card.Content>
    </Card.Root>

    {#if $dangerZoneEnabled}
      <Card.Root class="border-destructive/50 bg-destructive/15">
        <Card.Header class="space-y-1">
          <Card.Title class="text-destructive">{m.settings_danger_zone_title()}</Card.Title>
          <Card.Description
            >{m.settings_danger_zone_description()}</Card.Description
          >
        </Card.Header>
        <Card.Content class="space-y-3">
          <div
            class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 bg-card p-4"
          >
            <div class="space-y-1">
              <Label class="text-sm">{m.settings_danger_devtools_label()}</Label>
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
                    <br/>
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
            <strong>{m.settings_danger_warning() }</strong>
          </div>
          <Separator />

          <div class="text-xs text-muted-foreground">
            OS: {machineData?.Version} ({machineData?.OS})
            <br />
            Hostname: {machineData?.Hostname}
            <br />
            ID: {machineData?.HWID}
            <br />
            CPU: {machineData?.CPU.processors[0].model.trim()} ({machineData
              ?.CPU.total_hardware_threads} cores)
            <br />
            RAM: {Math.round(
              (machineData?.RAM.total_physical_bytes ?? 0) /
                (1024 * 1024 * 1024),
            )} GB
            <br />
            GPU: {machineData?.GPU.cards?.find(
              (c) => !(c.pci?.product?.name ?? "").includes("Virtual"),
            )?.pci?.product?.name ?? "N/A"}
            <br />
            <br />
            GUI: {config
              ? `${config.GUISemver} (${config.GUIReleaseChannel})`
              : "N/A"}
            <br />
            SDK: {config
              ? `${config.SDKDecoderSemver} (${config.SDKDecoderReleaseChannel})`
              : "N/A"}
          </div>
        </Card.Content>
      </Card.Root>
    {/if}

    <AlertDialog.Root bind:open={dangerWarningOpen}>
      <AlertDialog.Content>
        <AlertDialog.Header>
          <AlertDialog.Title>{m.settings_danger_alert_title()}</AlertDialog.Title>
          <AlertDialog.Description>
            {m.settings_danger_alert_description()}
          </AlertDialog.Description>
        </AlertDialog.Header>
        <AlertDialog.Footer>
          <AlertDialog.Action onclick={() => (dangerWarningOpen = false)}
            >{m.settings_danger_alert_understood()}</AlertDialog.Action
          >
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog.Root>
  </div>
</div>
