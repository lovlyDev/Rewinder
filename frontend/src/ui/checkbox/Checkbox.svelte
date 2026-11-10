<script lang="ts">import Icon from '../icon/Icon.svelte';export let checked = false; export let label: string;</script>
<label><input type="checkbox" bind:checked /><span class="box" aria-hidden="true"><Icon name="check" size={12} /></span><span>{label}</span></label>
<style>
 label { position: relative; display: inline-flex; align-items: center; gap: 8px; color: var(--muted); cursor: pointer; font-size: 12px; }
 input { position: absolute; width: 18px; height: 18px; opacity: 0; margin: 0; }
 .box { display: grid; place-items: center; width: 18px; height: 18px; background: var(--raised); border-radius: var(--small-radius); color: transparent; }
 input:checked + .box { background: var(--accent); color: var(--accent-ink); }
 input:focus-visible + .box { outline: 2px solid var(--accent); outline-offset: 3px; }
</style>
