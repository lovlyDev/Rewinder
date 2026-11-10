import { writable } from 'svelte/store';
let nativeHidden = false;
export const visible = writable(!document.hidden);
export function setVisibility(shown: boolean) {
 nativeHidden = !shown;
 update();
}
function update() {
 const shown = !nativeHidden && !document.hidden;
 document.documentElement.classList.toggle('hidden', !shown);
 visible.set(shown);
}
export function watchVisibility() {
 const focus = () => { nativeHidden = false; update(); };
 document.addEventListener('visibilitychange', update);
 window.addEventListener('focus', focus);
 return () => { document.removeEventListener('visibilitychange', update); window.removeEventListener('focus', focus); };
}
