import type { SnapshotMeta } from './types';

/**
 * Which snapshot "undo" should offer first (list is newest first).
 *
 * Normally one step back from the current state. But when the current state is
 * itself the result of a restore, one step back would just undo that restore;
 * then continue from the snapshot that was restored, one step further back.
 */
export function undoIndex(list: SnapshotMeta[]): number {
  if (list.length < 2) return 0;
  const from = list[0].restoredFrom;
  if (from) {
    const i = list.findIndex((s) => s.snapshotID === from);
    if (i >= 0) return Math.min(list.length - 1, i + 1);
  }
  return 1;
}
