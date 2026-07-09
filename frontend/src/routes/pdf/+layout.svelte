<script lang="ts">
  import type { LayoutProps } from "./$types";
  import { settingsStore } from "$lib/stores/settings.svelte.js";
  import TitleBar from "$lib/components/TitleBar.svelte";

  let { data, children }: LayoutProps = $props();

  function handleWheel(event: WheelEvent) {
    if (event.ctrlKey) {
      event.preventDefault();
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="app-layout" onwheel={handleWheel}>
  <TitleBar title="EMLy PDF Viewer" />

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
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--background);
  }
</style>