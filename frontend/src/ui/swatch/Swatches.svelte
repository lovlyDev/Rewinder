<script lang="ts">
  import Icon from '../icon/Icon.svelte';
  import Tooltip from '../tooltip/Tooltip.svelte';

  type Props = {
    value: string;
    presets: string[];
    label: string;
    customLabel: string;
    /** Live preview while the native picker is open. */
    oninput?: (v: string) => void;
    onchange?: (v: string) => void;
  };
  let { value = $bindable(), presets, label, customLabel, oninput, onchange }: Props = $props();

  const custom = $derived(!presets.includes(value.toLowerCase()));
</script>

<div class="swatches" role="radiogroup" aria-label={label}>
  {#each presets as color (color)}
    <button
      class="swatch"
      role="radio"
      aria-checked={value.toLowerCase() === color}
      aria-label={color}
      style="--c:{color}"
      onclick={() => {
        value = color;
        onchange?.(color);
      }}
    >
      <span class="dot"><Icon name="check" size={12} still /></span>
    </button>
  {/each}
  <Tooltip label={customLabel}>
    {#snippet children({ props })}
      <label {...props} class="swatch custom" class:on={custom} style="--c:{custom ? value : 'transparent'}">
        <input
          type="color"
          {value}
          aria-label={customLabel}
          oninput={(e) => {
            value = e.currentTarget.value;
            oninput?.(value);
          }}
          onchange={(e) => onchange?.(e.currentTarget.value)}
        />
        <span class="dot"><Icon name="palette" size={13} filled={custom} /></span>
      </label>
    {/snippet}
  </Tooltip>
</div>

<style>
  .swatches {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .swatch {
    position: relative;
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    border-radius: 999px;
    background: var(--c);
    cursor: pointer;
    transition: transform var(--quick) var(--ease);
  }
  .swatch:hover {
    transform: scale(1.08);
  }
  .swatch:active {
    transform: scale(0.92);
  }
  .swatch::after {
    content: '';
    position: absolute;
    inset: -4px;
    border-radius: 999px;
    box-shadow: 0 0 0 2px var(--c);
    opacity: 0;
    transform: scale(0.85);
    transition:
      opacity var(--quick) var(--ease),
      transform var(--duration) var(--ease);
  }
  .swatch[aria-checked='true']::after,
  .custom.on::after {
    opacity: 1;
    transform: none;
  }
  .dot {
    display: grid;
    place-items: center;
    color: #0d0b16;
    opacity: 0;
    transition: opacity var(--quick) var(--ease);
  }
  .swatch[aria-checked='true'] .dot {
    opacity: 1;
  }
  .custom {
    background: conic-gradient(from 90deg, #ff8f9c, #ffcf7a, #7ee2b0, #84b7ff, #a99bff, #ff8f9c);
  }
  .custom .dot {
    width: 22px;
    height: 22px;
    border-radius: 999px;
    background: var(--surface);
    color: var(--text);
    opacity: 1;
  }
  .custom.on .dot {
    background: var(--c);
    color: #0d0b16;
  }
  .custom.on::after {
    box-shadow: 0 0 0 2px var(--text);
  }
  input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
</style>
