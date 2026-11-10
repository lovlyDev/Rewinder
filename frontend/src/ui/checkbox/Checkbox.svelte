<script lang="ts">export let checked = false; export let label: string;</script>
<label><input type="checkbox" bind:checked /><span>{label}</span></label>
<style>label { display: inline-flex; align-items: center; gap: 8px; color: var(--muted); cursor: pointer; font-size: 12px; } input { accent-color: var(--accent); width: 16px; height: 16px; }</style>
