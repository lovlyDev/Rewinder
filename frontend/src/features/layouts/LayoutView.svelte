<script lang="ts">
  import { language, t } from '../../i18n';
  import { appName, when } from '../../lib/format';
  import {
    deleteLayout,
    layoutPreview,
    layouts,
    moveLayout,
    renameLayout,
    restoreLayout,
    restoringLayout,
    saveLayout,
    selectedLayoutID,
  } from '../../lib/layouts';
  import type { SnapshotLayout } from '../../lib/preview';
  import AppAvatar from '../../ui/avatar/AppAvatar.svelte';
  import Button from '../../ui/button/Button.svelte';
  import ConfirmButton from '../../ui/button/ConfirmButton.svelte';
  import IconButton from '../../ui/button/IconButton.svelte';
  import Empty from '../../ui/empty/Empty.svelte';
  import Kbd from '../../ui/kbd/Kbd.svelte';
  import Minimap from '../../ui/minimap/Minimap.svelte';

  const layout = $derived($layouts.find((l) => l.id === $selectedLayoutID));
  const index = $derived($layouts.findIndex((l) => l.id === $selectedLayoutID));

  let preview = $state<SnapshotLayout | null>(null);
  $effect(() => {
    const l = layout;
    if (!l) return;
    void layoutPreview(l.id, l.createdAtUTC).then((p) => {
      if (layout?.id === l.id) preview = p;
    });
  });

  let editing = $state(false);
  let draft = $state('');
  let input = $state<HTMLInputElement>();
  $effect(() => {
    if (editing) input?.select();
  });
  async function commitName() {
    if (layout && draft.trim() && draft.trim() !== layout.name) await renameLayout(layout.id, draft);
    editing = false;
  }
</script>

<main>
  {#if layout}
    <header>
      <div class="title">
        {#if editing}
          <input
            bind:this={input}
            bind:value={draft}
            maxlength="60"
            aria-label={$t('layoutName')}
            onblur={commitName}
            onkeydown={(e) => {
              if (e.key === 'Enter') void commitName();
              if (e.key === 'Escape') editing = false;
            }}
          />
        {:else}
          <button
            class="name"
            title={$t('rename')}
            onclick={() => {
              draft = layout.name;
              editing = true;
            }}>{layout.name}</button
          >
        {/if}
        <p>{$t('savedAt', { when: when(layout.createdAtUTC, $language) })}</p>
      </div>
      <div class="actions">
        <IconButton name="caretLeft" label={$t('moveUp')} disabled={index <= 0} onclick={() => moveLayout(layout.id, index - 1)} />
        <IconButton
          name="caretRight"
          label={$t('moveDown')}
          disabled={index >= $layouts.length - 1}
          onclick={() => moveLayout(layout.id, index + 1)}
        />
        <IconButton name="refresh" label={$t('updateLayout')} onclick={() => saveLayout(layout.name)} />
        <ConfirmButton icon="trash" label={$t('deleteLayout')} confirmLabel={$t('deleteHistoryConfirm')} onconfirm={() => deleteLayout(layout.id)} />
      </div>
    </header>

    <div class="stage"><Minimap layout={preview} label={$t('previewLabel')} /></div>

    <ul class="apps">
      {#each layout.apps as app, i (app.appID + i)}
        <li>
          <AppAvatar appID={app.appID} name={app.name} size={26} />
          <span class="app-name">{appName(app.name)}</span>
          <small>{$t('windows', { n: app.windows })}</small>
        </li>
      {/each}
    </ul>

    <footer>
      <div class="hint">
        {#if layout.shortcut}<Kbd keys={['Ctrl', 'Alt', `${layout.shortcut}`]} /><span>{$t('fromAnywhere')}</span>{/if}
      </div>
      <Button variant="primary" size="lg" icon="restore" disabled={!!$restoringLayout} onclick={() => restoreLayout(layout.id)}>
        {$t('restoreLayout')}
      </Button>
    </footer>
  {:else}
    <div class="center"><Empty icon="layouts" title={$t('noLayouts')} hint={$t('noLayoutsHint')} /></div>
  {/if}
</main>

<style>
  main {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 18px 12px 8px 16px;
    border-radius: var(--r-lg);
    background: color-mix(in srgb, var(--surface) 84%, transparent);
  }
  header {
    display: flex;
    align-items: center;
    gap: 14px;
  }
  .title {
    flex: 1;
    min-width: 0;
  }
  .name,
  .title input {
    display: block;
    max-width: 100%;
    margin-left: -8px;
    padding: 2px 8px;
    border-radius: var(--r-xs);
    font-size: 24px;
    font-weight: 650;
    letter-spacing: -0.025em;
    text-align: left;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: background-color var(--quick) var(--ease);
  }
  .name:hover {
    background: var(--hover);
  }
  .title input {
    width: 100%;
    background: var(--raised);
    outline: none;
    user-select: text;
  }
  .title p {
    color: var(--faint);
    font-size: 11px;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .stage {
    flex: 1;
    min-height: 160px;
    padding: 16px;
    border-radius: var(--r-lg);
    background: var(--raised);
  }
  .apps {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .apps li {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 38px;
    padding: 0 14px 0 6px;
    border-radius: 999px;
    background: var(--raised);
  }
  .app-name {
    font-weight: 550;
  }
  .apps small {
    color: var(--faint);
    font-size: 11px;
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
