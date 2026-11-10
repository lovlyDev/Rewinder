<script lang="ts">
  import { OpenOverlay, PauseTracking, ResumeTracking } from '../../../wailsjs/go/main/App';
  import { language, t } from '../../i18n';
  import { apps, selectedAppID, timeline } from '../../lib/apps';
  import { appName, clock, day, dayKey } from '../../lib/format';
  import { restore, restoring } from '../../lib/restore';
  import type { SnapshotMeta } from '../../lib/types';
  import { appPaused } from '../../lib/ui';
  import AppAvatar from '../../ui/avatar/AppAvatar.svelte';
  import Button from '../../ui/button/Button.svelte';
  import IconButton from '../../ui/button/IconButton.svelte';
  import Empty from '../../ui/empty/Empty.svelte';
  import Kbd from '../../ui/kbd/Kbd.svelte';
  import Segmented from '../../ui/segmented/Segmented.svelte';
  import SnapshotRow from './SnapshotRow.svelte';

  const PAGE = 60;
  let filter = $state<'all' | 'important'>('all');
  let limit = $state(PAGE);
  let selected = $state<string | null>(null);
  let list = $state<HTMLDivElement>();
  let sentinel = $state<HTMLDivElement>();

  const app = $derived($apps.find((a) => a.appID === $selectedAppID));
  const paused = $derived(!!($selectedAppID && $appPaused[$selectedAppID]));
  const important = (s: SnapshotMeta) => s.filesAdded > 0 || s.filesRemoved > 0;
  const filtered = $derived(filter === 'important' ? $timeline.filter(important) : $timeline);
  const shown = $derived(filtered.slice(0, limit));
  const newest = $derived($timeline[0]?.snapshotID);
  const chosen = $derived(filtered.find((s) => s.snapshotID === selected));

  // Days, newest first; rows keep the list's order.
  const days = $derived.by(() => {
    const groups: { key: number; items: SnapshotMeta[] }[] = [];
    for (const s of shown) {
      const key = dayKey(s.timestampUTC);
      if (groups.at(-1)?.key !== key) groups.push({ key, items: [] });
      groups.at(-1)!.items.push(s);
    }
    return groups;
  });

  // New app: first page again, and "one step back" (what undo means) preselected.
  $effect(() => {
    void $selectedAppID;
    limit = PAGE;
    selected = null;
    list?.scrollTo({ top: 0 });
  });
  $effect(() => {
    if (!filtered.some((s) => s.snapshotID === selected)) selected = (filtered[1] ?? filtered[0])?.snapshotID ?? null;
  });

  // Load more rows as the end of the list scrolls into view.
  $effect(() => {
    if (!sentinel || !list) return;
    const io = new IntersectionObserver((e) => e[0].isIntersecting && (limit += PAGE), { root: list, rootMargin: '300px' });
    io.observe(sentinel);
    return () => io.disconnect();
  });

  function go(id: string | undefined) {
    if (!id) return;
    selected = id;
    list?.querySelector<HTMLElement>(`[data-id="${CSS.escape(id)}"]`)?.focus({ preventScroll: false });
  }
  function keyboard(e: KeyboardEvent) {
    const i = shown.findIndex((s) => s.snapshotID === selected);
    if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
      e.preventDefault();
      go(shown[Math.max(0, Math.min(shown.length - 1, i + (e.key === 'ArrowDown' ? 1 : -1)))]?.snapshotID);
    } else if (e.key === 'Home' || e.key === 'End') {
      e.preventDefault();
      go((e.key === 'Home' ? shown[0] : shown.at(-1))?.snapshotID);
    } else if (e.key === 'Enter' && chosen) {
      e.preventDefault();
      doRestore();
    }
  }
  function doRestore() {
    if ($selectedAppID && chosen) void restore($selectedAppID, chosen.snapshotID);
  }
</script>

