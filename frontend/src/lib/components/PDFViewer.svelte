<script lang="ts">
  import { EmbedPDF } from '@embedpdf/core/svelte';
  import { createPluginRegistration } from '@embedpdf/core';
  import { usePdfiumEngine } from '@embedpdf/engines/svelte';
  import { ViewportPluginPackage } from '@embedpdf/plugin-viewport/svelte';
  import { ScrollPluginPackage } from '@embedpdf/plugin-scroll/svelte';
  import { DocumentManagerPluginPackage } from '@embedpdf/plugin-document-manager/svelte';
  import { RenderPluginPackage } from '@embedpdf/plugin-render/svelte';
  import { ZoomPluginPackage, ZoomMode } from '@embedpdf/plugin-zoom/svelte';
  import { RotatePluginPackage } from '@embedpdf/plugin-rotate/svelte';
  import PDFViewerContent from './PDFViewerContent.svelte';
  import * as m from '$lib/paraglide/messages.js';

  interface Props {
    src: string;
    filename?: string;
    height?: string;
    base64Data?: string;
  }

  let { src, filename = '', height = '100%', base64Data = '' }: Props = $props();

  const pdfEngine = usePdfiumEngine();

  const plugins = $derived([
    createPluginRegistration(DocumentManagerPluginPackage, {
      initialDocuments: [{ url: src }],
    }),
    createPluginRegistration(ViewportPluginPackage),
    createPluginRegistration(ScrollPluginPackage),
    createPluginRegistration(RenderPluginPackage),
    createPluginRegistration(ZoomPluginPackage, { defaultZoomLevel: ZoomMode.FitWidth }),
    createPluginRegistration(RotatePluginPackage),
  ]);
</script>

<div class="wrapper" style:height>
  {#if pdfEngine.isLoading || !pdfEngine.engine}
    <div class="loading-overlay">
      <div class="spinner"></div>
      <div>{m.pdf_loading()}</div>
    </div>
  {:else if pdfEngine.error}
    <div class="error-overlay">{pdfEngine.error.message}</div>
  {:else}
    <EmbedPDF engine={pdfEngine.engine} {plugins}>
      {#snippet children(ctx)}
        {#if ctx.activeDocumentId}
          <PDFViewerContent documentId={ctx.activeDocumentId} {filename} {base64Data} />
        {:else}
          <div class="loading-overlay">
            <div class="spinner"></div>
            <div>{m.pdf_loading()}</div>
          </div>
        {/if}
      {/snippet}
    </EmbedPDF>
  {/if}
</div>

<style>
  .wrapper {
    width: 100%;
    overflow: hidden;
    background: var(--background);
  }

  .loading-overlay,
  .error-overlay {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 10px;
    color: var(--foreground);
    background: var(--background);
  }

  .error-overlay {
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
