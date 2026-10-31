import { beforeEach, describe, expect, it } from 'vitest';
import type { AppName } from '#lib/types.js';
import type { ArrClient, QueueItem, WantedItem, WantedKind } from './arr.js';
import { runFailed, runSearch, runStalled, type JobContext } from './jobs.js';
import { Settings } from './settings.js';
import type { Config } from './settings-defs.js';
import { Store } from './store.js';

const MIN = 60_000;
const HOUR = 60 * MIN;

class FakeClient implements ArrClient {
	app: AppName = 'sonarr';
	queueItems: QueueItem[] = [];
	wantedItems: Record<WantedKind, WantedItem[]> = { missing: [], cutoff: [] };
	searches: number[][] = [];
	removed: { id: number; removeFromClient: boolean; blocklist: boolean }[] = [];
	unmonitored: number[] = [];

	status = async () => ({ version: '1' });
	health = async () => [];
	diskspace = async () => [];
	wantedTotal = async (k: WantedKind) => this.wantedItems[k].length;
	async wanted(kind: WantedKind, page: number, pageSize: number) {
		const all = this.wantedItems[kind];
		return { total: all.length, items: all.slice((page - 1) * pageSize, page * pageSize) };
	}
	queue = async () => this.queueItems;
	async search(ids: number[]) {
		this.searches.push(ids);
	}
	async removeQueueItem(id: number, opts: { removeFromClient: boolean; blocklist: boolean }) {
		this.removed.push({ id, ...opts });
		this.queueItems = this.queueItems.filter((q) => q.id !== id);
	}
	async unmonitor(id: number) {
		this.unmonitored.push(id);
	}
}

function queueItem(over: Partial<QueueItem> & { id: number }): QueueItem {
	return {
		release: `release-${over.id}`,
		label: `Show S01E0${over.id}`,
		itemId: over.id * 10,
		downloadId: `dl-${over.id}`,
		status: 'downloading',
		trackedDownloadStatus: 'ok',
		trackedDownloadState: 'downloading',
		message: null,
		size: 1000,
		sizeleft: 500,
		timeleft: null,
		...over
	};
}

let store: Store;
let settings: Settings;
let client: FakeClient;
let clock: number;

function ctx(env: Record<string, string> = {}): JobContext {
	settings = new Settings(store, env);
	const cfg: Config = settings.config();
	return { store, cfg, client: (app) => (app === 'sonarr' ? client : null), now: () => clock };
}

beforeEach(() => {
	store = new Store(':memory:');
	client = new FakeClient();
	clock = 1_000_000_000_000;
});

