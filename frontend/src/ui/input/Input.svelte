<script lang="ts">export let value = '';export let label: string;export let placeholder = '';</script>
<input type="text" aria-label={label} {placeholder} bind:value />
<style>input { min-width: 0; width: 100%; background: var(--raised); border-radius: var(--small-radius); padding: 10px 12px; } input::placeholder { color: var(--muted); }</style>
