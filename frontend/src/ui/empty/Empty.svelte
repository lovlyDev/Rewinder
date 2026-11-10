<script lang="ts">
  import Icon from '../icon/Icon.svelte';
  import type { IconName } from '../icon/icons';
  import { rise } from '../../lib/motion';

  type Props = { icon: IconName; title: string; hint?: string };
  let { icon, title, hint }: Props = $props();
</script>

<div class="empty" in:rise>
  <span class="ring"><Icon name={icon} size={26} /></span>
  <strong>{title}</strong>
  {#if hint}<p>{hint}</p>{/if}
</div>

<style>
  .empty {
    display: grid;
    justify-items: center;
    gap: 10px;
    padding: 48px 16px;
    text-align: center;
  }
  .ring {
    display: grid;
    place-items: center;
    width: 60px;
    height: 60px;
    margin-bottom: 4px;
    border-radius: 999px;
    background: var(--raised);
    color: var(--faint);
  }
  strong {
    font-weight: 600;
  }
  p {
    max-width: 260px;
    color: var(--muted);
    font-size: 12px;
  }
</style>
