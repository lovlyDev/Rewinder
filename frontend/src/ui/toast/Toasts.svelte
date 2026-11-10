<script lang="ts">
  import { flip } from 'svelte/animate';
  import { t } from '../../i18n';
  import { duration, rise } from '../../lib/motion';
  import { dismiss, toasts } from '../../lib/toasts';
  import Icon from '../icon/Icon.svelte';

  const icon = { info: 'restore', success: 'checkCircle', error: 'warning' } as const;
</script>

<div class="toasts" aria-live="polite">
  {#each $toasts as toast (toast.id)}
    <div class="toast {toast.kind}" transition:rise={{ y: 14 }} animate:flip={{ duration: duration(200) }}>
      <span class="badge"><Icon name={icon[toast.kind]} filled size={16} /></span>
      <span class="text">{$t(toast.code, toast.params)}</span>
      <button class="close" aria-label={$t('close')} onclick={() => dismiss(toast.id)}><Icon name="x" size={14} /></button>
    </div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    left: 50%;
    bottom: 18px;
    z-index: 500;
    display: grid;
    justify-items: center;
    gap: 8px;
    transform: translateX(-50%);
    pointer-events: none;
  }
  .toast {
    pointer-events: auto;
    display: flex;
    align-items: center;
    gap: 10px;
    max-width: min(520px, calc(100vw - 32px));
    padding: 6px 6px 6px 8px;
    border-radius: 999px;
    background: var(--overlay);
    box-shadow:
      0 16px 40px -10px #000c,
      0 0 0 1px var(--line);
    font-size: 13px;
  }
  .badge {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    border-radius: 999px;
    background: var(--accent-soft);
    color: var(--accent);
  }
  .success .badge {
    background: #7ee2b01f;
    color: var(--ok);
  }
  .error .badge {
    background: #ff8f9c1f;
    color: var(--danger);
  }
  .text {
    padding-right: 2px;
  }
  .close {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    border-radius: 999px;
    color: var(--faint);
    transition:
      background-color var(--quick) var(--ease),
      color var(--quick) var(--ease);
  }
  .close:hover {
    background: var(--hover);
    color: var(--text);
  }
</style>
