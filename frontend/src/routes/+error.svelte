<script lang="ts">
  import { page } from '$app/state';
  import * as m from '$lib/paraglide/messages.js';
</script>

<div class="page">
  <div class="error-container">
    <h1 class="error-code">{page.status}</h1>
    <p class="error-message">{page.error?.message || m.error_unexpected()}</p>
  </div>
</div>

<style>
  .page {
    height: 100%;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 12px;
    box-sizing: border-box;
    overflow: hidden;
  }

  .error-container {
    padding: 40px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    max-width: 400px;
    width: 100%;
  }

  .error-code {
    font-size: 64px;
    font-weight: 700;
    margin: 0 0 8px 0;
    opacity: 0.9;
    line-height: 1;
  }

  .error-message {
    font-size: 16px;
    opacity: 0.6;
    margin: 0 0 32px 0;
    line-height: 1.5;
  }
</style>
