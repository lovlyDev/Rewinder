// Mirrors internal/hotkey: MOD_ALT=1, MOD_CONTROL=2, MOD_SHIFT=4, MOD_WIN=8; key = virtual-key code.
export type Combination = { mods: number; key: number };

const MODS: [number, string][] = [
  [2, 'Ctrl'],
  [1, 'Alt'],
  [4, 'Shift'],
  [8, 'Win'],
];

export function keyName(vk: number): string {
  if (vk >= 112 && vk <= 135) return `F${vk - 111}`;
  return String.fromCharCode(vk);
}

export function parts(c: Combination): string[] {
  return [...MODS.filter(([bit]) => c.mods & bit).map(([, name]) => name), keyName(c.key)];
}

/** Reads a combination from a keydown, or null while only modifiers are held. */
export function fromEvent(e: KeyboardEvent): Combination | null {
  const code = e.code;
  const key = /^Key[A-Z]$/.test(code)
    ? code.charCodeAt(3)
    : /^Digit\d$/.test(code)
      ? code.charCodeAt(5)
      : /^F([1-9]|1\d|2[0-4])$/.test(code)
        ? 111 + Number(code.slice(1))
        : 0;
  const mods = (e.altKey ? 1 : 0) | (e.ctrlKey ? 2 : 0) | (e.shiftKey ? 4 : 0) | (e.metaKey ? 8 : 0);
  if (!key || !mods) return null;
  return { mods, key };
}

/** Held modifiers, for live feedback while recording. */
export function heldModifiers(e: KeyboardEvent): string[] {
  const held = (e.ctrlKey ? 2 : 0) | (e.altKey ? 1 : 0) | (e.shiftKey ? 4 : 0) | (e.metaKey ? 8 : 0);
  return MODS.filter(([bit]) => held & bit).map(([, name]) => name);
}
