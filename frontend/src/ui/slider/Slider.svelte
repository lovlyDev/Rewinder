<script lang="ts">export let value = 0; export let min = 0; export let max = 24; export let step = 1; export let label: string; export let unit = '';</script>
<label><span>{label}<output>{value}{unit}</output></span><input type="range" {min} {max} {step} bind:value on:input on:change /></label>
<style>label { display: grid; gap: 12px; } span { display: flex; justify-content: space-between; } output { color: var(--muted); font-variant-numeric: tabular-nums; } input { width: 100%; accent-color: var(--accent); cursor: pointer; }</style>
