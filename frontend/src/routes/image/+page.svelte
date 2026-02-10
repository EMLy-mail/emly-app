<script lang="ts">
  import { onMount } from "svelte";
  import type { PageData } from './$types';
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

  let { data }: { data: PageData } = $props();

  let imageData = $state("");
  let filename = $state("");
  let rotation = $state(0);
  let scale = $state(1);
  let error = $state("");
  let loading = $state(true);
  let translateX = $state(0);
  let translateY = $state(0);
  let imgElement = $state<HTMLImageElement>();
  let containerElement = $state<HTMLDivElement>();

  // Non-reactive state for drag calculations
  let isDragging = false;
  let startX = 0;
  let startY = 0;

  onMount(async () => {
    try {
      const result = data?.data
      if (result) {
        imageData = result.data;
        filename = result.filename;
        // Adjust title
        document.title = filename + " - EMLy Image Viewer";
        sidebarOpen.set(false);
      } else {
        toast.error("No image data provided");
        error = "No image data provided";
      }
    } catch (e) {
      error = "Failed to load image: " + e;
    } finally {
      loading = false;
    }
  });

  function fitToScreen() {
    if (!imgElement || !containerElement) return;

    const padding = 60; 
    const cw = containerElement.clientWidth - padding;
    const ch = containerElement.clientHeight - padding;
    const iw = imgElement.naturalWidth;
    const ih = imgElement.naturalHeight;

    if (!iw || !ih || !cw || !ch) return;

    const scaleW = cw / iw;
    const scaleH = ch / ih;

    scale = Math.min(scaleW, scaleH);
    // Ensure we don't end up with an invalid scale
    if (!Number.isFinite(scale) || scale <= 0) scale = 0.1;

    translateX = 0;
    translateY = 0;
  }

  function rotate(deg: number) {
    rotation += deg;
  }

  function zoom(factor: number) {
    scale = Math.max(0.01, scale + factor);
  }

  function reset() {
    rotation = 0;
    fitToScreen();
  }

  function downloadImage() {
    if (!imageData || !filename) return;

    const link = document.createElement("a");
    link.href = `data:image/png;base64,${imageData}`;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  function handleWheel(e: WheelEvent) {
    e.preventDefault();
    const delta = -e.deltaY * 0.001;
    scale = Math.max(0.01, Math.min(50, scale + delta));
  }

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return; // Only left click
    e.preventDefault();
    isDragging = true;
    startX = e.clientX - translateX;
    startY = e.clientY - translateY;
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isDragging) return;
    e.preventDefault();
    translateX = e.clientX - startX;
    translateY = e.clientY - startY;
  }

  function handleMouseUp() {
    isDragging = false;
  }
</script>

<div class="page-container">
  <!-- Toolbar -->
  <div class="toolbar">
    <h1 class="title" title={filename}>{filename || "Image Viewer"}</h1>

    <div class="controls">
       <button class="btn" onclick={() => downloadImage()} title="Download">
        <Download size="16" />
      </button>
      <div class="separator"></div>
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
      <button class="btn" onclick={reset} title="Reset">
        <AlignHorizontalSpaceAround size="16" />
      </button>
    </div>
  </div>

  <!-- Image Area -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    bind:this={containerElement}
    class="image-area"
    onwheel={handleWheel}
    onmousedown={handleMouseDown}
    onmousemove={handleMouseMove}
    onmouseup={handleMouseUp}
    onmouseleave={handleMouseUp}
    role="region"
    aria-label="Image View"
  >
    {#if loading}
      <div class="loading">Loading...</div>
    {:else if error}
      <div class="error-message">
        {error}
      </div>
    {:else if imageData}
      <div
        class="transform-layer"
        style="transform: translate({translateX}px, {translateY}px) scale({scale}) rotate({rotation}deg);"
      >
        <!-- svelte-ignore a11y_img_redundant_alt -->
        <img
          bind:this={imgElement}
          onload={fitToScreen}
          src={`data:image/png;base64,${imageData}`}
          alt={filename}
          class="viewer-img"
          draggable="false"
        />
      </div>
    {/if}
  </div>
</div>

<style>
  :global(body) {
    margin: 0;
    background: var(--background);
  }

  .page-container {
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--background);
    color: var(--foreground);
    font-family: system-ui, -apple-system, sans-serif;
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

  .image-area {
    flex: 1;
    background: var(--muted);
    position: relative;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: grab;
  }

  .image-area:active {
    cursor: grabbing;
  }

  .transform-layer {
    transition: transform 0.05s linear;
    transform-origin: center center;
    will-change: transform;
    display: flex; /* Ensures content centers */
  }

  .viewer-img {
    max-width: none;
    pointer-events: none;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    border-radius: 2px;
  }

  .loading {
    color: var(--muted-foreground);
    font-size: 14px;
  }

  .error-message {
    color: var(--destructive);
    background: var(--destructive-foreground);
    padding: 12px 16px;
    border-radius: 8px;
    border: 1px solid var(--destructive);
    font-size: 14px;
  }
</style>
