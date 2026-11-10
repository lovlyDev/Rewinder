import { get, writable } from 'svelte/store';
import { GetApps, GetTimeline, GetAppSummary } from '../../wailsjs/go/main/App';
import { visible } from './visibility';
import { toastError } from './toasts';
import type { AppSummary, SnapshotCreatedEvent, SnapshotMeta } from './types';
export const apps = writable<AppSummary[]>([]);
export const selectedAppID = writable<string | null>(null);
export const timeline = writable<SnapshotMeta[]>([]);
const pending = new Map<string, SnapshotMeta[]>();
let revision = 0;
let appCache: AppSummary[] = [];
let timelineCache: SnapshotMeta[] = [];
let refreshPromise: Promise<void> | undefined;
const merge = (a: SnapshotMeta[], b: SnapshotMeta[]) => Array.from(new Map([...a,...b].map(s => [s.snapshotID,s])).values()).sort((a,b) => b.timestampUTC-a.timestampUTC).slice(0,500);
function publish() { if(get(visible)) {apps.set(appCache);timeline.set(timelineCache);} }
export async function loadApps(): Promise<AppSummary[]> {
 try {const currentRevision=revision;const list=await GetApps();if(currentRevision===revision) appCache=list;publish();return appCache;}
 catch(e) {toastError('loadFailed',e);return appCache;}
}
export async function loadTimeline(appID: string): Promise<SnapshotMeta[]> {
 try {return await GetTimeline(appID);}catch(e){toastError('loadFailed',e);return [];}
}
export async function selectApp(appID: string) {
 selectedAppID.set(appID);timelineCache=[];publish();
 const list=await loadTimeline(appID);
 if(get(selectedAppID)===appID){timelineCache=merge(list,timelineCache);publish();}
}
export function refresh(): Promise<void> {
 if(refreshPromise)return refreshPromise;
 refreshPromise=(async () => {
  const list=await loadApps(); const current=get(selectedAppID);
  const id=current && list.some(a=>a.appID===current) ? current : list[0]?.appID;
  if(id && get(visible)) {
   if(id!==current) await selectApp(id);
   else {const list=await loadTimeline(id);if(get(selectedAppID)===id){timelineCache=merge(list,timelineCache);publish();}}
  }
 })().finally(()=>refreshPromise=undefined);
 return refreshPromise;
}
async function applySnapshots(appID: string, snapshots: SnapshotMeta[]) {
 revision++;
 if(get(selectedAppID)===appID) timelineCache=merge(timelineCache,snapshots);
 const latest=snapshots[0];
 appCache=appCache.map(a=>a.appID===appID ? {...a,lastActivityUTC:Math.max(a.lastActivityUTC,latest.timestampUTC)} as AppSummary : a).sort((a,b)=>b.lastActivityUTC-a.lastActivityUTC);
 publish();
 try {
  const summary=await GetAppSummary(appID);
  if(summary){appCache=[summary,...appCache.filter(a=>a.appID!==appID)].sort((a,b)=>b.lastActivityUTC-a.lastActivityUTC);publish();if(!get(selectedAppID)&&get(visible))await selectApp(appID);}
 } catch(e){toastError('loadFailed',e);}
}
export function onSnapshotCreated(p: SnapshotCreatedEvent) {
 if(!get(visible)){pending.set(p.appID,merge(pending.get(p.appID)??[],[p.snapshot]));return;}
 void applySnapshots(p.appID,[p.snapshot]);
}
visible.subscribe(shown=>{if(!shown)return;publish();for(const [id,list] of pending)void applySnapshots(id,list);pending.clear();});
