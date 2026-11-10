import { derived, get, writable } from 'svelte/store';
import en from './en';
import ru from './ru';
export type Key = keyof typeof en;
export const language = writable<'ru' | 'en'>(navigator.language.startsWith('ru') ? 'ru' : 'en');
export const t = derived(language, lang => (key: Key, params: Record<string, unknown> = {}) => {
 const text: string = (lang === 'ru' ? ru : en)[key];
 return text.replace(/\{(\w+)\}/g, (_, name: string) => String(params[name] ?? ''));
});
export function message(error: unknown): string {
 const code = error instanceof Error ? error.message : String(error);
 const key = code.split(':')[0] as Key;
 return get(t)(key in en ? key : 'unknownError');
}
