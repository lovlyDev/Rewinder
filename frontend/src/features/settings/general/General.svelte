<script lang="ts">
 import { onMount } from 'svelte';
 import { GetAutostart, SetAutostart, PauseTracking, ResumeTracking, ExitApp } from '../../../../wailsjs/go/main/App';
 import { t } from '../../../i18n';
 import { trackingState } from '../../../lib/ui';
 import { toastError } from '../../../lib/toasts';
 import Switch from '../../../ui/switch/Switch.svelte';
 import Button from '../../../ui/button/Button.svelte';
 import '../settings.css';
 let autostart=false;let busy=false;
 onMount(()=>{void GetAutostart().then(value=>autostart=value).catch(e=>toastError('saveFailed',e));});
 async function changeAutostart(e: CustomEvent<boolean>) {busy=true;try{await SetAutostart(e.detail);}catch(error){autostart=!e.detail;toastError('saveFailed',error);}finally{busy=false;}}
</script>
<div class="settings-section"><h2>{$t('general')}</h2>
 <div class="row"><div class="label"><span>{$t('autostart')}</span><span class="hint">{$t('autostartHint')}</span></div><Switch label={$t('autostart')} bind:checked={autostart} disabled={busy} on:change={changeAutostart} /></div>
 <div class="row"><div class="label"><span>{$t('tracking')}</span><span class="hint">{$t($trackingState)}</span></div><Switch label={$t('tracking')} checked={$trackingState==='active'} on:change={e=>e.detail?ResumeTracking(null):PauseTracking(null)} /></div>
 <div class="field"><span class="hint">{$t('exitHint')}</span><div><Button icon="power" on:click={()=>ExitApp()}>{$t('exit')}</Button></div></div>
</div>
