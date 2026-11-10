<script lang="ts">
  import logo from '../../../assets/logo.png';
  import { t } from '../../../i18n';
  import { appInfo, type Settings } from '../../../lib/preferences';
  import { checkForUpdate, installUpdate, updateInfo, updateProgress, updateState } from '../../../lib/update';
  import { BrowserOpenURL } from '../../../../wailsjs/runtime/runtime';
  import Button from '../../../ui/button/Button.svelte';
  import Select from '../../../ui/select/Select.svelte';
  import SettingRow from '../SettingRow.svelte';

  type Props = { draft: Settings; save: () => void };
  let { draft, save }: Props = $props();
  // Older settings files have no value: that means install automatically.
  $effect(() => {
    if (!draft.updates) draft.updates = 'install';
  });
  const dev = $derived($appInfo.version === 'dev');
</script>

<div class="hero">
  <img src={logo} alt="" width="56" height="56" draggable="false" />
  <div>
    <strong>Rewinder</strong>
    <span class="tabular">{$t('version')} {$appInfo.version}</span>
  </div>
  <div class="links">
    <Button variant="soft" size="sm" icon="github" onclick={() => BrowserOpenURL('https://github.com/lovlyDev/Rewinder')}>GitHub</Button>
    <Button variant="soft" size="sm" icon="user" onclick={() => BrowserOpenURL('https://lovlydev.ru')}>{$t('author')}</Button>
  </div>
</div>
<p class="privacy">{$t('privacy')}</p>


<SettingRow label={$t('updates')} hint={$t(dev ? 'updatesDev' : 'updatesHint')}>
  <Select
    label={$t('updates')}
    bind:value={draft.updates}
    onchange={save}
    options={[
      { value: 'install', label: $t('updatesInstall'), hint: $t('updatesInstallHint') },
      { value: 'auto', label: $t('updatesAuto'), hint: $t('updatesAutoHint') },
      { value: 'off', label: $t('updatesOff') },
    ]}
  />
</SettingRow>
<SettingRow
  label={$updateInfo?.available ? $t('updateReady', { version: $updateInfo.version }) : $t('updateCheck')}
  stacked={!!$updateInfo?.available}
>
  {#if $updateInfo?.available}
    <div class="update">
      {#if $updateInfo.notes}<p class="notes">{$updateInfo.notes}</p>{/if}
      {#if $updateState === 'downloading'}
        <div class="bar"><span style="transform:scaleX({$updateProgress / 100})"></span></div>
      {/if}
      <div class="update-actions">
        <button class="link" onclick={() => $updateInfo && BrowserOpenURL($updateInfo.url)}>{$t('releaseNotes')}</button>
        <Button
          variant="primary"
          icon="download"
          disabled={!$updateInfo.installable || $updateState !== 'idle'}
          onclick={installUpdate}
        >
          {$t($updateState === 'downloading' ? 'downloading' : $updateState === 'restarting' ? 'restarting' : 'installUpdate')}
        </Button>
      </div>
    </div>
  {:else}
    <Button icon="refresh" disabled={$updateState !== 'idle'} onclick={() => checkForUpdate()}>
      {$t($updateState === 'checking' ? 'checking' : 'checkNow')}
    </Button>
  {/if}
</SettingRow>
<SettingRow label={$t('log')}>
  <code>{$appInfo.logPath}</code>
</SettingRow>
<SettingRow label={$t('history')}>
  <code>{$appInfo.historyPath}</code>
</SettingRow>

<style>
  .update {
    display: grid;
    gap: 12px;
  }
  .notes {
    max-height: 120px;
    overflow: auto;
    color: var(--muted);
    font-size: 12px;
    white-space: pre-wrap;
    user-select: text;
  }
  .bar {
    height: 6px;
    border-radius: 999px;
    background: var(--overlay);
    overflow: hidden;
  }
  .bar span {
    display: block;
    height: 100%;
    background: var(--accent);
    transform-origin: left;
    transition: transform var(--quick) linear;
  }
  .update-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .update-actions .link {
    color: var(--accent);
    font-size: 12px;
    text-decoration: none;
  }
  .hero {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 6px 4px 14px;
  }
  img {
    border-radius: 14px;
  }
  .hero > div:not(.links) {
    display: grid;
    flex: 1;
  }
  .links {
    display: flex;
    gap: 8px;
  }
  strong {
    font-size: 22px;
    font-weight: 650;
    letter-spacing: -0.02em;
  }
  .hero span {
    color: var(--faint);
    font-size: 12px;
  }
  .privacy {
    margin-bottom: 14px;
    padding: 0 4px;
    color: var(--muted);
  }
  code {
    max-width: 380px;
    color: var(--muted);
    font: 11px/1.4 ui-monospace, 'Cascadia Code', Consolas, monospace;
    overflow-wrap: anywhere;
    text-align: right;
    user-select: text;
  }
</style>
