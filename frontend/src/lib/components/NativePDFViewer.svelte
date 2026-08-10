<!--
  Alternative PDF viewer built on the WebView2/Edge built-in PDF plugin.

  Unlike PdfjsViewer.svelte (which rasterizes pages with pdf.js), this one
  hands the document straight to Chromium's own PDF viewer through an
  <iframe type="application/pdf">. It has two modes, selected by the
  `useNativePdfToolbar` setting:

    - Native toolbar: Edge's own PDF chrome is left visible and does
      everything (zoom, rotate, page nav, search, print). A single static
      iframe, no state of ours involved.

    - EMLy toolbar: Edge's chrome is suppressed with `#toolbar=0&navpanes=0`
      and we drive the view ourselves — zoom through the `#zoom=` fragment,
      rotate by rewriting the document with pdf-lib.

  Rotation is view-only: the bytes are re-derived from the *original*
  document every time, and the file on disk is never touched.
-->
<script lang="ts">
  import { onDestroy, untrack } from "svelte";
  import { PDFDocument, degrees } from "pdf-lib";
  import { RotateCcw, RotateCw, ZoomIn, ZoomOut, Download } from "@lucide/svelte";
  import { saveAttachmentNatively } from "$lib/utils/attachment-download";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import * as m from "$lib/paraglide/messages.js";
  import "$lib/styles/viewer-toolbar.css";

  interface Props {
    src: string;
    filename?: string;
    height?: string;
    base64Data?: string;
  }

  let { src, filename = "", height = "100%", base64Data = "" }: Props = $props();

  const ZOOM_MIN = 25;
  const ZOOM_MAX = 400;
  const ZOOM_STEP = 25;

  /*
   * The PDF plugin needs a beat after `load` before it has painted the first
   * page. Swapping the moment `load` fires shows a white frame instead.
   */
  const PAINT_SETTLE_MS = 150;

  let nativeToolbar = $derived(!!settingsStore.settings.useNativePdfToolbar);

  let zoom = $state(100);
  let rotation = $state(0);
  let rotatedUrl = $state<string | null>(null);
  let busy = $state(false);
  let error = $state("");

  let originalBytes: Uint8Array | null = null;
  /* Rotated blob URLs replaced by a newer one, freed once they leave view. */
  let staleUrls: string[] = [];

  let displaySrc = $derived(
    `${rotatedUrl ?? src}#zoom=${zoom}&toolbar=0&navpanes=0`,
  );

  /*
   * Double buffer. Two stacked iframes, only one visible: a new URL always
   * loads into the hidden one, and they trade places only once it has
   * actually painted. Without this the view goes blank for the whole reload.
   */
  let srcA = $state("");
  let srcB = $state("");
  let frontIsA = $state(true);
  let hasFirstPaint = $state(false);
  let pendingUrl: string | null = null;
  let swapTimer: ReturnType<typeof setTimeout> | null = null;

  $effect(() => {
    const url = displaySrc;
    const native = nativeToolbar;

    /*
     * Only displaySrc/nativeToolbar may retrigger this. The buffer state is
     * read *and* written here, so tracking it too would feed back on itself.
     */
    untrack(() => {
      if (native || !url) return;

      if (!hasFirstPaint) {
        srcA = url;
        frontIsA = true;
        return;
      }
      if (url === (frontIsA ? srcA : srcB)) return;

      if (frontIsA) srcB = url;
      else srcA = url;
      pendingUrl = url;

      // A newer target supersedes a swap that was already queued.
      if (swapTimer) {
        clearTimeout(swapTimer);
        swapTimer = null;
      }
    });
  });

  function handleFrameLoad(isA: boolean) {
    if (!hasFirstPaint) {
      if (isA) hasFirstPaint = true;
      return;
    }
    // Ignore the visible buffer reloading, and any load we no longer want.
    if (isA === frontIsA) return;
    if (!pendingUrl || (isA ? srcA : srcB) !== pendingUrl) return;

    if (swapTimer) clearTimeout(swapTimer);
    swapTimer = setTimeout(() => {
      frontIsA = !frontIsA;
      pendingUrl = null;
      swapTimer = null;
      releaseStaleUrls();
    }, PAINT_SETTLE_MS);
  }

  function releaseStaleUrls() {
    for (const url of staleUrls) URL.revokeObjectURL(url);
    staleUrls = [];
  }

  async function ensureOriginalBytes(): Promise<Uint8Array> {
    if (originalBytes) return originalBytes;

    if (base64Data) {
      const binary = atob(base64Data);
      const bytes = new Uint8Array(binary.length);
      for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
      originalBytes = bytes;
    } else {
      const response = await fetch(src);
      originalBytes = new Uint8Array(await response.arrayBuffer());
    }
    return originalBytes;
  }

  async function rotateBy(delta: number) {
    const next = (((rotation + delta) % 360) + 360) % 360;
    try {
      busy = true;
      error = "";

      // Always re-derived from the original, so angles never compound.
      const doc = await PDFDocument.load(await ensureOriginalBytes());
      for (const page of doc.getPages()) page.setRotation(degrees(next));
      const out = await doc.save();

      const blob = new Blob([new Uint8Array(out)], { type: "application/pdf" });
      const previous = rotatedUrl;
      rotatedUrl = URL.createObjectURL(blob);
      rotation = next;
      // Freed on the next swap: the old frame may still be reading it.
      if (previous) staleUrls.push(previous);
    } catch (e) {
      error = m.pdf_error_loading() + e;
    } finally {
      busy = false;
    }
  }

  function zoomBy(delta: number) {
    zoom = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, zoom + delta));
  }

  function downloadPdf() {
    if (!base64Data) return;
    void saveAttachmentNatively(base64Data, filename || "document.pdf");
  }

  onDestroy(() => {
    if (swapTimer) clearTimeout(swapTimer);
    if (rotatedUrl) URL.revokeObjectURL(rotatedUrl);
    releaseStaleUrls();
  });
