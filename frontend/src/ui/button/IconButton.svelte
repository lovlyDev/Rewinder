<script lang="ts">
  import Icon from '../icon/Icon.svelte';
  import type { IconName } from '../icon/icons';
  import Tooltip from '../tooltip/Tooltip.svelte';

  type Props = {
    name: IconName;
    label: string;
    active?: boolean;
    disabled?: boolean;
    size?: 'sm' | 'md';
    tone?: 'default' | 'danger';
    side?: 'top' | 'bottom' | 'left' | 'right';
    onclick?: (e: MouseEvent) => void;
  };
  let { name, label, active = false, disabled = false, size = 'md', tone = 'default', side = 'bottom', onclick }: Props =
    $props();
</script>

<Tooltip {label} {side}>
  {#snippet children({ props })}
    <button
      {...props}
      class="icon-btn {size} {tone}"
      aria-label={label}
      aria-pressed={active || undefined}
      {disabled}
      {onclick}
    >
      <Icon {name} filled={active} size={size === 'sm' ? 16 : 18} />
    </button>
  {/snippet}
</Tooltip>

<style>
  .icon-btn {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    flex: none;
    border-radius: 999px;
    color: var(--muted);
    transition:
      transform var(--quick) var(--ease),
      background-color var(--quick) var(--ease),
      color var(--quick) var(--ease);
  }
  .sm {
    width: 30px;
    height: 30px;
  }
  .icon-btn:hover:not(:disabled) {
    background: var(--hover);
    color: var(--text);
  }
  .icon-btn:active:not(:disabled) {
    transform: scale(0.9);
    background: var(--press);
  }
  .icon-btn[aria-pressed='true'] {
    color: var(--accent);
    background: var(--accent-soft);
  }
  .danger:hover:not(:disabled) {
    background: #ff5d6e22;
    color: var(--danger);
  }
  .icon-btn:disabled {
    opacity: 0.35;
  }
</style>
