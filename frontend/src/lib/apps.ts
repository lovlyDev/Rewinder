import { get, writable } from 'svelte/store';
import { GetApps, GetTimeline } from '../../wailsjs/go/main/App.js';
import { toastError } from './toasts';
import type { AppSummary, SnapshotCreatedEvent, SnapshotMeta } from './types';

export const apps = writable<AppSummary[]>([]);
export const selectedAppID = writable<string | null>(null);
export const timeline = writable<SnapshotMeta[]>([]);

export async function loadApps(): Promise<AppSummary[]> {
  try {
    const list = await GetApps();
    apps.set(list);
    return list;
  } catch (e) {
    toastError('Could not load apps', e);
    return get(apps);
  }
}

export async function loadTimeline(appID: string): Promise<SnapshotMeta[]> {
  try {
    return await GetTimeline(appID);
  } catch (e) {
    toastError('Could not load timeline', e);
    return [];
  }
}

export async function selectApp(appID: string) {
  selectedAppID.set(appID);
  const list = await loadTimeline(appID);
  if (get(selectedAppID) === appID) timeline.set(list); // ignore a stale answer
}

// Loads the app list and keeps (or picks) a selected app.
export async function refresh() {
  const list = await loadApps();
  const current = get(selectedAppID);
  const id = current && list.some(a => a.appID === current) ? current : list[0]?.appID;
  if (id) await selectApp(id);
}

// Applies a new snapshot without reloading everything.
export function onSnapshotCreated(p: SnapshotCreatedEvent) {
  let known = false;
  apps.update(list =>
    list.map(a => {
      if (a.appID !== p.appID) return a;
      known = true;
      return { ...a, snapshotCount: a.snapshotCount + 1, lastActivityUTC: p.snapshot.timestampUTC } as AppSummary;
    }),
  );
  if (!known) void loadApps(); // a newly tracked app
  if (get(selectedAppID) === p.appID) timeline.update(list => [p.snapshot, ...list]);
}
