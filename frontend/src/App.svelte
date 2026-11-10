<script lang="ts">
  import { Tooltip } from 'bits-ui';
  import type { Component } from 'svelte';
  import { GetWindowVisible } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import AppList from './features/apps/AppList.svelte';
  import Overlay from './features/overlay/Overlay.svelte';
  import Timeline from './features/timeline/Timeline.svelte';
  import Titlebar from './features/titlebar/Titlebar.svelte';
  import { onSnapshotCreated, refresh } from './lib/apps';
  import { fade } from './lib/motion';
  import { loadPreferences } from './lib/preferences';
  import { onRestoreError } from './lib/restore';
  import { pushMessage, toastError } from './lib/toasts';
  import type { OverlayRequest, RestoreProgressEvent, TrackingStateChangedEvent } from './lib/types';
  import { appPaused, overlayRequest, requestOverlay, settingsOpen, trackingState, view } from './lib/ui';
  import { setVisibility, visible, watchVisibility } from './lib/visibility';
  import Toasts from './ui/toast/Toasts.svelte';

  // Settings are loaded on first open: they are not needed to show the history.
  let SettingsView = $state<Component | null>(null);
  $effect(() => {
    if (!$settingsOpen || SettingsView) return;
    import('./features/settings/Settings.svelte')
      .then((m) => (SettingsView = m.default))
      .catch((e) => {
        settingsOpen.set(false);
        toastError('unknownError', e);
      });
  });

  $effect(() => {
    const off = [
      EventsOn('onSnapshotCreated', onSnapshotCreated),
      EventsOn('onRestoreProgress', (p: RestoreProgressEvent) => {
        if (p.stage === 'done') pushMessage('success', p.params?.relaunched ? 'restore.doneRelaunched' : 'restore.done', p.params);
      }),
      EventsOn('onRestoreError', onRestoreError),
      EventsOn('onTrackingStateChanged', (p: TrackingStateChangedEvent) => {
        if (p.appID) appPaused.update((all) => ({ ...all, [p.appID!]: p.state === 'paused' }));
        else if (p.state === 'active' || p.state === 'paused' || p.state === 'error') trackingState.set(p.state);
      }),
      EventsOn('onOverlayOpenRequested', (p: OverlayRequest) => {
        settingsOpen.set(false);
        requestOverlay(p?.appID || undefined);
      }),
      EventsOn('onShowTimelineRequested', () => {
        view.set('timeline');
        setVisibility(true);
        void refresh();
      }),
      EventsOn('onShowSettingsRequested', () => {
        view.set('timeline');
        settingsOpen.set(true);
        setVisibility(true);
      }),
      EventsOn('onWindowVisibilityChanged', (shown: boolean) => setVisibility(shown)),
      watchVisibility(),
    ];
    // Reload whenever the window comes back on screen.
    let was = false;
    off.push(
      visible.subscribe((shown) => {
        if (shown && !was) void refresh();
        was = shown;
      }),
    );
    GetWindowVisible()
      .then(setVisibility)
      .catch(() => setVisibility(true));
    loadPreferences().catch((e) => toastError('settings.read', e));
    return () => off.forEach((f) => f());
  });
</script>

<Tooltip.Provider>
  {#if $visible}
    {#if $view === 'overlay'}
      {#key $overlayRequest.seq}<Overlay requestedAppID={$overlayRequest.appID} />{/key}
    {:else}
      <div class="shell" in:fade inert={$settingsOpen}>
        <Titlebar />
        <div class="body">
          <AppList />
          <Timeline />
        </div>
      </div>
      {#if SettingsView}<SettingsView />{/if}
    {/if}
    <Toasts />
  {/if}
</Tooltip.Provider>

<style>
  .shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
    /* a top scrim keeps the title bar readable over a custom picture */
    background-image: linear-gradient(#000a, #0000 120px), var(--background-image);
    background-position: center;
    background-size: cover;
  }
  .body {
    flex: 1;
    min-height: 0;
    display: flex;
    gap: 4px;
    padding: 0 10px 10px;
  }
</style>
