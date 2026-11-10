<script lang="ts">import { createEventDispatcher } from 'svelte';const dispatch=createEventDispatcher<{change:boolean}>();export let checked = false; export let label: string; export let disabled = false;</script>
<button role="switch" aria-checked={checked} aria-label={label} {disabled} on:click={() => {checked = !checked;dispatch('change',checked);}}><span class:on={checked} /></button>
<style>
 button { width: 42px; height: 25px; padding: 3px; border-radius: 999px; background: var(--hover); }
 button[aria-checked="true"] { background: var(--accent); }
 span { display: block; width: 19px; height: 19px; background: var(--text); border-radius: 999px; transition: transform var(--quick) var(--ease); }
 span.on { transform: translateX(17px); background: var(--accent-ink); }
</style>
