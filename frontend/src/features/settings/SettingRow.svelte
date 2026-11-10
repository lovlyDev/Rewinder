<script lang="ts">
  import type { Snippet } from 'svelte';

  type Props = { label: string; hint?: string; stacked?: boolean; children: Snippet };
  let { label, hint, stacked = false, children }: Props = $props();
</script>

<div class="row" class:stacked>
  <div class="text">
    <span class="label">{label}</span>
    {#if hint}<span class="hint">{hint}</span>{/if}
  </div>
  <div class="control">{@render children()}</div>
</div>

<style>
  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding: 14px 18px;
    border-radius: var(--r-sm);
    background: #16161b;
    /* controls inside a row sit one step lighter than the row itself */
    --raised: #202027;
    --overlay: #2b2b33;
    --hover: #ffffff12;
  }
  .row + :global(.row) {
    margin-top: 6px;
  }
  .stacked {
    flex-direction: column;
    align-items: stretch;
    gap: 14px;
  }
  .text {
    display: grid;
    gap: 3px;
    min-width: 0;
  }
  .label {
    font-weight: 550;
  }
  .hint {
    color: var(--faint);
    font-size: 12px;
  }
  .control {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: none;
  }
  .stacked .control {
    flex: auto;
  }
</style>
