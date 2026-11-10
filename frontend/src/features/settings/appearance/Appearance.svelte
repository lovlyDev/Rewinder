<script lang="ts">
  import { PickBackgroundImage } from '../../../../wailsjs/go/main/App';
  import { t } from '../../../i18n';
  import { previewPreferences, type Settings } from '../../../lib/preferences';
  import { toastError } from '../../../lib/toasts';
  import { setImage } from '../../../theme/theme';
  import Button from '../../../ui/button/Button.svelte';
  import Segmented from '../../../ui/segmented/Segmented.svelte';
  import Select from '../../../ui/select/Select.svelte';
  import Slider from '../../../ui/slider/Slider.svelte';
  import Swatches from '../../../ui/swatch/Swatches.svelte';
  import SettingRow from '../SettingRow.svelte';
  import ColorDot from './ColorDot.svelte';

  type Props = { draft: Settings; save: () => void };
  let { draft, save }: Props = $props();

  const presets = ['#a99bff', '#8b9dff', '#6fc3ff', '#5fe0c1', '#9be27a', '#ffcf7a', '#ff9e7a', '#ff8fbf'];
  const preview = () => previewPreferences(draft);
  let picking = $state(false);

  async function pickImage() {
    picking = true;
    try {
      const data = await PickBackgroundImage();
      if (!data) return;
      setImage(data);
      draft.theme.background.image = 'background.png';
      draft.theme.background.kind = 'image';
      save();
    } catch (e) {
      toastError('background.read', e);
    } finally {
      picking = false;
    }
  }
</script>

<SettingRow label={$t('accent')} hint={$t('accentHint')}>
  <Swatches
    label={$t('accent')}
    customLabel={$t('custom')}
    {presets}
    bind:value={draft.theme.accent}
    oninput={preview}
    onchange={save}
  />
</SettingRow>

<SettingRow label={$t('radius')} hint={$t('radiusHint')}>
  <Slider
    label={$t('radius')}
    bind:value={draft.theme.radius}
    max={24}
    format={(v) => `${v}px`}
    oninput={preview}
    oncommit={save}
  />
</SettingRow>

<SettingRow label={$t('density')}>
  <Segmented
    label={$t('density')}
    bind:value={draft.theme.density}
    onchange={save}
    options={[
      { value: 'compact', label: $t('compact') },
      { value: 'normal', label: $t('normal') },
    ]}
  />
</SettingRow>

<SettingRow label={$t('motion')} hint={$t('motionHint')}>
  <Select
    label={$t('motion')}
    bind:value={draft.theme.motion}
    onchange={save}
    options={[
      { value: 'fast', label: $t('fast') },
      { value: 'normal', label: $t('normal') },
      { value: 'off', label: $t('off'), hint: $t('offHint') },
    ]}
  />
</SettingRow>

<SettingRow label={$t('background')} hint={$t('backgroundHint')}>
  <Select
    label={$t('background')}
    bind:value={draft.theme.background.kind}
    onchange={save}
    options={[
      { value: 'color', label: $t('color') },
      { value: 'gradient', label: $t('gradient') },
      { value: 'image', label: $t('image') },
    ]}
  />
</SettingRow>

{#if draft.theme.background.kind === 'image'}
  <SettingRow label={$t('image')} hint={$t('imageHint')}>
    <Button icon="image" disabled={picking} onclick={pickImage}>{$t('chooseImage')}</Button>
  </SettingRow>
  <SettingRow label={$t('dim')}>
    <Slider
      label={$t('dim')}
      bind:value={draft.theme.background.dim}
      max={1}
      step={0.05}
      format={(v) => `${Math.round(v * 100)}%`}
      oninput={preview}
      oncommit={save}
    />
  </SettingRow>
  <SettingRow label={$t('blur')}>
    <Slider
      label={$t('blur')}
      bind:value={draft.theme.background.blur}
      max={32}
      format={(v) => `${v}px`}
      oninput={preview}
      oncommit={save}
    />
  </SettingRow>
{:else}
  <SettingRow label={$t(draft.theme.background.kind === 'gradient' ? 'gradientColors' : 'color')}>
    <ColorDot label={$t('color')} bind:value={draft.theme.background.color} oninput={preview} onchange={save} />
    {#if draft.theme.background.kind === 'gradient'}
      <ColorDot label={$t('gradient')} bind:value={draft.theme.background.endColor} oninput={preview} onchange={save} />
    {/if}
  </SettingRow>
{/if}
