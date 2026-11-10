<script lang="ts">
  import { onDestroy } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import Toasts from './components/Toasts.svelte';
  import TopBar from './components/TopBar.svelte';
  import { onSnapshotCreated, refresh } from './lib/apps';
  import { pushToast } from './lib/toasts';
  import { overlayRequest, requestOverlay, trackingState, view } from './lib/ui';
  import type { OverlayRequest, RestoreProgressEvent, TrackingStateChangedEvent } from './lib/types';
  import Overlay from './views/Overlay.svelte';
  import Settings from './views/Settings.svelte';
  import Timeline from './views/Timeline.svelte';

  // Backend events. Restore errors are shown by the caller of Restore, not here.
  const off = [
    EventsOn('onSnapshotCreated', onSnapshotCreated),
    EventsOn('onRestoreProgress', (p: RestoreProgressEvent) => {
      if (p.stage === 'done') pushToast('success', p.message);
    }),
    EventsOn('onTrackingStateChanged', (p: TrackingStateChangedEvent) => {
      if (p.state === 'active' || p.state === 'paused' || p.state === 'error') trackingState.set(p.state);
    }),
    EventsOn('onOverlayOpenRequested', (p: OverlayRequest) => requestOverlay(p?.appID || undefined)),
    EventsOn('onShowTimelineRequested', () => view.set('timeline')),
    EventsOn('onShowSettingsRequested', () => view.set('settings')),
  ];
  onDestroy(() => off.forEach(f => f()));

  void refresh();
</script>

{#if $view === 'overlay'}
  {#key $overlayRequest.seq}
    <Overlay requestedAppID={$overlayRequest.appID} />
  {/key}
{:else}
  <div class="shell">
    <TopBar />
    {#if $view === 'timeline'}
      <Timeline />
    {:else}
      <Settings />
    {/if}
  </div>
{/if}

<Toasts />

<style>
  .shell { height: 100vh; display: flex; flex-direction: column; }
</style>
