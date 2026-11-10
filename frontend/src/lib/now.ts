import { readable } from 'svelte/store';
import { visible } from './visibility';

/** Current time for "5 min ago" labels. Ticks every 30 s, and only while the window is visible. */
export const now = readable(Date.now(), (set) => {
  let timer: ReturnType<typeof setInterval> | undefined;
  const stop = visible.subscribe((shown) => {
    clearInterval(timer);
    timer = undefined;
    if (!shown) return;
    set(Date.now());
    timer = setInterval(() => set(Date.now()), 30_000);
  });
  return () => {
    stop();
    clearInterval(timer);
  };
});
