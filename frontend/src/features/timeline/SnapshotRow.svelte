<script lang="ts">
  import { language, t } from '../../i18n';
  import { ago, clock } from '../../lib/format';
  import { now } from '../../lib/now';
  import type { SnapshotMeta } from '../../lib/types';

  type Props = {
    snapshot: SnapshotMeta;
    selected: boolean;
    current: boolean;
    last: boolean;
    onselect: () => void;
    onrestore: () => void;
  };
  let { snapshot, selected, current, last, onselect, onrestore }: Props = $props();
</script>

<button
  class="row"
  class:last
  role="option"
  aria-selected={selected}
  data-id={snapshot.snapshotID}
  tabindex={selected ? 0 : -1}
  onclick={onselect}
  ondblclick={onrestore}
>
  <span class="rail"><span class="node" class:current></span></span>
  <span class="time">
    <strong class="tabular">{clock(snapshot.timestampUTC, $language)}</strong>
    <small>{ago(snapshot.timestampUTC, $language, $now)}</small>
  </span>
  <span class="meta">
    <span class="chip">{$t('windows', { n: snapshot.windowsCount })}</span>
    {#if snapshot.filesAdded}<span class="chip add">+{snapshot.filesAdded} {$t('filesShort')}</span>{/if}
    {#if snapshot.filesRemoved}<span class="chip del">−{snapshot.filesRemoved} {$t('filesShort')}</span>{/if}
  </span>
  {#if current}<span class="now">{$t('current')}</span>{/if}
</button>

<style>
  .row {
    position: relative;
    display: flex;
    align-items: center;
    gap: 16px;
    width: 100%;
    height: var(--row);
    padding: 0 16px 0 8px;
    border-radius: var(--r-sm);
    text-align: left;
    transition:
      background-color var(--quick) var(--ease),
      transform var(--quick) var(--ease);
  }
  .row:hover {
    background: var(--hover);
  }
  .row:active {
    transform: scale(0.99);
  }
  .row[aria-selected='true'] {
    background: var(--accent-soft);
  }

  /* Vertical time line running through every row of a day. */
  .rail {
    position: relative;
    display: grid;
    place-items: center;
    align-self: stretch;
    width: 24px;
    flex: none;
  }
  .rail::before {
    content: '';
    position: absolute;
    top: -4px;
    bottom: -4px;
    left: 50%;
    width: 2px;
    margin-left: -1px;
    background: var(--line);
  }
  .last .rail::before {
    bottom: 50%;
  }
  .node {
    position: relative;
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: var(--surface);
    box-shadow: inset 0 0 0 2px var(--faint);
    transition:
      box-shadow var(--quick) var(--ease),
      transform var(--duration) var(--ease);
  }
  .node.current {
    box-shadow: inset 0 0 0 2px var(--ok);
  }
  [aria-selected='true'] .node {
    background: var(--accent);
    box-shadow: 0 0 0 4px var(--accent-soft);
    transform: scale(1.15);
  }

  .time {
    display: grid;
    gap: 1px;
    width: 96px;
    flex: none;
  }
  strong {
    font-size: 15px;
    font-weight: 600;
    letter-spacing: -0.01em;
  }
  small {
    color: var(--faint);
    font-size: 11px;
  }
  .meta {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    gap: 6px;
    min-width: 0;
  }
  .chip {
    padding: 3px 9px;
    border-radius: 999px;
    background: var(--raised);
    color: var(--muted);
    font-size: 11px;
    font-weight: 550;
  }
  [aria-selected='true'] .chip {
    background: #ffffff10;
    color: var(--text);
  }
  .add {
    color: var(--ok);
  }
  .del {
    color: var(--danger);
  }
  .now {
    padding: 3px 9px;
    border-radius: 999px;
    background: #7ee2b01a;
    color: var(--ok);
    font-size: 11px;
    font-weight: 600;
  }
</style>
