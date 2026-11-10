<script lang="ts">
 import { apps, selectApp, selectedAppID } from '../../lib/apps';
 import { t } from '../../i18n';
 import Icon from '../../ui/icon/Icon.svelte';
 import Input from '../../ui/input/Input.svelte';
 import Scroll from '../../ui/scroll/Scroll.svelte';
 import { reveal } from '../../lib/motion';
 let query='';let limit=40;
 $: query,limit=40;
 $: filtered=$apps.filter(a=>a.name.toLocaleLowerCase().includes(query.toLocaleLowerCase()));
</script>
<aside>
 <div class="heading"><h2>{$t('apps')}</h2><span>{$apps.length}</span></div>
 <div class="search"><Icon name="magnifying-glass" size={18} /><Input label={$t('search')} placeholder={$t('search')} bind:value={query} /></div>
 <Scroll label={$t('apps')} on:more={() => limit+=40}>
  <div class="list">
   {#each filtered.slice(0,limit) as app (app.appID)}
    <button class="app" aria-pressed={$selectedAppID===app.appID} on:click={() => selectApp(app.appID)} in:reveal>
     <span class="app-icon"><Icon name="app-window" filled={$selectedAppID===app.appID} size={22} /></span>
     <span class="name"><strong>{app.name}</strong><small>{$t('snapshots',{n:app.snapshotCount})}</small></span>
     {#if $selectedAppID===app.appID}<span class="selected" />{/if}
    </button>
   {:else}<div class="empty"><Icon name="app-window" size={28} /><p>{$t('noApps')}</p></div>{/each}
  </div>
 </Scroll>
</aside>
<style>
 aside { width: 260px; flex: none; display: flex; flex-direction: column; min-height: 0; gap: 16px; padding: var(--space); background: var(--surface); border-radius: var(--radius); }
 .heading { display: flex; align-items: center; justify-content: space-between; } h2 { font-size: 13px; font-weight: 600; }.heading>span { color: var(--muted); font-size: 12px; }
 .search { display: flex; align-items: center; position: relative; }.search :global(svg) { position: absolute; left: 10px; color: var(--muted); }.search :global(input) { padding-left: 35px; font-size: 12px; }
 .list { display: grid; gap: 5px; }.app { width: 100%; text-align: left; display: flex; align-items: center; gap: 11px; min-height: var(--row); padding: 10px; background: transparent; border-radius: var(--small-radius); transition: transform var(--quick) var(--ease); }
 .app:hover { background: var(--raised); }.app:active { transform: scale(.97); }.app[aria-pressed="true"] { background: var(--accent-soft); }
 .app-icon { color: var(--muted); }.app[aria-pressed="true"] .app-icon { color: var(--accent); }.name { display: grid; gap: 4px; flex: 1; min-width: 0; }strong { font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }small { color: var(--muted); font-size: 11px; }
 .selected { width: 5px; height: 5px; border-radius: 999px; background: var(--accent); }.empty { display: grid; justify-items: center; gap: 12px; padding: 40px 4px; color: var(--muted); text-align: center; font-size: 12px; }
</style>