</script>

<div class="viewer-container" style:height>
  <div class="toolbar">
    <h1 class="title" title={filename}>{filename || m.pdf_viewer_title()}</h1>

    <div class="controls">
      <button
        class="btn"
        onclick={downloadPdf}
        disabled={!base64Data}
        title={m.mail_download_btn_title()}
      >
        <Download size="16" />
      </button>

      <!-- Edge's own toolbar already offers these; duplicating them would
           only desync our state from the plugin's. -->
      {#if !nativeToolbar}
        <div class="separator"></div>
        <button
          class="btn"
          onclick={() => zoomBy(-ZOOM_STEP)}
          disabled={zoom <= ZOOM_MIN}
          title={m.pdf_zoom_out()}
        >
          <ZoomOut size="16" />
        </button>
        <span class="zoom-info">{zoom}%</span>
        <button
          class="btn"
          onclick={() => zoomBy(ZOOM_STEP)}
          disabled={zoom >= ZOOM_MAX}
          title={m.pdf_zoom_in()}
        >
          <ZoomIn size="16" />
        </button>
        <div class="separator"></div>
        <button
          class="btn"
          onclick={() => rotateBy(-90)}
          disabled={busy}
          title={m.pdf_rotate_left()}
        >
          <RotateCcw size="16" />
        </button>
        <button
          class="btn"
          onclick={() => rotateBy(90)}
          disabled={busy}
          title={m.pdf_rotate_right()}
        >
          <RotateCw size="16" />
        </button>
      {/if}
    </div>
  </div>

  <div class="content-area">
    {#if nativeToolbar}
      <iframe
        class="pdf-frame front"
        src="{src}#toolbar=1"
        title={filename || m.pdf_viewer_title()}
      ></iframe>
    {:else}
      <!--
        Each frame gets its own {#key}: only the *hidden* buffer's URL ever
        changes, so only that element is torn down and rebuilt. That is
        required, not incidental — swapping just the hash on a live frame is a
        same-document navigation, which the PDF plugin answers by ignoring the
        new `#zoom=` outright. Rebuilding the hidden frame is free; the
        visible one is never keyed away mid-view.
      -->
      {#key srcA}
        {#if srcA}
          <iframe
            class="pdf-frame"
            class:front={frontIsA}
            src={srcA}
            title={filename || m.pdf_viewer_title()}
            onload={() => handleFrameLoad(true)}
          ></iframe>
        {/if}
      {/key}
      {#key srcB}
        {#if srcB}
          <iframe
            class="pdf-frame"
            class:front={!frontIsA}
            src={srcB}
            title={filename || m.pdf_viewer_title()}
            onload={() => handleFrameLoad(false)}
          ></iframe>
        {/if}
      {/key}
    {/if}

    {#if error}
      <div class="error-overlay">{error}</div>
    {/if}
  </div>
</div>

<style>
  .viewer-container {
    display: flex;
    flex-direction: column;
    width: 100%;
    background: var(--background);
    color: var(--foreground);
    user-select: none;
  }

  .zoom-info {
    font-size: 13px;
    min-width: 44px;
    text-align: center;
    color: var(--foreground);
    opacity: 0.8;
  }

  .content-area {
    flex: 1;
    overflow: hidden;
    position: relative;
    background: var(--muted);
  }

  .pdf-frame {
    position: absolute;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    border: none;
    opacity: 0;
    pointer-events: none;
    z-index: 0;
  }

  .pdf-frame.front {
    opacity: 1;
    pointer-events: auto;
    z-index: 1;
  }

  .error-overlay {
    position: absolute;
    inset: auto 0 0 0;
    padding: 10px 16px;
    background: var(--card);
    border-top: 1px solid var(--border);
    color: var(--destructive);
    font-size: 13px;
    z-index: 2;
  }
</style>
