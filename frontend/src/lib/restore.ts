import { Restore } from '../../wailsjs/go/main/App.js';
import { pushToast, toastError } from './toasts';

// The success message comes from the backend's "done" progress event (see App.svelte).
export async function restore(appID: string, snapshotID: string) {
  pushToast('info', 'Restoring…');
  try {
    await Restore(appID, snapshotID);
  } catch (e) {
    toastError('Restore failed', e);
  }
}
