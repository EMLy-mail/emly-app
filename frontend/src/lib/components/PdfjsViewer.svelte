<!--
  Default PDF viewer ("builtin"), drawing pages with pdf.js (pdfjs-dist) onto
  our own canvases. NativePDFViewer.svelte is the alternative engine.

  Two earlier attempts are worth not repeating: @embedpdf (pdfium/wasm) needed
  a build-time patch to survive Svelte 5.56's rest_props change and pulled a
  wasm payload for what is drawn here in one file, and `svelte-pdf` owns its
  toolbar and keeps zoom and rotation private - `scale` is read once at init
  and `rotation` is not a prop at all - so EMLy's toolbar could not drive it.
  Talking to pdf.js directly is less code than working around either.

  Pages scroll continuously. Only pages near the viewport hold a painted
  canvas; the rest keep their box size so the scrollbar stays honest, and paint
  when scrolled into range. Rotation is view-only - the attachment bytes are
  never touched.
-->
<script lang="ts" module>
  import * as pdfjs from "pdfjs-dist";
  import pdfWorkerUrl from "pdfjs-dist/build/pdf.worker.mjs?url";

  /*
   * Vite emits the worker as a real asset and hands back its final URL, in dev
   * and in the production build alike. Left to itself pdf.js guesses a path
   * relative to the importing module, which inside a bundle is wrong, and the
   * failure only shows up at runtime as "Setting up fake worker failed".
   */
  pdfjs.GlobalWorkerOptions.workerSrc = pdfWorkerUrl;
</script>

