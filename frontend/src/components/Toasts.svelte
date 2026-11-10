<script lang="ts">
  import { toasts } from '../lib/toasts';
</script>

<div class="toasts" aria-live="polite">
  {#each $toasts as t (t.id)}
    <div class="toast {t.kind}">{t.text}</div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    bottom: 14px;
    left: 14px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    z-index: 1000;
  }
  .toast {
    width: min(340px, calc(100vw - 28px));
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--panel-solid);
    padding: 6px 10px;
    font-size: 12px;
  }
  .toast.success { border-color: var(--ok); }
  .toast.error { border-color: var(--bad); }
</style>
