import { get, writable } from 'svelte/store';
import { errorKey, isKey } from '../i18n';
import type { Key } from '../i18n';
import { visible } from './visibility';

export type ToastKind = 'info' | 'success' | 'error';
export type Toast = { id: number; kind: ToastKind; code: Key; params: Record<string, unknown> };

const MAX = 3;
const LIFETIME = { info: 2600, success: 3200, error: 6000 } as const;

export const toasts = writable<Toast[]>([]);

let nextID = 0;
const timers = new Map<number, ReturnType<typeof setTimeout>>();

function arm(toast: Toast) {
  // While the window is hidden nothing is shown and nothing ticks.
  if (!get(visible) || timers.has(toast.id)) return;
  timers.set(toast.id, setTimeout(() => dismiss(toast.id), LIFETIME[toast.kind]));
}

visible.subscribe((shown) => {
  if (shown) get(toasts).forEach(arm);
  else {
    timers.forEach(clearTimeout);
    timers.clear();
  }
});

export function dismiss(id: number) {
  clearTimeout(timers.get(id));
  timers.delete(id);
  toasts.update((list) => list.filter((t) => t.id !== id));
}

export function pushMessage(kind: ToastKind, code: Key, params: Record<string, unknown> = {}) {
  const toast = { id: ++nextID, kind, code, params };
  const list = [...get(toasts), toast];
  for (const old of list.slice(0, Math.max(0, list.length - MAX))) {
    clearTimeout(timers.get(old.id));
    timers.delete(old.id);
  }
  toasts.set(list.slice(-MAX));
  arm(toast);
}

export function toastError(what: string, error: unknown) {
  pushMessage('error', errorKey(error, isKey(what) ? what : 'unknownError'));
}