<script lang="ts">
  import { onDestroy } from "svelte";
  import type { PDFDocumentProxy, RenderTask } from "pdfjs-dist";
  import {
    RotateCcw,
    RotateCw,
    ZoomIn,
    ZoomOut,
    Download,
    ChevronLeft,
    ChevronRight,
  } from "@lucide/svelte";
  import { saveAttachmentNatively } from "$lib/utils/attachment-download";
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
  /** How far outside the viewport a page still gets painted. */
  const PREPAINT_MARGIN = "300px";

  let zoom = $state(100);
  let rotation = $state(0);
  let pageCount = $state(0);
  let currentPage = $state(1);
  let loading = $state(true);
  let error = $state("");

  let viewportEl = $state<HTMLDivElement | null>(null);
  let canvasEls = $state<HTMLCanvasElement[]>([]);
  /** Page box at scale 1, with the page's own /Rotate already applied. */
  let naturalSizes = $state<{ width: number; height: number }[]>([]);

  let doc: PDFDocumentProxy | null = null;
  let observer: IntersectionObserver | null = null;
  const tasks = new Map<number, RenderTask>();
  /** Which zoom+rotation each page was last painted at. */
  const paintedAt = new Map<number, string>();
  const onScreen = new Set<number>();

  let scale = $derived(zoom / 100);
  let viewKey = $derived(`${zoom}|${rotation}`);

  /**
   * Box a page occupies on screen. Our own rotation swaps the axes at 90 and
   * 270; the page's intrinsic rotation is already baked into naturalSizes.
   */
  function boxOf(index: number) {
    const size = naturalSizes[index];
    if (!size) return { width: 0, height: 0 };
    const swapped = rotation % 180 !== 0;
    return {
      width: Math.round((swapped ? size.height : size.width) * scale),
      height: Math.round((swapped ? size.width : size.height) * scale),
    };
  }

  $effect(() => {
    const url = src;
    let cancelled = false;

    loading = true;
    error = "";

    void (async () => {
      try {
        const loaded = await pdfjs.getDocument({ url }).promise;
        if (cancelled) return void loaded.loadingTask.destroy();

        // Metadata only, but it has to happen up front: without every page's
        // size the placeholders are wrong and the scrollbar lies.
        const sizes: { width: number; height: number }[] = [];
        for (let n = 1; n <= loaded.numPages; n++) {
          const page = await loaded.getPage(n);
          const view = page.getViewport({ scale: 1 });
          sizes.push({ width: view.width, height: view.height });
        }
        if (cancelled) return void loaded.loadingTask.destroy();

        doc = loaded;
        naturalSizes = sizes;
        pageCount = loaded.numPages;
        loading = false;
      } catch (e) {
        if (!cancelled) {
          error = m.pdf_error_loading() + e;
          loading = false;
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  });

  // Watch the page boxes so pages paint as they scroll near the viewport.
  $effect(() => {
    const root = viewportEl;
    const count = pageCount;
    if (!root || !count) return;

    const io = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const index = Number((entry.target as HTMLElement).dataset.index);
          if (entry.isIntersecting) {
            onScreen.add(index);
            void paintPage(index);
          } else {
            onScreen.delete(index);
          }
        }
      },
      { root, rootMargin: PREPAINT_MARGIN, threshold: 0 },
    );

    for (const el of root.querySelectorAll<HTMLElement>(".page")) io.observe(el);
    observer = io;

    return () => {
      io.disconnect();
      if (observer === io) observer = null;
    };
  });

  // A new zoom or rotation invalidates every painted page at once.
  $effect(() => {
    viewKey;
    paintedAt.clear();
    for (const index of onScreen) void paintPage(index);
  });

  /*
   * Derived from the scroll position rather than from the observer: the
   * observer deliberately fires early, PREPAINT_MARGIN before a page is on
   * screen, so its idea of "visible" would leave the counter a page behind.
   */
  function syncCurrentPage() {
    const root = viewportEl;
    if (!root || !pageCount) return;

    const middle = root.scrollTop + root.clientHeight / 2;
    let visible = 1;
    for (const el of root.querySelectorAll<HTMLElement>(".page")) {
      if (el.offsetTop > middle) break;
      visible = Number(el.dataset.index) + 1;
    }
    currentPage = visible;
  }

  async function paintPage(index: number) {
    const pdf = doc;
    const canvas = canvasEls[index];
    if (!pdf || !canvas) return;

    // Snapshot the view: everything below is async and the user may zoom again.
    const key = viewKey;
    const atScale = scale;
    const atRotation = rotation;
    if (paintedAt.get(index) === key) return;

    tasks.get(index)?.cancel();
    tasks.delete(index);

    try {
      const page = await pdf.getPage(index + 1);
      // getViewport's rotation is absolute, so the page's own /Rotate has to be
      // carried along or landscape pages would snap upright.
      const view = page.getViewport({
        scale: atScale,
        rotation: (page.rotate + atRotation) % 360,
      });

      const ratio = window.devicePixelRatio || 1;
      canvas.width = Math.floor(view.width * ratio);
      canvas.height = Math.floor(view.height * ratio);
      canvas.style.width = `${Math.floor(view.width)}px`;
      canvas.style.height = `${Math.floor(view.height)}px`;

      const context = canvas.getContext("2d");
      if (!context) return;

      const task = page.render({
        canvasContext: context,
        viewport: view,
        transform: ratio === 1 ? undefined : [ratio, 0, 0, ratio, 0, 0],
      });
      tasks.set(index, task);
      await task.promise;
      paintedAt.set(index, key);
    } catch (e) {
      // Cancelling in-flight renders is how zooming quickly is meant to work.
      if ((e as { name?: string })?.name !== "RenderingCancelledException") {
        error = m.pdf_error_rendering() + e;
      }
    } finally {
      tasks.delete(index);
    }
  }

  function goToPage(target: number) {
    const clamped = Math.min(pageCount, Math.max(1, target));
    const el = viewportEl?.querySelector<HTMLElement>(
      `.page[data-index="${clamped - 1}"]`,
    );
    if (!el) return;
    viewportEl?.scrollTo({ top: el.offsetTop - 8, behavior: "smooth" });
    currentPage = clamped;
  }

  function zoomBy(delta: number) {
    zoom = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, zoom + delta));
  }

  function rotateBy(delta: number) {
    rotation = (((rotation + delta) % 360) + 360) % 360;
  }

  function downloadPdf() {
    if (!base64Data) return;
    void saveAttachmentNatively(base64Data, filename || "document.pdf");
  }

  onDestroy(() => {
    for (const task of tasks.values()) task.cancel();
    tasks.clear();
    observer?.disconnect();
    void doc?.loadingTask.destroy();
    doc = null;
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
      <div class="separator"></div>
      <button
        class="btn"
        onclick={() => goToPage(currentPage - 1)}
        disabled={currentPage <= 1}
        title={m.pdf_prev_page()}
      >
        <ChevronLeft size="16" />
      </button>
      <span class="page-info">{currentPage} / {pageCount || 1}</span>
      <button
        class="btn"
        onclick={() => goToPage(currentPage + 1)}
        disabled={currentPage >= pageCount}
        title={m.pdf_next_page()}
      >
        <ChevronRight size="16" />
      </button>
      <div class="separator"></div>
      <button
        class="btn"
        onclick={() => zoomBy(-ZOOM_STEP)}
        disabled={zoom <= ZOOM_MIN}
        title={m.pdf_zoom_out()}
      >
        <ZoomOut size="16" />
      </button>
      <span class="page-info">{zoom}%</span>
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
        title={m.pdf_rotate_left()}
      >
        <RotateCcw size="16" />
      </button>
      <button
        class="btn"
        onclick={() => rotateBy(90)}
        title={m.pdf_rotate_right()}
      >
        <RotateCw size="16" />
      </button>
    </div>
  </div>

  <div class="content-area" bind:this={viewportEl} onscroll={syncCurrentPage}>
    {#if loading}
      <div class="state-overlay">
        <div class="spinner"></div>
        <div>{m.pdf_loading()}</div>
      </div>
    {:else if error}
      <div class="state-overlay error">{error}</div>
    {:else}
      {#each { length: pageCount }, index (index)}
        {@const box = boxOf(index)}
        <div
          class="page"
          data-index={index}
          style:width="{box.width}px"
          style:height="{box.height}px"
        >
          <canvas bind:this={canvasEls[index]}></canvas>
        </div>
      {/each}
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

  .page-info {
    font-size: 13px;
    min-width: 48px;
    text-align: center;
    color: var(--foreground);
    opacity: 0.8;
  }

  .content-area {
    flex: 1;
    overflow: auto;
    position: relative;
    background: var(--muted);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 12px 0;
  }

  .page {
    flex: none;
    background: #fff;
    box-shadow: 0 1px 4px rgb(0 0 0 / 0.35);
  }

  .page canvas {
    display: block;
  }

  .state-overlay {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    flex: 1;
    gap: 10px;
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

  .content-area::-webkit-scrollbar {
    width: 10px;
    height: 10px;
  }

  .content-area::-webkit-scrollbar-track {
    background: transparent;
  }

  .content-area::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 6px;
  }

  .content-area::-webkit-scrollbar-thumb:hover {
    background: var(--muted-foreground);
  }
</style>
