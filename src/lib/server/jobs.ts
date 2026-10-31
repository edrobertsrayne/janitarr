import { APPS, type AppName, type JobName } from '#lib/types.js';
import type { ArrClient, QueueItem, WantedKind } from './arr';
import { groupByDownload, isFailed, isStalled } from './classify';
import type { Config } from './settings-defs';
import type { Store } from './store';

export interface JobContext {
	store: Store;
	cfg: Config;
	/** Returns null when the app has no URL/API key configured. */
	client(app: AppName): ArrClient | null;
	now(): number;
}

const WANTED_PAGE = 100;
const WANTED_MAX_PAGES = 50;
const LIST_PREVIEW = 5;

type Log = (level: 'info' | 'warn' | 'error', message: string) => void;

function logger(ctx: JobContext, job: JobName, app: AppName): Log {
	const dryRun = ctx.cfg.general.dryRun;
	return (level, message) => ctx.store.log({ job, app, level, dryRun, message }, ctx.now());
}

function preview(labels: string[]): string {
	const shown = [...new Set(labels)].slice(0, LIST_PREVIEW).join(', ');
	const more = new Set(labels).size - LIST_PREVIEW;
	return more > 0 ? `${shown} +${more} more` : shown;
}

/** Run `fn` for every configured app; a failure in one app does not stop the other. */
async function forEachApp(
	ctx: JobContext,
	job: JobName,
	fn: (app: AppName, client: ArrClient, log: Log) => Promise<void>
): Promise<void> {
	for (const app of APPS) {
		const client = ctx.client(app);
		if (!client) continue;
		const log = logger(ctx, job, app);
		try {
			await fn(app, client, log);
		} catch (e) {
			log('error', `${job} failed: ${(e as Error).message}`);
		}
	}
}

// --- search ---

export async function runSearch(ctx: JobContext): Promise<void> {
	const { search, general } = ctx.cfg;
	const kinds: WantedKind[] = [];
	if (search.missing) kinds.push('missing');
	if (search.cutoff) kinds.push('cutoff');

	await forEachApp(ctx, 'search', async (app, client, log) => {
		const skip = ctx.store.disabledItems(app);
		const last = ctx.store.lastSearched(app);
		const cooldownMs = search.cooldownHours * 3_600_000;
		const now = ctx.now();

		for (const kind of kinds) {
			const picked: { id: number; label: string }[] = [];
			for (let page = 1; page <= WANTED_MAX_PAGES && picked.length < search.maxPerCycle; page++) {
				const res = await client.wanted(kind, page, WANTED_PAGE);
				for (const item of res.items) {
					if (picked.length >= search.maxPerCycle) break;
					if (!item.searchable || skip.has(item.id)) continue;
					const prev = last.get(item.id);
					if (prev !== undefined && now - prev < cooldownMs) continue;
					picked.push(item);
				}
				if (page * WANTED_PAGE >= res.total || res.items.length === 0) break;
			}
			if (picked.length === 0) continue;

			const what = `${picked.length} ${kind === 'missing' ? 'missing' : 'cutoff-unmet'}`;
			const names = preview(picked.map((p) => p.label));
			if (general.dryRun) {
				log('info', `Would search ${what}: ${names}`);
				continue;
			}
			await client.search(picked.map((p) => p.id));
			ctx.store.markSearched(
				app,
				picked.map((p) => p.id),
				now
			);
			log('info', `Searched ${what}: ${names}`);
		}
	});
}

// --- stalled ---

export async function runStalled(ctx: JobContext): Promise<void> {
	const { stalled, general, failed } = ctx.cfg;

	await forEachApp(ctx, 'stalled', async (app, client, log) => {
		const queue = await client.queue();
		const progress = ctx.store.syncProgress(app, queue, ctx.now(), stalled.minProgressPercent);
		const hits = queue.filter((q) =>
			isStalled(q, progress.get(q.id), ctx.now(), stalled, failed.importBlocked)
		);

		for (const group of groupByDownload(hits)) {
			const names = preview(group.map((q) => q.label));
			if (general.dryRun) {
				log('info', `Would remove stalled download ${group[0].release} (${names})`);
				continue;
			}
			await client.removeQueueItem(group[0].id, {
				removeFromClient: stalled.removeFromClient,
				blocklist: stalled.blocklist
			});
			const ids = researchIds(group, ctx.store.disabledItems(app), stalled.research);
			if (ids.length) await client.search(ids);
			log(
				'info',
				`Removed stalled download ${group[0].release} (${names})${stalled.blocklist ? ', blocklisted' : ''}${ids.length ? ', searching again' : ''}`
			);
		}
	});
}

// --- failed ---

export async function runFailed(ctx: JobContext): Promise<void> {
	const { failed, general } = ctx.cfg;
	const decayMs = failed.strikeDecayDays * 86_400_000;

	await forEachApp(ctx, 'failed', async (app, client, log) => {
		const queue = await client.queue();
		const hits = queue.filter((q) => isFailed(q, failed.importBlocked));

		for (const group of groupByDownload(hits)) {
			const names = preview(group.map((q) => q.label));
			const reason = group[0].message ? ` – ${group[0].message}` : '';
			if (general.dryRun) {
				log('info', `Would clear failed download ${group[0].release} (${names})${reason}`);
				continue;
			}

			// Blocklist the release, free the client, and strike each affected item.
			await client.removeQueueItem(group[0].id, { removeFromClient: true, blocklist: true });
			const exhausted: QueueItem[] = [];
			const retry: number[] = [];
			const seen = new Set<number>();
			for (const q of group) {
				if (q.itemId === null || seen.has(q.itemId)) continue;
				seen.add(q.itemId);
				const count = ctx.store.addStrike(app, q.itemId, q.label, ctx.now(), decayMs);
				if (count >= failed.strikes) exhausted.push(q);
				else retry.push(q.itemId);
			}

			for (const q of exhausted) {
				ctx.store.disableItem(app, q.itemId!);
				if (failed.unmonitorAtThreshold) {
					try {
						await client.unmonitor(q.itemId!);
					} catch (e) {
						log('error', `Could not unmonitor ${q.label}: ${(e as Error).message}`);
					}
				}
				log(
					'warn',
					`${q.label} reached ${failed.strikes} strikes; no longer searching${failed.unmonitorAtThreshold ? ' and unmonitored' : ''}`
				);
			}
			const skipped = new Set(ctx.store.disabledItems(app));
			const ids = retry.filter((id) => !skipped.has(id));
			if (ids.length) await client.search(ids);
			log(
				'info',
				`Cleared failed download ${group[0].release} (${names})${reason}; blocklisted${ids.length ? ', searching again' : ''}`
			);
		}
	});
}

/** Item ids to search again after removal: unique, known, and not given up on. */
function researchIds(group: QueueItem[], disabled: Set<number>, enabled: boolean): number[] {
	if (!enabled) return [];
	const ids = new Set<number>();
	for (const q of group) if (q.itemId !== null && !disabled.has(q.itemId)) ids.add(q.itemId);
	return [...ids];
}

export const JOB_RUNNERS: Record<JobName, (ctx: JobContext) => Promise<void>> = {
	search: runSearch,
	stalled: runStalled,
	failed: runFailed
};
