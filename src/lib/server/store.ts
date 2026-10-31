import { mkdirSync } from 'node:fs';
import { dirname } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import type { ActivityEntry, AppName, JobName, StrikeRow } from '#lib/types.js';

const ACTIVITY_KEEP = 2000;
/** Upper bound for the minimum-progress threshold, so huge releases are not held to an unrealistic speed. */
const MAX_MIN_PROGRESS_BYTES = 50 * 1024 * 1024;

export interface Progress {
	firstSeen: number;
	/** Timestamp at which `sizeleft` last changed. */
	since: number;
}

export class Store {
	private db: DatabaseSync;

	constructor(path: string) {
		if (path !== ':memory:') mkdirSync(dirname(path), { recursive: true });
		this.db = new DatabaseSync(path);
		this.db.exec(`
			PRAGMA journal_mode = WAL;
			CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
			CREATE TABLE IF NOT EXISTS activity (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				ts INTEGER NOT NULL,
				job TEXT NOT NULL,
				app TEXT,
				level TEXT NOT NULL,
				dry_run INTEGER NOT NULL,
				message TEXT NOT NULL
			);
			CREATE TABLE IF NOT EXISTS strikes (
				app TEXT NOT NULL,
				item_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				count INTEGER NOT NULL,
				last_ts INTEGER NOT NULL,
				disabled INTEGER NOT NULL DEFAULT 0,
				PRIMARY KEY (app, item_id)
			);
			CREATE TABLE IF NOT EXISTS searched (
				app TEXT NOT NULL,
				item_id INTEGER NOT NULL,
				ts INTEGER NOT NULL,
				PRIMARY KEY (app, item_id)
			);
			CREATE TABLE IF NOT EXISTS queue_progress (
				app TEXT NOT NULL,
				queue_id INTEGER NOT NULL,
				sizeleft REAL NOT NULL,
				first_seen INTEGER NOT NULL,
				since INTEGER NOT NULL,
				PRIMARY KEY (app, queue_id)
			);
		`);
	}

	close(): void {
		this.db.close();
	}

	private tx<T>(fn: () => T): T {
		this.db.exec('BEGIN');
		try {
			const r = fn();
			this.db.exec('COMMIT');
			return r;
		} catch (e) {
			this.db.exec('ROLLBACK');
			throw e;
		}
	}

	// --- settings overrides (JSON-encoded values) ---

	overrides(): Map<string, string> {
		const rows = this.db.prepare('SELECT key, value FROM settings').all() as {
			key: string;
			value: string;
		}[];
		return new Map(rows.map((r) => [r.key, r.value]));
	}

	setOverride(key: string, json: string): void {
		this.db
			.prepare(
				'INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value'
			)
			.run(key, json);
	}

	deleteOverride(key: string): void {
		this.db.prepare('DELETE FROM settings WHERE key = ?').run(key);
	}

	// --- activity log ---

	log(
		e: {
			job: JobName | 'system';
			app: AppName | null;
			level: 'info' | 'warn' | 'error';
			dryRun?: boolean;
			message: string;
		},
		now: number
	): void {
		this.db
			.prepare(
				'INSERT INTO activity (ts, job, app, level, dry_run, message) VALUES (?, ?, ?, ?, ?, ?)'
			)
			.run(now, e.job, e.app, e.level, e.dryRun ? 1 : 0, e.message);
		this.db
			.prepare('DELETE FROM activity WHERE id <= (SELECT MAX(id) FROM activity) - ?')
			.run(ACTIVITY_KEEP);
	}

	activity(limit: number): ActivityEntry[] {
		const rows = this.db
			.prepare(
				'SELECT id, ts, job, app, level, dry_run, message FROM activity ORDER BY id DESC LIMIT ?'
			)
			.all(limit) as {
			id: number;
			ts: number;
			job: ActivityEntry['job'];
			app: AppName | null;
			level: ActivityEntry['level'];
			dry_run: number;
			message: string;
		}[];
		return rows.map((r) => ({
			id: r.id,
			ts: r.ts,
			job: r.job,
			app: r.app,
			level: r.level,
			dryRun: r.dry_run === 1,
			message: r.message
		}));
	}

	// --- strikes ---

	/**
	 * Record a failure against an item and return the new strike count. A count
	 * whose last failure is older than `decayMs` (when > 0) starts over.
	 */
	addStrike(app: AppName, itemId: number, title: string, now: number, decayMs: number): number {
		return this.tx(() => {
			const row = this.db
				.prepare('SELECT count, last_ts FROM strikes WHERE app = ? AND item_id = ?')
				.get(app, itemId) as { count: number; last_ts: number } | undefined;
			const base = row && !(decayMs > 0 && now - row.last_ts > decayMs) ? row.count : 0;
			const count = base + 1;
			this.db
				.prepare(
					`INSERT INTO strikes (app, item_id, title, count, last_ts) VALUES (?, ?, ?, ?, ?)
					 ON CONFLICT(app, item_id) DO UPDATE SET title = excluded.title, count = excluded.count, last_ts = excluded.last_ts`
				)
				.run(app, itemId, title, count, now);
			return count;
		});
	}

