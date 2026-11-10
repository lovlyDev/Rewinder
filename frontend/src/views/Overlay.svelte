<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte';
  import { HideOverlay, LogUI } from '../../wailsjs/go/main/App.js';
  import { loadApps, loadTimeline } from '../lib/apps';
  import { when, windows } from '../lib/format';
  import { restore } from '../lib/restore';
  import { view } from '../lib/ui';
  import type { SnapshotMeta } from '../lib/types';

  /** The app the user was in when they pressed the hotkey. */
  export let requestedAppID: string | undefined;

  let el: HTMLDivElement;
  let appID: string | null = null;
  let appName = '';
  let snapshots: SnapshotMeta[] = [];
  let index = 0;

  $: current = snapshots[index];

  async function load() {
    const list = await loadApps();
    // If the app the user was in has no history, fall back to the most recent app.
    const app = list.find(a => a.appID === requestedAppID) ?? list[0];
    appID = app?.appID ?? null;
    appName = app?.name ?? '';
    snapshots = appID ? await loadTimeline(appID) : [];
    // The newest snapshot is the current state; start one step back, which is what "undo" means.
    index = Math.min(1, Math.max(0, snapshots.length - 1));
    await tick();
    window.focus();
    el?.focus();
    void LogUI(`overlay open app=${appID} requested=${requestedAppID ?? ''} snapshots=${snapshots.length} index=${index}`);
  }

  function close() {
    view.set('timeline');
    void HideOverlay();
  }

  // Hide first, so restored windows are not covered by the overlay and can take focus.
  function restoreCurrent() {
    const id = appID;
    const snap = current;
    void LogUI(`overlay restore app=${id} index=${index} snapshot=${snap?.snapshotID ?? 'none'}`);
    close();
    if (id && snap) void restore(id, snap.snapshotID);
  }

  const step = (d: number) => (index = Math.max(0, Math.min(snapshots.length - 1, index + d)));

  function onKey(e: KeyboardEvent) {
    const actions: Record<string, () => void> = {
      Escape: close,
      Enter: restoreCurrent,
      ArrowLeft: () => step(-1),
      ArrowRight: () => step(1),
    };
    const act = actions[e.key];
    if (!act) return;
    e.preventDefault();
    if (e.key === 'Escape' || e.key === 'Enter') void LogUI(`overlay key ${e.key}`);
    act();
  }

  function onWheel(e: WheelEvent) {
    e.preventDefault();
    step(e.deltaY > 0 ? 1 : -1);
  }

  onMount(() => {
    window.addEventListener('keydown', onKey, { capture: true });
    window.addEventListener('wheel', onWheel, { passive: false });
    void load();
  });
  onDestroy(() => {
    window.removeEventListener('keydown', onKey, { capture: true });
    window.removeEventListener('wheel', onWheel);
  });
</script>

<div class="overlay" role="dialog" aria-label="Quick restore" tabindex="-1" bind:this={el}>
  <div class="card">
    <div class="header">
      <div class="title">Restore{appName ? ` ${appName}` : ''}</div>
      <div class="hint">Scroll = time • ← → = step • Enter = restore • Esc = cancel</div>
    </div>

    {#if !current}
      <div class="empty">No snapshots yet</div>
    {:else}
      <div class="axis">
        {#each snapshots as s, i (s.snapshotID)}
          <button class="tick" class:active={i === index} on:click={() => (index = i)} aria-label="Snapshot {when(s.timestampUTC)}" />
        {/each}
      </div>
      <div class="preview">
        <div class="when">{when(current.timestampUTC)}</div>
        <div class="meta">
          <span>{windows(current.windowsCount)}</span>
          {#if index === 0}<span>current state</span>{/if}
          <span>{current.filesAdded}+ / {current.filesRemoved}− files</span>
        </div>
      </div>
      <div class="actions">
        <button class="btn ghost" on:click={close}>Cancel</button>
        <button class="btn primary" on:click={restoreCurrent}>Restore</button>
      </div>
    {/if}
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    display: grid;
    place-items: center;
    background: var(--scrim);
    outline: none;
  }
  .card {
    width: min(980px, calc(100vw - 32px));
    border-radius: 16px;
    border: 1px solid var(--border-strong);
    background: var(--panel-solid);
    padding: 16px;
  }
  .header { display: flex; justify-content: space-between; align-items: baseline; gap: 12px; margin-bottom: 14px; }
  .title { font-weight: 800; font-size: 18px; }
  .hint { color: var(--muted); font-size: 12px; }
  .axis {
    display: flex;
    gap: 6px;
    padding: 10px;
    border-radius: 12px;
    border: 1px solid var(--border);
    background: var(--panel);
    overflow: auto;
  }
  .tick {
    min-width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1px solid var(--border-strong);
    background: var(--panel);
    cursor: pointer;
  }
  .tick.active { border-color: var(--accent); background: var(--accent); }
  .preview { margin-top: 12px; display: flex; justify-content: space-between; align-items: baseline; }
  .when { font-weight: 800; }
  .meta { color: var(--muted); font-size: 12px; display: flex; gap: 12px; }
  .actions { margin-top: 14px; display: flex; justify-content: flex-end; gap: 10px; }
</style>
