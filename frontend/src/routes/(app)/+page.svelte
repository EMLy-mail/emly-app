<script lang="ts">
  import MailViewer from "$lib/components/MailViewer.svelte";
  import PDFTabContent from "$lib/components/PDFTabContent.svelte";
  import ImageViewer from "$lib/components/ImageViewer.svelte";
  import TabStrip from "$lib/components/TabStrip.svelte";
  import { mailState } from "$lib/stores/mail-state.svelte";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import { sidebarOpen } from "$lib/stores/app";
  import * as m from "$lib/paraglide/messages.js";
  import { toast } from "svelte-sonner";
  import { Loader2 } from "@lucide/svelte";
  import {
    openAndLoadEmail,
    isPecOpenBlocked,
    loadEmailFromPath,
    isEmailFile,
  } from "$lib/utils/mail";
  import { onDestroy, onMount } from "svelte";
  import {
    EventsOn,
    EventsEmit,
    WindowShow,
    WindowUnminimise,
  } from "$lib/wailsjs/runtime/runtime";
  import { trace } from "$lib/utils/startupTrace";

  let { data } = $props();

  let isAddingTab = $state(false);

  let unregisterLaunchArgs = () => {};

  onMount(() => {
    trace("fe_page_mount_start");
    if (data.email) {
      if (isPecOpenBlocked(data.email)) return;
      if (settingsStore.settings.enableTabMode) {
        mailState.addTab(data.email, data.filePath);
        sidebarOpen.set(false);
      } else {
        mailState.setParams(data.email, data.filePath);
      }
      trace("fe_page_mount_done");
    } else if (data.loadError) {
      toast.error(m.mail_error_opening());
    }

    // Handles files opened while the app is already running (second-instance
    // launch). Registered once here at the page level - not per-tab - so it
    // fires regardless of which tab (mail or attachment) currently has focus.
    unregisterLaunchArgs = EventsOn("launchArgs", async (args: string[]) => {
      if (!args || args.length === 0 || isAddingTab) return;

      for (const arg of args) {
        if (isEmailFile(arg)) {
          isAddingTab = true;
          trace("fe_launch_args_read_start");

          const result = await loadEmailFromPath(arg);
          trace(
            "fe_launch_args_read_done",
            `attachments=${result.email?.attachments?.length ?? 0} body_len=${result.email?.body?.length ?? 0}`,
          );

          if (result.success && result.email) {
            if (isPecOpenBlocked(result.email)) {
              isAddingTab = false;
              break;
            }
            if (settingsStore.settings.enableTabMode) {
              mailState.addTab(result.email, result.filePath);
            } else {
              mailState.setParams(result.email, result.filePath);
            }
            sidebarOpen.set(false);
            WindowUnminimise();
            WindowShow();
            EventsEmit("bringOnTop");
          } else if (result.error) {
            toast.error(m.mail_error_opening());
          }

          isAddingTab = false;
          break;
        }
      }
    });
  });

  onDestroy(() => {
    unregisterLaunchArgs();
    if (!settingsStore.settings.enableTabMode) {
      mailState.getAllTabs().forEach((tab) => {
        if (tab.id !== mailState.getActiveTabId()) {
          mailState.removeTab(tab.id);
        }
      });
    }
  });

  async function openNewTab() {
    if (isAddingTab) return;
    isAddingTab = true;

    const result = await openAndLoadEmail();

    if (!result.cancelled && result.success && result.email) {
      if (!isPecOpenBlocked(result.email)) {
        mailState.addTab(result.email, result.filePath);
        sidebarOpen.set(false);
      }
    } else if (result.error) {
      toast.error(m.mail_error_opening());
    }

    isAddingTab = false;
  }

  let showTabs = $derived(
    settingsStore.settings.enableTabMode === true && mailState.tabs.length > 0,
  );
</script>

<div class="page">
  <section
    class="center"
    aria-label={m.page_overview_label()}
    id="main-content-app"
  >
    {#if showTabs}
      <!-- Windows 11 Explorer-style tabbed panel -->
      <div class="tabbed-panel" style="position: relative;">
        {#if isAddingTab}
          <div class="loading-overlay">
            <Loader2 class="spinner" size="40" />
          </div>
        {/if}
        <!-- Tab strip -->
        <TabStrip onAddTab={openNewTab} {isAddingTab} />

        <!-- Tab content panels - all mounted, shown/hidden via display -->
        <div class="tab-content-area">
          {#each mailState.tabs as tab (tab.id)}
            <div
              class="tab-panel"
              role="tabpanel"
              style:display={tab.id === mailState.activeTabId ? "flex" : "none"}
            >
              {#if tab.type === "email"}
                <MailViewer
                  emailData={tab.email}
                  tabId={tab.id}
                  embedded={true}
                />
              {:else if tab.type === "pdf" && tab.id === mailState.activeTabId}
                <PDFTabContent base64Data={tab.base64Data} filename={tab.filename} />
              {:else if tab.type === "image" && tab.id === mailState.activeTabId}
                <ImageViewer base64Data={tab.base64Data} filename={tab.filename} />
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <MailViewer />
    {/if}
  </section>
</div>

<style>
  .page {
    height: 100%;
    min-height: 0;
    display: flex;
    gap: 12px;
    padding: 12px;
    box-sizing: border-box;
    overflow: hidden;
  }

  .center {
    flex: 1 1 auto;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  /* ── Unified tabbed panel ── */
  .tabbed-panel {
    flex: 1 1 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: var(--card);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
  }

  /* ── Loading overlay ── */
  .loading-overlay {
    position: absolute;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    background: color-mix(in srgb, var(--background) 60%, transparent);
    backdrop-filter: blur(4px);
    border-radius: 14px;
  }

  :global(.spinner) {
    animation: spin 0.8s linear infinite;
    color: var(--muted-foreground);
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to   { transform: rotate(360deg); }
  }

  /* ── Tab content area ── */
  .tab-content-area {
    flex: 1 1 0;
    min-height: 0;
    position: relative;
    display: flex;
    flex-direction: column;
  }

  .tab-panel {
    flex: 1 1 0;
    min-height: 0;
    flex-direction: column;
  }
</style>