<main>
  {#if app}
    <header>
      <AppAvatar appID={app.appID} name={app.name} size={44} />
      <div class="title">
        <h1>{appName(app.name)}</h1>
        <p title={app.executablePath}>{app.executablePath}</p>
      </div>
      <div class="actions">
        <IconButton
          name={paused ? 'play' : 'pause'}
          label={$t(paused ? 'resumeApp' : 'pauseApp')}
          active={paused}
          onclick={() => (paused ? ResumeTracking($selectedAppID) : PauseTracking($selectedAppID))}
        />
        <Button variant="soft" icon="lightning" onclick={() => OpenOverlay(app.appID)}>{$t('quickRestore')}</Button>
      </div>
    </header>

    <div class="toolbar">
      <Segmented
        label={$t('filter')}
        bind:value={filter}
        options={[
          { value: 'all', label: $t('all') },
          { value: 'important', label: $t('withFiles') },
        ]}
      />
      <span class="count tabular">{$t('snapshots', { n: filtered.length })}</span>
    </div>

    <div class="list" bind:this={list} role="listbox" aria-label={$t('timeline')} tabindex="-1" onkeydown={keyboard}>
      {#each days as group (group.key)}
        <section>
          <h3>{day(group.key, $language)}</h3>
          {#each group.items as snapshot, i (snapshot.snapshotID)}
            <SnapshotRow
              {snapshot}
              selected={snapshot.snapshotID === selected}
              current={snapshot.snapshotID === newest}
              last={i === group.items.length - 1}
              onselect={() => (selected = snapshot.snapshotID)}
              onrestore={() => {
                selected = snapshot.snapshotID;
                doRestore();
              }}
            />
          {/each}
        </section>
      {:else}
        <Empty icon="clock" title={$t('noSnapshots')} hint={$t(filter === 'all' ? 'noSnapshotsHint' : 'noImportantHint')} />
      {/each}
      {#if limit < filtered.length}<div class="sentinel" bind:this={sentinel}></div>{/if}
    </div>

    <footer>
      <div class="hint">
        {#if chosen}
          <span>{$t('restoreTo', { time: clock(chosen.timestampUTC, $language) })}</span>
          <Kbd keys={['Enter']} />
        {/if}
      </div>
      <Button variant="primary" size="lg" icon="restore" disabled={!chosen || $restoring} onclick={doRestore}>
        {$t('restore')}
      </Button>
    </footer>
  {:else}
    <div class="center"><Empty icon="clock" title={$t('selectApp')} hint={$t('selectAppHint')} /></div>
  {/if}
</main>

<style>
  main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 18px 8px 8px 12px;
    border-radius: var(--r-lg);
    background: color-mix(in srgb, var(--surface) 84%, transparent);
  }
  header {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 0 8px;
  }
  .title {
    flex: 1;
    min-width: 0;
  }
  h1 {
    font-size: 24px;
    font-weight: 650;
    letter-spacing: -0.025em;
    line-height: 1.15;
  }
  .title p {
    overflow: hidden;
    color: var(--faint);
    font-size: 11px;
    text-overflow: ellipsis;
    white-space: nowrap;
    user-select: text;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 8px;
  }
  .count {
    color: var(--faint);
    font-size: 12px;
  }
  .list {
    flex: 1;
    min-height: 0;
    padding: 0 4px 0 0;
    overflow-y: auto;
    outline: none;
  }
  section + section {
    margin-top: 10px;
  }
  h3 {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 6px 12px;
    border-radius: var(--r-xs);
    background: color-mix(in srgb, var(--surface) 96%, transparent);
    color: var(--muted);
    font-size: 11px;
    font-weight: 650;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .sentinel {
    height: 1px;
  }
  footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 14px;
    padding: 12px 12px 12px 20px;
    border-radius: var(--r-lg);
    background: var(--raised);
  }
  .hint {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-right: auto;
    color: var(--muted);
  }
  .center {
    flex: 1;
    display: grid;
    place-items: center;
  }
</style>
