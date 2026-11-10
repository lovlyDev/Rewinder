<script lang="ts">
  import { Dialog } from 'bits-ui';
  import type { Snippet } from 'svelte';
  import { fade, pop } from '../../lib/motion';

  type Props = { open: boolean; label: string; children: Snippet };
  let { open = $bindable(), label, children }: Props = $props();
  let panel = $state<HTMLElement>();
</script>

<!-- Focus trap, Esc, outside click and focus return come from bits-ui. -->
<Dialog.Root bind:open>
  <Dialog.Portal>
    <Dialog.Overlay forceMount>
      {#snippet child({ props, open })}
        {#if open}<div {...props} class="scrim" transition:fade></div>{/if}
      {/snippet}
    </Dialog.Overlay>
    <!-- Focus the panel itself, not its first button: otherwise a mouse-opened
         dialog greets the user with a focus ring on some control. -->
    <Dialog.Content
      forceMount
      onOpenAutoFocus={(e) => {
        e.preventDefault();
        panel?.focus();
      }}
    >
      {#snippet child({ props, open })}
        {#if open}
          <div {...props} bind:this={panel} tabindex="-1" class="panel" aria-label={label} transition:pop={{ from: 0.965 }}>
            <Dialog.Title class="sr-only">{label}</Dialog.Title>
            {@render children()}
          </div>
        {/if}
      {/snippet}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: 200;
    background: var(--scrim);
  }
  .panel {
    position: fixed;
    z-index: 201;
    inset: 0;
    margin: auto;
    width: min(880px, calc(100vw - 48px));
    height: min(620px, calc(100vh - 48px));
    border-radius: var(--r-lg);
    background: var(--surface);
    box-shadow:
      0 40px 120px -20px #000,
      0 0 0 1px var(--line);
    overflow: hidden;
    outline: none;
  }
  :global(.sr-only) {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
</style>
