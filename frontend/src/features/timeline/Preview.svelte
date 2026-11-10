<script lang="ts">
  import { language, t } from '../../i18n';
  import { day, clockSeconds } from '../../lib/format';
  import { snapshotLayout, type SnapshotLayout } from '../../lib/preview';
  import type { SnapshotMeta } from '../../lib/types';
  import Minimap from '../../ui/minimap/Minimap.svelte';

  type Props = { appID: string; snapshot: SnapshotMeta };
  let { appID, snapshot }: Props = $props();

  let layout = $state<SnapshotLayout | null>(null);
  $effect(() => {
    const id = snapshot.snapshotID;
    void snapshotLayout(appID, id).then((l) => {
      if (id === snapshot.snapshotID) layout = l;
    });
  });

  // Topmost first for the list (the layout is bottom-most first for drawing).
  const windows = $derived(layout ? [...layout.windows].reverse() : []);
</script>

<aside class="preview">
  <div class="stage"><Minimap {layout} label={$t('previewLabel')} /></div>
  <div class="when">
    <strong class="tabular">{clockSeconds(snapshot.timestampUTC, $language)}</strong>
    <span>{day(snapshot.timestampUTC, $language)}</span>
  </div>
  <ul>
    {#each windows.slice(0, 5) as w, i (i)}
      <li class:dim={w.minimized}>
        <span class="dot" class:fg={w.foreground}></span>
        <span class="title">{w.title || $t('untitled')}</span>
        {#if w.minimized}<small>{$t('minimizedShort')}</small>{:else if w.maximized}<small>{$t('maximizedShort')}</small>{/if}
      </li>
    {/each}
    {#if windows.length > 5}<li class="more">+{windows.length - 5}</li>{/if}
  </ul>
</aside>

<style>
  .preview {
    display: flex;
    flex-direction: column;
    gap: 14px;
    width: 250px;
    flex: none;
    padding: 14px;
    border-radius: var(--r-lg);
    background: var(--raised);
    overflow: hidden;
  }
  .stage {
    display: grid;
    place-items: center;
    height: 128px;
    padding: 10px;
    border-radius: var(--r-sm);
    background: var(--surface);
  }
  .when {
    display: grid;
    gap: 1px;
  }
  .when strong {
    font-size: 20px;
    font-weight: 650;
    letter-spacing: -0.02em;
  }
  .when span {
    color: var(--muted);
    font-size: 12px;
  }
  ul {
    display: grid;
    gap: 6px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  li {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    font-size: 12px;
  }
  li.dim {
    color: var(--faint);
  }
  .dot {
    width: 6px;
    height: 6px;
    flex: none;
    border-radius: 999px;
    background: var(--faint);
  }
  .dot.fg {
    background: var(--accent);
  }
  .title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  small {
    color: var(--faint);
    font-size: 11px;
  }
  .more {
    color: var(--faint);
    padding-left: 14px;
  }
</style>
