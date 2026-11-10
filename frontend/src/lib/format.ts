import { get } from 'svelte/store';
import { language, t } from '../i18n';
export function when(utcMillis: number) {return new Intl.DateTimeFormat(get(language),{dateStyle:'short',timeStyle:'medium'}).format(utcMillis);}
export function windows(n: number) {return get(t)('windows',{n});}
