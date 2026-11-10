<script lang="ts">
 import { get } from 'svelte/store';
 import { onDestroy } from 'svelte';
 import { EventsOn } from '../wailsjs/runtime/runtime';
 import Titlebar from './features/titlebar/Titlebar.svelte';
 import Timeline from './features/timeline/Timeline.svelte';
 import Overlay from './features/overlay/Overlay.svelte';
 import Settings from './features/settings/Settings.svelte';
 import Toasts from './ui/toast/Toasts.svelte';
 import { onSnapshotCreated, refresh } from './lib/apps';
 import { pushToast, toastError } from './lib/toasts';
 import { loadPreferences } from './lib/preferences';
 import { visible, setVisibility, watchVisibility } from './lib/visibility';
 import { overlayRequest, requestOverlay, trackingState, view, settingsOpen, appPaused } from './lib/ui';
 import { t } from './i18n';
 import type { OverlayRequest, RestoreProgressEvent, TrackingStateChangedEvent } from './lib/types';
 const off=[
  EventsOn('onSnapshotCreated',onSnapshotCreated),
  EventsOn('onRestoreProgress',(p:RestoreProgressEvent)=>{if(p.stage==='done')pushToast('success',get(t)('restore.done',p.params)+(p.params?.relaunched?' '+get(t)('restore.relaunched'):''));}),
  EventsOn('onTrackingStateChanged',(p:TrackingStateChangedEvent)=>{
   if(p.appID){appPaused.update(all=>({...all,[p.appID!]:p.state==='paused'}));return;}
   if(p.state==='active'||p.state==='paused'||p.state==='error')trackingState.set(p.state);
  }),
  EventsOn('onOverlayOpenRequested',(p:OverlayRequest)=>{settingsOpen.set(false);requestOverlay(p?.appID||undefined);}),
  EventsOn('onShowTimelineRequested',()=>{view.set('timeline');setVisibility(true);void refresh();}),
  EventsOn('onShowSettingsRequested',()=>{view.set('timeline');settingsOpen.set(true);setVisibility(true);}),
  EventsOn('onWindowVisibilityChanged',(shown:boolean)=>setVisibility(shown)),
  watchVisibility(),
 ];
 let wasVisible=false;
 off.push(visible.subscribe(shown=>{if(shown&&!wasVisible)void refresh();wasVisible=shown;}));
 onDestroy(()=>off.forEach(f=>f()));
 void loadPreferences().catch(e=>toastError('saveFailed',e));
</script>

{#if $visible}
 {#if $view==='overlay'}
  {#key $overlayRequest.seq}<Overlay requestedAppID={$overlayRequest.appID} />{/key}
 {:else}
  <div class="shell"><Titlebar /><Timeline /></div>
  {#if $settingsOpen}<Settings />{/if}
 {/if}
 <Toasts />
{/if}
<style>.shell { height: 100vh; display: flex; flex-direction: column; background-image: var(--background-image); background-size: cover; background-position: center; }</style>
