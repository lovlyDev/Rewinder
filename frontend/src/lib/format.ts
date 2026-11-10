const dateTime = new Intl.DateTimeFormat(undefined, { dateStyle: 'short', timeStyle: 'medium' });

export function when(utcMillis: number): string {
  return dateTime.format(utcMillis);
}

export function windows(n: number): string {
  return n === 1 ? '1 window' : `${n} windows`;
}
