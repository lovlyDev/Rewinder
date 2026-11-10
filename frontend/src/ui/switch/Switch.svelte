<script lang="ts">
  import { Switch } from 'bits-ui';

  type Props = { checked: boolean; label: string; disabled?: boolean; onchange?: (checked: boolean) => void };
  let { checked = $bindable(), label, disabled = false, onchange }: Props = $props();
</script>

<Switch.Root
  class="switch"
  {checked}
  {disabled}
  aria-label={label}
  onCheckedChange={(v) => {
    checked = v;
    onchange?.(v);
  }}
>
  <Switch.Thumb class="thumb" />
</Switch.Root>

<style>
  :global(.switch) {
    position: relative;
    width: 44px;
    height: 26px;
    flex: none;
    border-radius: 999px;
    background: var(--overlay);
    transition:
      background-color var(--duration) var(--ease),
      transform var(--quick) var(--ease);
  }
  :global(.switch:active:not(:disabled)) {
    transform: scale(0.95);
  }
  :global(.switch[data-state='checked']) {
    background: var(--accent);
  }
  :global(.switch:disabled) {
    opacity: 0.4;
  }
  :global(.thumb) {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 20px;
    height: 20px;
    border-radius: 999px;
    background: #fff;
    box-shadow: 0 2px 6px #0006;
    transition: transform var(--duration) var(--ease);
  }
  :global(.thumb[data-state='checked']) {
    transform: translateX(18px);
  }
</style>
