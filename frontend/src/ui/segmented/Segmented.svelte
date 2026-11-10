<script lang="ts" generics="T extends string">
  type Option = { value: T; label: string };
  type Props = { options: Option[]; value: T; label: string; onchange?: (value: T) => void };
  let { options, value = $bindable(), label, onchange }: Props = $props();

  const index = $derived(Math.max(0, options.findIndex((o) => o.value === value)));

  function pick(v: T) {
    if (v === value) return;
    value = v;
    onchange?.(v);
  }
  function keyboard(e: KeyboardEvent) {
    const d = e.key === 'ArrowRight' ? 1 : e.key === 'ArrowLeft' ? -1 : 0;
    if (!d) return;
    e.preventDefault();
    pick(options[(index + d + options.length) % options.length].value);
  }
</script>

<!-- Equal-width segments, so the indicator moves with a single translateX. -->
<div class="seg" role="radiogroup" aria-label={label} style="--n:{options.length};--i:{index}" onkeydown={keyboard} tabindex="-1">
  <span class="indicator"></span>
  {#each options as option (option.value)}
    <button
      role="radio"
      aria-checked={option.value === value}
      tabindex={option.value === value ? 0 : -1}
      onclick={() => pick(option.value)}>{option.label}</button
    >
  {/each}
</div>

<style>
  .seg {
    position: relative;
    display: grid;
    grid-template-columns: repeat(var(--n), 1fr);
    padding: 3px;
    border-radius: 999px;
    background: var(--raised);
    outline: none;
  }
  .indicator {
    position: absolute;
    top: 3px;
    bottom: 3px;
    left: 3px;
    width: calc((100% - 6px) / var(--n));
    border-radius: 999px;
    background: var(--overlay);
    box-shadow: 0 2px 8px #0006;
    transform: translateX(calc(var(--i) * 100%));
    transition: transform var(--duration) var(--ease);
  }
  button {
    position: relative;
    height: 30px;
    padding: 0 14px;
    border-radius: 999px;
    color: var(--muted);
    font-size: 12px;
    font-weight: 550;
    white-space: nowrap;
    transition: color var(--quick) var(--ease);
  }
  button:hover,
  button[aria-checked='true'] {
    color: var(--text);
  }
</style>
