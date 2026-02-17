<script lang="ts">
  import { bugReportDialogOpen } from "$lib/stores/app";
  import * as m from "$lib/paraglide/messages.js";
  import * as Dialog from "$lib/components/ui/dialog/index.js";
  import { Button, buttonVariants } from "$lib/components/ui/button/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import { Textarea } from "$lib/components/ui/textarea/index.js";
  import { CheckCircle, Copy, FolderOpen, Camera, Loader2, CloudUpload, AlertTriangle } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import { TakeScreenshot, SubmitBugReport, OpenFolderInExplorer, GetConfig } from "$lib/wailsjs/go/main/App";
  import { browser } from "$app/environment";
  import { dev } from "$app/environment";

  // Bug report form state
  let userName = $state("");
  let userEmail = $state("");
  let bugDescription = $state("");
  // Auto-fill form in dev mode
  $effect(() => {
    if (dev && $bugReportDialogOpen && !userName) {
      userName = "Test User";
      userEmail = "test@example.com";
      bugDescription = "This is a test bug report submitted from development mode.";
    }
  });
  // Bug report screenshot state
  let screenshotData = $state("");
  let isCapturing = $state(false);

  // Bug report system data
  let localStorageData = $state("");
  let configData = $state("");

  // Bug report UI state
  let isSubmitting = $state(false);
  let isSuccess = $state(false);
  let resultZipPath = $state("");
  let uploadedToServer = $state(false);
  let serverReportId = $state(0);
  let uploadError = $state("");
  let canSubmit: boolean = $derived(
    bugDescription.trim().length > 0 && userName.trim().length > 0 && userEmail.trim().length > 0 && !isSubmitting && !isCapturing
  );

  // Bug report dialog effects
  $effect(() => {
    if ($bugReportDialogOpen) {
      // Capture screenshot immediately when dialog opens
      captureScreenshot();
      // Capture localStorage data
      captureLocalStorage();
      // Capture config.ini data
      captureConfig();
    } else {
      // Reset form when dialog closes
      resetBugReportForm();
    }
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

  function captureLocalStorage() {
    if (!browser) return;
    try {
      const data: Record<string, string> = {};
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key) {
          data[key] = localStorage.getItem(key) || "";
        }
      }
      localStorageData = JSON.stringify(data, null, 2);
      console.log("localStorage data captured");
    } catch (err) {
      console.error("Failed to capture localStorage:", err);
      localStorageData = "Error capturing localStorage";
    }
  }

  async function captureConfig() {
    try {
      const config = await GetConfig();
      configData = JSON.stringify(config, null, 2);
      console.log("Config data captured");
    } catch (err) {
      console.error("Failed to capture config:", err);
      configData = "Error capturing config";
    }
  }

  function resetBugReportForm() {
    userName = "";
    userEmail = "";
    bugDescription = "";
    screenshotData = "";
    localStorageData = "";
    configData = "";
    isCapturing = false;
    isSubmitting = false;
    isSuccess = false;
    resultZipPath = "";
    uploadedToServer = false;
    serverReportId = 0;
    uploadError = "";
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
        screenshotData: screenshotData,
        localStorageData: localStorageData,
        configData: configData
      });

      resultZipPath = result.zipPath;
      uploadedToServer = result.uploaded;
      serverReportId = result.reportId;
      uploadError = result.uploadError;
      isSuccess = true;
      console.log("Bug report created:", result.zipPath, "uploaded:", result.uploaded);
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
</script>

<Dialog.Root bind:open={$bugReportDialogOpen}>
  <Dialog.Content class="sm:max-w-125 w-full max-h-[80vh] overflow-y-auto custom-scrollbar">
    {#if isSuccess}
      <!-- Success State -->
      <Dialog.Header>
        <Dialog.Title class="flex items-center gap-2">
          {#if uploadedToServer}
            <CloudUpload class="h-5 w-5 text-green-500" />
            {m.bugreport_uploaded_title()}
          {:else}
            <CheckCircle class="h-5 w-5 text-green-500" />
            {m.bugreport_success_title()}
          {/if}
        </Dialog.Title>
        <Dialog.Description>
          {#if uploadedToServer}
            {m.bugreport_uploaded_success({ reportId: String(serverReportId) })}
          {:else}
            {m.bugreport_success_message()}
          {/if}
        </Dialog.Description>
      </Dialog.Header>

      <div class="grid gap-4 py-4">
        {#if uploadError}
          <div class="flex items-start gap-2 bg-yellow-500/10 border border-yellow-500/30 rounded-md p-3">
            <AlertTriangle class="h-4 w-4 text-yellow-500 mt-0.5 shrink-0" />
            <p class="text-sm text-yellow-600 dark:text-yellow-400">{m.bugreport_upload_failed()}</p>
          </div>
        {/if}

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
              class="min-h-30"
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
          <Button type="submit" disabled={!canSubmit}>
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

<style>
  :global(.custom-scrollbar::-webkit-scrollbar) {
    width: 6px;
    height: 6px;
  }

  :global(.custom-scrollbar::-webkit-scrollbar-track) {
    background: transparent;
  }

  :global(.custom-scrollbar::-webkit-scrollbar-thumb) {
    background: var(--border);
    border-radius: 6px;
  }

  :global(.custom-scrollbar::-webkit-scrollbar-thumb:hover) {
    background: var(--muted-foreground);
  }

  :global(.custom-scrollbar::-webkit-scrollbar-corner) {
    background: transparent;
  }
</style>
