<script lang="ts">
  import type { LayoutProps } from "./$types";
  import TitleBar from "$lib/components/TitleBar.svelte";

  let { data, children }: LayoutProps = $props();
</script>

<div class="app-layout">
  <TitleBar title="EMLy Viewer" />

  <!-- Content -->
  <main class="content">
    {@render children()}
  </main>
</div>

<style>
  :global(body) {
    margin: 0;
    background: var(--background);
    color: var(--foreground);
    overflow: hidden;
  }

  .app-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
    background: var(--background);
    color: var(--foreground);
  }

  .content {
    flex: 1;
    overflow: hidden;
    position: relative;
    background: var(--background);
  }
</style>
