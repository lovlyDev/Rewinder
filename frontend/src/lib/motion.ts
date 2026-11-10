export const easing = (x: number) => 1 - Math.pow(1 - x, 3);
export function duration(): number {
 if (document.documentElement.dataset.motion === 'off' || matchMedia('(prefers-reduced-motion: reduce)').matches || document.documentElement.classList.contains('hidden')) return 0;
 return document.documentElement.dataset.motion === 'fast' ? 120 : 180;
}
export function reveal(_node: Element, { scale = false } = {}) {
 return { duration: duration(), easing, css: (t: number) => `opacity:${t};transform:${scale ? `scale(${.96 + .04*t})` : `translateY(${(1-t)*6}px)`}` };
}
