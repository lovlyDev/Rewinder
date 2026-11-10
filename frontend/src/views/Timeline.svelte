<script lang="ts">
  import { PauseTracking, ResumeTracking } from '../../wailsjs/go/main/App.js';
  import AppList from '../components/AppList.svelte';
  import SnapshotList from '../components/SnapshotList.svelte';
  import { selectedAppID, timeline } from '../lib/apps';
  import { restore } from '../lib/restore';
  import { requestOverlay } from '../lib/ui';
  import type { SnapshotMeta } from '../lib/types';

  const PAGE = 20;
  let limit = PAGE;
  let onlyImportant = true;
  let selectedID: string | null = null;

  const important = (s: SnapshotMeta) => s.windowsCount > 0 || s.filesAdded > 0 || s.filesRemoved > 0;

  // a different app starts from the first page again
  $: $selectedAppID, (limit = PAGE);
  $: filtered = onlyImportant ? $timeline.filter(important) : $timeline;
  $: visible = filtered.slice(0, limit);
  $: if (!visible.some(s => s.snapshotID === selectedID)) selectedID = visible[0]?.snapshotID ?? null;
</script>

<div class="layout">
  <AppList />

  <main class="main">
    <div class="header">
      <div class="h">Timeline</div>
      <div class="actions">
        <button class="btn ghost" disabled={!$selectedAppID} on:click={() => PauseTracking($selectedAppID)}>Pause app</button>
        <button class="btn ghost" disabled={!$selectedAppID} on:click={() => ResumeTracking($selectedAppID)}>Resume app</button>
        <button class="btn ghost" on:click={() => PauseTracking(null)}>Pause all</button>
        <button class="btn ghost" on:click={() => ResumeTracking(null)}>Resume all</button>
      </div>
    </div>

    <div class="body">
      {#if !$selectedAppID}
        <div class="empty">Select an app</div>
      {:else if $timeline.length === 0}
        <div class="empty">No snapshots for this app yet</div>
      {:else}
        <div class="controls">
          <label class="toggle">
            <input type="checkbox" bind:checked={onlyImportant} />
            <span>Only important changes</span>
          </label>
          <div>Showing {visible.length} of {filtered.length}</div>
        </div>
        <SnapshotList snapshots={visible} bind:selectedID />
        {#if filtered.length > visible.length}
          <button class="btn ghost more" on:click={() => (limit += PAGE)}>Show more</button>
        {/if}
        <div class="bottom">
          <button class="btn ghost" on:click={() => requestOverlay($selectedAppID ?? undefined)}>Quick overlay</button>
          <button
            class="btn primary"
            disabled={!selectedID}
            on:click={() => $selectedAppID && selectedID && restore($selectedAppID, selectedID)}
          >
            Restore selected
          </button>
        </div>
      {/if}
    </div>
  </main>
</div>

<style>
  .layout { flex: 1; display: grid; grid-template-columns: 320px 1fr; min-height: 0; }
  .main { min-height: 0; overflow: hidden; display: flex; flex-direction: column; }
  .header {
    padding: 12px 14px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--border);
  }
  .actions { display: flex; gap: 8px; flex-wrap: wrap; justify-content: flex-end; }
  .body { flex: 1; min-height: 0; display: flex; flex-direction: column; padding: 12px; }
  .controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--muted);
  }
  .toggle { display: inline-flex; align-items: center; gap: 8px; cursor: pointer; }
  .toggle input { accent-color: var(--accent); }
  .more { align-self: center; margin: 10px 0 2px; }
  .bottom {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
</style>
