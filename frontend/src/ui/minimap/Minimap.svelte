<script lang="ts">
  import type { SnapshotLayout } from '../../lib/preview';

  type Props = { layout: SnapshotLayout | null; label: string };
  let { layout, label }: Props = $props();

  type Box = { x: number; y: number; w: number; h: number };

  // Everything is placed in percent of the bounding box of monitors and windows,
  // so the picture scales to whatever space the parent gives it.
  const view = $derived.by(() => {
    if (!layout || !layout.monitors.length) return null;
    const rects = [...layout.monitors.map((m) => m.rect), ...layout.windows.map((w) => w.rect)];
    const x0 = Math.min(...rects.map((r) => r.x));
    const y0 = Math.min(...rects.map((r) => r.y));
    const x1 = Math.max(...rects.map((r) => r.x + r.w));
    const y1 = Math.max(...rects.map((r) => r.y + r.h));
    const W = Math.max(1, x1 - x0);
    const H = Math.max(1, y1 - y0);
    const box = (r: { x: number; y: number; w: number; h: number }): Box => ({
      x: ((r.x - x0) / W) * 100,
      y: ((r.y - y0) / H) * 100,
      w: (r.w / W) * 100,
      h: (r.h / H) * 100,
    });
    return {
      ratio: W / H,
      monitors: layout.monitors.map((m) => ({ ...box(m.rect), primary: m.primary })),
      windows: layout.windows.map((w) => ({ ...box(w.rect), title: w.title, minimized: w.minimized, foreground: w.foreground })),
    };
  });

  // Fit the picture inside the parent while keeping the desktop's proportions.
  let boxW = $state(0);
  let boxH = $state(0);
  const size = $derived.by(() => {
    const ratio = view?.ratio ?? 16 / 9;
    const w = Math.min(boxW, boxH * ratio);
    return { w, h: w / ratio };
  });
</script>

<div class="fit" bind:clientWidth={boxW} bind:clientHeight={boxH}>
<div class="map" role="img" aria-label={label} style="width:{size.w}px;height:{size.h}px">
  {#if view}
    {#each view.monitors as m, i (i)}
      <span class="monitor" style="left:{m.x}%;top:{m.y}%;width:{m.w}%;height:{m.h}%"></span>
    {/each}
    {#each view.windows as w, i (i)}
      <span
        class="window"
        class:minimized={w.minimized}
        class:foreground={w.foreground}
        title={w.title}
        style="left:{w.x}%;top:{w.y}%;width:{w.w}%;height:{w.h}%"
      ><i></i></span>
    {/each}
  {/if}
</div>
</div>

<style>
  .fit {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
  }
  .map {
    position: relative;
  }
  .monitor {
    position: absolute;
    border-radius: 6px;
    background: var(--sunken);
    box-shadow: inset 0 0 0 1px var(--line);
  }
  .window {
    position: absolute;
    border-radius: 4px;
    background: color-mix(in srgb, var(--accent) 22%, transparent);
    box-shadow: inset 0 0 0 1.5px color-mix(in srgb, var(--accent) 55%, transparent);
    transition:
      left var(--duration) var(--ease),
      top var(--duration) var(--ease),
      width var(--duration) var(--ease),
      height var(--duration) var(--ease),
      opacity var(--duration) var(--ease);
  }
  /* a thin title bar, so the shape reads as a window */
  .window i {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    height: min(18%, 7px);
    border-radius: 4px 4px 0 0;
    background: color-mix(in srgb, var(--accent) 45%, transparent);
  }
  .window.foreground {
    background: color-mix(in srgb, var(--accent) 38%, transparent);
    box-shadow: inset 0 0 0 1.5px var(--accent);
  }
  .window.foreground i {
    background: var(--accent);
  }
  .window.minimized {
    opacity: 0.35;
  }
</style>
