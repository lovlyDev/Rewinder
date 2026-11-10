<script lang="ts">
  import { t } from '../../i18n';
  import { layouts, loadLayouts, saveLayout, selectedLayoutID } from '../../lib/layouts';
  import { rise } from '../../lib/motion';
  import Button from '../../ui/button/Button.svelte';
  import Icon from '../../ui/icon/Icon.svelte';
  import Kbd from '../../ui/kbd/Kbd.svelte';

  let naming = $state(false);
  let name = $state('');
  let saving = $state(false);
  let input = $state<HTMLInputElement>();

  $effect(() => {
    void loadLayouts();
  });
  $effect(() => {
    if (naming) input?.focus();
  });

  const selectedIndex = $derived($layouts.findIndex((l) => l.id === $selectedLayoutID));

  async function save() {
    if (!name.trim() || saving) return;
    saving = true;
    if (await saveLayout(name)) {
      naming = false;
      name = '';
    }
    saving = false;
  }
</script>

<div class="pane">
  {#if naming}
    <form
      class="namer"
      in:rise
      onsubmit={(e) => {
        e.preventDefault();
        void save();
      }}
    >
      <input
        bind:this={input}
        bind:value={name}
        maxlength="60"
        placeholder={$t('layoutNamePlaceholder')}
        aria-label={$t('layoutName')}
        onkeydown={(e) => e.key === 'Escape' && (naming = false)}
      />
      <div class="namer-actions">
        <Button variant="ghost" size="sm" type="button" onclick={() => (naming = false)}>{$t('cancel')}</Button>
        <Button variant="primary" size="sm" type="submit" disabled={!name.trim() || saving}>{$t('save')}</Button>
      </div>
    </form>
  {:else}
    <Button variant="soft" icon="plus" onclick={() => (naming = true)}>{$t('saveCurrentLayout')}</Button>
  {/if}

  <div class="list" role="listbox" aria-label={$t('layoutsTitle')}>
    {#if selectedIndex >= 0}<span class="highlight" style="--i:{selectedIndex}"></span>{/if}
    {#each $layouts as layout (layout.id)}
      <button
        class="item"
        role="option"
        aria-selected={layout.id === $selectedLayoutID}
        onclick={() => selectedLayoutID.set(layout.id)}
      >
        <span class="badge"><Icon name="layouts" size={18} filled={layout.id === $selectedLayoutID} /></span>
        <span class="name">
          <strong>{layout.name}</strong>
          <small>{$t('appsCount', { n: layout.apps.length })}</small>
        </span>
        {#if layout.shortcut}<Kbd keys={[`${layout.shortcut}`]} />{/if}
      </button>
    {/each}
  </div>
</div>

<style>
  .pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 0;
  }
  .pane > :global(.btn) {
    width: 100%;
  }
  .namer {
    display: grid;
    gap: 8px;
    padding: 10px;
    border-radius: var(--r-sm);
    background: var(--raised);
  }
  .namer input {
    height: 36px;
    padding: 0 12px;
    border-radius: 999px;
    background: var(--overlay);
    outline: none;
    user-select: text;
    transition: box-shadow var(--duration) var(--ease);
  }
  .namer input:focus {
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .namer-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
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
  .item {
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
  .item:hover:not([aria-selected='true']) {
    background: var(--hover);
  }
  .item:active {
    transform: scale(0.98);
  }
  .badge {
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    border-radius: 9px;
    background: var(--accent-soft);
    color: var(--accent);
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
    color: var(--faint);
    font-size: 11px;
  }
</style>
