<script lang="ts">
  import { t } from '../../../i18n';
  import type { Combination } from '../../../lib/hotkey';
  import { preferences, saveHotkey } from '../../../lib/preferences';
  import { pushMessage, toastError } from '../../../lib/toasts';
  import Kbd from '../../../ui/kbd/Kbd.svelte';
  import SettingRow from '../SettingRow.svelte';
  import HotkeyRecorder from './HotkeyRecorder.svelte';

  let busy = $state(false);

  async function record(c: Combination) {
    busy = true;
    try {
      await saveHotkey(c.mods, c.key);
      pushMessage('success', 'hotkeySaved');
    } catch (e) {
      toastError('hotkey.taken', e);
    } finally {
      busy = false;
    }
  }
</script>

{#if $preferences}
  <SettingRow label={$t('hotkey')} hint={$t('hotkeyRowHint')} stacked>
    <HotkeyRecorder value={$preferences.hotkey} {busy} onrecord={record} />
  </SettingRow>
{/if}
<SettingRow label={$t('overlayKeys')} hint={$t('overlayKeysHint')}>
  <span class="legend">
    <Kbd keys={['←', '→']} />
    <Kbd keys={['Enter']} />
    <Kbd keys={['Esc']} />
  </span>
</SettingRow>

<style>
  .legend {
    display: flex;
    gap: 8px;
  }
</style>
