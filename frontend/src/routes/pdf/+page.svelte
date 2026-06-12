<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { PageData } from "./$types";
  import { sidebarOpen } from "$lib/stores/app";
  import { toast } from "svelte-sonner";
  import * as m from "$lib/paraglide/messages.js";
  import { logger } from "$lib/utils/logger";
  import PDFViewer from "$lib/components/PDFViewer.svelte";

  let { data }: { data: PageData } = $props();

  let filename = $state("");
  let blobUrl = $state<string | null>(null);
  let base64Data = $state("");
  let error = $state("");

  onMount(() => {
    try {
      const result = data?.data;
      if (result) {
        const binaryString = window.atob(result.data);
        const len = binaryString.length;
        const bytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        filename = result.filename;
        base64Data = result.data;
        logger.info("pdf_viewer: data received", { filename, sizeBytes: len });
        document.title = filename + " - EMLy PDF Viewer";
        sidebarOpen.set(false);

        const blob = new Blob([bytes], { type: "application/pdf" });
        blobUrl = URL.createObjectURL(blob);
      } else {
        logger.warn("pdf_viewer: no data received");
        toast.error(m.pdf_error_no_data());
        error = m.pdf_error_no_data_desc();
      }
    } catch (e) {
      logger.error("pdf_viewer: mount error", { error: String(e) });
      error = m.pdf_error_loading() + e;
    }
  });

  onDestroy(() => {
    if (blobUrl) URL.revokeObjectURL(blobUrl);
  });
</script>

{#if error}
  <div class="state-page error">
    <div>{error}</div>
  </div>
{:else if blobUrl}
  <PDFViewer src={blobUrl} {filename} {base64Data} />
{:else}
  <div class="state-page">
    <div class="spinner"></div>
    <div>{m.pdf_loading()}</div>
  </div>
{/if}

<style>
  .state-page {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 10px;
    background: var(--background);
    color: var(--foreground);
  }

  .state-page.error {
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
