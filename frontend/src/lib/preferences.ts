import { get, writable } from 'svelte/store';
import { GetSettings, SaveSettings, SetHotkey, GetAppInfo, GetBackgroundImage } from '../../wailsjs/go/main/App';
import type { settings } from '../../wailsjs/go/models';
import { applyTheme, setImage } from '../theme/theme';
import { toastError } from './toasts';
export const preferences = writable<settings.Settings | null>(null);
export const appInfo = writable({ version: '', logPath: '', historyPath: '', systemLanguage: 'en' });
let queue = Promise.resolve();
export async function loadPreferences() {
 const [s, info, image] = await Promise.all([GetSettings(), GetAppInfo(), GetBackgroundImage().catch(e=>{toastError('background.read',e);return '';})]);
 appInfo.set(info); setImage(image); preferences.set(s); await applyTheme(s, info.systemLanguage);
}
// Serialize writes: rapid slider input must never persist an older value last.
export async function savePreferences(s: settings.Settings): Promise<void> {
 const copy = JSON.parse(JSON.stringify(s)) as settings.Settings;
 preferences.set(copy); void applyTheme(copy, get(appInfo).systemLanguage);
 const result = queue.then(async () => {
  // Theme/language edits must not undo a hotkey change queued just before them.
  copy.hotkey=(await GetSettings()).hotkey;
  await SaveSettings(copy);
 });
 queue = result.catch(() => {});
 return result.catch(async error => { await queue; const actual = await GetSettings(); preferences.set(actual); await applyTheme(actual, get(appInfo).systemLanguage); throw error; });
}
export function saveHotkey(mods:number,key:number): Promise<void> {
 const result=queue.then(async()=>{await SetHotkey(mods,key);preferences.set(await GetSettings());});
 queue=result.catch(()=>{});
 return result;
}
