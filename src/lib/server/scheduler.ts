import { JOBS, type JobName, type JobState } from '#lib/types.js';
import { createClient } from './arr';
import { JOB_RUNNERS } from './jobs';
import type { Settings } from './settings';
import type { Config } from './settings-defs';
import type { Store } from './store';

const TICK_MS = 5_000;
const STARTUP_DELAY_MS = 10_000;

interface RunInfo {
	running: boolean;
	lastStart: number | null;
	lastEnd: number | null;
	lastError: string | null;
}

export class Scheduler {
	private runs: Record<JobName, RunInfo> = {
		search: { running: false, lastStart: null, lastEnd: null, lastError: null },
		stalled: { running: false, lastStart: null, lastEnd: null, lastError: null },
		failed: { running: false, lastStart: null, lastEnd: null, lastError: null }
	};
	private timer: ReturnType<typeof setInterval> | null = null;
	private bootTime = Date.now();

	constructor(
		private store: Store,
		private settings: Settings
	) {}

	start(): void {
		if (this.timer) return;
		this.bootTime = Date.now();
		this.timer = setInterval(() => this.tick(), TICK_MS);
		this.timer.unref();
	}

	stop(): void {
		clearInterval(this.timer ?? undefined);
		this.timer = null;
	}

	private nextRun(job: JobName, cfg: Config): number {
		const interval = cfg[job].intervalMinutes * 60_000;
		return (this.runs[job].lastEnd ?? this.bootTime + STARTUP_DELAY_MS - interval) + interval;
	}

	private tick(): void {
		const cfg = this.settings.config();
		const now = Date.now();
		for (const job of JOBS) {
			if (cfg[job].enabled && !this.runs[job].running && now >= this.nextRun(job, cfg)) {
				void this.run(job);
			}
		}
	}

	/** Run a job now. Rejects if it is already running. Job errors are logged, not thrown. */
	async run(job: JobName): Promise<void> {
		const info = this.runs[job];
		if (info.running) throw new Error(`${job} is already running`);
		info.running = true;
		info.lastStart = Date.now();
		info.lastError = null;
		try {
			const cfg = this.settings.config();
			await JOB_RUNNERS[job]({
				store: this.store,
				cfg,
				client: (app) => createClient(app, cfg),
				now: Date.now
			});
		} catch (e) {
			info.lastError = (e as Error).message;
			this.store.log(
				{ job, app: null, level: 'error', message: `${job} crashed: ${info.lastError}` },
				Date.now()
			);
		} finally {
			info.running = false;
			info.lastEnd = Date.now();
		}
	}

	states(): JobState[] {
		const cfg = this.settings.config();
		return JOBS.map((name) => ({
			name,
			enabled: cfg[name].enabled,
			running: this.runs[name].running,
			intervalMinutes: cfg[name].intervalMinutes,
			lastStart: this.runs[name].lastStart,
			lastEnd: this.runs[name].lastEnd,
			lastError: this.runs[name].lastError,
			nextRun: this.nextRun(name, cfg)
		}));
	}
}
