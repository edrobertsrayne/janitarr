export type AppName = 'sonarr' | 'radarr';
export const APPS: readonly AppName[] = ['sonarr', 'radarr'];

export type JobName = 'search' | 'stalled' | 'failed';
export const JOBS: readonly JobName[] = ['search', 'stalled', 'failed'];

export type SettingType = 'string' | 'number' | 'boolean';
export type SettingValue = string | number | boolean;
export type SettingSource = 'env' | 'ui' | 'default';

/** A setting as presented to the UI. Secret values are masked. */
export interface SettingView {
	key: string;
	env: string;
	group: string;
	label: string;
	description: string;
	type: SettingType;
	secret: boolean;
	value: SettingValue;
	default: SettingValue;
	source: SettingSource;
	/** True when the value comes from an environment variable and cannot be edited. */
	locked: boolean;
}

export interface ActivityEntry {
	id: number;
	ts: number;
	job: JobName | 'system';
	app: AppName | null;
	level: 'info' | 'warn' | 'error';
	dryRun: boolean;
	message: string;
}

export interface JobState {
	name: JobName;
	enabled: boolean;
	running: boolean;
	intervalMinutes: number;
	lastStart: number | null;
	lastEnd: number | null;
	lastError: string | null;
	nextRun: number;
}

export interface QueueRow {
	id: number;
	title: string;
	release: string;
	status: string;
	state: string | null;
	progress: number;
	size: number;
	timeleft: string | null;
	flag: 'stalled' | 'failed' | null;
	message: string | null;
	strikes: number;
}

export interface StrikeRow {
	app: AppName;
	itemId: number;
	title: string;
	count: number;
	disabled: boolean;
	lastTs: number;
}

export interface InstanceStatus {
	app: AppName;
	configured: boolean;
	online: boolean;
	error: string | null;
	version: string | null;
	health: { type: string; message: string }[];
	disks: { path: string; free: number; total: number }[];
	counts: {
		missing: number;
		cutoff: number;
		queue: number;
		stalled: number;
		failed: number;
	} | null;
	queue: QueueRow[];
}

export interface StatusPayload {
	now: number;
	dryRun: boolean;
	/** Failures before an item is given up on (failed.strikes). */
	strikeLimit: number;
	dryRunLocked: boolean;
	jobs: JobState[];
	instances: InstanceStatus[];
	strikes: StrikeRow[];
	activity: ActivityEntry[];
}
