import type { ipcapi } from '../../wailsjs/go/models';

export type AppSummary = ipcapi.AppSummary;
export type SnapshotMeta = ipcapi.SnapshotMeta;

export type View = 'timeline' | 'overlay' | 'settings';
export type TrackingState = 'active' | 'paused' | 'error';

// Payloads of backend events (see internal/ipcapi/types.go).
export type SnapshotCreatedEvent = { appID: string; snapshot: SnapshotMeta; occurredAtUTC: number };
export type RestoreProgressEvent = { appID: string; snapshotID: string; stage: string; percent: number; message: string; code: string; params?: Record<string, unknown> };
export type TrackingStateChangedEvent = { appID?: string | null; state: string; reason?: string; atUTC: number };
export type OverlayRequest = { appID?: string };
