import { get, writable } from 'svelte/store';
import { visible } from '../../lib/visibility';
import { message } from '../../i18n';
export type Toast = {id: number; kind: 'info' | 'success' | 'error'; text: string};
export const toasts = writable<Toast[]>([]);
let pending: Toast[] = [];
let nextID = 0;
let timer: ReturnType<typeof setTimeout> | undefined;
function schedule() {
 clearTimeout(timer);timer=undefined;
 if (!get(visible)) return;
 toasts.set(pending);
 if(pending.length) timer=setTimeout(() => {pending=pending.slice(1);schedule();},4000);
}
visible.subscribe(() => schedule());
export function dismiss(id: number) {pending=pending.filter(t => t.id!==id);schedule();}
export function pushToast(kind: Toast['kind'], text: string) { pending=[...pending,{id:++nextID,kind,text}].slice(-4);schedule(); }
export function toastError(_what: string, error: unknown) {pushToast('error', message(error));}
