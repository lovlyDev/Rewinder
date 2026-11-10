<script lang="ts">
 import { onDestroy } from 'svelte';
 import { GetWindowVisible } from '../wailsjs/go/main/App';
 import { EventsOn } from '../wailsjs/runtime/runtime';
 import Titlebar from './features/titlebar/Titlebar.svelte';
 import Timeline from './features/timeline/Timeline.svelte';
 import Overlay from './features/overlay/Overlay.svelte';
 import Toasts from './ui/toast/Toasts.svelte';
 import { onSnapshotCreated, refresh } from './lib/apps';
 import { pushMessage, toastError } from './lib/toasts';
 import { loadPreferences } from './lib/preferences';
 import { visible, setVisibility, watchVisibility } from './lib/visibility';
 import { overlayRequest, requestOverlay, trackingState, view, settingsOpen, appPaused } from './lib/ui';
 import { onRestoreError } from './lib/restore';
 import { fade } from './lib/motion';
 import type { OverlayRequest, RestoreProgressEvent, TrackingStateChangedEvent } from './lib/types';
 let SettingsComponent: typeof import('./features/settings/Settings.svelte').default | null = null;
 let loadingSettings = false;
 $: if($settingsOpen && !SettingsComponent && !loadingSettings) {loadingSettings=true;void import('./features/settings/Settings.svelte').then(module=>SettingsComponent=module.default).catch(e=>{settingsOpen.set(false);toastError('saveFailed',e);}).finally(()=>loadingSettings=false);}
 function inert(node:HTMLElement,disabled:boolean) {const update=(v:boolean)=>{if(v)node.setAttribute('inert','');else node.removeAttribute('inert');};update(disabled);return {update};}
 const off=[
  EventsOn('onSnapshotCreated',onSnapshotCreated),
  EventsOn('onRestoreProgress',(p:RestoreProgressEvent)=>{if(p.stage==='done')pushMessage('success',p.params?.relaunched?'restore.doneRelaunched':'restore.done',p.params);}),
  EventsOn('onRestoreError',onRestoreError),
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
 void GetWindowVisible().then(setVisibility).catch(e=>toastError('saveFailed',e));
 void loadPreferences().catch(e=>toastError('saveFailed',e));
</script>

{#if $visible}
 {#if $view==='overlay'}
  {#key $overlayRequest.seq}<Overlay requestedAppID={$overlayRequest.appID} />{/key}
 {:else}
  <div class="shell" use:inert={$settingsOpen} in:fade><Titlebar /><Timeline /></div>
  {#if $settingsOpen && SettingsComponent}<svelte:component this={SettingsComponent} />{/if}
 {/if}
 <Toasts />
{/if}
<style>.shell { height: 100vh; display: flex; flex-direction: column; background-image: var(--background-image); background-size: cover; background-position: center; }</style>
