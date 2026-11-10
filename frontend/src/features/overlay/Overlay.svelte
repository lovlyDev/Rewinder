<script lang="ts">
 import { onMount, tick } from 'svelte';
 import { HideOverlay, LogUI, ShowTimelineWindow } from '../../../wailsjs/go/main/App';
 import { loadApps, loadTimeline } from '../../lib/apps';
 import { restore } from '../../lib/restore';
 import { when } from '../../lib/format';
 import { t } from '../../i18n';
 import type { SnapshotMeta } from '../../lib/types';
 import Icon from '../../ui/icon/Icon.svelte';
 import IconButton from '../../ui/button/IconButton.svelte';
 import Button from '../../ui/button/Button.svelte';
 export let requestedAppID: string|undefined;
 let el: HTMLElement;let appID: string|undefined;let appName='';let snapshots: SnapshotMeta[]=[];let index=0;let alive=true;let busy=false;
 $: current=snapshots[index];
 async function load() {
  const list=await loadApps();const app=list.find(a=>a.appID===requestedAppID)??list[0];
  const history=app?await loadTimeline(app.appID):[];if(!alive)return;
  appID=app?.appID;appName=app?.name??'';snapshots=history;index=Math.min(1,Math.max(0,history.length-1));
  await tick();window.focus();el?.focus();void LogUI(`overlay open app=${appID} requested=${requestedAppID??''} snapshots=${history.length} index=${index}`);
 }
 async function close() {if(busy)return;busy=true;await HideOverlay();}
 async function restoreCurrent() {
  if(busy)return;busy=true;const id=appID;const snapshot=current;
  void LogUI(`overlay restore app=${id} index=${index} snapshot=${snapshot?.snapshotID??'none'}`);
  await HideOverlay();if(id&&snapshot)void restore(id,snapshot.snapshotID);
 }
 function step(d: number) { index=Math.max(0,Math.min(snapshots.length-1,index+d)); }
 function keyboard(e: KeyboardEvent) {
  if(e.key==='Enter' && e.target instanceof HTMLButtonElement)return;
  const actions: Record<string,()=>void>={Escape:()=>void close(),Enter:()=>void restoreCurrent(),ArrowLeft:()=>step(-1),ArrowRight:()=>step(1)};
  if(!actions[e.key])return;e.preventDefault();void LogUI(`overlay key ${e.key}`);actions[e.key]();
 }
 function wheel(e: WheelEvent) {e.preventDefault();if(e.deltaY)step(e.deltaY>0?1:-1);}
 onMount(()=>{window.addEventListener('keydown',keyboard,true);window.addEventListener('wheel',wheel,{passive:false});void load();return()=>{alive=false;window.removeEventListener('keydown',keyboard,true);window.removeEventListener('wheel',wheel);};});
</script>
<section class="overlay" role="dialog" aria-label={$t('overlay')} tabindex="-1" bind:this={el}>
 <header><span class="brand"><Icon name="clock-counter-clockwise" /><strong>{appName||$t('overlay')}</strong></span><span class="drag" /><Button icon="app-window" on:click={() => ShowTimelineWindow()}>{$t('openApp')}</Button><IconButton name="x" label={$t('close')} on:click={close} /></header>
 {#if current}
  <div class="scale-row"><IconButton name="arrow-left" label={$t('timeline')} disabled={index===0} on:click={() => step(-1)} />
   <div class="scale"><div class="marks" /><div class="marker-track"><span class="marker" style="transform:translateX({snapshots.length>1?index/(snapshots.length-1)*100:0}%)"><i /></span></div><input type="range" min="0" max={snapshots.length-1} bind:value={index} aria-label={$t('timeline')} /></div>
   <IconButton name="arrow-right" label={$t('timeline')} disabled={index===snapshots.length-1} on:click={() => step(1)} /></div>
  <div class="preview"><time>{when(current.timestampUTC)}</time><span>{$t('windows',{n:current.windowsCount})}</span>{#if index===0}<small>{$t('current')}</small>{/if}</div>
 {:else}<div class="empty"><Icon name="clock-counter-clockwise" />{$t('noSnapshots')}</div>{/if}
 <footer><span>{$t('overlayHint')}</span><Button primary icon="arrow-counter-clockwise" disabled={!current||busy} on:click={restoreCurrent}>{$t('restore')}</Button></footer>
</section>
<style>
 .overlay { height: 100vh; padding: 14px 22px; background: var(--surface); display: flex; flex-direction: column; border-radius: var(--radius); outline: none; }header { display: flex; align-items: center; gap: 8px; --wails-draggable: drag; }.brand { display: flex; align-items: center; gap: 10px; }.brand :global(svg) { color: var(--accent); }strong { font-weight: 600; }.drag { flex: 1; align-self: stretch; }
 .scale-row { display: flex; align-items: center; gap: 14px; margin-top: 12px; }.scale { flex: 1; position: relative; height: 34px; }.marks { position: absolute; inset: 8px 0; background: repeating-linear-gradient(90deg,var(--muted) 0px,var(--muted) 1px,transparent 1px,transparent 12px); opacity: .35; }.marker-track { position: absolute; inset: 0 1px; pointer-events: none; }.marker { display: block; width: 100%; height: 100%; transition: transform var(--quick) var(--ease); }.marker i { display: block; width: 3px; height: 34px; background: var(--accent); border-radius: 999px; transform: translateX(-50%); }.scale input { position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }.scale:focus-within { outline: 2px solid var(--accent); outline-offset: 4px; border-radius: var(--small-radius); }
 .preview { display: flex; align-items: center; justify-content: center; gap: 14px; margin-top: 8px; }.preview span,.preview small { color: var(--muted); font-size: 12px; }time { font-variant-numeric: tabular-nums; font-weight: 500; }footer { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: auto; padding-top: 10px; }footer>span { font-size: 11px; color: var(--muted); }.empty { display: flex; flex: 1; align-items: center; justify-content: center; gap: 10px; color: var(--muted); }
</style>
