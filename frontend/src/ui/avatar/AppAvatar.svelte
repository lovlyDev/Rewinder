<script lang="ts">
  import { appIcon } from '../../lib/appIcons';

  type Props = { appID: string; name: string; size?: number };
  let { appID, name, size = 32 }: Props = $props();

  let src = $state('');
  let loaded = $state(false);
  $effect(() => {
    const id = appID;
    src = '';
    loaded = false;
    void appIcon(id).then((url) => {
      if (id === appID) src = url;
    });
  });

  // Fallback: the app's initial on a tint derived from its name.
  const hue = $derived([...name].reduce((h, c) => (h * 31 + c.charCodeAt(0)) % 360, 7));
  const initial = $derived(name.replace(/\.exe$/i, '').charAt(0).toUpperCase() || '?');
</script>

<span class="avatar" style="--size:{size}px;--hue:{hue}" aria-hidden="true">
  {#if src}
    <img {src} alt="" class:loaded onload={() => (loaded = true)} draggable="false" />
  {:else}
    <span class="initial">{initial}</span>
  {/if}
</span>

<style>
  .avatar {
    position: relative;
    display: grid;
    place-items: center;
    width: var(--size);
    height: var(--size);
    flex: none;
    border-radius: calc(var(--size) * 0.28);
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    opacity: 0;
    transform: scale(0.85);
    transition:
      opacity var(--duration) var(--ease),
      transform var(--duration) var(--ease);
  }
  img.loaded {
    opacity: 1;
    transform: none;
  }
  .initial {
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    border-radius: inherit;
    background: hsl(var(--hue) 45% 22%);
    color: hsl(var(--hue) 80% 82%);
    font-weight: 700;
    font-size: calc(var(--size) * 0.44);
  }
</style>
