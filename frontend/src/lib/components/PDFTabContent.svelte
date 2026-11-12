<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import NativePDFViewer from "./NativePDFViewer.svelte";
  import PdfjsViewer from "./PdfjsViewer.svelte";
  import { settingsStore } from "#lib/stores/settings.svelte.js";
  import * as m from "#lib/paraglide/messages.js";
  import {
    type AttachmentSource,
    attachmentSourceToBlob,
  } from "#lib/utils/mail/attachment-source.js";

  interface Props {
    source: AttachmentSource;
    filename: string;
  }

  let { source, filename }: Props = $props();

  let blobUrl = $state<string | null>(null);
  let error = $state("");
  let destroyed = false;

  onMount(async () => {
    // The PDF is fetched as raw bytes (AssetServer) straight into a Blob:
    // no base64 copy of the document lives in the renderer.
    try {
      const blob = await attachmentSourceToBlob(source, "application/pdf");
      if (destroyed) return;
      blobUrl = URL.createObjectURL(blob);
    } catch (e) {
      error = m.pdf_error_loading() + e;
    }
  });

  onDestroy(() => {
    destroyed = true;
    if (blobUrl) URL.revokeObjectURL(blobUrl);
  });
</script>

{#if error}
  <div class="state">
    <span class="error">{error}</span>
  </div>
{:else if blobUrl}
  {#if settingsStore.settings.pdfRenderer === "native"}
    <NativePDFViewer src={blobUrl} {filename} saveSource={source} height="100%" />
  {:else}
    <PdfjsViewer src={blobUrl} {filename} saveSource={source} height="100%" />
  {/if}
{:else}
  <div class="state">
    <div class="spinner"></div>
    <span>{m.pdf_loading()}</span>
  </div>
{/if}

<style>
  .state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 10px;
    background: var(--background);
    color: var(--foreground);
  }

  .error {
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
    to { transform: rotate(360deg); }
  }
</style>
