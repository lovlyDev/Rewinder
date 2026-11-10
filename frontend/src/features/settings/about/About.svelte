<script lang="ts">
  import logo from '../../../assets/logo.png';
  import { t } from '../../../i18n';
  import { appInfo } from '../../../lib/preferences';
  import SettingRow from '../SettingRow.svelte';
</script>

<div class="hero">
  <img src={logo} alt="" width="56" height="56" draggable="false" />
  <div>
    <strong>Rewinder</strong>
    <span class="tabular">{$t('version')} {$appInfo.version}</span>
  </div>
</div>
<p class="privacy">{$t('privacy')}</p>
<SettingRow label={$t('log')}>
  <code>{$appInfo.logPath}</code>
</SettingRow>
<SettingRow label={$t('history')}>
  <code>{$appInfo.historyPath}</code>
</SettingRow>

<style>
  .hero {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 6px 4px 14px;
  }
  img {
    border-radius: 14px;
  }
  .hero div {
    display: grid;
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
