import { get } from 'svelte/store';
import { language, t } from '../i18n';
export function when(utcMillis: number, locale = get(language)) {return new Intl.DateTimeFormat(locale,{dateStyle:'short',timeStyle:'medium'}).format(utcMillis);}
export function windows(n: number) {return get(t)('windows',{n});}
