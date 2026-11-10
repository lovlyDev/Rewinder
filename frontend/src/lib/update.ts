import { writable } from 'svelte/store';
import { CheckForUpdate, InstallUpdate } from '../../wailsjs/go/main/App';
import type { main } from '../../wailsjs/go/models';
import { pushMessage, toastError } from './toasts';

export type UpdateInfo = main.UpdateInfo;

export const updateInfo = writable<UpdateInfo | null>(null);
export const updateState = writable<'idle' | 'checking' | 'downloading' | 'restarting'>('idle');
export const updateProgress = writable(0);

export async function checkForUpdate(quiet = false) {
  updateState.set('checking');
  try {
    const info = await CheckForUpdate();
    updateInfo.set(info);
    if (!quiet && !info.available) pushMessage('success', 'upToDate', { version: info.current });
  } catch (e) {
    if (!quiet) toastError('update.check', e);
  } finally {
    updateState.set('idle');
  }
}

export async function installUpdate() {
  updateState.set('downloading');
  updateProgress.set(0);
  try {
    await InstallUpdate();
    updateState.set('restarting');
  } catch (e) {
    updateState.set('idle');
    toastError('update.failed', e);
  }
}

export function onUpdateAvailable(info: UpdateInfo) {
  updateInfo.set(info);
  pushMessage('info', 'updateAvailable', { version: info.version });
}
