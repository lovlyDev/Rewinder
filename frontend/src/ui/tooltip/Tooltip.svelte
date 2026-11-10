<script lang="ts">export let text: string;</script>
<span class="tooltip"><slot /><span class="tip" role="tooltip">{text}</span></span>
<style>
 .tooltip { display: inline-flex; position: relative; }
 .tip { pointer-events: none; position: absolute; top: calc(100% + 7px); right: 0; z-index: 200; white-space: nowrap; padding: 6px 9px; background: var(--hover); border-radius: var(--small-radius); font-size: 12px; opacity: 0; transform: translateY(-3px); transition: opacity var(--quick) var(--ease), transform var(--quick) var(--ease); }
 .tooltip:hover .tip, .tooltip:focus-within .tip { opacity: 1; transform: translateY(0); transition-delay: 350ms; }
</style>
