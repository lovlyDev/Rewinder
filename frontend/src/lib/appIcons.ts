import { GetAppIcon } from '../../wailsjs/go/main/App';

// appID -> data URL ('' = no icon). Go caches too; this keeps repeat renders free.
const cache = new Map<string, Promise<string>>();

export function appIcon(appID: string): Promise<string> {
  const hit = cache.get(appID);
  if (hit) return hit;
  const p = GetAppIcon(appID).catch(() => '');
  cache.set(appID, p);
  return p;
}
