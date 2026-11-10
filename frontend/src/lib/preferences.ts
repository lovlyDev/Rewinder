import { get, writable } from 'svelte/store';
import { GetAppInfo, GetBackgroundImage, GetSettings, SaveSettings, SetHotkey } from '../../wailsjs/go/main/App';
import type { settings } from '../../wailsjs/go/models';
import { applyTheme, setImage } from '../theme/theme';
import { toastError } from './toasts';

export type Settings = settings.Settings;

export const preferences = writable<Settings | null>(null);
export const appInfo = writable({
  version: '',
  logPath: '',
  historyPath: '',
  systemLanguage: 'en',
  elevated: false,
  elevationDeclined: false,
});

// Writes run one after another, so a fast slider can never persist an older value last.
let queue: Promise<unknown> = Promise.resolve();

export async function loadPreferences() {
  const [s, info, image] = await Promise.all([
    GetSettings(),
    GetAppInfo(),
    GetBackgroundImage().catch((e) => {
      toastError('background.read', e);
      return '';
    }),
  ]);
  appInfo.set(info);
  setImage(image);
  preferences.set(s);
  await applyTheme(s, info.systemLanguage);
}

/** Applies a change immediately and persists it; on failure the saved state comes back. */
export function savePreferences(next: Settings): Promise<void> {
  const copy = structuredClone(next) as Settings;
  preferences.set(copy);
  void applyTheme(copy, get(appInfo).systemLanguage);
  const result = queue.then(async () => {
    // The hotkey and the admin mode have their own calls; never let a theme edit undo them.
    const saved = await GetSettings();
    copy.hotkey = saved.hotkey;
    copy.runAsAdmin = saved.runAsAdmin;
    await SaveSettings(copy);
  });
  queue = result.catch(() => {});
  return result.catch(async (error) => {
    await queue;
    const actual = await GetSettings();
    preferences.set(actual);
    await applyTheme(actual, get(appInfo).systemLanguage);
    throw error;
  });
}

/** Live preview without saving (slider drag, colour picker). */
export function previewPreferences(next: Settings) {
  void applyTheme(next, get(appInfo).systemLanguage);
}

export function saveHotkey(mods: number, key: number): Promise<void> {
  const result = queue.then(async () => {
    await SetHotkey(mods, key);
    preferences.set(await GetSettings());
  });
  queue = result.catch(() => {});
  return result;
}
