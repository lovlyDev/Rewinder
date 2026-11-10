<script lang="ts">
 import { createEventDispatcher } from 'svelte';
 import { t } from '../../i18n';
 const dispatch = createEventDispatcher<{record: {mods: number; key: number}}>();
 export let value: { mods: number; key: number };
 export let disabled = false;
 let recording = false;
 $: parts = [[2,'Ctrl'],[1,'Alt'],[4,'Shift'],[8,'Win']].filter(([bit]) => value.mods & Number(bit)).map(([,name]) => name);
 $: key = value.key >= 112 ? `F${value.key-111}` : String.fromCharCode(value.key);
 function record(e: KeyboardEvent) {
  if (e.key === 'Tab' || e.key === 'Escape') { recording = false; return; }
  e.preventDefault();e.stopPropagation();
  const code = e.code;
  const key = /^Key[A-Z]$/.test(code) ? code.charCodeAt(3) : /^Digit\d$/.test(code) ? code.charCodeAt(5) : /^F([1-9]|1\d|2[0-4])$/.test(code) ? 111+Number(code.slice(1)) : 0;
  const mods = (e.altKey ? 1 : 0) | (e.ctrlKey ? 2 : 0) | (e.shiftKey ? 4 : 0) | (e.metaKey ? 8 : 0);
  if (!key || !mods || e.repeat) return;
  dispatch('record', {mods, key}); recording = false;
 }
</script>
<input readonly aria-label={$t('hotkey')} {disabled} value={recording ? $t('record') : [...parts,key].join(' + ')} on:focus={() => recording=true} on:blur={() => recording=false} on:keydown={record} />
<style>input { width: 100%; padding: 14px; background: var(--raised); border-radius: var(--small-radius); text-align: center; font-size: 18px; }</style>
