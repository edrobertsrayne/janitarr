import {
	APPS,
	type AppName,
	type InstanceStatus,
	type QueueRow,
	type StatusPayload
} from '#lib/types.js';
import { createClient, type ArrClient, type QueueItem } from './arr';
import { isFailed, isStalled } from './classify';
import type { Scheduler } from './scheduler';
import type { Settings } from './settings';
import type { Config } from './settings-defs';
import type { Store } from './store';

const ACTIVITY_LIMIT = 100;

function messageOf(e: unknown): string {
	return e instanceof Error ? e.message : String(e);
}

async function instanceStatus(
	app: AppName,
	client: ArrClient | null,
	cfg: Config,
	store: Store,
	now: number
): Promise<InstanceStatus> {
	const base: InstanceStatus = {
		app,
		configured: client !== null,
		online: false,
		error: null,
		version: null,
		health: [],
		disks: [],
		counts: null,
		queue: []
	};
	if (!client) return base;

	// Version doubles as the reachability probe; everything else degrades independently.
	let version: string;
	try {
		version = (await client.status()).version;
	} catch (e) {
		return { ...base, error: messageOf(e) };
	}
	const [health, disks, missing, cutoff, queue] = await Promise.allSettled([
		client.health(),
		client.diskspace(),
		client.wantedTotal('missing'),
		client.wantedTotal('cutoff'),
		client.queue()
	]);

	const strikes = store.strikeCounts(app);
	const progress = store.progress(app);
	const rows: QueueRow[] = queue.status === 'fulfilled' ? queue.value.map((q) => toRow(q)) : [];

	function toRow(q: QueueItem): QueueRow {
		const failed = isFailed(q, cfg.failed.importBlocked);
		const stalled = isStalled(q, progress.get(q.id), now, cfg.stalled, cfg.failed.importBlocked);
		return {
			id: q.id,
			title: q.label,
			release: q.release,
			status: q.status,
			state: q.trackedDownloadState,
			progress: q.size > 0 ? Math.min(1, Math.max(0, 1 - q.sizeleft / q.size)) : 0,
			size: q.size,
			timeleft: q.timeleft,
			idleMs: progress.has(q.id) ? now - progress.get(q.id)!.since : null,
			flag: failed ? 'failed' : stalled ? 'stalled' : null,
			message: q.message,
			strikes: q.itemId === null ? 0 : (strikes.get(q.itemId) ?? 0)
		};
	}

	const countsOk =
		missing.status === 'fulfilled' && cutoff.status === 'fulfilled' && queue.status === 'fulfilled';
	const errors = [health, disks, missing, cutoff, queue]
		.filter((r): r is PromiseRejectedResult => r.status === 'rejected')
		.map((r) => messageOf(r.reason));

	return {
		...base,
		online: true,
		version,
		error: errors[0] ?? null,
		health: health.status === 'fulfilled' ? health.value : [],
		disks: disks.status === 'fulfilled' ? disks.value : [],
		counts: countsOk
			? {
					missing: missing.value,
					cutoff: cutoff.value,
					queue: rows.length,
					stalled: rows.filter((r) => r.flag === 'stalled').length,
					failed: rows.filter((r) => r.flag === 'failed').length
				}
			: null,
		queue: rows
	};
}

export async function buildStatus(
	store: Store,
	settings: Settings,
	scheduler: Scheduler
): Promise<StatusPayload> {
	const cfg = settings.config();
	const now = Date.now();
	const instances = await Promise.all(
		APPS.map((app) => instanceStatus(app, createClient(app, cfg), cfg, store, now))
	);
	return {
		now,
		dryRun: cfg.general.dryRun,
		dryRunLocked: settings.isLocked('general.dryRun'),
		strikeLimit: cfg.failed.strikes,
		jobs: scheduler.states(),
		instances,
		strikes: store.listStrikes(),
		activity: store.activity(ACTIVITY_LIMIT)
	};
}
