<script lang="ts">
    import { mailState } from "$lib/stores/mail-state.svelte";
    import type { AppTab } from "$lib/stores/mail-state.svelte";
    import * as m from "$lib/paraglide/messages.js";
    import {
        X,
        Plus,
        Mail,
        FileText,
        Image,
        ChevronLeft,
        ChevronRight,
    } from "@lucide/svelte";
    import * as ContextMenu from "$lib/components/ui/context-menu/index.js";

    let { onAddTab, isAddingTab }: { onAddTab: () => void; isAddingTab: boolean } =
        $props();

    let tabStripEl: HTMLDivElement | null = $state(null);
    let canScrollLeft = $state(false);
    let canScrollRight = $state(false);
    let hasOverflow = $state(false);

    function updateScrollState() {
        if (!tabStripEl) return;
        hasOverflow = tabStripEl.scrollWidth > tabStripEl.clientWidth + 1;
        canScrollLeft = tabStripEl.scrollLeft > 0;
        canScrollRight =
            tabStripEl.scrollLeft <
            tabStripEl.scrollWidth - tabStripEl.clientWidth - 1;
    }

    function scrollTabs(direction: number) {
        if (!tabStripEl) return;
        tabStripEl.scrollBy({ left: direction * 160, behavior: "smooth" });
    }

    function handleTabStripWheel(e: WheelEvent) {
        if (!tabStripEl) return;
        e.preventDefault();
        const delta =
            Math.abs(e.deltaY) > Math.abs(e.deltaX) ? e.deltaY : e.deltaX;
        tabStripEl.scrollBy({ left: delta, behavior: "auto" });
    }

    $effect(() => {
        mailState.tabs.length;
        setTimeout(updateScrollState, 0);
    });

    function getTabLabel(tab: AppTab): string {
        if (tab.type === "email") {
            const s = tab.email.subject || m.mail_subject_no_subject();
            return s.length > 24 ? s.slice(0, 24) + "…" : s;
        }
        const s = tab.filename;
        return s.length > 24 ? s.slice(0, 24) + "…" : s;
    }

    function closeTab(id: string, e: MouseEvent) {
        e.stopPropagation();
        mailState.removeTab(id);
    }

    function handleGlobalKeydown(e: KeyboardEvent) {
        if (e.ctrlKey && e.key === "w") {
            e.preventDefault();
            const activeTab = mailState.tabs.find(
                (tab) => tab.id === mailState.activeTabId
            );
            if (activeTab) mailState.removeTab(activeTab.id);
        } else if (e.ctrlKey && e.key === "Tab") {
            e.preventDefault();
            const currentIndex = mailState.tabs.findIndex(
                (tab) => tab.id === mailState.activeTabId
            );
            if (currentIndex !== -1) {
                const nextIndex = (currentIndex + 1) % mailState.tabs.length;
                mailState.setActiveTab(mailState.tabs[nextIndex].id);
            }
        }
    }
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<div class="tab-strip-wrapper">
    <!-- Left scroll arrow -->
    {#if hasOverflow}
        <button
            class="tab-scroll-btn"
            class:inactive={!canScrollLeft}
            onclick={() => scrollTabs(-1)}
            aria-label={m.tabs_scroll_left()}
            tabindex={-1}
        >
            <ChevronLeft size="13" strokeWidth={2.5} />
        </button>
    {/if}

    <!-- Scrollable tabs area -->
    <div
        class="tab-strip"
        role="tablist"
        tabindex="0"
        bind:this={tabStripEl}
        onscroll={updateScrollState}
        onwheel={handleTabStripWheel}
    >
        {#each mailState.tabs as tab, i (tab.id)}
            {@const isActive = tab.id === mailState.activeTabId}
            <ContextMenu.Root>
                <ContextMenu.Trigger
                    class={isActive ? "tab-item active" : "tab-item"}
                    role="tab"
                    aria-selected={isActive}
                    onclick={() => mailState.setActiveTab(tab.id)}
                    onkeydown={(e: KeyboardEvent) =>
                        e.key === "Enter" && mailState.setActiveTab(tab.id)}
                >
                    <span class="tab-icon">
                        {#if tab.type === "pdf"}
                            <FileText size="11" strokeWidth={2} />
                        {:else if tab.type === "image"}
                            <Image size="11" strokeWidth={2} />
                        {:else}
                            <Mail size="11" strokeWidth={2} />
                        {/if}
                    </span>
                    <span class="tab-label">{getTabLabel(tab)}</span>
                    <button
                        class="tab-close"
                        tabindex={-1}
                        aria-label={m.tabs_close_tab_label()}
                        onclick={(e) => closeTab(tab.id, e)}
                    >
                        <X size="11" strokeWidth={2.5} />
                    </button>
                </ContextMenu.Trigger>
                <ContextMenu.Content>
                    <ContextMenu.Item
                        onclick={() => mailState.removeTab(tab.id)}
                    >
                        {m.tabs_context_close()}
                    </ContextMenu.Item>
                    <ContextMenu.Item
                        disabled={mailState.tabs.length <= 1}
                        onclick={() => mailState.closeOtherTabs(tab.id)}
                    >
                        {m.tabs_context_close_others()}
                    </ContextMenu.Item>
                    <ContextMenu.Separator />
                    <ContextMenu.Item
                        disabled={i === 0}
                        onclick={() => mailState.closeTabsToLeft(tab.id)}
                    >
                        {m.tabs_context_close_left()}
                    </ContextMenu.Item>
                    <ContextMenu.Item
                        disabled={i === mailState.tabs.length - 1}
                        onclick={() => mailState.closeTabsToRight(tab.id)}
                    >
                        {m.tabs_context_close_right()}
                    </ContextMenu.Item>
                    <ContextMenu.Separator />
                    <ContextMenu.Item onclick={() => mailState.closeAllTabs()}>
                        {m.tabs_context_close_all()}
                    </ContextMenu.Item>
                </ContextMenu.Content>
            </ContextMenu.Root>
        {/each}

        <!-- Add tab button — inside scroll area, right after the last tab -->
        <button
            class="tab-add"
            onclick={onAddTab}
            disabled={isAddingTab}
            aria-label={m.tabs_new_tab()}
            title={m.tabs_new_tab()}
        >
            <Plus size="14" strokeWidth={2} />
        </button>
    </div>

    <!-- Right scroll arrow -->
    {#if hasOverflow}
        <button
            class="tab-scroll-btn"
            class:inactive={!canScrollRight}
            onclick={() => scrollTabs(1)}
            aria-label={m.tabs_scroll_right()}
            tabindex={-1}
        >
            <ChevronRight size="13" strokeWidth={2.5} />
        </button>
    {/if}
</div>

<style>
    /* ── Tab strip wrapper (the dark chrome bar) ── */
    .tab-strip-wrapper {
        flex-shrink: 0;
        display: flex;
        align-items: center;
        gap: 2px;
        padding: 6px 6px 0 6px;
        background: color-mix(in srgb, var(--background) 80%, var(--card) 20%);
        border-bottom: 1px solid var(--border);
    }

    /* ── Scrollable tabs area ── */
    .tab-strip {
        flex: 1 1 0;
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 2px;
        overflow-x: auto;
        overflow-y: hidden;
        scrollbar-width: none;
    }

    .tab-strip::-webkit-scrollbar {
        display: none;
    }

    /* ── Scroll arrows ── */
    .tab-scroll-btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 22px;
        height: 22px;
        border-radius: 5px;
        border: none;
        background: transparent;
        color: var(--foreground);
        cursor: pointer;
        flex-shrink: 0;
        transition:
            background 0.1s,
            opacity 0.15s;
        opacity: 0.75;
    }

    .tab-scroll-btn:hover {
        background: var(--muted);
        opacity: 1;
    }

    .tab-scroll-btn.inactive {
        opacity: 0.22;
        cursor: default;
        pointer-events: none;
    }

    /* ── Individual tab ── */
    /* :global() because .tab-item is applied via a prop to <ContextMenu.Trigger>,
       whose rendered element lives outside this component's scoped-CSS template */
    :global(.tab-item) {
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

    :global(.tab-item):hover {
        background: color-mix(in srgb, var(--muted) 70%, transparent);
        color: var(--foreground);
    }

    :global(.tab-item.active) {
        background: var(--card);
        border-color: var(--border);
        color: var(--foreground);
        z-index: 1;
    }

    .tab-icon {
        display: inline-flex;
        align-items: center;
        flex-shrink: 0;
        opacity: 0.6;
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

    :global(.tab-item):hover .tab-close,
    :global(.tab-item.active) .tab-close {
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
</style>
