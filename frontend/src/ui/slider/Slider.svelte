<script lang="ts">export let value = 0; export let min = 0; export let max = 24; export let step = 1; export let label: string; export let unit = '';</script>
<label><span>{label}<output>{value}{unit}</output></span><input type="range" {min} {max} {step} bind:value on:input on:change /></label>
<style>
 label { display: grid; gap: 12px; } span { display: flex; justify-content: space-between; } output { color: var(--muted); font-variant-numeric: tabular-nums; }
 input { appearance: none; width: 100%; height: 22px; margin: 0; background: transparent; cursor: pointer; }
 input::-webkit-slider-runnable-track { height: 4px; border-radius: 999px; background: var(--hover); }
 input::-webkit-slider-thumb { appearance: none; margin-top: -6px; width: 16px; height: 16px; border-radius: 999px; background: var(--accent); }
</style>
