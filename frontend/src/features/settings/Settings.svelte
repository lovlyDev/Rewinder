<script lang="ts">
 import { preferences, savePreferences } from '../../lib/preferences';
 import { settingsOpen } from '../../lib/ui';
 import { t, message } from '../../i18n';
 import type { Key } from '../../i18n';
 import type { IconName } from '../../ui/icon/icons';
 import Modal from '../../ui/modal/Modal.svelte';
 import IconButton from '../../ui/button/IconButton.svelte';
 import Icon from '../../ui/icon/Icon.svelte';
 import General from './general/General.svelte';
 import Hotkeys from './hotkeys/Hotkeys.svelte';
 import Appearance from './appearance/Appearance.svelte';
 import Language from './language/Language.svelte';
 import About from './about/About.svelte';
 import { fade } from '../../lib/motion';
 const sections: {key:Key;icon:IconName}[]=[{key:'general',icon:'gear'},{key:'hotkeys',icon:'keyboard'},{key:'appearance',icon:'palette'},{key:'language',icon:'globe'},{key:'about',icon:'info'}];
 let section: Key='general';let error='';
 $: draft=$preferences;
 async function change() {if(!draft)return;error='';try{await savePreferences(draft);}catch(e){error=message(e);}}
</script>
<Modal label={$t('settings')} close={()=>settingsOpen.set(false)}>
 <div class="settings">
  <nav aria-label={$t('settings')}><strong>{$t('settings')}</strong>{#each sections as item}<button aria-pressed={section===item.key} on:click={()=>section=item.key}><Icon name={item.icon} filled={section===item.key} />{$t(item.key)}</button>{/each}</nav>
  <div class="content"><div class="close"><IconButton name="x" label={$t('close')} on:click={()=>settingsOpen.set(false)} /></div>
   <div class="section-host">{#if draft}{#key section}<div class="section" transition:fade>
    {#if section==='general'}<General />{:else if section==='hotkeys'}<Hotkeys value={draft} />{:else if section==='appearance'}<Appearance bind:value={draft} on:change={change} />{:else if section==='language'}<Language bind:value={draft} on:change={change} />{:else}<About />{/if}
   </div>{/key}{/if}</div>
   {#if error}<p class="error" role="alert">{error}</p>{/if}
  </div>
 </div>
</Modal>
<style>
 .settings { display: grid; grid-template-columns: 205px 1fr; height: 100%; }nav { background: var(--raised); padding: 26px 12px; display: flex; flex-direction: column; gap: 5px; }nav>strong { font-size: 16px; font-weight: 600; padding: 0 12px 24px; }nav>button { display: flex; align-items: center; gap: 11px; padding: 12px; text-align: left; background: transparent; color: var(--muted); border-radius: var(--small-radius); transition: transform var(--quick) var(--ease); }nav>button:hover { background: var(--hover); color: var(--text); }nav>button:active { transform: scale(.97); }nav>button[aria-pressed="true"] { background: var(--accent-soft); color: var(--accent); }.content { min-width: 0; overflow: auto; position: relative; padding: 30px; }.close { position: absolute; right: 14px; top: 14px; }.section-host { display: grid; }.section { grid-area: 1 / 1; padding-top: 4px; }.error { color: var(--danger); margin-top: 20px; font-size: 12px; }
</style>
