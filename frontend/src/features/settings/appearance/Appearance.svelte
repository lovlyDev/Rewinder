<script lang="ts">
 import { createEventDispatcher } from 'svelte';
 import { PickBackgroundImage } from '../../../../wailsjs/go/main/App';
 import type { settings } from '../../../../wailsjs/go/models';
 import { setImage, applyTheme } from '../../../theme/theme';
 import { appInfo } from '../../../lib/preferences';
 import { t, message } from '../../../i18n';
 import Slider from '../../../ui/slider/Slider.svelte';
 import Button from '../../../ui/button/Button.svelte';
 import '../settings.css';
 export let value: settings.Settings;
 const dispatch=createEventDispatcher<{change:void}>();
 const presets=['#a99bff','#84b7ff','#85dbb1','#ffc280','#f9a3c7'];
 let error='';let busy=false;
 const change=()=>dispatch('change');
 function preview(){void applyTheme(value,$appInfo.systemLanguage);}
 async function pick(){busy=true;error='';try{const data=await PickBackgroundImage();if(!data)return;setImage(data);value.theme.background.image='background.png';value.theme.background.kind='image';change();}catch(e){error=message(e);}finally{busy=false;}}
</script>
<div class="settings-section"><h2>{$t('appearance')}</h2>
 <div class="field"><span>{$t('accent')}</span><div class="swatches">{#each presets as color}<button class="swatch" style="--swatch:{color}" aria-label={color} aria-pressed={value.theme.accent===color} on:click={()=>{value.theme.accent=color;change();}}><span /></button>{/each}<input class="color" type="color" aria-label={$t('custom')} bind:value={value.theme.accent} on:input={preview} on:change={change} /></div></div>
 <Slider label={$t('radius')} bind:value={value.theme.radius} unit="px" on:input={preview} on:change={change} />
 <div class="field"><span>{$t('density')}</span><div class="choices">{#each ['compact','normal'] as density}<button class="choice" aria-pressed={value.theme.density===density} on:click={()=>{value.theme.density=density;change();}}>{$t(density==='compact'?'compact':'normal')}</button>{/each}</div></div>
 <div class="field"><span>{$t('motion')}</span><div class="choices">{#each ['fast','normal','off'] as motion}<button class="choice" aria-pressed={value.theme.motion===motion} on:click={()=>{value.theme.motion=motion;change();}}>{$t(motion==='fast'?'fast':motion==='off'?'off':'normal')}</button>{/each}</div></div>
 <div class="field"><span>{$t('background')}</span><div class="choices">{#each ['color','gradient','image'] as kind}<button class="choice" aria-pressed={value.theme.background.kind===kind} on:click={()=>{value.theme.background.kind=kind;change();}}>{$t(kind==='color'?'color':kind==='gradient'?'gradient':'image')}</button>{/each}</div></div>
 {#if value.theme.background.kind==='image'}
  <div><Button icon="folder" disabled={busy} on:click={pick}>{$t('chooseImage')}</Button></div>
  <Slider label={$t('dim')} bind:value={value.theme.background.dim} max={1} step={.05} on:input={preview} on:change={change} />
  <Slider label={$t('blur')} bind:value={value.theme.background.blur} max={32} unit="px" on:input={preview} on:change={change} />
 {:else}<div class="colors"><label>{$t('color')}<input class="color" type="color" bind:value={value.theme.background.color} on:input={preview} on:change={change} /></label>{#if value.theme.background.kind==='gradient'}<label>{$t('gradient')}<input class="color" type="color" bind:value={value.theme.background.endColor} on:input={preview} on:change={change} /></label>{/if}</div>{/if}
 {#if error}<p class="error" role="alert">{error}</p>{/if}
</div>
<style>
 .swatches { display: flex; align-items: center; gap: 10px; }.swatch { width: 30px; height: 30px; border-radius: 999px; background: var(--swatch); display: grid; place-items: center; transition: transform var(--quick) var(--ease); }.swatch:active { transform: scale(.97); }.swatch[aria-pressed="true"] span { width: 8px; height: 8px; border-radius: 999px; background: var(--accent-ink); }
 .color { width: 38px; height: 30px; background: var(--raised); padding: 3px; border-radius: var(--small-radius); cursor: pointer; }.color::-webkit-color-swatch-wrapper { padding: 0; }.color::-webkit-color-swatch { border: 0; border-radius: var(--small-radius); }.colors { display: flex; gap: 24px; }.colors label { display: flex; align-items: center; gap: 12px; font-size: 12px; }
</style>
