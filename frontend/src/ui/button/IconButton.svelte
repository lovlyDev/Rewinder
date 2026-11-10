<script lang="ts">
 import './button.css';
 import Icon from '../icon/Icon.svelte';
 import Tooltip from '../tooltip/Tooltip.svelte';
 import type { IconName } from '../icon/icons';
 export let name: IconName;
 export let label: string;
 export let active = false;
 export let disabled = false;
</script>
<Tooltip text={label}><button class="button icon" aria-label={label} aria-pressed={active} {disabled} on:click><Icon {name} filled={active} /></button></Tooltip>
