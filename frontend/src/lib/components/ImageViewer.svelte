<script lang="ts">
  import { onMount } from "svelte";
  import {
    RotateCcw,
    RotateCw,
    ZoomIn,
    ZoomOut,
    AlignHorizontalSpaceAround,
    Download,
    Loader2,
  } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages.js";
  import { saveAttachmentNatively } from "$lib/utils/attachment-download";
  import { toDisplayableImageSrc, mimeTypeForFilename } from "$lib/utils/image-decode";
  import "$lib/styles/viewer-toolbar.css";

  interface Props {
    base64Data: string;
    filename: string;
  }

  let { base64Data, filename }: Props = $props();

  let rotation = $state(0);
  let scale = $state(1);
  let translateX = $state(0);
  let translateY = $state(0);
  let imgElement = $state<HTMLImageElement>();
  let containerElement = $state<HTMLDivElement>();

  let displaySrc = $state("");
  let decoding = $state(true);
  let decodeError = $state("");
  /** Set only for HEIC/HEIF: the converted JPEG to offer on download,
   *  since most systems can't open the original format directly. */
  let downloadOverride: { base64: string; filename: string } | undefined;

  let isDragging = false;
  let startX = 0;
  let startY = 0;

  onMount(async () => {
    // Convert to a displayable <img> src: a fast-path data: URI for
    // regular formats, or a decoded JPEG for HEIC/HEIF (which the browser
    // can't render natively).
    try {
      const result = await toDisplayableImageSrc(
        base64Data,
        filename,
        mimeTypeForFilename(filename),
      );
      displaySrc = result.src;
      downloadOverride = result.download;
    } catch (e) {
      decodeError = m.image_error_loading() + String(e);
    } finally {
      decoding = false;
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
    if (!Number.isFinite(scale) || scale <= 0) scale = 0.1;
    translateX = 0;
    translateY = 0;
  }

  function rotate(deg: number) {
    rotation += deg;
    clampTranslate();
  }

  // Keep translate anchored to the container center as scale changes,
  // otherwise the image drifts based on whatever pan offset was already applied.
  function zoomTo(newScaleRaw: number) {
    const newScale = Math.max(0.01, Math.min(50, newScaleRaw));
    const ratio = newScale / scale;
    translateX *= ratio;
    translateY *= ratio;
    scale = newScale;
    clampTranslate();
  }

  function zoom(factor: number) {
    zoomTo(scale + factor);
  }

  function reset() {
    rotation = 0;
    fitToScreen();
  }

  function downloadImage() {
    if (downloadOverride) {
      void saveAttachmentNatively(downloadOverride.base64, downloadOverride.filename);
      return;
    }
    if (!base64Data || !filename) return;
    void saveAttachmentNatively(base64Data, filename);
  }

  function handleWheel(e: WheelEvent) {
    e.preventDefault();
    const delta = -e.deltaY * 0.001;
    zoomTo(scale + delta);
  }

  function clampTranslate() {
    if (!imgElement || !containerElement) return;
    const rotated = (((rotation % 180) + 180) % 180) !== 0;
    const iw = rotated ? imgElement.naturalHeight : imgElement.naturalWidth;
    const ih = rotated ? imgElement.naturalWidth : imgElement.naturalHeight;
    const scaledW = iw * scale;
    const scaledH = ih * scale;
    const cw = containerElement.clientWidth;
    const ch = containerElement.clientHeight;
    const maxX = Math.max(0, (scaledW - cw) / 2);
    const maxY = Math.max(0, (scaledH - ch) / 2);
    translateX = Math.min(maxX, Math.max(-maxX, translateX));
    translateY = Math.min(maxY, Math.max(-maxY, translateY));
  }

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
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
    clampTranslate();
  }

  function handleMouseUp() {
    isDragging = false;
  }
</script>

<div class="viewer">
  <div class="toolbar">
    <h1 class="title" title={filename}>{filename || m.image_viewer_title()}</h1>
    <div class="controls">
      <button class="btn" onclick={downloadImage} title={m.mail_download_btn_title()}>
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
      <button class="btn" onclick={reset} title={m.image_reset_btn()}>
        <AlignHorizontalSpaceAround size="16" />
      </button>
    </div>
  </div>

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
    aria-label={m.image_viewer_area_label()}
  >
    {#if decoding}
      <div class="loading">
        <Loader2 size="20" class="spin" />
        <span>{m.layout_loading_text()}</span>
      </div>
    {:else if decodeError}
      <div class="error-message">
        {decodeError}
      </div>
    {:else if displaySrc}
      <div
        class="transform-layer"
        style="transform: translate({translateX}px, {translateY}px) scale({scale}) rotate({rotation}deg);"
      >
        <!-- svelte-ignore a11y_img_redundant_alt -->
        <img
          bind:this={imgElement}
          onload={fitToScreen}
          src={displaySrc}
          alt={filename}
          class="viewer-img"
          draggable="false"
        />
      </div>
    {/if}
  </div>
</div>

<style>
  .viewer {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    background: var(--background);
    color: var(--foreground);
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
    display: flex;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    color: var(--muted-foreground);
    font-size: 14px;
  }

  .loading :global(.spin) {
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error-message {
    color: var(--destructive);
    background: var(--destructive-foreground);
    padding: 12px 16px;
    border-radius: 8px;
    border: 1px solid var(--destructive);
    font-size: 14px;
  }

  .viewer-img {
    max-width: none;
    pointer-events: none;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    border-radius: 2px;
  }
</style>
