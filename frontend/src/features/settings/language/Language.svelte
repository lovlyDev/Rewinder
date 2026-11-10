<script lang="ts">
 import { createEventDispatcher } from 'svelte';
 import type { settings } from '../../../../wailsjs/go/models';
 import { t } from '../../../i18n';
 import '../settings.css';
 export let value: settings.Settings;
 const dispatch=createEventDispatcher<{change:void}>();
</script>
<div class="settings-section"><h2>{$t('language')}</h2><label class="field"><span>{$t('language')}</span><select bind:value={value.language} on:change={()=>dispatch('change')}><option value="system">{$t('system')}</option><option value="ru">Русский</option><option value="en">English</option></select></label></div>
