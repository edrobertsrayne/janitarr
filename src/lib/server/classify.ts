import type { QueueItem } from './arr';
import type { Config } from './settings-defs';
import type { Progress } from './store';

const FAILED_STATES: Record<string, true> = {
	failed: true,
	failedPending: true,
	importFailed: true
};

/** A download or import that has failed and will not recover on its own. */
export function isFailed(q: QueueItem, importBlocked: boolean): boolean {
	const state = q.trackedDownloadState;
	if (state === 'imported' || state === 'ignored') return false;
	if (state === 'importBlocked') return importBlocked;
	if (state && FAILED_STATES[state]) return true;
	return q.status === 'failed' || q.trackedDownloadStatus === 'error';
}

/**
 * A download that has not made meaningful progress for `timeoutMinutes`, or is
 * flagged by the *arr as stalled. Both need to outlast the grace period.
 * A `queued` item with size 0 is treated as an active download: that is how a
 * torrent stuck fetching metadata looks, and it will not start by itself.
 */
export function isStalled(
	q: QueueItem,
	progress: Progress | undefined,
	now: number,
	cfg: Config['stalled'],
	importBlocked: boolean
): boolean {
	if (!progress || isFailed(q, importBlocked)) return false;
	const noMetadata = q.status === 'queued' && q.size === 0;
	if (q.status !== 'downloading' && q.status !== 'warning' && !noMetadata) return false;
	if (q.trackedDownloadState && q.trackedDownloadState !== 'downloading') return false;
	if (now - progress.firstSeen < cfg.graceMinutes * 60_000) return false;
	if (now - progress.since >= cfg.timeoutMinutes * 60_000) return true;
	return q.status === 'warning' && /stalled/i.test(q.message ?? '');
}

/** Queue records that belong to one download (e.g. a season pack) are removed together. */
export function groupByDownload(items: QueueItem[]): QueueItem[][] {
	const groups = new Map<string, QueueItem[]>();
	for (const q of items) {
		const key = q.downloadId ?? `queue:${q.id}`;
		const g = groups.get(key);
		if (g) g.push(q);
		else groups.set(key, [q]);
	}
	return [...groups.values()];
}
