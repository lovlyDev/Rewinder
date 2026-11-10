import { cubicOut } from 'svelte/easing';
import type { TransitionConfig } from 'svelte/transition';

/** Animation length for JS transitions, following the user's motion setting. */
export function duration(base = 220): number {
  const root = document.documentElement;
  if (root.dataset.motion === 'off' || root.classList.contains('hidden')) return 0;
  if (matchMedia('(prefers-reduced-motion: reduce)').matches) return 0;
  return root.dataset.motion === 'fast' ? Math.round(base * 0.65) : base;
}

// expo-out: fast start, soft landing
const ease = (t: number) => (t === 1 ? 1 : 1 - Math.pow(2, -10 * t));

export function rise(_node: Element, { y = 6, delay = 0 } = {}): TransitionConfig {
  return {
    delay,
    duration: duration(),
    easing: ease,
    css: (t) => `opacity:${t};transform:translateY(${(1 - t) * y}px)`,
  };
}

export function pop(_node: Element, { from = 0.96 } = {}): TransitionConfig {
  return {
    duration: duration(),
    easing: ease,
    css: (t) => `opacity:${t};transform:scale(${from + (1 - from) * t})`,
  };
}

export function fade(_node: Element, { base = 160 } = {}): TransitionConfig {
  return { duration: duration(base), easing: cubicOut, css: (t) => `opacity:${t}` };
}
