<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import PDFViewer from "./PDFViewer.svelte";
  import * as m from "$lib/paraglide/messages.js";

  interface Props {
    base64Data: string;
    filename: string;
  }

  let { base64Data, filename }: Props = $props();

  let blobUrl = $state<string | null>(null);
  let error = $state("");

  onMount(() => {
    try {
      const binaryString = window.atob(base64Data);
      const len = binaryString.length;
      const bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      const blob = new Blob([bytes], { type: "application/pdf" });
      blobUrl = URL.createObjectURL(blob);
    } catch (e) {
      error = m.pdf_error_loading() + e;
    }
  });

  onDestroy(() => {
    if (blobUrl) URL.revokeObjectURL(blobUrl);
  });
</script>

{#if error}
  <div class="state">
    <span class="error">{error}</span>
  </div>
{:else if blobUrl}
  <PDFViewer src={blobUrl} {filename} height="100%" />
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
