<script lang="ts">
 import { createEventDispatcher } from 'svelte';
 import { saveHotkey } from '../../../lib/preferences';
 import { t, message } from '../../../i18n';
 import HotkeyInput from '../../../ui/input/HotkeyInput.svelte';
 import type { settings } from '../../../../wailsjs/go/models';
 import '../settings.css';
 export let value: settings.Settings;
 const dispatch=createEventDispatcher<{updated:void}>();
 let error='';let busy=false;
 async function record(e: CustomEvent<{mods:number;key:number}>) {
  error='';busy=true;
  try{await saveHotkey(e.detail.mods,e.detail.key);dispatch('updated');}
  catch(e){error=message(e);}finally{busy=false;}
 }
</script>
<div class="settings-section"><h2>{$t('hotkeys')}</h2><div class="field"><span>{$t('hotkey')}</span><HotkeyInput value={value.hotkey} disabled={busy} on:record={record} /><p class="hint">{$t('hotkeyHint')}</p>{#if error}<p class="error" role="alert">{error}</p>{/if}</div></div>