	disableItem(app: AppName, itemId: number): void {
		this.db
			.prepare('UPDATE strikes SET disabled = 1 WHERE app = ? AND item_id = ?')
			.run(app, itemId);
	}

	clearStrikes(app: AppName, itemId: number): void {
		this.db.prepare('DELETE FROM strikes WHERE app = ? AND item_id = ?').run(app, itemId);
	}

	strikeCounts(app: AppName): Map<number, number> {
		const rows = this.db.prepare('SELECT item_id, count FROM strikes WHERE app = ?').all(app) as {
			item_id: number;
			count: number;
		}[];
		return new Map(rows.map((r) => [r.item_id, r.count]));
	}

	disabledItems(app: AppName): Set<number> {
		const rows = this.db
			.prepare('SELECT item_id FROM strikes WHERE app = ? AND disabled = 1')
			.all(app) as { item_id: number }[];
		return new Set(rows.map((r) => r.item_id));
	}

	listStrikes(): StrikeRow[] {
		const rows = this.db
			.prepare(
				'SELECT app, item_id, title, count, last_ts, disabled FROM strikes ORDER BY disabled DESC, last_ts DESC'
			)
			.all() as {
			app: AppName;
			item_id: number;
			title: string;
			count: number;
			last_ts: number;
			disabled: number;
		}[];
		return rows.map((r) => ({
			app: r.app,
			itemId: r.item_id,
			title: r.title,
			count: r.count,
			disabled: r.disabled === 1,
			lastTs: r.last_ts
		}));
	}

	// --- search cooldown ---

	lastSearched(app: AppName): Map<number, number> {
		const rows = this.db.prepare('SELECT item_id, ts FROM searched WHERE app = ?').all(app) as {
			item_id: number;
			ts: number;
		}[];
		return new Map(rows.map((r) => [r.item_id, r.ts]));
	}

	markSearched(app: AppName, itemIds: number[], now: number): void {
		this.tx(() => {
			const stmt = this.db.prepare(
				`INSERT INTO searched (app, item_id, ts) VALUES (?, ?, ?)
				 ON CONFLICT(app, item_id) DO UPDATE SET ts = excluded.ts`
			);
			for (const id of itemIds) stmt.run(app, id, now);
		});
	}

	// --- queue progress tracking ---

	/**
	 * Record the current queue. `since` is reset when `sizeleft` has dropped by
	 * at least `minProgressPercent` of `size` (capped at 50 MiB) since the last
	 * reset, or when it grew (e.g. the size became known); smaller drops
	 * accumulate against the same baseline.
	 * Rows for items no longer queued are dropped.
	 */
	syncProgress(
		app: AppName,
		items: { id: number; size: number; sizeleft: number }[],
		now: number,
		minProgressPercent: number
	): Map<number, Progress> {
		return this.tx(() => {
			const existing = new Map(
				(
					this.db
						.prepare(
							'SELECT queue_id, sizeleft, first_seen, since FROM queue_progress WHERE app = ?'
						)
						.all(app) as { queue_id: number; sizeleft: number; first_seen: number; since: number }[]
				).map((r) => [r.queue_id, r])
			);
			const upsert = this.db.prepare(
				`INSERT INTO queue_progress (app, queue_id, sizeleft, first_seen, since) VALUES (?, ?, ?, ?, ?)
				 ON CONFLICT(app, queue_id) DO UPDATE SET sizeleft = excluded.sizeleft, since = excluded.since`
			);
			const result = new Map<number, Progress>();
			const live = new Set<number>();
			for (const it of items) {
				live.add(it.id);
				const prev = existing.get(it.id);
				const firstSeen = prev?.first_seen ?? now;
				const threshold = Math.min((it.size * minProgressPercent) / 100, MAX_MIN_PROGRESS_BYTES);
				const moved =
					prev !== undefined &&
					it.sizeleft !== prev.sizeleft &&
					(it.sizeleft > prev.sizeleft || prev.sizeleft - it.sizeleft >= threshold);
				const since = prev && !moved ? prev.since : now;
				if (!prev || moved) upsert.run(app, it.id, it.sizeleft, firstSeen, since);
				result.set(it.id, { firstSeen, since });
			}
			const del = this.db.prepare('DELETE FROM queue_progress WHERE app = ? AND queue_id = ?');
			for (const id of existing.keys()) if (!live.has(id)) del.run(app, id);
			return result;
		});
	}

	progress(app: AppName): Map<number, Progress> {
		const rows = this.db
			.prepare('SELECT queue_id, first_seen, since FROM queue_progress WHERE app = ?')
			.all(app) as { queue_id: number; first_seen: number; since: number }[];
		return new Map(rows.map((r) => [r.queue_id, { firstSeen: r.first_seen, since: r.since }]));
	}
}
