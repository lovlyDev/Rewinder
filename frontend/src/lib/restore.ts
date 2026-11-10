import { get, writable } from 'svelte/store';
import { Restore } from '../../wailsjs/go/main/App';
import { errorKey } from '../i18n';
import { pushMessage, toastError } from './toasts';

export const restoring = writable(false);

// Snapshots whose failure the backend already reported through onRestoreError,
// so the rejected promise of Restore does not show a second toast.
const reported = new Set<string>();

export function onRestoreError(p: { snapshotID: string; code: string; params?: Record<string, unknown> }) {
  reported.add(p.snapshotID);
  pushMessage('error', errorKey(p.code), p.params);
}

export async function restore(appID: string, snapshotID: string) {
  if (get(restoring)) return;
  restoring.set(true);
  pushMessage('info', 'restoring');
  try {
    await Restore(appID, snapshotID);
  } catch (e) {
    if (!reported.has(snapshotID)) toastError('restore.failed', e);
  } finally {
    reported.delete(snapshotID);
    restoring.set(false);
  }
}
