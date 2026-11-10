<script lang="ts">
 import { t } from '../../../i18n';
 import { appInfo } from '../../../lib/preferences';
 import Icon from '../../../ui/icon/Icon.svelte';
 import '../settings.css';
</script>
<div class="settings-section"><h2>{$t('about')}</h2><div class="logo"><Icon name="clock-counter-clockwise" size={44} /><strong>Rewinder</strong></div><div class="row"><span>{$t('version')}</span><span class="hint">{$appInfo.version}</span></div><div class="field"><span>{$t('log')}</span><code>{$appInfo.logPath}</code></div><div class="field"><span>{$t('history')}</span><code>{$appInfo.historyPath}</code></div><p class="hint">{$t('privacy')}</p></div>
<style>.logo { display: flex; align-items: center; gap: 16px; color: var(--accent); }.logo strong { font-size: 26px; color: var(--text); font-weight: 600; }code { color: var(--muted); font-size: 12px; overflow-wrap: anywhere; user-select: text; }</style>
