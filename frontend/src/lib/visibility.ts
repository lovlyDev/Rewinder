import { writable } from 'svelte/store';

// Whether the window is actually on screen. Wails' WindowHide does not always
// make document.hidden true, so the backend also reports its own show/hide.
let nativeHidden = false;
export const visible = writable(!document.hidden);

function update() {
  const shown = !nativeHidden && !document.hidden;
  document.documentElement.classList.toggle('hidden', !shown);
  visible.set(shown);
}

export function setVisibility(shown: boolean) {
  nativeHidden = !shown;
  update();
}

export function watchVisibility() {
  const focus = () => setVisibility(true);
  document.addEventListener('visibilitychange', update);
  window.addEventListener('focus', focus);
  return () => {
    document.removeEventListener('visibilitychange', update);
    window.removeEventListener('focus', focus);
  };
}
