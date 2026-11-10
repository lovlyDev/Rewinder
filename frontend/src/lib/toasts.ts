import { writable } from 'svelte/store';

export type Toast = { id: number; kind: 'info' | 'success' | 'error'; text: string };

const MAX = 5;
const LIFETIME_MS = { info: 2000, success: 2500, error: 5000 };

let nextID = 0;
export const toasts = writable<Toast[]>([]);

export function pushToast(kind: Toast['kind'], text: string) {
  const id = ++nextID;
  toasts.update(list => [...list, { id, kind, text }].slice(-MAX));
  setTimeout(() => toasts.update(list => list.filter(t => t.id !== id)), LIFETIME_MS[kind]);
}

export function errorText(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

export function toastError(what: string, e: unknown) {
  pushToast('error', `${what}: ${errorText(e)}`);
}
