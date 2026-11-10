<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { FocusOverlay, HideOverlay, LogUI, ShowTimelineWindow } from '../../../wailsjs/go/main/App';
  import { language, t } from '../../i18n';
  import { loadApps, loadTimeline, selectApp } from '../../lib/apps';
  import { ago, appName, clockSeconds, day } from '../../lib/format';
  import { now } from '../../lib/now';
  import { restore } from '../../lib/restore';
  import type { AppSummary, SnapshotMeta } from '../../lib/types';
  import { rise } from '../../lib/motion';
  import AppAvatar from '../../ui/avatar/AppAvatar.svelte';
  import Button from '../../ui/button/Button.svelte';
  import IconButton from '../../ui/button/IconButton.svelte';
  import Empty from '../../ui/empty/Empty.svelte';
  import Kbd from '../../ui/kbd/Kbd.svelte';
  import Scrubber from './Scrubber.svelte';

  /** The app the user was in when they pressed the hotkey. */
  let { requestedAppID }: { requestedAppID?: string } = $props();

  let el: HTMLElement;
  let app = $state<AppSummary>();
  let snapshots = $state<SnapshotMeta[]>([]);
  let index = $state(0);
  let loaded = $state(false);
  let busy = false;
  let alive = true;

  const current = $derived(snapshots[index]);

  async function load() {
    const list = await loadApps();
    // If the app the user was in has no history yet, fall back to the most recent one.
    const found = list.find((a) => a.appID === requestedAppID) ?? list[0];
    const history = found ? await loadTimeline(found.appID) : [];
    if (!alive) return;
    app = found;
    snapshots = history;
    // The newest snapshot is the current state; start one step back, which is what "undo" means.
    index = Math.min(1, Math.max(0, history.length - 1));
    loaded = true;
    await tick();
    // The UI is not drawn while the window is hidden, so ask Go to activate it now
    // that the overlay exists; otherwise keys could land before it is ready.
    await FocusOverlay();
    window.focus();
    el?.focus();
    void LogUI(`overlay open app=${app?.appID} requested=${requestedAppID ?? ''} snapshots=${history.length} index=${index}`);
  }

  async function close() {
    void LogUI('overlay close');
    if (busy) return;
    busy = true;
    await HideOverlay();
  }

  async function openApp() {
    void LogUI('overlay open-app');
    if (busy) return;
    busy = true;
    if (app) void selectApp(app.appID);
    await ShowTimelineWindow();
  }

  // Hide first, so the restored windows are not covered by the overlay and can take focus.
  async function restoreCurrent() {
    if (busy || !current) return;
    busy = true;
    const id = app?.appID;
    const snap = current;
    void LogUI(`overlay restore app=${id} index=${index} snapshot=${snap.snapshotID}`);
    await HideOverlay();
    if (id) void restore(id, snap.snapshotID);
  }

  // index 0 = now; a larger index = further back in time.
  const step = (older: number) => (index = Math.max(0, Math.min(snapshots.length - 1, index + older)));

  function keyboard(e: KeyboardEvent) {
    const actions: Record<string, () => void> = {
      Escape: () => void close(),
      Enter: () => void restoreCurrent(),
      ArrowLeft: () => step(1),
      ArrowRight: () => step(-1),
      Home: () => step(snapshots.length),
      End: () => step(-snapshots.length),
    };
    const act = actions[e.key];
    if (!act) return;
    e.preventDefault();
    if (e.key === 'Escape' || e.key === 'Enter') void LogUI(`overlay key ${e.key}`);
    act();
  }

  function wheel(e: WheelEvent) {
    e.preventDefault();
    if (e.deltaY) step(e.deltaY > 0 ? 1 : -1);
  }

  onMount(() => {
    window.addEventListener('keydown', keyboard, true);
    window.addEventListener('wheel', wheel, { passive: false });
    void load();
    return () => {
      alive = false;
      window.removeEventListener('keydown', keyboard, true);
      window.removeEventListener('wheel', wheel);
    };
  });
