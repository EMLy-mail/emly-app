<script lang="ts">
  import { onMount, untrack } from "svelte";
  import type { PageData } from "./$types";
  import {
    RotateCcw,
    RotateCw,
    ZoomIn,
    ZoomOut,
    AlignHorizontalSpaceAround,
    Download
  } from "@lucide/svelte";
  import { sidebarOpen } from "$lib/stores/app";
  import { toast } from "svelte-sonner";
  import * as m from "$lib/paraglide/messages.js";
  import * as pdfjsLib from "pdfjs-dist";
  import pdfWorker from "pdfjs-dist/build/pdf.worker.min.mjs?url";

  if (typeof Promise.withResolvers === "undefined") {
    // @ts-ignore
    Promise.withResolvers = function () {
      let resolve, reject;
      const promise = new Promise((res, rej) => {
        resolve = res;
        reject = rej;
      });
      return { promise, resolve, reject };
    };
  }

  // Set worker source
  pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker;

  let { data }: { data: PageData } = $props();

  let pdfData = $state<Uint8Array | null>(null);
  let filename = $state("");
  let rotation = $state(0);
  let scale = $state(1.5); // Default scale
  let error = $state("");
  let loading = $state(true);

  let pdfDoc = $state<pdfjsLib.PDFDocumentProxy | null>(null);
  let pageNum = $state(1);
  let totalPages = $state(0);
  let canvasRef = $state<HTMLCanvasElement>();
  let canvasContainerRef = $state<HTMLDivElement>();
  let renderTask = $state<pdfjsLib.RenderTask | null>(null);

  onMount(async () => {
    try {
      const result = data?.data;
      if (result) {
        // Decode Base64 to Uint8Array
        const binaryString = window.atob(result.data);
        const len = binaryString.length;
        const bytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        pdfData = bytes;
        filename = result.filename;
        // Adjust title
        document.title = filename + " - EMLy PDF Viewer";
        sidebarOpen.set(false);

        await loadPDF();
      } else {
        toast.error(m.pdf_error_no_data());
        error = m.pdf_error_no_data_desc();
        loading = false;
      }
    } catch (e) {
      error = "Failed to load PDF: " + e;
      loading = false;
    }
  });

  async function loadPDF() {
    if (!pdfData) return;

    // Set a timeout to prevent infinite loading
    const timeout = setTimeout(() => {
      if (loading) {
        loading = false;
        error = m.pdf_error_timeout();
        toast.error(error);
      }
    }, 10000);

    try {
      const loadingTask = pdfjsLib.getDocument({ data: pdfData });
      pdfDoc = await loadingTask.promise;
      totalPages = pdfDoc.numPages;
      pageNum = 1;
      await renderPage(pageNum);
      loading = false;
    } catch (e) {
      console.error(e);
      error = m.pdf_error_parsing() + e;
      loading = false;
    } finally {
      clearTimeout(timeout);
    }
  }

  async function renderPage(num: number) {
    if (!pdfDoc || !canvasRef) return;

    if (renderTask) {
      // Cancel previous render if any and await its cleanup
      renderTask.cancel();
      try {
        await renderTask.promise;
      } catch (e) {
        // Expected cancellation error
      }
    }

    try {
      const page = await pdfDoc.getPage(num);

      // Calculate scale if needed or use current scale
      // We apply rotation to the viewport
      const viewport = page.getViewport({ scale: scale, rotation: rotation });

      const canvas = canvasRef;
      const context = canvas.getContext("2d");

      if (!context) return;

      canvas.height = viewport.height;
      canvas.width = viewport.width;

      const renderContext = {
        canvasContext: context,
        viewport: viewport,
      };

      // Cast to any to avoid type mismatch with PDF.js definitions
      const task = page.render(renderContext as any);
      renderTask = task;
      await task.promise;
    } catch (e: any) {
      if (e.name !== "RenderingCancelledException") {
        console.error(e);
        toast.error(m.pdf_error_rendering() + e.message);
      }
    }
  }

  function fitToWidth() {
    if (!pdfDoc || !canvasContainerRef) return;
    // We need to fetch page to get dimensions
    loading = true;
    pdfDoc.getPage(pageNum).then((page) => {
      const containerWidth = canvasContainerRef!.clientWidth - 40; // padding
      const viewport = page.getViewport({ scale: 1, rotation: rotation });
      scale = containerWidth / viewport.width;
      renderPage(pageNum).then(() => {
        loading = false;
      });
    });
  }

  $effect(() => {
    // Re-render when scale or rotation changes
    // Access them here to ensure dependency tracking since renderPage is untracked
    // We also track pageNum to ensure we re-render if it changes via other means, 
    // although navigation functions usually call renderPage manually.
    const _deps = [scale, rotation, pageNum];

    if (pdfDoc) {
      // Untrack renderPage because it reads and writes to renderTask,
      // which would otherwise cause an infinite loop.
      untrack(() => renderPage(pageNum));
    }
  });

  function rotate(deg: number) {
    rotation = (rotation + deg + 360) % 360;
  }

  function zoom(factor: number) {
    scale = Math.max(0.1, Math.min(5.0, scale + factor));
  }

  function nextPage() {
    if (pageNum >= totalPages) return;
    pageNum++;
    renderPage(pageNum);
  }

  function prevPage() {
    if (pageNum <= 1) return;
    pageNum--;
    renderPage(pageNum);
  }

  function downloadPDF() {
    if (!pdfData) return;
    try {
      // @ts-ignore
      const blob = new Blob([pdfData], { type: "application/pdf" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename || "document.pdf";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (e) {
      toast.error("Failed to download PDF: " + e);
    }
  }
</script>

<div class="viewer-container">
  {#if loading}
    <div class="loading-overlay">
      <div class="spinner"></div>
      <div>{m.pdf_loading()}</div>
    </div>
  {/if}

  {#if error}
    <div class="error-overlay">
      <div class="error-message">{error}</div>
    </div>
  {/if}

  <div class="toolbar">
    <h1 class="title" title={filename}>{filename || m.pdf_viewer_title()}</h1>

    <div class="controls">
      <button class="btn" onclick={() => downloadPDF()} title={m.mail_download_btn_title()}>
        <Download size="16" />
      </button>
      <div class="separator"></div>
      <button class="btn" onclick={() => zoom(0.1)} title={m.pdf_zoom_in()}>
        <ZoomIn size="16" />
      </button>
      <button class="btn" onclick={() => zoom(-0.1)} title={m.pdf_zoom_out()}>
        <ZoomOut size="16" />
      </button>
      <div class="separator"></div>
      <button class="btn" onclick={() => rotate(-90)} title={m.pdf_rotate_left()}>
        <RotateCcw size="16" />
      </button>
      <button class="btn" onclick={() => rotate(90)} title={m.pdf_rotate_right()}>
        <RotateCw size="16" />
      </button>
      <div class="separator"></div>
      <button class="btn" onclick={fitToWidth} title={m.pdf_fit_width()}>
        <AlignHorizontalSpaceAround size="16" />
      </button>
    </div>
  </div>

  <div class="canvas-container" bind:this={canvasContainerRef}>
    <canvas bind:this={canvasRef}></canvas>
  </div>
</div>

<style>
  .viewer-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--background);
    position: relative;
    user-select: none;
    color: var(--foreground);
  }

  .loading-overlay {
    position: absolute;
    inset: 0;
    z-index: 50;
    display: flex;
    flex-direction: column; /* Match the requested style */
    gap: 10px;
    align-items: center;
    justify-content: center;
    background: var(--background);
    color: var(--foreground);
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

  .error-overlay {
    position: absolute;
    inset: 0;
    z-index: 40;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--background);
    color: var(--destructive);
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
  }

  .btn:hover {
    background: var(--accent);
    color: var(--accent-foreground);
  }

  .canvas-container {
    flex: 1;
    overflow: auto;
    display: flex;
    justify-content: center;
    align-items: flex-start; /* scroll from top */
    padding: 20px;
    background: var(--muted);
  }

  canvas {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    max-width: none; /* Allow canvas to be larger than container */
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
