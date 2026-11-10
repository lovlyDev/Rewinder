<script lang="ts">
  import { when, windows } from '../lib/format';
  import type { SnapshotMeta } from '../lib/types';

  export let snapshots: SnapshotMeta[];
  export let selectedID: string | null;
</script>

<div class="list" role="list">
  {#each snapshots as s (s.snapshotID)}
    <button class="snap" class:active={selectedID === s.snapshotID} on:click={() => (selectedID = s.snapshotID)}>
      <div class="when">{when(s.timestampUTC)}</div>
      <div class="meta">
        <span>{windows(s.windowsCount)}</span>
        <span>{s.filesAdded}+ / {s.filesRemoved}− files</span>
      </div>
    </button>
  {/each}
</div>

<style>
  .list {
    flex: 1;
    min-height: 0;
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .snap {
    text-align: left;
    padding: 10px;
    border-radius: 12px;
    border: 1px solid var(--border);
    background: var(--panel);
    cursor: pointer;
  }
  .snap.active { border-color: var(--accent); background: var(--accent-soft); }
  .when { font-weight: 700; }
  .meta { color: var(--muted); font-size: 12px; display: flex; gap: 12px; margin-top: 6px; }
</style>
