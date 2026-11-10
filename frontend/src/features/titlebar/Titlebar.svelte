<script lang="ts">
 import { HideWindow, MinimiseWindow, PauseTracking, ResumeTracking } from '../../../wailsjs/go/main/App';
 import Icon from '../../ui/icon/Icon.svelte';
 import IconButton from '../../ui/button/IconButton.svelte';
 import { t } from '../../i18n';
 import { settingsOpen, trackingState } from '../../lib/ui';
</script>
<header>
 <div class="brand"><Icon name="clock-counter-clockwise" /><strong>Rewinder</strong><span class="status" class:paused={$trackingState!=='active'} aria-label={$t($trackingState)} /></div>
 <span class="drag" />
 <nav aria-label={$t('settings')}>
  <IconButton name={$trackingState==='paused'?'play':'pause'} label={$t($trackingState==='paused'?'resume':'pause')} on:click={() => $trackingState==='paused'?ResumeTracking(null):PauseTracking(null)} />
  <IconButton name="gear" label={$t('settings')} active={$settingsOpen} on:click={() => settingsOpen.set(true)} />
  <span class="spacer" />
  <IconButton name="minus" label={$t('minimise')} on:click={() => MinimiseWindow()} />
  <IconButton name="x" label={$t('close')} on:click={() => HideWindow()} />
 </nav>
</header>
<style>
 header { display: flex; align-items: center; min-height: 54px; padding: 0 12px 0 20px; background: var(--surface); --wails-draggable: drag; }
 .brand { display: flex; align-items: center; gap: 10px; } .brand :global(svg) { color: var(--accent); }
 strong { font-size: 14px; font-weight: 600; } .drag { flex: 1; align-self: stretch; }
 nav { display: flex; gap: 4px; align-items: center; } .spacer { width: 12px; }
 .status { width: 6px; height: 6px; border-radius: 999px; background: var(--ok); margin-left: 4px; }.status.paused { background: var(--muted); }
</style>
