import { GetSnapshotLayout } from '../../wailsjs/go/main/App';
import type { ipcapi } from '../../wailsjs/go/models';

export type SnapshotLayout = ipcapi.SnapshotLayout;

// Snapshots never change, so their layouts are cached for the session.
const cache = new Map<string, Promise<SnapshotLayout | null>>();

export function snapshotLayout(appID: string, snapshotID: string): Promise<SnapshotLayout | null> {
  const hit = cache.get(snapshotID);
  if (hit) return hit;
  const p = GetSnapshotLayout(appID, snapshotID).catch(() => null);
  cache.set(snapshotID, p);
  if (cache.size > 200) cache.delete(cache.keys().next().value as string);
  return p;
}
