<script lang="ts">
 import './button.css';
 import Icon from '../icon/Icon.svelte';
 import type { IconName } from '../icon/icons';
 export let primary = false;
 export let disabled = false;
 export let icon: IconName | undefined = undefined;
</script>
<button class="button" class:primary {disabled} on:click>{#if icon}<Icon name={icon} />{/if}<slot /></button>
