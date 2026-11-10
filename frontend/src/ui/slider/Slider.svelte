<script lang="ts">
  import { Slider } from 'bits-ui';

  type Props = {
    value: number;
    label: string;
    min?: number;
    max: number;
    step?: number;
    format?: (v: number) => string;
    /** Every move: for live preview. */
    oninput?: (v: number) => void;
    /** Pointer released or key pressed: for saving. */
    oncommit?: (v: number) => void;
  };
  let { value = $bindable(), label, min = 0, max, step = 1, format = String, oninput, oncommit }: Props = $props();
</script>

<div class="slider">
  <Slider.Root
    type="single"
    class="root"
    {value}
    {min}
    {max}
    {step}
    aria-label={label}
    onValueChange={(v) => {
      value = v;
      oninput?.(v);
    }}
    onValueCommit={(v) => oncommit?.(v)}
  >
    <span class="track"><Slider.Range class="range" /></span>
    <Slider.Thumb index={0} class="knob" aria-label={label} />
  </Slider.Root>
  <output class="tabular">{format(value)}</output>
</div>

<style>
  .slider {
    display: flex;
    align-items: center;
    gap: 14px;
    width: 260px;
  }
  :global(.root) {
    position: relative;
    flex: 1;
    display: flex;
    align-items: center;
    height: 26px;
    touch-action: none;
    cursor: pointer;
  }
  .track {
    position: relative;
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: var(--overlay);
    overflow: hidden;
  }
  :global(.range) {
    position: absolute;
    height: 100%;
    background: var(--accent);
    border-radius: 999px;
  }
  :global(.knob) {
    display: block;
    width: 18px;
    height: 18px;
    border-radius: 999px;
    background: #fff;
    box-shadow:
      0 2px 8px #0008,
      0 0 0 0 var(--accent-glow);
    transition:
      transform var(--quick) var(--ease),
      box-shadow var(--duration) var(--ease);
  }
  :global(.knob:hover),
  :global(.knob:focus-visible) {
    box-shadow:
      0 2px 8px #0008,
      0 0 0 6px var(--accent-soft);
    outline: none;
  }
  :global(.knob:active) {
    transform: scale(1.12);
  }
  output {
    min-width: 42px;
    text-align: right;
    color: var(--muted);
    font-size: 12px;
  }
</style>
