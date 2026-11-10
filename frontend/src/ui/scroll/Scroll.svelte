<script lang="ts">
 import { createEventDispatcher } from 'svelte';
 export let label: string;
 const dispatch = createEventDispatcher<{more: void}>();
 function scroll(e: Event) { const el=e.currentTarget as HTMLElement;if(el.scrollTop+el.clientHeight>=el.scrollHeight-180) dispatch('more'); }
</script>
<section class="scroll" aria-label={label} on:scroll={scroll}><slot /></section>
<style>.scroll { flex: 1; min-height: 0; overflow: auto; padding: 3px; }</style>
