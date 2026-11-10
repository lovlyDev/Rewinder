<script lang="ts">
  import { onMount } from 'svelte';
  import { ExitApp, GetAutostart, PauseTracking, ResumeTracking, SetAutostart, SetRunAsAdmin } from '../../../../wailsjs/go/main/App';
  import { appInfo, preferences } from '../../../lib/preferences';
  import { t } from '../../../i18n';
  import { pushMessage, toastError } from '../../../lib/toasts';
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

  let admin = $state(false);
  let adminBusy = $state(false);
  $effect(() => {
    admin = !!$preferences?.runAsAdmin;
  });
  const adminHint = $derived(
    $appInfo.elevationDeclined ? 'adminDeclined' : $appInfo.elevated ? 'adminOn' : admin ? 'adminNextStart' : 'adminHint',
  );

  async function changeAdmin(v: boolean) {
    adminBusy = true;
    try {
      await SetRunAsAdmin(v);
      preferences.update((p) => (p ? ({ ...p, runAsAdmin: v } as typeof p) : p));
      if (!v && $appInfo.elevated) pushMessage('info', 'adminOffLater');
    } catch (e) {
      admin = !v;
      toastError('admin.failed', e);
    } finally {
      adminBusy = false;
    }
  }

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
<SettingRow label={$t('runAsAdmin')} hint={$t(adminHint)}>
  <Switch label={$t('runAsAdmin')} bind:checked={admin} disabled={adminBusy} onchange={changeAdmin} />
</SettingRow>
<SettingRow label={$t('exit')} hint={$t('exitHint')}>
  <Button variant="danger" icon="power" onclick={() => ExitApp()}>{$t('exitShort')}</Button>
</SettingRow>
