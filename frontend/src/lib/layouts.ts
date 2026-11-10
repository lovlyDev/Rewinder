import { get, writable } from 'svelte/store';
import {
  DeleteLayout,
  GetLayoutPreview,
  GetLayouts,
  MoveLayout,
  RenameLayout,
  RestoreLayout,
  SaveLayout,
} from '../../wailsjs/go/main/App';
import type { services } from '../../wailsjs/go/models';
import type { SnapshotLayout } from './preview';
import { pushMessage, toastError } from './toasts';

export type LayoutSummary = services.LayoutSummary;

export const layouts = writable<LayoutSummary[]>([]);
export const selectedLayoutID = writable<string | null>(null);
export const restoringLayout = writable<string | null>(null);

export async function loadLayouts() {
  try {
    const list = (await GetLayouts()) ?? [];
    layouts.set(list);
    const sel = get(selectedLayoutID);
    if (sel && !list.some((l) => l.id === sel)) selectedLayoutID.set(list[0]?.id ?? null);
    if (!sel && list.length) selectedLayoutID.set(list[0].id);
  } catch (e) {
    toastError('loadFailed', e);
  }
}

export async function saveLayout(name: string) {
  try {
    const l = await SaveLayout(name);
    await loadLayouts();
    selectedLayoutID.set(l.id);
    pushMessage('success', 'layoutSaved', { name: l.name, n: l.apps.length });
    return true;
  } catch (e) {
    toastError('layout.failed', e);
    return false;
  }
}

export async function renameLayout(id: string, name: string) {
  try {
    await RenameLayout(id, name);
    await loadLayouts();
  } catch (e) {
    toastError('layout.name', e);
  }
}

export async function deleteLayout(id: string) {
  try {
    await DeleteLayout(id);
    await loadLayouts();
  } catch (e) {
    toastError('layout.missing', e);
  }
}

export async function moveLayout(id: string, to: number) {
  try {
    await MoveLayout(id, to);
    await loadLayouts();
  } catch (e) {
    toastError('layout.missing', e);
  }
}

export async function restoreLayout(id: string) {
  if (get(restoringLayout)) return;
  restoringLayout.set(id);
  try {
    // success / partial results arrive through the onLayoutRestored event
    await RestoreLayout(id);
  } catch (e) {
    toastError('layout.failed', e);
  } finally {
    restoringLayout.set(null);
  }
}

const previews = new Map<string, Promise<SnapshotLayout | null>>();
export function layoutPreview(id: string, version: number): Promise<SnapshotLayout | null> {
  // version: the layout's creation time, so a re-saved layout gets a fresh picture
  const key = `${id}:${version}`;
  let p = previews.get(key);
  if (!p) {
    p = GetLayoutPreview(id).catch(() => null);
    previews.set(key, p);
  }
  return p;
}

export function onLayoutRestored(r: { name: string; restored: number; total: number; failed: string[] }) {
  if (r.restored === r.total) pushMessage('success', 'layoutRestored', { name: r.name, n: r.restored });
  else if (r.restored > 0) pushMessage('info', 'layoutPartial', { name: r.name, n: r.restored, total: r.total, failed: r.failed.join(', ') });
}
