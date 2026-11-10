<script lang="ts">
  import { onMount } from 'svelte';
  import { ExitApp, GetAutostart, SetAutostart } from '../../wailsjs/go/main/App.js';
  import { pushToast, toastError } from '../lib/toasts';
  import { trackingState } from '../lib/ui';

  let autostart = false;

  onMount(async () => {
    try {
      autostart = await GetAutostart();
    } catch (e) {
      toastError('Could not load settings', e);
    }
  });

  async function toggleAutostart() {
    try {
      await SetAutostart(!autostart);
      autostart = !autostart;
      pushToast('success', autostart ? 'Autostart enabled' : 'Autostart disabled');
    } catch (e) {
      toastError('Could not change autostart', e);
    }
  }
</script>

<div class="settings">
  <div class="card">
    <div class="h">Settings</div>
    <div class="row"><div class="label">Tracking</div><div>{$trackingState}</div></div>
    <div class="row"><div class="label">Hotkey</div><div>Ctrl + Alt + Z</div></div>
    <div class="row">
      <div class="label">Autostart</div>
      <div>
        <button class="btn" class:primary={autostart} class:ghost={!autostart} on:click={toggleAutostart}>
          {autostart ? 'Enabled' : 'Disabled'}
        </button>
      </div>
    </div>
    <div class="row">
      <div class="label">Privacy</div>
      <div>Window layout, open folders and Office documents only. Stored locally, never sent anywhere.</div>
    </div>
    <div class="row">
      <div class="label">Application</div>
      <div><button class="btn ghost" on:click={() => ExitApp()}>Exit completely</button></div>
    </div>
  </div>
</div>

<style>
  .settings { padding: 20px; overflow: auto; }
  .card {
    margin: 0 auto;
    max-width: 680px;
    border: 1px solid var(--border);
    background: var(--panel);
    border-radius: 14px;
    padding: 14px;
  }
  .row {
    display: grid;
    grid-template-columns: 160px 1fr;
    gap: 12px;
    padding: 10px 0;
    border-top: 1px solid var(--border);
  }
  .h + .row { border-top: none; }
  .label { color: var(--muted); }
</style>
