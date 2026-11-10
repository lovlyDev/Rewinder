import type { settings } from '../../wailsjs/go/models';
import { language } from '../i18n';
let image = '';
let cacheKey = '';
let rendered = '';
let generation = 0;
export function setImage(data: string) { image = data; cacheKey = ''; }
export async function applyTheme(s: settings.Settings, systemLanguage = 'en') {
 const root = document.documentElement;
 const theme = s.theme;
 root.style.setProperty('--accent', theme.accent);
 root.style.setProperty('--accent-soft', `${theme.accent}18`);
 root.style.setProperty('--radius', `${theme.radius}px`);
 root.style.setProperty('--space', theme.density === 'compact' ? '10px' : '16px');
 root.style.setProperty('--row', theme.density === 'compact' ? '54px' : '66px');
 root.style.setProperty('--duration', theme.motion === 'off' ? '0ms' : theme.motion === 'fast' ? '120ms' : '180ms');
 root.style.setProperty('--quick', theme.motion === 'off' ? '0ms' : '140ms');
 root.dataset.motion = theme.motion;
 const lang = s.language === 'system' ? systemLanguage : s.language;
 language.set(lang === 'ru' ? 'ru' : 'en'); root.lang = lang;
 const b = theme.background;
 root.style.setProperty('--bg', b.color);
 const current = ++generation;
 let background = 'none';
 if (b.kind === 'gradient') background = `linear-gradient(135deg, ${b.color}, ${b.endColor})`;
 if (b.kind === 'image' && image) {
  const key = `${b.blur}:${b.dim}:${image.length}`;
  if (cacheKey !== key) {
   const img = new Image(); img.src = image; await img.decode();
   const ratio = Math.min(1, 1600 / img.width, 1200 / img.height);
   const canvas = document.createElement('canvas'); canvas.width = Math.round(img.width*ratio); canvas.height = Math.round(img.height*ratio);
   const ctx = canvas.getContext('2d'); if (!ctx) return;
   ctx.filter = `blur(${b.blur}px)`;
   ctx.drawImage(img, -b.blur*2, -b.blur*2, canvas.width+b.blur*4, canvas.height+b.blur*4);
   ctx.filter = 'none'; ctx.fillStyle = `rgba(0,0,0,${b.dim})`;ctx.fillRect(0,0,canvas.width,canvas.height);
   if (current !== generation) return;
   rendered = canvas.toDataURL('image/jpeg', .85); cacheKey = key;
  }
  background = `url("${rendered}")`;
 }
 if (current === generation) root.style.setProperty('--background-image', background);
}
