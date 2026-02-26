<script lang="ts">
  import MailViewer from "$lib/components/MailViewer.svelte";
  import { mailState } from "$lib/stores/mail-state.svelte";
  import * as m from "$lib/paraglide/messages.js";

  let { data } = $props();

  $effect(() => {
    if (data.email) {
      mailState.setParams(data.email);
    }
  });
</script>

<div class="page">
  <section class="center" aria-label={m.page_overview_label()} id="main-content-app">
    <MailViewer />
  </section>
</div>

<style>
  .page {
    height: 100%;
    min-height: 0;
    display: flex;
    gap: 12px;
    padding: 12px;
    box-sizing: border-box;
    overflow: hidden;
  }

  .center {
    flex: 1 1 auto;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
</style>
