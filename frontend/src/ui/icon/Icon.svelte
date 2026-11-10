<script lang="ts">
  import { icons, type IconName } from './icons';

  /** still: never switch to the filled variant (marks such as a check). */
  type Props = { name: IconName; filled?: boolean; size?: number; still?: boolean };
  let { name, filled = false, size = 18, still = false }: Props = $props();
</script>

<!-- Outline and fill are stacked and cross-faded, so an icon "fills in" when its
     control is pressed, selected or active. -->
<svg class:filled class:still width={size} height={size} viewBox="0 0 256 256" fill="currentColor" aria-hidden="true">
  <g class="regular">{@html icons[name][0]}</g>
  <g class="fill">{@html icons[name][1]}</g>
</svg>

<style>
  svg {
    flex: none;
    display: block;
  }
  g {
    transition: opacity var(--quick) var(--ease);
  }
  .fill {
    opacity: 0;
  }
  .filled .fill,
  :global(:is(button, [role="option"], [role="tab"]):is(:active, [aria-pressed="true"], [aria-selected="true"], [data-state="checked"], [data-selected])) svg:not(.still) .fill {
    opacity: 1;
  }
  .filled .regular,
  :global(:is(button, [role="option"], [role="tab"]):is(:active, [aria-pressed="true"], [aria-selected="true"], [data-state="checked"], [data-selected])) svg:not(.still) .regular {
    opacity: 0;
  }
</style>
