<script lang="ts">
  import { HideWindow, MinimiseWindow, PauseTracking, ResumeTracking } from '../../../wailsjs/go/main/App';
  import logo from '../../assets/logo.png';
  import { t } from '../../i18n';
  import { settingsOpen, trackingState } from '../../lib/ui';
  import IconButton from '../../ui/button/IconButton.svelte';
  import Tooltip from '../../ui/tooltip/Tooltip.svelte';

  const paused = $derived($trackingState === 'paused');
  const failed = $derived($trackingState === 'error');
</script>

<header>
  <div class="brand">
    <img src={logo} alt="" width="22" height="22" draggable="false" />
    <strong>Rewinder</strong>
  </div>

  <Tooltip label={$t(paused ? 'resume' : 'pause')}>
    {#snippet children({ props })}
      <button
        {...props}
        class="status"
        class:paused
        class:failed
        aria-label={$t(paused ? 'resume' : 'pause')}
        onclick={() => (paused ? ResumeTracking(null) : PauseTracking(null))}
      >
        <span class="dot"></span>
        <span>{$t(failed ? 'trackingError' : paused ? 'trackingPaused' : 'trackingOn')}</span>
      </button>
    {/snippet}
  </Tooltip>

  <span class="drag"></span>

  <div class="actions">
    <IconButton name="gear" label={$t('settings')} active={$settingsOpen} onclick={() => settingsOpen.set(true)} />
    <span class="gap"></span>
    <IconButton name="minus" label={$t('minimise')} onclick={() => MinimiseWindow()} />
    <IconButton name="x" label={$t('hideToTray')} tone="danger" onclick={() => HideWindow()} />
  </div>
</header>

<style>
  header {
    display: flex;
    align-items: center;
    gap: 14px;
    height: 52px;
    flex: none;
    /* buttons sit inside the rounded window corner, never against its edge */
    padding: 0 8px 0 18px;
    --wails-draggable: drag;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  img {
    display: block;
    border-radius: 6px;
  }
  strong {
    font-size: 14px;
    font-weight: 650;
    letter-spacing: -0.01em;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 26px;
    padding: 0 11px 0 10px;
    border-radius: 999px;
    background: #7ee2b014;
    color: var(--ok);
    font-size: 11px;
    font-weight: 600;
    transition:
      background-color var(--quick) var(--ease),
      color var(--quick) var(--ease),
      transform var(--quick) var(--ease);
  }
  .status:hover {
    background: #7ee2b026;
  }
  .status:active {
    transform: scale(0.95);
  }
  .dot {
    position: relative;
    width: 7px;
    height: 7px;
    border-radius: 999px;
    background: currentColor;
  }
  .dot::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    background: currentColor;
    /* a few pulses when recording starts, then still: an endless animation costs CPU */
    animation: ping 2.4s var(--ease) 3;
  }
  @keyframes ping {
    0% {
      transform: scale(1);
      opacity: 0.6;
    }
    70%,
    100% {
      transform: scale(2.6);
      opacity: 0;
    }
  }
  .paused {
    background: var(--hover);
    color: var(--muted);
  }
  .paused:hover {
    background: var(--press);
  }
  .paused .dot::after,
  .failed .dot::after {
    animation: none;
  }
  .failed {
    background: #ff8f9c14;
    color: var(--danger);
  }
  .drag {
    flex: 1;
    align-self: stretch;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .gap {
    width: 8px;
  }
</style>
