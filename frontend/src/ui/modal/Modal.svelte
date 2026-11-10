<script lang="ts">
 import { onMount } from 'svelte';
 import { fade } from 'svelte/transition';
 import { reveal, duration } from '../../lib/motion';
 export let label: string;
 export let close: () => void;
 let panel: HTMLDivElement;
 function keyboard(e: KeyboardEvent) {
  if (e.key === 'Escape') { e.preventDefault();e.stopImmediatePropagation();close(); }
  if (e.key !== 'Tab') return;
  const items = Array.from(panel.querySelectorAll<HTMLElement>('button:not(:disabled),input:not(:disabled),select:not(:disabled),[tabindex="0"]')).filter(el => el.getClientRects().length > 0);
  const first = items[0]; const last=items[items.length-1];
  if (!first) {e.preventDefault(); panel.focus();return;}
  if (e.shiftKey && (document.activeElement === first || document.activeElement === panel)) { e.preventDefault();last.focus(); }
  else if (!e.shiftKey && document.activeElement === last) {e.preventDefault();first.focus();}
 }
 onMount(() => {
  const previous = document.activeElement as HTMLElement | null;
  panel.focus();window.addEventListener('keydown',keyboard,true);
  return () => {window.removeEventListener('keydown',keyboard,true);if(previous?.isConnected)previous.focus();};
 });
</script>
<div class="layer" transition:fade={{duration: duration()}}>
 <button class="backdrop" aria-label={label} tabindex="-1" on:click={close} />
 <div class="panel" role="dialog" aria-modal="true" aria-label={label} tabindex="-1" bind:this={panel} transition:reveal={{scale:true}}><slot /></div>
</div>
<style>
 .layer { position: fixed; inset: 0; display: grid; place-items: center; z-index: 100; padding: 24px; }
 .backdrop { position: absolute; inset: 0; background: var(--scrim); border-radius: 0; }
 .panel { position: relative; width: min(830px,100%); height: min(650px,100%); background: var(--surface); border-radius: var(--radius); outline: none; overflow: hidden; }
</style>
