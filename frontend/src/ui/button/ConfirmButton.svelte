<script lang="ts">
  import Icon from '../icon/Icon.svelte';
  import type { IconName } from '../icon/icons';
  import Tooltip from '../tooltip/Tooltip.svelte';

  // A destructive icon button that asks for a second click instead of a dialog:
  // the first click arms it and shows what will happen, the second one acts.
  type Props = { icon: IconName; label: string; confirmLabel: string; onconfirm: () => void };
  let { icon, label, confirmLabel, onconfirm }: Props = $props();

  let armed = $state(false);
  let timer: ReturnType<typeof setTimeout> | undefined;

  function click() {
    if (armed) {
      clearTimeout(timer);
      armed = false;
      onconfirm();
      return;
    }
    armed = true;
    timer = setTimeout(() => (armed = false), 3500);
  }
  $effect(() => () => clearTimeout(timer));
</script>

<Tooltip label={armed ? confirmLabel : label}>
  {#snippet children({ props })}
    <button {...props} class="confirm" class:armed aria-label={armed ? confirmLabel : label} onclick={click} onblur={() => (armed = false)}>
      <Icon name={icon} filled={armed} size={18} />
      <span class="text">{confirmLabel}</span>
    </button>
  {/snippet}
</Tooltip>

<style>
  .confirm {
    display: inline-flex;
    align-items: center;
    gap: 0;
    height: 36px;
    padding: 0 9px;
    border-radius: 999px;
    color: var(--muted);
    transition:
      background-color var(--quick) var(--ease),
      color var(--quick) var(--ease),
      gap var(--duration) var(--ease),
      padding var(--duration) var(--ease),
      transform var(--quick) var(--ease);
  }
  .confirm:hover {
    background: var(--hover);
    color: var(--text);
  }
  .confirm:active {
    transform: scale(0.94);
  }
  .text {
    max-width: 0;
    overflow: hidden;
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
    opacity: 0;
    transition:
      max-width var(--duration) var(--ease),
      opacity var(--quick) var(--ease);
  }
  .armed {
    gap: 8px;
    padding: 0 14px 0 11px;
    background: #ff5d6e22;
    color: var(--danger);
  }
  .armed:hover {
    background: #ff5d6e33;
    color: var(--danger);
  }
  .armed .text {
    max-width: 180px;
    opacity: 1;
  }
</style>
