<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';
  import Icon from '../icon/Icon.svelte';
  import type { IconName } from '../icon/icons';

  type Props = HTMLButtonAttributes & {
    variant?: 'primary' | 'soft' | 'ghost' | 'danger';
    size?: 'sm' | 'md' | 'lg';
    icon?: IconName;
    children?: Snippet;
  };
  let { variant = 'soft', size = 'md', icon, children, class: extra = '', ...rest }: Props = $props();
</script>

<button class="btn {variant} {size} {extra}" {...rest}>
  {#if icon}<Icon name={icon} size={size === 'sm' ? 16 : 18} />{/if}
  {#if children}<span>{@render children()}</span>{/if}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    height: 38px;
    padding: 0 16px;
    border-radius: 999px;
    font-weight: 550;
    white-space: nowrap;
    transition:
      transform var(--quick) var(--ease),
      background-color var(--quick) var(--ease),
      color var(--quick) var(--ease),
      box-shadow var(--duration) var(--ease),
      opacity var(--quick) var(--ease);
  }
  .btn:active:not(:disabled) {
    transform: scale(0.96);
  }
  .btn:disabled {
    opacity: 0.4;
  }
  .sm {
    height: 32px;
    padding: 0 12px;
    font-size: 12px;
    gap: 6px;
  }
  .lg {
    height: 44px;
    padding: 0 22px;
    font-size: 14px;
  }

  .soft {
    background: var(--raised);
  }
  .soft:hover:not(:disabled) {
    background: var(--overlay);
  }
  .ghost {
    color: var(--muted);
  }
  .ghost:hover:not(:disabled) {
    background: var(--hover);
    color: var(--text);
  }
  .primary {
    background: var(--accent);
    color: var(--accent-ink);
    box-shadow: 0 6px 22px -6px var(--accent-glow);
  }
  .primary:hover:not(:disabled) {
    box-shadow: 0 8px 30px -4px var(--accent-glow);
    filter: brightness(1.06);
  }
  .danger {
    background: #ff8f9c14;
    color: var(--danger);
  }
  .danger:hover:not(:disabled) {
    background: #ff8f9c24;
  }
</style>
