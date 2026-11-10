<script lang="ts" generics="T extends string">
  import { Select } from 'bits-ui';
  import { pop } from '../../lib/motion';
  import Icon from '../icon/Icon.svelte';

  type Option = { value: T; label: string; hint?: string };
  type Props = {
    options: Option[];
    value: T;
    label: string;
    onchange?: (value: T) => void;
  };
  let { options, value = $bindable(), label, onchange }: Props = $props();

  const current = $derived(options.find((o) => o.value === value));
</script>

<Select.Root
  type="single"
  {value}
  items={options}
  onValueChange={(v) => {
    value = v as T;
    onchange?.(v as T);
  }}
>
  <Select.Trigger class="trigger" aria-label={label}>
    <span class="value">{current?.label ?? ''}</span>
    <span class="caret"><Icon name="caretDown" size={14} /></span>
  </Select.Trigger>
  <Select.Portal>
    <Select.Content sideOffset={6} collisionPadding={12} forceMount>
      {#snippet child({ wrapperProps, props, open })}
        {#if open}
          <div {...wrapperProps}>
            <div {...props} class="menu" transition:pop={{ from: 0.95 }}>
              <Select.Viewport>
                {#each options as option (option.value)}
                  <Select.Item value={option.value} label={option.label} class="item">
                    {#snippet children({ selected })}
                      <span class="text">
                        <span>{option.label}</span>
                        {#if option.hint}<small>{option.hint}</small>{/if}
                      </span>
                      <span class="check" class:on={selected}><Icon name="check" size={15} still /></span>
                    {/snippet}
                  </Select.Item>
                {/each}
              </Select.Viewport>
            </div>
          </div>
        {/if}
      {/snippet}
    </Select.Content>
  </Select.Portal>
</Select.Root>

<style>
  :global(.trigger) {
    display: inline-flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-width: 190px;
    height: 38px;
    padding: 0 12px 0 16px;
    border-radius: 999px;
    background: var(--raised);
    font-weight: 500;
    transition:
      background-color var(--quick) var(--ease),
      transform var(--quick) var(--ease);
  }
  :global(.trigger:hover) {
    background: var(--overlay);
  }
  :global(.trigger:active) {
    transform: scale(0.98);
  }
  .caret {
    color: var(--muted);
    transition: transform var(--duration) var(--ease);
  }
  :global(.trigger[data-state='open']) .caret {
    transform: rotate(180deg);
  }

  .menu {
    min-width: var(--bits-select-anchor-width);
    max-height: min(320px, var(--bits-select-content-available-height));
    overflow: auto;
    padding: 6px;
    border-radius: var(--r-sm);
    background: var(--overlay);
    box-shadow:
      0 18px 48px -8px #000c,
      0 0 0 1px var(--line);
    transform-origin: var(--bits-select-content-transform-origin);
    z-index: 400;
  }
  :global(.item) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 9px 10px 9px 12px;
    border-radius: var(--r-xs);
    cursor: pointer;
    outline: none;
    transition: background-color var(--quick) var(--ease);
  }
  :global(.item[data-highlighted]) {
    background: var(--hover);
  }
  :global(.item[data-selected]) {
    color: var(--accent);
  }
  .text {
    display: grid;
    gap: 2px;
  }
  small {
    color: var(--faint);
    font-size: 11px;
  }
  .check {
    opacity: 0;
    transform: scale(0.6);
    transition:
      opacity var(--quick) var(--ease),
      transform var(--duration) var(--ease);
  }
  .check.on {
    opacity: 1;
    transform: none;
  }
</style>
