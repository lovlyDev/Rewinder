<script lang="ts">
 import { PauseTracking, ResumeTracking } from '../../../wailsjs/go/main/App';
 import Apps from '../apps/Apps.svelte';
 import { apps, selectedAppID, timeline } from '../../lib/apps';
 import { appPaused } from '../../lib/ui';
 import { restore, restoring } from '../../lib/restore';
 import { when } from '../../lib/format';
 import { t } from '../../i18n';
 import Button from '../../ui/button/Button.svelte';
 import IconButton from '../../ui/button/IconButton.svelte';
 import Icon from '../../ui/icon/Icon.svelte';
 import Scroll from '../../ui/scroll/Scroll.svelte';
 import Checkbox from '../../ui/checkbox/Checkbox.svelte';
 import { reveal } from '../../lib/motion';
 let limit=30;let important=false;let selected: string|null=null;
 $: $selectedAppID,limit=30;
 $: filtered=important?$timeline.filter(s=>s.windowsCount>0||s.filesAdded>0||s.filesRemoved>0):$timeline;
 $: if(!filtered.some(s=>s.snapshotID===selected))selected=filtered[0]?.snapshotID??null;
 $: app=$apps.find(a=>a.appID===$selectedAppID);
 $: paused=!!($selectedAppID && $appPaused[$selectedAppID]);
</script>
<div class="layout">
 <Apps />
 <main>
  <header><div><span class="eyebrow">{$t('timeline')}</span><h1>{app?.name??'Rewinder'}</h1></div><IconButton name={paused?'play':'pause'} label={$t(paused?'resumeApp':'pauseApp')} disabled={!app} on:click={() => paused?ResumeTracking($selectedAppID):PauseTracking($selectedAppID)} /></header>
  <div class="controls"><Checkbox label={$t('important')} bind:checked={important} /><small>{$t('snapshots',{n:filtered.length})}</small></div>
  <Scroll label={$t('timeline')} on:more={() => limit+=30}>
   <div class="list">
    {#each filtered.slice(0,limit) as snapshot,i (snapshot.snapshotID)}
     <button class="snapshot" aria-pressed={selected===snapshot.snapshotID} on:click={() => selected=snapshot.snapshotID} in:reveal>
      <span class="time-icon"><Icon name="clock-counter-clockwise" filled={selected===snapshot.snapshotID} size={22} /></span>
      <span class="details"><strong>{$t('windows',{n:snapshot.windowsCount})}</strong><small>{$t('files',{added:snapshot.filesAdded,removed:snapshot.filesRemoved})}</small></span>
      <span class="time"><time>{when(snapshot.timestampUTC)}</time>{#if i===0}<small>{$t('current')}</small>{/if}</span>
     </button>
    {:else}<div class="empty"><Icon name="clock-counter-clockwise" size={36} /><p>{$t(app?'noSnapshots':'selectApp')}</p></div>{/each}
   </div>
   {#if limit<filtered.length}<div class="more"><Button on:click={() => limit+=30}>{$t('more')}</Button></div>{/if}
  </Scroll>
  <footer><span>{$t('privacy')}</span><Button primary icon="arrow-counter-clockwise" disabled={!selected||$restoring} on:click={() => $selectedAppID&&selected&&restore($selectedAppID,selected)}>{$t('restore')}</Button></footer>
 </main>
</div>
<style>
 .layout { flex: 1; min-height: 0; display: flex; gap: 12px; padding: 14px; }main { flex: 1; min-width: 0; display: flex; flex-direction: column; padding: 12px; gap: 16px; }
 header { display: flex; justify-content: space-between; align-items: center; padding: 8px 4px; }.eyebrow { color: var(--muted); font-size: 12px; }h1 { font-size: 26px; font-weight: 600; margin-top: 4px; letter-spacing: -.7px; }
 .controls { display: flex; align-items: center; justify-content: space-between; padding: 0 4px; }small { font-size: 11px; color: var(--muted); }.list { display: grid; gap: 7px; }
 .snapshot { display: flex; align-items: center; gap: 14px; width: 100%; min-height: var(--row); text-align: left; padding: 14px 16px; border-radius: var(--radius); background: var(--surface); transition: transform var(--quick) var(--ease); }.snapshot:hover { background: var(--raised); }.snapshot:active { transform: scale(.97); }.snapshot[aria-pressed="true"] { background: var(--accent-soft); }
 .time-icon { color: var(--muted); }.snapshot[aria-pressed="true"] .time-icon { color: var(--accent); }.details { display: grid; gap: 4px; flex: 1; }strong { font-size: 13px; font-weight: 500; }.time { display: grid; justify-items: end; gap: 4px; }time { font-size: 12px; font-variant-numeric: tabular-nums; }
 .empty { color: var(--muted); display: grid; justify-items: center; gap: 14px; padding: 80px 0; }.more { text-align: center; padding: 12px; }footer { display: flex; align-items: center; justify-content: space-between; gap: 16px; }footer>span { color: var(--muted); font-size: 11px; max-width: 260px; }
</style>
