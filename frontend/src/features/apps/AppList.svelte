<script lang="ts">
  import { language, t } from '../../i18n';
  import { apps, selectApp, selectedAppID } from '../../lib/apps';
  import { ago, appName } from '../../lib/format';
  import { now } from '../../lib/now';
  import { appPaused } from '../../lib/ui';
  import AppAvatar from '../../ui/avatar/AppAvatar.svelte';
  import Empty from '../../ui/empty/Empty.svelte';
  import Icon from '../../ui/icon/Icon.svelte';
  import SearchField from '../../ui/search/SearchField.svelte';

  let query = $state('');
  let list: HTMLDivElement;

  const filtered = $derived.by(() => {
    const q = query.trim().toLocaleLowerCase();
    return q ? $apps.filter((a) => a.name.toLocaleLowerCase().includes(q)) : $apps;
  });
  const selectedIndex = $derived(filtered.findIndex((a) => a.appID === $selectedAppID));

  // ↑/↓ move through the list like a native list box.
  function keyboard(e: KeyboardEvent) {
    const d = e.key === 'ArrowDown' ? 1 : e.key === 'ArrowUp' ? -1 : 0;
    if (!d || !filtered.length) return;
    e.preventDefault();
    const next = filtered[Math.max(0, Math.min(filtered.length - 1, selectedIndex + d))];
    void selectApp(next.appID);
    list.querySelector<HTMLElement>(`[data-id="${CSS.escape(next.appID)}"]`)?.focus();
  }
</script>

<div class="pane">
  <div class="head">
    <h2>{$t('apps')}</h2>
    <span class="count tabular">{$apps.length}</span>
  </div>
  <SearchField bind:value={query} placeholder={$t('search')} clearLabel={$t('clear')} />

  <div class="list" bind:this={list} role="listbox" aria-label={$t('apps')} tabindex="-1" onkeydown={keyboard}>
    {#if selectedIndex >= 0}
      <!-- One highlight that glides between rows instead of each row lighting up. -->
      <span class="highlight" style="--i:{selectedIndex}"></span>
    {/if}
    {#each filtered as app (app.appID)}
      {@const selected = app.appID === $selectedAppID}
      <button
        class="app"
        role="option"
        aria-selected={selected}
        data-id={app.appID}
        tabindex={selected || (selectedIndex < 0 && app === filtered[0]) ? 0 : -1}
        onclick={() => selectApp(app.appID)}
      >
        <AppAvatar appID={app.appID} name={app.name} size={30} />
        <span class="name">
          <strong>{appName(app.name)}</strong>
          <small>
            {$t('snapshots', { n: app.snapshotCount })}
            <span class="sep">·</span>
            {ago(app.lastActivityUTC, $language, $now)}
          </small>
        </span>
        {#if app.excluded}<span class="paused" title={$t('notTracked')}><Icon name="eyeSlash" filled size={12} /></span>
        {:else if $appPaused[app.appID]}<span class="paused"><Icon name="pause" filled size={12} /></span>{/if}
      </button>
    {:else}
      {#if query}
        <Empty icon="magnifyingGlass" title={$t('noResults')} />
      {:else}
        <Empty icon="appWindow" title={$t('noApps')} hint={$t('noAppsHint')} />
      {/if}
    {/each}
  </div>
</div>

<style>
  .pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 14px;
    min-height: 0;
  }
  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    padding: 0 4px;
  }
  h2 {
    font-size: 13px;
    font-weight: 650;
  }
  .count {
    color: var(--faint);
    font-size: 12px;
  }
  .list {
    --gap: 4px;
    position: relative;
    display: flex;
    flex-direction: column;
    gap: var(--gap);
    min-height: 0;
    margin-right: -6px;
    padding-right: 4px;
    overflow-y: auto;
    outline: none;
  }
  .highlight {
    position: absolute;
    left: 0;
    right: 4px;
    top: 0;
    height: var(--row);
    border-radius: var(--r-sm);
    background: var(--raised);
    transform: translateY(calc(var(--i) * (var(--row) + var(--gap))));
    transition: transform var(--duration) var(--ease);
    pointer-events: none;
  }
  .highlight::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    width: 3px;
    height: 18px;
    margin-top: -9px;
    border-radius: 0 3px 3px 0;
    background: var(--accent);
  }
  .app {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    flex: none;
    height: var(--row);
    padding: 0 12px;
    border-radius: var(--r-sm);
    text-align: left;
    transition:
      background-color var(--quick) var(--ease),
      transform var(--quick) var(--ease);
  }
  .app:hover:not([aria-selected='true']) {
    background: var(--hover);
  }
  .app:active {
    transform: scale(0.98);
  }
  .name {
    display: grid;
    gap: 2px;
    flex: 1;
    min-width: 0;
  }
  strong {
    overflow: hidden;
    font-weight: 550;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  small {
    overflow: hidden;
    color: var(--faint);
    font-size: 11px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  [aria-selected='true'] small {
    color: var(--muted);
  }
  .sep {
    margin: 0 3px;
  }
  .paused {
    display: grid;
    place-items: center;
    width: 20px;
    height: 20px;
    border-radius: 999px;
    background: var(--hover);
    color: var(--muted);
  }
</style>
