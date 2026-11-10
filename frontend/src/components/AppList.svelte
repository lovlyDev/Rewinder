<script lang="ts">
  import { apps, refresh, selectApp, selectedAppID } from '../lib/apps';
  import { when } from '../lib/format';
</script>

<aside class="sidebar" aria-label="Applications">
  <div class="header">
    <div class="h">Apps</div>
    <button class="btn small" on:click={refresh}>Refresh</button>
  </div>
  <div class="cards">
    {#each $apps as a (a.appID)}
      <button class="card" class:active={$selectedAppID === a.appID} on:click={() => selectApp(a.appID)}>
        <div class="title">{a.name}</div>
        <div class="meta">
          <span>{a.snapshotCount} snapshots</span>
          <span>{when(a.lastActivityUTC)}</span>
        </div>
      </button>
    {:else}
      <div class="empty">No tracked apps yet. Use your apps and snapshots will appear.</div>
    {/each}
  </div>
</aside>

<style>
  .sidebar {
    border-right: 1px solid var(--border);
    padding: 12px;
    min-height: 0;
    overflow: auto;
  }
  .header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
  .cards { display: flex; flex-direction: column; gap: 8px; }
  .card {
    text-align: left;
    padding: 10px;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: var(--panel);
    cursor: pointer;
  }
  .card.active { border-color: var(--accent); background: var(--accent-soft); }
  .title { font-weight: 700; }
  .meta { display: flex; justify-content: space-between; margin-top: 6px; color: var(--muted); font-size: 12px; }
</style>
