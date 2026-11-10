<script lang="ts">
  import { Tooltip } from 'bits-ui';
  import type { Snippet } from 'svelte';
  import { pop } from '../../lib/motion';

  type Props = {
    label: string;
    side?: 'top' | 'bottom' | 'left' | 'right';
    /** Renders the trigger; spread `props` onto the focusable element. */
    children: Snippet<[{ props: Record<string, unknown> }]>;
  };
  let { label, side = 'bottom', children }: Props = $props();
</script>

<Tooltip.Root delayDuration={350}>
  <Tooltip.Trigger>
    {#snippet child({ props })}
      {@render children({ props })}
    {/snippet}
  </Tooltip.Trigger>
  <Tooltip.Portal>
    <Tooltip.Content {side} sideOffset={8} collisionPadding={8} forceMount>
      {#snippet child({ wrapperProps, props, open })}
        {#if open}
          <div {...wrapperProps}>
            <div {...props} class="tip" transition:pop={{ from: 0.94 }}>{label}</div>
          </div>
        {/if}
      {/snippet}
    </Tooltip.Content>
  </Tooltip.Portal>
</Tooltip.Root>

<style>
  .tip {
    padding: 6px 10px;
    border-radius: var(--r-xs);
    background: var(--overlay);
    color: var(--text);
    font-size: 12px;
    font-weight: 500;
    white-space: nowrap;
    box-shadow: 0 8px 24px #0008;
    pointer-events: none;
  }
</style>