describe('failed job', () => {
	const failedItem = (id: number, over: Partial<QueueItem> = {}) =>
		queueItem({
			id,
			trackedDownloadStatus: 'error',
			trackedDownloadState: 'failedPending',
			...over
		});

	it('blocklists, clears from the client and searches again on the first failure', async () => {
		client.queueItems = [failedItem(1)];
		await runFailed(ctx());
		expect(client.removed).toEqual([{ id: 1, removeFromClient: true, blocklist: true }]);
		expect(client.searches).toEqual([[10]]);
		expect(store.strikeCounts('sonarr').get(10)).toBe(1);
	});

	it('gives up at the strike limit: no search, unmonitored, and skipped by later searches', async () => {
		const c = ctx({ JANITARR_FAILED_STRIKES: '2' });
		client.queueItems = [failedItem(1)];
		await runFailed(c);
		client.queueItems = [failedItem(2, { itemId: 10 })];
		await runFailed(c);

		expect(client.searches).toEqual([[10]]); // only the first failure re-searches
		expect(client.unmonitored).toEqual([10]);
		expect(store.disabledItems('sonarr').has(10)).toBe(true);
	});

	it('keeps the item monitored when unmonitoring is off', async () => {
		const c = ctx({ JANITARR_FAILED_STRIKES: '1', JANITARR_FAILED_UNMONITOR: 'false' });
		client.queueItems = [failedItem(1)];
		await runFailed(c);
		expect(client.unmonitored).toEqual([]);
		expect(store.disabledItems('sonarr').has(10)).toBe(true);
	});

	it('forgets old strikes after the decay period', async () => {
		const c = ctx({ JANITARR_FAILED_STRIKES: '2', JANITARR_FAILED_STRIKE_DECAY_DAYS: '1' });
		client.queueItems = [failedItem(1)];
		await runFailed(c);
		clock += 25 * HOUR;
		client.queueItems = [failedItem(2, { itemId: 10 })];
		await runFailed(c);

		expect(store.strikeCounts('sonarr').get(10)).toBe(1);
		expect(client.searches).toEqual([[10], [10]]);
		expect(client.unmonitored).toEqual([]);
	});

	it('removes a multi-episode download once but strikes every episode', async () => {
		client.queueItems = [
			failedItem(1, { downloadId: 'pack', itemId: 11 }),
			failedItem(2, { downloadId: 'pack', itemId: 12 })
		];
		await runFailed(ctx());
		expect(client.removed).toHaveLength(1);
		expect(client.searches).toEqual([[11, 12]]);
		expect([...store.strikeCounts('sonarr').keys()].sort()).toEqual([11, 12]);
	});

	it('in a pack only gives up on the episodes that hit the limit', async () => {
		const c = ctx({ JANITARR_FAILED_STRIKES: '2' });
		client.queueItems = [failedItem(1, { itemId: 11 })];
		await runFailed(c);
		client.searches = [];
		client.queueItems = [
			failedItem(2, { downloadId: 'pack', itemId: 11 }),
			failedItem(3, { downloadId: 'pack', itemId: 12 })
		];
		await runFailed(c);
		expect(client.unmonitored).toEqual([11]);
		expect(client.searches).toEqual([[12]]);
	});

	it('treats import-blocked as failed only when enabled', async () => {
		const blocked = queueItem({
			id: 1,
			trackedDownloadStatus: 'warning',
			trackedDownloadState: 'importBlocked'
		});
		client.queueItems = [blocked];
		await runFailed(ctx({ JANITARR_FAILED_IMPORT_BLOCKED: 'false' }));
		expect(client.removed).toEqual([]);
		await runFailed(ctx());
		expect(client.removed).toHaveLength(1);
	});

	it('leaves healthy downloads alone', async () => {
		client.queueItems = [queueItem({ id: 1 })];
		await runFailed(ctx());
		expect(client.removed).toEqual([]);
		expect(store.listStrikes()).toEqual([]);
	});

	it('dry run changes nothing but says what it would do', async () => {
		client.queueItems = [failedItem(1)];
		await runFailed(ctx({ JANITARR_DRY_RUN: 'true' }));
		expect(client.removed).toEqual([]);
		expect(client.searches).toEqual([]);
		expect(store.listStrikes()).toEqual([]);
		const [entry] = store.activity(1);
		expect(entry.dryRun).toBe(true);
		expect(entry.message).toMatch(/^Would clear/);
	});
});

