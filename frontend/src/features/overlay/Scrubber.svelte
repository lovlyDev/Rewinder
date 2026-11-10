<script lang="ts">
  import { Spring } from 'svelte/motion';
  import { duration } from '../../lib/motion';

  type Props = { count: number; index: number; label: string; onpick: (index: number) => void };
  let { count, index, label, onpick }: Props = $props();

  // Time runs left to right: oldest snapshot on the left, the current state on the right.
  const at = (i: number) => (count > 1 ? (count - 1 - i) / (count - 1) : 1);
  const pos = new Spring(1, { stiffness: 0.18, damping: 0.78 });
  $effect(() => {
    void pos.set(at(index), { instant: duration() === 0 });
  });

  // At most ~120 tick marks; denser timelines are sampled.
  const ticks = $derived.by(() => {
    const step = Math.max(1, Math.ceil(count / 120));
    const out: number[] = [];
    for (let i = 0; i < count; i += step) out.push(i);
    return out;
  });

  let track: HTMLDivElement;
  let dragging = false;
  function pick(e: PointerEvent) {
    const r = track.getBoundingClientRect();
    const x = Math.min(1, Math.max(0, (e.clientX - r.left) / r.width));
    onpick(Math.round((1 - x) * (count - 1)));
  }
</script>

<div
  class="track"
  bind:this={track}
  role="slider"
  tabindex="-1"
  aria-label={label}
  aria-valuemin={0}
  aria-valuemax={count - 1}
  aria-valuenow={count - 1 - index}
  onpointerdown={(e) => {
    dragging = true;
    track.setPointerCapture(e.pointerId);
    pick(e);
  }}
  onpointermove={(e) => dragging && pick(e)}
  onpointerup={() => (dragging = false)}
>
  <span class="line"></span>
  <span class="fill" style="transform:scaleX({pos.current})"></span>
  {#each ticks as i (i)}
    <span class="tick" class:passed={i >= index} style="left:{at(i) * 100}%"></span>
  {/each}
  <span class="thumb" style="left:{pos.current * 100}%"><span></span></span>
</div>

<style>
  .track {
    position: relative;
    flex: 1;
    height: 36px;
    margin: 0 10px;
    cursor: pointer;
    touch-action: none;
    outline: none;
  }
  .line,
  .fill {
    position: absolute;
    left: 0;
    right: 0;
    top: 50%;
    height: 4px;
    margin-top: -2px;
    border-radius: 999px;
    background: var(--overlay);
  }
  .fill {
    background: linear-gradient(90deg, transparent, var(--accent-glow));
    transform-origin: left;
  }
  .tick {
    position: absolute;
    top: 50%;
    width: 4px;
    height: 4px;
    margin: -2px 0 0 -2px;
    border-radius: 999px;
    background: var(--faint);
  }
  .tick.passed {
    background: var(--muted);
  }
  .thumb {
    position: absolute;
    top: 50%;
    width: 0;
    height: 0;
  }
  .thumb span {
    position: absolute;
    left: -9px;
    top: -9px;
    width: 18px;
    height: 18px;
    border-radius: 999px;
    background: var(--accent);
    box-shadow:
      0 0 0 5px var(--accent-soft),
      0 4px 14px var(--accent-glow);
  }
</style>
