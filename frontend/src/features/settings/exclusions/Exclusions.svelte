<script lang="ts">
  import { onMount } from 'svelte';
  import { flip } from 'svelte/animate';
  import { GetExclusions, SetExcluded } from '../../../../wailsjs/go/main/App';
  import { t } from '../../../i18n';
  import { apps, loadApps } from '../../../lib/apps';
  import { appName } from '../../../lib/format';
  import { duration, pop } from '../../../lib/motion';
  import { toastError } from '../../../lib/toasts';
  import Icon from '../../../ui/icon/Icon.svelte';
  import Select from '../../../ui/select/Select.svelte';
  import SettingRow from '../SettingRow.svelte';

  let excluded = $state<string[]>([]);
  let pick = $state('');

  const exeOf = (path: string) => path.split('\\').pop()?.toLowerCase() ?? '';
  // Tracked apps that can still be excluded, one entry per executable.
  const candidates = $derived.by(() => {
    const seen = new Set(excluded);
    const out: { value: string; label: string }[] = [];
    for (const a of $apps) {
      const exe = exeOf(a.executablePath);
      if (!exe || seen.has(exe)) continue;
      seen.add(exe);
      out.push({ value: exe, label: appName(a.name) });
    }
    return out;
  });

  async function reload() {
    try {
      excluded = (await GetExclusions()) ?? [];
    } catch (e) {
      toastError('settings.read', e);
    }
  }
  async function set(exe: string, on: boolean) {
    try {
      await SetExcluded(exe, on);
      await reload();
      await loadApps();
    } catch (e) {
      toastError('settings.write', e);
    }
  }
  onMount(reload);
</script>

<SettingRow label={$t('addExclusion')} hint={$t('addExclusionHint')}>
  {#if candidates.length}
    <Select
      label={$t('addExclusion')}
      bind:value={pick}
      options={[{ value: '', label: $t('chooseApp') }, ...candidates]}
      onchange={(exe) => {
        if (exe) void set(exe, true);
        pick = '';
      }}
    />
  {:else}
    <span class="none">{$t('nothingToExclude')}</span>
  {/if}
</SettingRow>

<SettingRow label={$t('exclusions')} hint={$t('exclusionsListHint')} stacked>
  {#if excluded.length}
    <div class="chips">
      {#each excluded as exe (exe)}
        <span class="chip" transition:pop animate:flip={{ duration: duration(200) }}>
          <Icon name="eyeSlash" size={14} />
          {exe}
          <button aria-label={$t('trackAgain')} onclick={() => set(exe, false)}><Icon name="x" size={12} /></button>
        </span>
      {/each}
    </div>
  {:else}
    <span class="none">{$t('noExclusions')}</span>
  {/if}
</SettingRow>

<p class="note">{$t('systemExclusions')}</p>

<style>
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 32px;
    padding: 0 4px 0 12px;
    border-radius: 999px;
    background: var(--raised);
    font-size: 12px;
    font-weight: 550;
  }
  .chip button {
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    border-radius: 999px;
    color: var(--muted);
    transition:
      background-color var(--quick) var(--ease),
      color var(--quick) var(--ease);
  }
  .chip button:hover {
    background: var(--hover);
    color: var(--text);
  }
  .none {
    color: var(--faint);
    font-size: 12px;
  }
  .note {
    margin-top: 14px;
    padding: 0 4px;
    color: var(--faint);
    font-size: 12px;
  }
</style>
