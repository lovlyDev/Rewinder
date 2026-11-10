import { writable } from 'svelte/store';
import type { TrackingState, View } from './types';

export const view = writable<View>('timeline');
export const trackingState = writable<TrackingState>('active');
export const settingsOpen = writable(false);
/** Section the settings dialog should show next time it opens. */
export const settingsSection = writable<string | null>(null);
export const appPaused = writable<Record<string, boolean>>({});
/** What the sidebar lists and the main area shows: app history or saved layouts. */
export const sidebarMode = writable<'apps' | 'layouts'>('apps');

// Each hotkey press creates a new request, so the overlay reloads even if it is already open.
export const overlayRequest = writable<{ appID?: string; seq: number }>({ seq: 0 });

export function requestOverlay(appID?: string) {
  overlayRequest.update(r => ({ appID, seq: r.seq + 1 }));
  view.set('overlay');
}
