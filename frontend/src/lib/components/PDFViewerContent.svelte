<script lang="ts">
  import {
    RotateCcw,
    RotateCw,
    ZoomIn,
    ZoomOut,
    AlignHorizontalSpaceAround,
    Download,
    ChevronLeft,
    ChevronRight,
  } from "@lucide/svelte";
  import { useZoom, ZoomMode } from "@embedpdf/plugin-zoom/svelte";
  import { useScroll, Scroller } from "@embedpdf/plugin-scroll/svelte";
  import type { PageLayout } from "@embedpdf/plugin-scroll/svelte";
  import { Viewport } from "@embedpdf/plugin-viewport/svelte";
  import { DocumentContent } from "@embedpdf/plugin-document-manager/svelte";
  import { RenderLayer } from "@embedpdf/plugin-render/svelte";
  import { useRotate } from "@embedpdf/plugin-rotate/svelte";
  import * as m from "$lib/paraglide/messages.js";
  import { Rotate } from "@embedpdf/plugin-rotate/svelte";

  let readerCSSStylesheet = `height: 100%; width: 100%; overflow: auto; background: var(--muted); ::-webkit-scrollbar{width:10px;height:10px}::-webkit-scrollbar-track{background:transparent}::-webkit-scrollbar-thumb{background:var(--border);border-radius:6px}::-webkit-scrollbar-thumb:hover{background:var(--muted-foreground)}::-webkit-scrollbar-corner{background:transparent}`;

  interface Props {
    documentId: string;
    filename?: string;
    src: string;
  }

  let { documentId, filename = "", src }: Props = $props();

  const zoom = useZoom(() => documentId);
  const scroll = useScroll(() => documentId);
  const rotate = useRotate(() => documentId);
</script>

{#snippet renderPage(page: PageLayout)}
  <div
    style:width="{page.width}px"
    style:height="{page.height}px"
    style:position="relative"
  >
    <Rotate {documentId} pageIndex={page.pageIndex}>
      <RenderLayer {documentId} pageIndex={page.pageIndex} />
    </Rotate>
  </div>
{/snippet}

<div class="viewer-container">
  <div class="toolbar">
    <h1 class="title" title={filename}>{filename || m.pdf_viewer_title()}</h1>

    <div class="controls">
      <a
        class="btn"
        href={src}
        download={filename || "document.pdf"}
        title={m.mail_download_btn_title()}
      >
        <Download size="16" />
      </a>
      <div class="separator"></div>
      <button
        class="btn"
        onclick={() => scroll.provides?.scrollToPreviousPage()}
        disabled={scroll.state.currentPage <= 1}
        title={m.pdf_prev_page()}
      >
        <ChevronLeft size="16" />
      </button>
      <span class="page-info"
        >{scroll.state.currentPage} / {scroll.state.totalPages}</span
      >
      <button
        class="btn"
        onclick={() => scroll.provides?.scrollToNextPage()}
        disabled={scroll.state.currentPage >= scroll.state.totalPages}
        title={m.pdf_next_page()}
      >
        <ChevronRight size="16" />
      </button>
      <div class="separator"></div>
      <button
        class="btn"
        onclick={() => zoom.provides?.zoomIn()}
        title={m.pdf_zoom_in()}
      >
        <ZoomIn size="16" />
      </button>
      <button
        class="btn"
        onclick={() => zoom.provides?.zoomOut()}
        title={m.pdf_zoom_out()}
      >
        <ZoomOut size="16" />
      </button>
      <div class="separator"></div>
      <button
        class="btn"
        onclick={() => rotate.provides?.rotateBackward()}
        title={m.pdf_rotate_left()}
      >
        <RotateCcw size="16" />
      </button>
      <button
        class="btn"
        onclick={() => rotate.provides?.rotateForward()}
        title={m.pdf_rotate_right()}
      >
        <RotateCw size="16" />
      </button>
      <div class="separator"></div>
      <button
        class="btn"
        onclick={() => zoom.provides?.requestZoom(ZoomMode.FitWidth)}
        title={m.pdf_fit_width()}
      >
        <AlignHorizontalSpaceAround size="16" />
      </button>
    </div>
  </div>

  <div class="content-area">
    <DocumentContent {documentId}>
      {#snippet children({ isLoading, isLoaded, isError })}
        {#if isLoading}
          <div class="state-overlay">
            <div class="spinner"></div>
            <div>{m.pdf_loading()}</div>
          </div>
        {:else if isError}
          <div class="state-overlay error">
            <div>{m.pdf_error_loading("")}</div>
          </div>
        {:else if isLoaded}
          <Viewport {documentId} style={readerCSSStylesheet}>
            <Scroller {documentId} {renderPage} />
          </Viewport>
        {/if}
      {/snippet}
    </DocumentContent>
  </div>
</div>

<style>
  .viewer-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--background);
    color: var(--foreground);
    user-select: none;
  }

  .toolbar {
    height: 50px;
    background: var(--card);
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    flex-shrink: 0;
    z-index: 10;
  }

  .title {
    font-size: 14px;
    font-weight: 500;
    opacity: 0.9;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 50%;
  }

  .controls {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .separator {
    width: 1px;
    height: 18px;
    background: var(--border);
    margin: 0 4px;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    padding: 0;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--muted);
    color: var(--foreground);
    cursor: pointer;
    transition: all 0.2s;
    text-decoration: none;
  }

  .btn:hover {
    background: var(--accent);
    color: var(--accent-foreground);
  }

  .btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .btn:disabled:hover {
    background: var(--muted);
    color: var(--foreground);
  }

  .page-info {
    font-size: 13px;
    min-width: 48px;
    text-align: center;
    color: var(--foreground);
    opacity: 0.8;
  }

  .content-area {
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .state-overlay {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 10px;
    background: var(--background);
    color: var(--foreground);
  }

  .state-overlay.error {
    color: var(--destructive);
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

  ::-webkit-scrollbar {
    width: 10px;
    height: 10px;
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
</style>
