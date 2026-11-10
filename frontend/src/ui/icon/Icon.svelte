<script lang="ts">
 import { icons } from './icons';
 import type { IconName } from './icons';
 export let name: IconName;
 export let filled = false;
 export let size = 20;
</script>
<svg class:filled width={size} height={size} viewBox="0 0 256 256" fill="currentColor" aria-hidden="true">
 <path class="regular" d={icons[name].regular} /><path class="fill" d={icons[name].fill} />
</svg>
<style>
 svg { flex: none; } path { transition: opacity var(--quick) var(--ease); }
 .fill { opacity: 0; } .regular { opacity: 1; }
 .filled .fill, :global(button:active) .fill, :global(button[aria-pressed="true"]) .fill, :global(button[aria-selected="true"]) .fill { opacity: 1; }
 .filled .regular, :global(button:active) .regular, :global(button[aria-pressed="true"]) .regular, :global(button[aria-selected="true"]) .regular { opacity: 0; }
</style>
