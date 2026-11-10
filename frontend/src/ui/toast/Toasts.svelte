<script lang="ts">
 import { toasts, dismiss } from './toasts';
 import { t } from '../../i18n';
 import IconButton from '../button/IconButton.svelte';
 import { reveal } from '../../lib/motion';
</script>
<div class="toasts" aria-live="polite">
 {#each $toasts as toast (toast.id)}<div class="toast {toast.kind}" transition:reveal><span>{toast.code?$t(toast.code,toast.params):toast.text}</span><IconButton name="x" label={$t('close')} on:click={() => dismiss(toast.id)} /></div>{/each}
</div>
<style>
 .toasts { position: fixed; right: 18px; bottom: 16px; z-index: 300; display: grid; gap: 8px; pointer-events: none; }
 .toast { pointer-events: auto; display: flex; align-items: center; gap: 12px; max-width: 480px; background: var(--raised); border-radius: var(--radius); padding: 6px 8px 6px 16px; font-size: 13px; }
 .error { color: var(--danger); } .success { color: var(--ok); }
</style>
