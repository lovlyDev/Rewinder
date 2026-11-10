<script lang="ts">
  type Props = { keys: string[]; size?: 'sm' | 'lg'; pulse?: boolean };
  let { keys, size = 'sm', pulse = false }: Props = $props();
</script>

<span class="combo {size}" class:pulse>
  {#each keys as key, i (key + i)}
    <kbd>{key}</kbd>
  {/each}
</span>

<style>
  .combo {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  kbd {
    display: inline-grid;
    place-items: center;
    min-width: 22px;
    height: 22px;
    padding: 0 6px;
    border-radius: 7px;
    background: var(--overlay);
    box-shadow: inset 0 -2px 0 #0007;
    color: var(--muted);
    font: 600 11px/1 var(--font);
  }
  .lg {
    gap: 8px;
  }
  .lg kbd {
    min-width: 44px;
    height: 44px;
    padding: 0 14px;
    border-radius: 12px;
    font-size: 16px;
    color: var(--text);
    box-shadow: inset 0 -3px 0 #0008;
  }
  .pulse kbd {
    animation: pulse 1.2s var(--ease) infinite;
  }
  @keyframes pulse {
    50% {
      opacity: 0.45;
    }
  }
</style>
