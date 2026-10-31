import type { AppName } from '#lib/types.js';
import type { Config } from './settings-defs';

export interface QueueItem {
	/** Queue record id. */
	id: number;
	/** Release name. */
	release: string;
	/** Human title: "Series S01E02" / "Movie (2020)". Falls back to the release name. */
	label: string;
	/** Episode id (Sonarr) or movie id (Radarr); absent for unmatched downloads. */
	itemId: number | null;
	downloadId: string | null;
	status: string;
	trackedDownloadStatus: string | null;
	trackedDownloadState: string | null;
	/** errorMessage plus all statusMessages, joined. */
	message: string | null;
	size: number;
	sizeleft: number;
	timeleft: string | null;
}

export interface WantedItem {
	id: number;
	label: string;
	/** False for e.g. unreleased movies that cannot be searched yet. */
	searchable: boolean;
}

export type WantedKind = 'missing' | 'cutoff';

export interface ArrClient {
	readonly app: AppName;
	status(): Promise<{ version: string }>;
	health(): Promise<{ type: string; message: string }[]>;
	diskspace(): Promise<{ path: string; free: number; total: number }[]>;
	wantedTotal(kind: WantedKind): Promise<number>;
	wanted(
		kind: WantedKind,
		page: number,
		pageSize: number
	): Promise<{ items: WantedItem[]; total: number }>;
	queue(): Promise<QueueItem[]>;
	search(itemIds: number[]): Promise<void>;
	removeQueueItem(
		id: number,
		opts: { removeFromClient: boolean; blocklist: boolean }
	): Promise<void>;
	unmonitor(itemId: number): Promise<void>;
}

const TIMEOUT_MS = 30_000;
const QUEUE_PAGE = 200;
const QUEUE_MAX_PAGES = 25;

type Json = Record<string, unknown>;

// *arr responses are untyped JSON; these accessors narrow without trusting the shape.
function obj(v: unknown): Json {
	return v && typeof v === 'object' && !Array.isArray(v) ? (v as Json) : {};
}
function list(v: unknown): Json[] {
	return Array.isArray(v) ? v.map(obj) : [];
}
function str(v: unknown): string | null {
	return typeof v === 'string' && v !== '' ? v : null;
}
function num(v: unknown): number {
	return typeof v === 'number' && Number.isFinite(v) ? v : 0;
}

export class HttpArrClient implements ArrClient {
	private base: string;

	constructor(
		readonly app: AppName,
		url: string,
		private apiKey: string,
		private fetchFn: typeof fetch = fetch
	) {
		this.base = url.replace(/\/+$/, '');
	}

	private async req(
		method: string,
		path: string,
		body?: unknown,
		allow404 = false
	): Promise<unknown> {
		const res = await this.fetchFn(`${this.base}/api/v3${path}`, {
			method,
			headers: {
				'X-Api-Key': this.apiKey,
				...(body !== undefined ? { 'Content-Type': 'application/json' } : {})
			},
			body: body !== undefined ? JSON.stringify(body) : undefined,
			signal: AbortSignal.timeout(TIMEOUT_MS)
		});
		if (allow404 && res.status === 404) return null;
		if (!res.ok) {
			const text = await res.text().catch(() => '');
			throw new Error(
				`${this.app} ${method} ${path} → HTTP ${res.status}${text ? `: ${text.slice(0, 200)}` : ''}`
			);
		}
		const text = await res.text();
		return text ? JSON.parse(text) : null;
	}

	async status() {
		return { version: str(obj(await this.req('GET', '/system/status')).version) ?? 'unknown' };
	}

	async health() {
		return list(await this.req('GET', '/health')).map((r) => ({
			type: str(r.type) ?? 'warning',
			message: str(r.message) ?? ''
		}));
	}

	async diskspace() {
		return list(await this.req('GET', '/diskspace')).map((r) => ({
			path: str(r.path) ?? str(r.label) ?? '',
			free: num(r.freeSpace),
			total: num(r.totalSpace)
		}));
	}

	private wantedPath(kind: WantedKind, page: number, pageSize: number): string {
		const extra = this.app === 'sonarr' ? '&includeSeries=true' : '';
		return `/wanted/${kind}?page=${page}&pageSize=${pageSize}&monitored=true${extra}`;
	}

	async wantedTotal(kind: WantedKind) {
		return num(obj(await this.req('GET', this.wantedPath(kind, 1, 1))).totalRecords);
	}

