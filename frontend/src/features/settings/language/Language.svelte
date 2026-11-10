<script lang="ts">
  import { t } from '../../../i18n';
  import type { Settings } from '../../../lib/preferences';
  import Select from '../../../ui/select/Select.svelte';
  import SettingRow from '../SettingRow.svelte';

  type Props = { draft: Settings; save: () => void };
  let { draft, save }: Props = $props();
</script>

<SettingRow label={$t('language')} hint={$t('languageHint')}>
  <Select
    label={$t('language')}
    bind:value={draft.language}
    onchange={save}
    options={[
      { value: 'system', label: $t('system'), hint: $t('systemHint') },
      { value: 'ru', label: 'Русский' },
      { value: 'en', label: 'English' },
    ]}
  />
</SettingRow>