</script>

<div class="overlay" role="dialog" aria-label={$t('quickRestore')} tabindex="-1" bind:this={el}>
  <header>
    {#if app}<AppAvatar appID={app.appID} name={app.name} size={28} />{/if}
    <div class="who">
      <strong>{app ? appName(app.name) : 'Rewinder'}</strong>
      <small>{$t('quickRestore')}</small>
    </div>
    <span class="drag"></span>
    <Button variant="ghost" size="sm" icon="arrowSquareOut" onclick={openApp}>{$t('openApp')}</Button>
    <IconButton name="x" size="sm" label={$t('close')} onclick={close} />
  </header>

  {#if current}
    <div class="moment">
      {#key current.snapshotID}
        <div class="when" in:rise={{ y: 4 }}>
          <span class="time tabular">{clockSeconds(current.timestampUTC, $language)}</span>
          <span class="sub">{day(current.timestampUTC, $language)} · {ago(current.timestampUTC, $language, $now)}</span>
        </div>
      {/key}
      <div class="tags">
        <span class="chip">{$t('windows', { n: current.windowsCount })}</span>
        {#if index === 0}<span class="chip now">{$t('current')}</span>{/if}
      </div>
    </div>

    <div class="scrub">
      <IconButton name="caretLeft" size="sm" label={$t('older')} disabled={index >= snapshots.length - 1} onclick={() => step(1)} />
      <Scrubber count={snapshots.length} {index} label={$t('timeline')} onpick={(i) => (index = i)} />
      <IconButton name="caretRight" size="sm" label={$t('newer')} disabled={index === 0} onclick={() => step(-1)} />
    </div>

    <footer>
      <span class="keys">
        <Kbd keys={['←', '→']} /><span>{$t('keyStep')}</span>
        <Kbd keys={['Esc']} /><span>{$t('keyClose')}</span>
      </span>
      <Button variant="primary" icon="restore" disabled={index === 0} onclick={restoreCurrent}>
        {$t('restore')}
        <Kbd keys={['Enter']} />
      </Button>
    </footer>
  {:else if loaded}
    <div class="empty"><Empty icon="clock" title={$t('noSnapshots')} hint={$t('noSnapshotsHint')} /></div>
  {/if}
</div>

<style>
  .overlay {
    display: flex;
    flex-direction: column;
    gap: 10px;
    height: 100vh;
    padding: 12px 14px 14px 18px;
    background: var(--surface);
    outline: none;
    overflow: hidden;
  }
  header {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 34px;
    flex: none;
    --wails-draggable: drag;
  }
  .who {
    display: grid;
    line-height: 1.2;
  }
  .who strong {
    font-weight: 650;
  }
  .who small {
    color: var(--faint);
    font-size: 11px;
  }
  .drag {
    flex: 1;
    align-self: stretch;
  }
  .moment {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 12px;
    padding: 0 4px;
  }
  .when {
    display: grid;
  }
  .time {
    font-size: 34px;
    font-weight: 650;
    letter-spacing: -0.035em;
    line-height: 1.05;
  }
  .sub {
    color: var(--muted);
    font-size: 12px;
  }
  .tags {
    display: flex;
    gap: 6px;
    padding-bottom: 4px;
  }
  .chip {
    padding: 4px 10px;
    border-radius: 999px;
    background: var(--raised);
    color: var(--muted);
    font-size: 11px;
    font-weight: 550;
  }
  .chip.now {
    background: #7ee2b01a;
    color: var(--ok);
  }
  .scrub {
    display: flex;
    align-items: center;
  }
  footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: auto;
  }
  .keys {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--faint);
    font-size: 11px;
  }
  .keys span + :global(.combo) {
    margin-left: 10px;
  }
  footer :global(.btn.primary kbd) {
    background: #0000002a;
    box-shadow: none;
    color: inherit;
  }
  .empty {
    flex: 1;
    display: grid;
    place-items: center;
    margin-top: -24px;
  }
</style>