describe('stalled job', () => {
	const env = { JANITARR_STALLED_TIMEOUT_MINUTES: '30', JANITARR_STALLED_GRACE_MINUTES: '5' };

	it('removes a download whose remaining size has not moved for the timeout, then searches again', async () => {
		const c = ctx(env);
		client.queueItems = [queueItem({ id: 1 })];
		await runStalled(c);
		expect(client.removed).toEqual([]); // first sighting only starts the clock

		clock += 29 * MIN;
		await runStalled(c);
		expect(client.removed).toEqual([]);

		clock += 2 * MIN;
		await runStalled(c);
		expect(client.removed).toEqual([{ id: 1, removeFromClient: true, blocklist: true }]);
		expect(client.searches).toEqual([[10]]);
		expect(store.listStrikes()).toEqual([]); // stalled is not a strike
	});

	it('restarts the clock whenever the download makes progress', async () => {
		const c = ctx(env);
		client.queueItems = [queueItem({ id: 1, sizeleft: 500 })];
		await runStalled(c);
		clock += 29 * MIN;
		client.queueItems = [queueItem({ id: 1, sizeleft: 400 })];
		await runStalled(c);
		clock += 29 * MIN;
		await runStalled(c);
		expect(client.removed).toEqual([]);
	});

	it('acts on an *arr stalled warning only once the grace period has passed', async () => {
		const c = ctx(env);
		client.queueItems = [
			queueItem({
				id: 1,
				status: 'warning',
				trackedDownloadStatus: 'warning',
				message: 'The download is stalled with no connections'
			})
		];
		await runStalled(c);
		clock += 3 * MIN;
		await runStalled(c);
		expect(client.removed).toEqual([]);
		clock += 3 * MIN;
		await runStalled(c);
		expect(client.removed).toHaveLength(1);
	});

	it('honours the blocklist / client / search switches', async () => {
		const c = ctx({
			...env,
			JANITARR_STALLED_BLOCKLIST: 'false',
			JANITARR_STALLED_REMOVE_FROM_CLIENT: 'false',
			JANITARR_STALLED_RESEARCH: 'false'
		});
		client.queueItems = [queueItem({ id: 1 })];
		await runStalled(c);
		clock += 31 * MIN;
		await runStalled(c);
		expect(client.removed).toEqual([{ id: 1, removeFromClient: false, blocklist: false }]);
		expect(client.searches).toEqual([]);
	});

	it('does not touch paused or queued downloads', async () => {
		const c = ctx(env);
		client.queueItems = [
			queueItem({ id: 1, status: 'paused' }),
			queueItem({ id: 2, status: 'queued' })
		];
		await runStalled(c);
		clock += 2 * HOUR;
		await runStalled(c);
		expect(client.removed).toEqual([]);
	});

	it('does not search again for items that were given up on', async () => {
		const c = ctx(env);
		store.addStrike('sonarr', 10, 'x', clock, 0);
		store.disableItem('sonarr', 10);
		client.queueItems = [queueItem({ id: 1 })];
		await runStalled(c);
		clock += 31 * MIN;
		await runStalled(c);
		expect(client.removed).toHaveLength(1);
		expect(client.searches).toEqual([]);
	});
});

describe('search job', () => {
	const wanted = (n: number, from = 1): WantedItem[] =>
		Array.from({ length: n }, (_, i) => ({
			id: from + i,
			label: `Item ${from + i}`,
			searchable: true
		}));

	it('respects the per-cycle cap and the cooldown across cycles', async () => {
		const c = ctx({ JANITARR_SEARCH_MAX_PER_CYCLE: '3', JANITARR_SEARCH_COOLDOWN_HOURS: '24' });
		client.wantedItems.missing = wanted(5);
		await runSearch(c);
		await runSearch(c);
		expect(client.searches).toEqual([
			[1, 2, 3],
			[4, 5]
		]);

		await runSearch(c);
		expect(client.searches).toHaveLength(2); // everything is cooling down

		clock += 25 * HOUR;
		await runSearch(c);
		expect(client.searches[2]).toEqual([1, 2, 3]);
	});

	it('applies the cap separately to missing and cutoff and honours the toggles', async () => {
		client.wantedItems.missing = wanted(2);
		client.wantedItems.cutoff = wanted(2, 100);
		await runSearch(ctx());
		expect(client.searches).toEqual([
			[1, 2],
			[100, 101]
		]);

		client.searches = [];
		clock += 100 * HOUR;
		await runSearch(ctx({ JANITARR_SEARCH_CUTOFF: 'false' }));
		expect(client.searches).toEqual([[1, 2]]);
	});

	it('skips unsearchable items and items given up on', async () => {
		store.addStrike('sonarr', 2, 'x', clock, 0);
		store.disableItem('sonarr', 2);
		client.wantedItems.missing = [
			{ id: 1, label: 'a', searchable: false },
			{ id: 2, label: 'b', searchable: true },
			{ id: 3, label: 'c', searchable: true }
		];
		await runSearch(ctx());
		expect(client.searches).toEqual([[3]]);
	});

	it('pages past ineligible items to fill the cap', async () => {
		const c = ctx({ JANITARR_SEARCH_MAX_PER_CYCLE: '2' });
		store.markSearched(
			'sonarr',
			wanted(150).map((w) => w.id),
			clock
		);
		client.wantedItems.missing = wanted(160);
		await runSearch(c);
		expect(client.searches).toEqual([[151, 152]]);
	});

	it('dry run searches nothing and leaves no cooldown behind', async () => {
		client.wantedItems.missing = wanted(2);
		await runSearch(ctx({ JANITARR_DRY_RUN: 'true' }));
		expect(client.searches).toEqual([]);
		expect(store.lastSearched('sonarr').size).toBe(0);
	});
});
