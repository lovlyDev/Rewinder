import type { settings } from '../../wailsjs/go/models';
import { language } from '../i18n';

let image = '';
let cacheKey = '';
let rendered = '';
let generation = 0;

export function setImage(data: string) {
  image = data;
  cacheKey = '';
}

/** Dark text on light accents, light text on dark ones. */
function inkFor(hex: string): string {
  const n = parseInt(hex.slice(1), 16);
  const [r, g, b] = [(n >> 16) & 255, (n >> 8) & 255, n & 255];
  return 0.299 * r + 0.587 * g + 0.114 * b > 150 ? '#0d0b16' : '#ffffff';
}

const motion = {
  fast: { quick: '90ms', duration: '150ms' },
  normal: { quick: '140ms', duration: '220ms' },
  off: { quick: '0ms', duration: '0ms' },
} as const;

/** Applies the user's appearance settings to :root. Cheap; safe to call on every slider tick. */
export async function applyTheme(s: settings.Settings, systemLanguage = 'en') {
  const root = document.documentElement;
  const theme = s.theme;
  const set = (name: string, value: string) => root.style.setProperty(name, value);

  set('--accent', theme.accent);
  set('--accent-soft', `${theme.accent}1f`);
  set('--accent-glow', `${theme.accent}40`);
  set('--accent-ink', inkFor(theme.accent));
  set('--radius', `${theme.radius}px`);
  set('--space', theme.density === 'compact' ? '12px' : '16px');
  set('--row', theme.density === 'compact' ? '50px' : '60px');
  const m = motion[theme.motion as keyof typeof motion] ?? motion.normal;
  set('--quick', m.quick);
  set('--duration', m.duration);
  root.dataset.motion = theme.motion;

  const lang = s.language === 'system' ? systemLanguage : s.language;
  language.set(lang === 'ru' ? 'ru' : 'en');
  root.lang = lang;

  const b = theme.background;
  set('--bg', b.color);
  const current = ++generation;
  let background = 'none';
  if (b.kind === 'gradient') background = `radial-gradient(120% 90% at 0% 0%, ${b.endColor}, ${b.color} 70%)`;
  if (b.kind === 'image' && image) {
    // Blur and darken once into a static picture: animating a CSS blur would be
    // far too costly with WebView2 rendering on the CPU.
    const key = `${b.blur}:${b.dim}:${image.length}`;
    if (cacheKey !== key) {
      const img = new Image();
      img.src = image;
      await img.decode();
      const ratio = Math.min(1, 1600 / img.width, 1200 / img.height);
      const canvas = document.createElement('canvas');
      canvas.width = Math.round(img.width * ratio);
      canvas.height = Math.round(img.height * ratio);
      const ctx = canvas.getContext('2d');
      if (!ctx) return;
      ctx.filter = `blur(${b.blur}px)`;
      ctx.drawImage(img, -b.blur * 2, -b.blur * 2, canvas.width + b.blur * 4, canvas.height + b.blur * 4);
      ctx.filter = 'none';
      ctx.fillStyle = `rgba(0,0,0,${b.dim})`;
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      if (current !== generation) return;
      rendered = canvas.toDataURL('image/jpeg', 0.85);
      cacheKey = key;
    }
    background = `url("${rendered}")`;
  }
  if (current === generation) set('--background-image', background);
}
