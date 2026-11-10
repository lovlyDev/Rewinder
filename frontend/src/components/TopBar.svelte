<script lang="ts">
  import { trackingState, view } from '../lib/ui';
</script>

<header class="topbar">
  <div class="brand">
    <div class="dot {$trackingState}" title="Tracking: {$trackingState}"></div>
    <div class="name">Rewinder</div>
  </div>
  <nav class="actions">
    <button class="btn ghost" class:on={$view === 'timeline'} on:click={() => view.set('timeline')}>Timeline</button>
    <button class="btn ghost" class:on={$view === 'settings'} on:click={() => view.set('settings')}>Settings</button>
  </nav>
</header>

<style>
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    background: var(--panel);
  }
  .brand { display: flex; gap: 10px; align-items: center; }
  .name { font-weight: 700; }
  .dot { width: 10px; height: 10px; border-radius: 50%; background: var(--ok); }
  .dot.paused { background: var(--warn); }
  .dot.error { background: var(--bad); }
  .actions { display: flex; gap: 8px; }
  .on { border-color: var(--accent); }
</style>
