<script lang="ts">
  import { onMount } from 'svelte';
  import { ExitApp, GetAutostart, PauseTracking, ResumeTracking, SetAutostart } from '../../../../wailsjs/go/main/App';
  import { t } from '../../../i18n';
  import { toastError } from '../../../lib/toasts';
  import { trackingState } from '../../../lib/ui';
  import Button from '../../../ui/button/Button.svelte';
  import Switch from '../../../ui/switch/Switch.svelte';
  import SettingRow from '../SettingRow.svelte';

  let autostart = $state(false);
  let busy = $state(false);

  onMount(() => {
    GetAutostart()
      .then((v) => (autostart = v))
      .catch((e) => toastError('saveFailed', e));
  });

  async function changeAutostart(v: boolean) {
    busy = true;
    try {
      await SetAutostart(v);
    } catch (e) {
      autostart = !v;
      toastError('saveFailed', e);
    } finally {
      busy = false;
    }
  }
</script>

<SettingRow label={$t('autostart')} hint={$t('autostartHint')}>
  <Switch label={$t('autostart')} bind:checked={autostart} disabled={busy} onchange={changeAutostart} />
</SettingRow>
<SettingRow label={$t('tracking')} hint={$t($trackingState)}>
  <Switch
    label={$t('tracking')}
    checked={$trackingState === 'active'}
    onchange={(on) => (on ? ResumeTracking(null) : PauseTracking(null))}
  />
</SettingRow>
<SettingRow label={$t('exit')} hint={$t('exitHint')}>
  <Button variant="danger" icon="power" onclick={() => ExitApp()}>{$t('exitShort')}</Button>
</SettingRow>
