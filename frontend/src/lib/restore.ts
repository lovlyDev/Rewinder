import { get, writable } from 'svelte/store';
import { Restore } from '../../wailsjs/go/main/App';
import { t } from '../i18n';
import { pushToast, toastError } from './toasts';
export const restoring = writable(false);
export async function restore(appID: string,snapshotID: string) {
 if(get(restoring))return;
 restoring.set(true);pushToast('info',get(t)('restoring'));
 try {await Restore(appID,snapshotID);}catch(e){toastError('restore.failed',e);}finally{restoring.set(false);}
}
