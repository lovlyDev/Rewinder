import { get, writable } from 'svelte/store';
import { GetAppSummary, GetApps, GetTimeline } from '../../wailsjs/go/main/App';
import { toastError } from './toasts';
import type { AppSummary, SnapshotCreatedEvent, SnapshotMeta } from './types';
import { visible } from './visibility';

export const apps = writable<AppSummary[]>([]);
export const selectedAppID = writable<string | null>(null);
export const timeline = writable<SnapshotMeta[]>([]);

// While the window is hidden, stores are not touched (no DOM work at all);
// changes collect here and are published when it is shown again.
let appCache: AppSummary[] = [];
let timelineCache: SnapshotMeta[] = [];
const pending = new Map<string, SnapshotMeta[]>();
let revision = 0;
const summaryRevisions = new Map<string, number>();
let refreshing: Promise<void> | undefined;

const newestFirst = (a: { lastActivityUTC: number }, b: { lastActivityUTC: number }) => b.lastActivityUTC - a.lastActivityUTC;

/** Union by snapshot ID, newest first, capped like the backend. */
function merge(a: SnapshotMeta[], b: SnapshotMeta[]) {
  return Array.from(new Map([...a, ...b].map((s) => [s.snapshotID, s])).values())
    .sort((x, y) => y.timestampUTC - x.timestampUTC)
    .slice(0, 500);
}

function publish() {
  if (!get(visible)) return;
  apps.set(appCache);
  timeline.set(timelineCache);
}

export async function loadApps(): Promise<AppSummary[]> {
  try {
    const at = revision;
    const list = await GetApps();
    if (at === revision) appCache = list; // a snapshot arrived meanwhile: keep its newer counts
    publish();
  } catch (e) {
    toastError('loadFailed', e);
  }
  return appCache;
}

export async function loadTimeline(appID: string): Promise<SnapshotMeta[]> {
  try {
    return await GetTimeline(appID);
  } catch (e) {
    toastError('loadFailed', e);
    return [];
  }
}

export async function selectApp(appID: string) {
  if (get(selectedAppID) !== appID) {
    selectedAppID.set(appID);
    timelineCache = [];
    publish();
  }
  const list = await loadTimeline(appID);
  if (get(selectedAppID) === appID) {
    timelineCache = merge(list, timelineCache);
    publish();
  }
}

/** Reloads everything; concurrent calls share one request. */
export function refresh(): Promise<void> {
  refreshing ??= (async () => {
    const list = await loadApps();
    const current = get(selectedAppID);
    const id = current && list.some((a) => a.appID === current) ? current : list[0]?.appID;
    if (id && get(visible)) await selectApp(id);
  })().finally(() => (refreshing = undefined));
  return refreshing;
}

async function applySnapshots(appID: string, snapshots: SnapshotMeta[]) {
  revision++;
  const mine = (summaryRevisions.get(appID) ?? 0) + 1;
  summaryRevisions.set(appID, mine);
  if (get(selectedAppID) === appID) timelineCache = merge(timelineCache, snapshots);
  const latest = snapshots[0].timestampUTC;
  appCache = appCache
    .map((a) => (a.appID === appID ? ({ ...a, lastActivityUTC: Math.max(a.lastActivityUTC, latest) } as AppSummary) : a))
    .sort(newestFirst);
  publish();
  try {
    const summary = await GetAppSummary(appID);
    if (summaryRevisions.get(appID) !== mine || !summary) return;
    appCache = [summary, ...appCache.filter((a) => a.appID !== appID)].sort(newestFirst);
    publish();
    if (!get(selectedAppID) && get(visible)) await selectApp(appID);
  } catch (e) {
    toastError('loadFailed', e);
  }
}

export function onSnapshotCreated(p: SnapshotCreatedEvent) {
  if (!get(visible)) {
    pending.set(p.appID, merge(pending.get(p.appID) ?? [], [p.snapshot]));
    return;
  }
  void applySnapshots(p.appID, [p.snapshot]);
}

visible.subscribe((shown) => {
  if (!shown) return;
  publish();
  for (const [id, list] of pending) void applySnapshots(id, list);
  pending.clear();
});
