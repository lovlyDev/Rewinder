import { get } from 'svelte/store';
import { language, t } from '../i18n';

const cache = new Map<string, Intl.DateTimeFormat | Intl.RelativeTimeFormat>();
function fmt<T extends Intl.DateTimeFormat | Intl.RelativeTimeFormat>(key: string, make: () => T): T {
  let f = cache.get(key) as T | undefined;
  if (!f) cache.set(key, (f = make()));
  return f;
}

/** "21:10" */
export function clock(ms: number, locale = get(language)) {
  return fmt(`c${locale}`, () => new Intl.DateTimeFormat(locale, { hour: '2-digit', minute: '2-digit' })).format(ms);
}

/** "21:10:04" */
export function clockSeconds(ms: number, locale = get(language)) {
  return fmt(`s${locale}`, () => new Intl.DateTimeFormat(locale, { timeStyle: 'medium' })).format(ms);
}

/** "10.11.2026, 21:10:04" */
export function when(ms: number, locale = get(language)) {
  return fmt(`w${locale}`, () => new Intl.DateTimeFormat(locale, { dateStyle: 'short', timeStyle: 'medium' })).format(ms);
}

const startOfDay = (ms: number) => new Date(ms).setHours(0, 0, 0, 0);

/** "Today", "Yesterday" or "8 November". */
export function day(ms: number, locale = get(language), now = Date.now()) {
  const diff = Math.round((startOfDay(now) - startOfDay(ms)) / 86_400_000);
  if (diff === 0) return get(t)('today');
  if (diff === 1) return get(t)('yesterday');
  return fmt(`d${locale}`, () => new Intl.DateTimeFormat(locale, { day: 'numeric', month: 'long', weekday: 'short' })).format(ms);
}

export const dayKey = (ms: number) => startOfDay(ms);

/** "just now", "5 min ago", "2 h ago", "3 days ago". */
export function ago(ms: number, locale = get(language), now = Date.now()) {
  const s = Math.max(0, Math.round((now - ms) / 1000));
  if (s < 45) return get(t)('justNow');
  const rtf = fmt(`r${locale}`, () => new Intl.RelativeTimeFormat(locale, { numeric: 'auto', style: 'short' }));
  if (s < 3600) return rtf.format(-Math.round(s / 60), 'minute');
  if (s < 86_400) return rtf.format(-Math.round(s / 3600), 'hour');
  return rtf.format(-Math.round(s / 86_400), 'day');
}

/** "ChatGPT.exe" → "ChatGPT" */
export function appName(name: string) {
  return name.replace(/\.exe$/i, '');
}
