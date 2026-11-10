<script lang="ts">
  import { t } from '../../../i18n';
  import { fromEvent, heldModifiers, parts, type Combination } from '../../../lib/hotkey';
  import Button from '../../../ui/button/Button.svelte';
  import Kbd from '../../../ui/kbd/Kbd.svelte';

  type Props = { value: Combination; busy?: boolean; onrecord: (c: Combination) => void };
  let { value, busy = false, onrecord }: Props = $props();

  let recording = $state(false);
  let held = $state<string[]>([]);

  const shown = $derived(recording ? (held.length ? [...held, '…'] : ['…']) : parts(value));

  function start() {
    held = [];
    recording = true;
  }
  function stop() {
    recording = false;
    held = [];
  }

  // Listen on window in the capture phase while recording, so no other shortcut
  // (dialog Esc, list keys) sees the keys being recorded.
  $effect(() => {
    if (!recording) return;
    const down = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (e.key === 'Escape') return stop();
      held = heldModifiers(e);
      const combo = fromEvent(e);
      if (combo && !e.repeat) {
        stop();
        onrecord(combo);
      }
    };
    const up = (e: KeyboardEvent) => (held = heldModifiers(e));
    window.addEventListener('keydown', down, true);
    window.addEventListener('keyup', up, true);
    window.addEventListener('blur', stop);
    return () => {
      window.removeEventListener('keydown', down, true);
      window.removeEventListener('keyup', up, true);
      window.removeEventListener('blur', stop);
    };
  });
</script>

<div class="recorder" class:recording>
  <Kbd keys={shown} size="lg" pulse={recording} />
  <p>{$t(recording ? 'record' : 'hotkeyHint')}</p>
  {#if recording}
    <Button variant="ghost" onclick={stop}>{$t('cancel')}</Button>
  {:else}
    <Button variant="soft" icon="keyboard" disabled={busy} onclick={start}>{$t('change')}</Button>
  {/if}
</div>

<style>
  .recorder {
    display: grid;
    justify-items: center;
    gap: 14px;
    padding: 26px 18px 22px;
    border-radius: var(--r-sm);
    background: var(--sunken);
    text-align: center;
    transition: box-shadow var(--duration) var(--ease);
  }
  .recording {
    box-shadow: inset 0 0 0 2px var(--accent);
  }
  p {
    max-width: 340px;
    color: var(--muted);
    font-size: 12px;
  }
</style>