	async wanted(kind: WantedKind, page: number, pageSize: number) {
		const r = obj(await this.req('GET', this.wantedPath(kind, page, pageSize)));
		return {
			total: num(r.totalRecords),
			items: list(r.records).map((rec): WantedItem =>
				this.app === 'sonarr'
					? {
							id: num(rec.id),
							label: episodeLabel(
								obj(rec.series).title,
								rec.seasonNumber,
								rec.episodeNumber,
								rec.title
							),
							searchable: true
						}
					: {
							id: num(rec.id),
							label: movieLabel(rec.title, rec.year),
							searchable: rec.isAvailable !== false
						}
			)
		};
	}

	async queue(): Promise<QueueItem[]> {
		const include =
			this.app === 'sonarr' ? 'includeSeries=true&includeEpisode=true' : 'includeMovie=true';
		const out: QueueItem[] = [];
		for (let page = 1; page <= QUEUE_MAX_PAGES; page++) {
			const r = obj(await this.req('GET', `/queue?page=${page}&pageSize=${QUEUE_PAGE}&${include}`));
			const records = list(r.records);
			for (const q of records) out.push(this.mapQueue(q));
			if (page * QUEUE_PAGE >= num(r.totalRecords) || records.length === 0) break;
		}
		return out;
	}

	private mapQueue(q: Json): QueueItem {
		const release = str(q.title) ?? '';
		const messages = [
			str(q.errorMessage),
			...list(q.statusMessages).flatMap((m) => [
				str(m.title),
				...(Array.isArray(m.messages) ? m.messages.map(str) : [])
			])
		].filter((m): m is string => m !== null);
		const sonarr = this.app === 'sonarr';
		const series = obj(q.series);
		const episode = obj(q.episode);
		const movie = obj(q.movie);
		const label = sonarr
			? q.series && q.episode
				? episodeLabel(series.title, episode.seasonNumber, episode.episodeNumber)
				: (str(series.title) ?? release)
			: q.movie
				? movieLabel(movie.title, movie.year)
				: release;
		const itemId = sonarr ? q.episodeId : q.movieId;
		return {
			id: num(q.id),
			release,
			label,
			itemId: typeof itemId === 'number' && itemId > 0 ? itemId : null,
			downloadId: str(q.downloadId),
			status: str(q.status) ?? '',
			trackedDownloadStatus: str(q.trackedDownloadStatus),
			trackedDownloadState: str(q.trackedDownloadState),
			message: messages.length ? messages.join(' · ') : null,
			size: num(q.size),
			sizeleft: num(q.sizeleft),
			timeleft: str(q.timeleft)
		};
	}

	async search(itemIds: number[]) {
		if (itemIds.length === 0) return;
		await this.req(
			'POST',
			'/command',
			this.app === 'sonarr'
				? { name: 'EpisodeSearch', episodeIds: itemIds }
				: { name: 'MoviesSearch', movieIds: itemIds }
		);
	}

	async removeQueueItem(id: number, opts: { removeFromClient: boolean; blocklist: boolean }) {
		// skipRedownload: Janitarr triggers the re-search itself so it can honour strikes.
		await this.req(
			'DELETE',
			`/queue/${id}?removeFromClient=${opts.removeFromClient}&blocklist=${opts.blocklist}&skipRedownload=true`,
			undefined,
			true // already gone, e.g. removed together with another record of the same download
		);
	}

	async unmonitor(itemId: number) {
		if (this.app === 'sonarr') {
			await this.req('PUT', '/episode/monitor', { episodeIds: [itemId], monitored: false });
		} else {
			await this.req('PUT', '/movie/editor', { movieIds: [itemId], monitored: false });
		}
	}
}

function pad(n: unknown): string {
	return String(n ?? 0).padStart(2, '0');
}

function episodeLabel(series: unknown, season: unknown, episode: unknown, title?: unknown): string {
	const base = `${series ?? 'Unknown series'} S${pad(season)}E${pad(episode)}`;
	return typeof title === 'string' && title ? `${base} – ${title}` : base;
}

function movieLabel(title: unknown, year: unknown): string {
	return year ? `${title} (${year})` : String(title ?? 'Unknown movie');
}

/** Build a client for `app` from config, or null when URL/API key are not both set. */
export function createClient(app: AppName, cfg: Config): ArrClient | null {
	const { url, apiKey } = cfg[app];
	return url && apiKey ? new HttpArrClient(app, url, apiKey) : null;
}
