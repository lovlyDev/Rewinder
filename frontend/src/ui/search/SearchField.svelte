<script lang="ts">
  import Icon from '../icon/Icon.svelte';

  type Props = { value: string; placeholder: string; clearLabel: string };
  let { value = $bindable(), placeholder, clearLabel }: Props = $props();
  let input: HTMLInputElement;
</script>

<label class="search" class:filled={value !== ''}>
  <Icon name="magnifyingGlass" size={16} />
  <input bind:this={input} bind:value {placeholder} aria-label={placeholder} spellcheck="false" />
  <button
    class="clear"
    aria-label={clearLabel}
    tabindex={value ? 0 : -1}
    onclick={() => {
      value = '';
      input.focus();
    }}
  >
    <Icon name="x" size={12} />
  </button>
</label>

<style>
  .search {
    display: flex;
    align-items: center;
    gap: 10px;
    height: 38px;
    padding: 0 6px 0 14px;
    border-radius: 999px;
    background: var(--raised);
    color: var(--faint);
    transition:
      background-color var(--quick) var(--ease),
      box-shadow var(--duration) var(--ease);
  }
  .search:focus-within {
    background: var(--overlay);
    box-shadow: 0 0 0 3px var(--accent-soft);
    color: var(--muted);
  }
  input {
    flex: 1;
    min-width: 0;
    height: 100%;
    outline: none;
    color: var(--text);
    user-select: text;
  }
  input::placeholder {
    color: var(--faint);
  }
  .clear {
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    border-radius: 999px;
    background: var(--hover);
    color: var(--muted);
    opacity: 0;
    transform: scale(0.6);
    pointer-events: none;
    transition:
      opacity var(--quick) var(--ease),
      transform var(--duration) var(--ease);
  }
  .filled .clear {
    opacity: 1;
    transform: none;
    pointer-events: auto;
  }
</style>
