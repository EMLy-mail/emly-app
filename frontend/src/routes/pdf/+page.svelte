<script lang="ts">
  import { onMount } from "svelte";
  import type { PageData } from "./$types";
  import {
    RotateCcw,
    RotateCw,
    ZoomIn,
    ZoomOut,
    AlignHorizontalSpaceAround,
  } from "@lucide/svelte";
  import { sidebarOpen } from "$lib/stores/app";
  import { toast } from "svelte-sonner";
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
        toast.error("No PDF data provided");
        error =
          "No PDF data provided. Please open this window from the main EMLy application.";
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
        error =
          "Timeout loading PDF. The worker might have failed to initialize.";
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
      error = "Error parsing PDF: " + e;
      loading = false;
    } finally {
      clearTimeout(timeout);
    }
  }

  async function renderPage(num: number) {
    if (!pdfDoc || !canvasRef) return;

    if (renderTask) {
      await renderTask.promise.catch(() => {}); // Cancel previous render if any (though we wait usually)
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
      await page.render(renderContext as any).promise;
    } catch (e: any) {
      if (e.name !== "RenderingCancelledException") {
        console.error(e);
        toast.error("Error rendering page: " + e.message);
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
    // Access them here to ensure dependency tracking since renderPage is async
    const _deps = [scale, rotation];

    if (pdfDoc) {
      renderPage(pageNum);
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
</script>

<div class="viewer-container">
  {#if loading}
    <div class="loading-overlay">
      <div class="spinner"></div>
      <div>Loading PDF...</div>
    </div>
  {/if}

  {#if error}
    <div class="error-overlay">
      <div class="error-message">{error}</div>
    </div>
  {/if}

  <div class="toolbar">
    <h1 class="title" title={filename}>{filename || "Image Viewer"}</h1>

    <div class="controls">
      <button class="btn" onclick={() => zoom(0.1)} title="Zoom In">
        <ZoomIn size="16" />
      </button>
      <button class="btn" onclick={() => zoom(-0.1)} title="Zoom Out">
        <ZoomOut size="16" />
      </button>
      <div class="separator"></div>
      <button class="btn" onclick={() => rotate(-90)} title="Rotate Left">
        <RotateCcw size="16" />
      </button>
      <button class="btn" onclick={() => rotate(90)} title="Rotate Right">
        <RotateCw size="16" />
      </button>
      <div class="separator"></div>
      <button class="btn" onclick={fitToWidth} title="Reset">
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
    background: #1e1e1e;
    position: relative;
    user-select: none;
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
    background: #111;
    color: white;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 2px solid rgba(255, 255, 255, 0.1);
    border-top-color: rgba(255, 255, 255, 0.8);
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
    background: #111;
    color: #ef4444;
  }

  .toolbar {
    height: 50px;
    background: #000;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
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
    background: rgba(255, 255, 255, 0.15);
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
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.06);
    color: rgba(255, 255, 255, 0.85);
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
  }

  .separator {
    width: 1px;
    height: 24px;
    background: rgba(255, 255, 255, 0.1);
  }

  .canvas-container {
    flex: 1;
    overflow: auto;
    display: flex;
    justify-content: center;
    align-items: flex-start; /* scroll from top */
    padding: 20px;
    background: #333; /* Dark background for contrast */
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
    background: rgba(255, 255, 255, 0.1);
    border-radius: 6px;
  }

  ::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  ::-webkit-scrollbar-corner {
    background: transparent;
  }
</style>
