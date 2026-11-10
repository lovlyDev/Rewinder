<script lang="ts">
  import { t } from '../../i18n';
  import { sidebarMode } from '../../lib/ui';
  import Segmented from '../../ui/segmented/Segmented.svelte';
  import AppList from '../apps/AppList.svelte';
  import LayoutList from '../layouts/LayoutList.svelte';
</script>

<aside>
  <Segmented
    label={$t('sidebarMode')}
    bind:value={$sidebarMode}
    options={[
      { value: 'apps', label: $t('apps') },
      { value: 'layouts', label: $t('layoutsTitle') },
    ]}
  />
  {#if $sidebarMode === 'apps'}<AppList />{:else}<LayoutList />{/if}
</aside>

<style>
  aside {
    width: 280px;
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 14px;
    min-height: 0;
    padding: 12px 10px 10px 14px;
    border-radius: var(--r-lg);
    background: color-mix(in srgb, var(--surface) 84%, transparent);
  }
</style>
