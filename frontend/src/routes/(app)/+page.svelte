<script lang="ts">
  import MailViewer from "$lib/components/MailViewer.svelte";
  import { mailState } from "$lib/stores/mail-state.svelte";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import { sidebarOpen } from "$lib/stores/app";
  import * as m from "$lib/paraglide/messages.js";
  import { toast } from "svelte-sonner";
  import { X, Plus } from "@lucide/svelte";
  import { openAndLoadEmail } from "$lib/utils/mail";
  import { onDestroy, onMount } from "svelte";

  let { data } = $props();

  let isAddingTab = $state(false);

  onMount(() => {
    if (data.email) {
      if (settingsStore.settings.enableTabMode) {
        mailState.addTab(data.email);
        sidebarOpen.set(false);
      } else {
        mailState.setParams(data.email);
      }
    } else if (data.loadError) {
      toast.error(m.mail_error_opening());
    }
  });

  onDestroy(() => {
    if (!settingsStore.settings.enableTabMode) {
      mailState.getAllTabs().forEach((tab) => {
        if (tab.id !== mailState.getActiveTabId()) {
          mailState.removeTab(tab.id);
        }
      });
    }
  });

  function truncateSubject(subject: string | undefined): string {
    const s = subject || m.mail_subject_no_subject();
    return s.length > 24 ? s.slice(0, 24) + "…" : s;
  }

  function closeTab(id: string, e: MouseEvent) {
    e.stopPropagation();
    mailState.removeTab(id);
  }

  async function openNewTab() {
    if (isAddingTab) return;
    isAddingTab = true;

    const result = await openAndLoadEmail();

    if (!result.cancelled && result.success && result.email) {
      mailState.addTab(result.email);
      sidebarOpen.set(false);
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
      <div class="tabbed-panel">
        <!-- Tab strip -->
        <div class="tab-strip" role="tablist">
          {#each mailState.tabs as tab (tab.id)}
            {@const isActive = tab.id === mailState.activeTabId}
            <!-- svelte-ignore a11y_interactive_supports_focus -->
            <div
              class="tab-item"
              class:active={isActive}
              role="tab"
              aria-selected={isActive}
              onclick={() => mailState.setActiveTab(tab.id)}
              onkeydown={(e) =>
                e.key === "Enter" && mailState.setActiveTab(tab.id)}
            >
              <span class="tab-label">{truncateSubject(tab.email.subject)}</span
              >
              <button
                class="tab-close"
                tabindex={-1}
                aria-label={m.tabs_close_tab_label()}
                onclick={(e) => closeTab(tab.id, e)}
              >
                <X size="11" strokeWidth={2.5} />
              </button>
            </div>
          {/each}

          <button
            class="tab-add"
            onclick={openNewTab}
            disabled={isAddingTab}
            aria-label="Apri nuova scheda"
            title="Apri nuova scheda"
          >
            <Plus size="14" strokeWidth={2} />
          </button>
        </div>

        <!-- Tab content panels - all mounted, shown/hidden via display -->
        <div class="tab-content-area">
          {#each mailState.tabs as tab (tab.id)}
            <div
              class="tab-panel"
              role="tabpanel"
              style:display={tab.id === mailState.activeTabId ? "flex" : "none"}
            >
              <MailViewer
                emailData={tab.email}
                tabId={tab.id}
                embedded={true}
              />
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

  /* ── Tab strip (the dark chrome bar) ── */
  .tab-strip {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 6px 8px 0 8px;
    background: color-mix(in srgb, var(--background) 80%, var(--card) 20%);
    border-bottom: 1px solid var(--border);
    overflow-x: auto;
    overflow-y: visible;
    scrollbar-width: none;
  }

  .tab-strip::-webkit-scrollbar {
    display: none;
  }

  /* ── Individual tab ── */
  .tab-item {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 30px;
    padding: 0 6px 0 10px;
    border-radius: 8px 8px 0 0;
    border: 1px solid transparent;
    border-bottom: none;
    background: transparent;
    color: var(--muted-foreground);
    font-size: 12px;
    font-weight: 500;
    white-space: nowrap;
    cursor: pointer;
    user-select: none;
    min-width: 80px;
    max-width: 200px;
    transition:
      background 0.1s,
      color 0.1s;
    /* extend 1px down to cover the strip's bottom border when active */
    margin-bottom: -1px;
    padding-bottom: 1px;
  }

  .tab-item:hover {
    background: color-mix(in srgb, var(--muted) 70%, transparent);
    color: var(--foreground);
  }

  .tab-item.active {
    background: var(--card);
    border-color: var(--border);
    color: var(--foreground);
    z-index: 1;
  }

  .tab-label {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tab-close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border-radius: 4px;
    flex-shrink: 0;
    opacity: 0.45;
    transition:
      opacity 0.1s,
      background 0.1s,
      color 0.1s;
  }

  .tab-item:hover .tab-close,
  .tab-item.active .tab-close {
    opacity: 0.7;
  }

  .tab-close:hover {
    opacity: 1 !important;
    background: var(--destructive);
    color: #fff;
    border-radius: 4px;
  }

  /* ── Add tab button ── */
  .tab-add {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 6px;
    border: none;
    background: transparent;
    color: var(--muted-foreground);
    cursor: pointer;
    flex-shrink: 0;
    margin-left: 2px;
    transition:
      background 0.1s,
      color 0.1s;
  }

  .tab-add:hover:not(:disabled) {
    background: var(--muted);
    color: var(--foreground);
  }

  .tab-add:disabled {
    opacity: 0.4;
    cursor: not-allowed;
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
