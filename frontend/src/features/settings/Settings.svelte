<script lang="ts">
  import { t, type Key } from '../../i18n';
  import { preferences, savePreferences, type Settings } from '../../lib/preferences';
  import { toastError } from '../../lib/toasts';
  import { settingsOpen } from '../../lib/ui';
  import { fade } from '../../lib/motion';
  import IconButton from '../../ui/button/IconButton.svelte';
  import Dialog from '../../ui/dialog/Dialog.svelte';
  import Icon from '../../ui/icon/Icon.svelte';
  import type { IconName } from '../../ui/icon/icons';
  import About from './about/About.svelte';
  import Appearance from './appearance/Appearance.svelte';
  import General from './general/General.svelte';
  import Language from './language/Language.svelte';
  import Shortcuts from './shortcuts/Shortcuts.svelte';

  type Section = 'general' | 'hotkeys' | 'appearance' | 'language' | 'about';
  const sections: { key: Section; icon: IconName; hint: Key }[] = [
    { key: 'general', icon: 'sliders', hint: 'generalHint' },
    { key: 'hotkeys', icon: 'keyboard', hint: 'hotkeysHint' },
    { key: 'appearance', icon: 'palette', hint: 'appearanceHint' },
    { key: 'language', icon: 'globe', hint: 'languageSectionHint' },
    { key: 'about', icon: 'info', hint: 'aboutHint' },
  ];
  let section = $state<Section>('general');
  const index = $derived(sections.findIndex((s) => s.key === section));

  // A local copy the controls edit; every change is applied live and saved.
  let draft = $state<Settings | null>(null);
  $effect(() => {
    if ($settingsOpen && $preferences) draft = structuredClone($state.snapshot($preferences)) as Settings;
  });

  async function save() {
    if (!draft) return;
    try {
      await savePreferences($state.snapshot(draft) as Settings);
    } catch (e) {
      toastError('settings.write', e);
    }
  }
</script>

<Dialog bind:open={$settingsOpen} label={$t('settings')}>
  <div class="settings">
    <nav aria-label={$t('settings')}>
      <h2>{$t('settings')}</h2>
      <div class="items" style="--i:{index}">
        <span class="highlight"></span>
        {#each sections as item (item.key)}
          <button aria-pressed={section === item.key} onclick={() => (section = item.key)}>
            <Icon name={item.icon} filled={section === item.key} />
            <span>{$t(item.key)}</span>
          </button>
        {/each}
      </div>
    </nav>

    <div class="content">
      <div class="close"><IconButton name="x" label={$t('close')} onclick={() => settingsOpen.set(false)} /></div>
      {#key section}
        <div class="section" in:fade={{ base: 180 }}>
          <header>
            <h1>{$t(section)}</h1>
            <p>{$t(sections[index].hint)}</p>
          </header>
          {#if section === 'general'}
            <General />
          {:else if section === 'hotkeys'}
            <Shortcuts />
          {:else if section === 'appearance' && draft}
            <Appearance {draft} {save} />
          {:else if section === 'language' && draft}
            <Language {draft} {save} />
          {:else if section === 'about'}
            <About />
          {/if}
        </div>
      {/key}
    </div>
  </div>
</Dialog>

<style>
  .settings {
    display: grid;
    grid-template-columns: 220px 1fr;
    height: 100%;
  }
  nav {
    padding: 22px 12px;
    background: var(--sunken);
  }
  h2 {
    padding: 0 12px 18px;
    font-size: 15px;
    font-weight: 650;
  }
  .items {
    --h: 40px;
    --gap: 4px;
    position: relative;
    display: grid;
    gap: var(--gap);
  }
  .highlight {
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    height: var(--h);
    border-radius: 999px;
    background: var(--accent-soft);
    transform: translateY(calc(var(--i) * (var(--h) + var(--gap))));
    transition: transform var(--duration) var(--ease);
  }
  nav button {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    height: var(--h);
    padding: 0 14px;
    border-radius: 999px;
    color: var(--muted);
    font-weight: 550;
    text-align: left;
    transition:
      color var(--quick) var(--ease),
      background-color var(--quick) var(--ease),
      transform var(--quick) var(--ease);
  }
  nav button:hover:not([aria-pressed='true']) {
    background: var(--hover);
    color: var(--text);
  }
  nav button:active {
    transform: scale(0.97);
  }
  nav button[aria-pressed='true'] {
    color: var(--accent);
  }
  .content {
    position: relative;
    min-width: 0;
    overflow-y: auto;
    padding: 24px 26px 26px;
  }
  .close {
    position: absolute;
    top: 14px;
    right: 14px;
    z-index: 2;
  }
  header {
    margin-bottom: 18px;
    padding: 0 4px;
  }
  h1 {
    font-size: 22px;
    font-weight: 650;
    letter-spacing: -0.02em;
  }
  header p {
    margin-top: 2px;
    color: var(--faint);
  }
</style>
