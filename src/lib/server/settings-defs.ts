import type { SettingType, SettingValue } from '#lib/types.js';

export interface SettingDef {
	key: string;
	env: string;
	group: string;
	label: string;
	description: string;
	type: SettingType;
	default: SettingValue;
	secret?: boolean;
	/** Numbers only: smallest accepted value. */
	min?: number;
}

export interface Config {
	general: { dryRun: boolean };
	sonarr: { url: string; apiKey: string };
	radarr: { url: string; apiKey: string };
	search: {
		enabled: boolean;
		intervalMinutes: number;
		missing: boolean;
		cutoff: boolean;
		maxPerCycle: number;
		cooldownHours: number;
	};
	stalled: {
		enabled: boolean;
		intervalMinutes: number;
		timeoutMinutes: number;
		minProgressPercent: number;
		graceMinutes: number;
		removeFromClient: boolean;
		blocklist: boolean;
		research: boolean;
	};
	failed: {
		enabled: boolean;
		intervalMinutes: number;
		importBlocked: boolean;
		strikes: number;
		strikeDecayDays: number;
		unmonitorAtThreshold: boolean;
	};
}

export const DEFS: readonly SettingDef[] = [
	{
		key: 'general.dryRun',
		env: 'JANITARR_DRY_RUN',
		group: 'General',
		label: 'Dry run',
		description:
			'Log what would be done without searching, removing, blocklisting or unmonitoring anything.',
		type: 'boolean',
		default: false
	},

	{
		key: 'sonarr.url',
		env: 'SONARR_URL',
		group: 'Sonarr',
		label: 'URL',
		description: 'Base URL, e.g. http://sonarr:8989',
		type: 'string',
		default: ''
	},
	{
		key: 'sonarr.apiKey',
		env: 'SONARR_API_KEY',
		group: 'Sonarr',
		label: 'API key',
		description: 'Settings → General → API key in Sonarr.',
		type: 'string',
		default: '',
		secret: true
	},
	{
		key: 'radarr.url',
		env: 'RADARR_URL',
		group: 'Radarr',
		label: 'URL',
		description: 'Base URL, e.g. http://radarr:7878',
		type: 'string',
		default: ''
	},
	{
		key: 'radarr.apiKey',
		env: 'RADARR_API_KEY',
		group: 'Radarr',
		label: 'API key',
		description: 'Settings → General → API key in Radarr.',
		type: 'string',
		default: '',
		secret: true
	},

	{
		key: 'search.enabled',
		env: 'JANITARR_SEARCH_ENABLED',
		group: 'Search',
		label: 'Enabled',
		description: 'Periodically trigger searches for missing and cutoff-unmet items.',
		type: 'boolean',
		default: true
	},
	{
		key: 'search.intervalMinutes',
		env: 'JANITARR_SEARCH_INTERVAL_MINUTES',
		group: 'Search',
		label: 'Interval (minutes)',
		description: 'Time between search cycles.',
		type: 'number',
		default: 60,
		min: 1
	},
	{
		key: 'search.missing',
		env: 'JANITARR_SEARCH_MISSING',
		group: 'Search',
		label: 'Search missing',
		description: 'Search monitored episodes/movies that have no file.',
		type: 'boolean',
		default: true
	},
	{
		key: 'search.cutoff',
		env: 'JANITARR_SEARCH_CUTOFF',
		group: 'Search',
		label: 'Search cutoff unmet',
		description: 'Search items whose file has not reached the quality cutoff.',
		type: 'boolean',
		default: true
	},
	{
		key: 'search.maxPerCycle',
		env: 'JANITARR_SEARCH_MAX_PER_CYCLE',
		group: 'Search',
		label: 'Max items per cycle',
		description: 'Cap per instance and per list (missing / cutoff) in each cycle.',
		type: 'number',
		default: 10,
		min: 1
	},
	{
		key: 'search.cooldownHours',
		env: 'JANITARR_SEARCH_COOLDOWN_HOURS',
		group: 'Search',
		label: 'Per-item cooldown (hours)',
		description: 'Do not search the same item again within this period.',
		type: 'number',
		default: 72,
		min: 0
	},

	{
		key: 'stalled.enabled',
		env: 'JANITARR_STALLED_ENABLED',
		group: 'Stalled downloads',
		label: 'Enabled',
		description: 'Remove downloads that have stopped making progress.',
		type: 'boolean',
		default: true
	},
	{
		key: 'stalled.intervalMinutes',
		env: 'JANITARR_STALLED_INTERVAL_MINUTES',
		group: 'Stalled downloads',
		label: 'Interval (minutes)',
		description: 'Time between queue checks.',
		type: 'number',
		default: 5,
		min: 1
	},
	{
		key: 'stalled.timeoutMinutes',
		env: 'JANITARR_STALLED_TIMEOUT_MINUTES',
		group: 'Stalled downloads',
		label: 'No-progress timeout (minutes)',
		description:
			'A download that has not made meaningful progress (see Minimum progress) for this long is stalled. Also applies to queued items whose size is still unknown, e.g. a torrent that never got its metadata.',
		type: 'number',
		default: 30,
		min: 1
	},
	{
		key: 'stalled.minProgressPercent',
		env: 'JANITARR_STALLED_MIN_PROGRESS_PERCENT',
		group: 'Stalled downloads',
		label: 'Minimum progress (% of size)',
		description:
			'Progress only restarts the timeout once the download has moved on by at least this share of its size (at most 50 MB) since the last restart. Stops a trickle of data from keeping a dead download alive. 0 = any progress counts.',
		type: 'number',
		default: 1,
		min: 0
	},
	{
		key: 'stalled.graceMinutes',
		env: 'JANITARR_STALLED_GRACE_MINUTES',
		group: 'Stalled downloads',
		label: 'Grace period (minutes)',
		description:
			'Downloads younger than this are never treated as stalled, including when the *arr itself reports a stalled warning.',
		type: 'number',
		default: 10,
		min: 0
	},
	{
		key: 'stalled.removeFromClient',
		env: 'JANITARR_STALLED_REMOVE_FROM_CLIENT',
		group: 'Stalled downloads',
		label: 'Remove from download client',
		description: 'Also delete the download (and its data) from the download client.',
		type: 'boolean',
		default: true
	},
	{
		key: 'stalled.blocklist',
		env: 'JANITARR_STALLED_BLOCKLIST',
		group: 'Stalled downloads',
		label: 'Blocklist release',
		description: 'Add the stalled release to the blocklist so it is not grabbed again.',
		type: 'boolean',
		default: true
	},
	{
		key: 'stalled.research',
		env: 'JANITARR_STALLED_RESEARCH',
		group: 'Stalled downloads',
		label: 'Search again',
		description: 'Trigger a new search for the affected item after removal.',
		type: 'boolean',
		default: true
	},

	{
		key: 'failed.enabled',
		env: 'JANITARR_FAILED_ENABLED',
		group: 'Failed downloads and imports',
		label: 'Enabled',
		description:
			'Clear failed downloads/imports, blocklist the release, search again, and count a strike against the item.',
		type: 'boolean',
		default: true
	},
	{
		key: 'failed.intervalMinutes',
		env: 'JANITARR_FAILED_INTERVAL_MINUTES',
		group: 'Failed downloads and imports',
		label: 'Interval (minutes)',
		description: 'Time between queue checks.',
		type: 'number',
		default: 5,
		min: 1
	},
	{
		key: 'failed.importBlocked',
		env: 'JANITARR_FAILED_IMPORT_BLOCKED',
		group: 'Failed downloads and imports',
		label: 'Treat "import blocked" as failed',
		description: 'Downloads that completed but cannot be imported (e.g. no eligible files).',
		type: 'boolean',
		default: true
	},
	{
		key: 'failed.strikes',
		env: 'JANITARR_FAILED_STRIKES',
		group: 'Failed downloads and imports',
		label: 'Strikes before giving up',
		description:
			'At this many failures an item is no longer searched and (optionally) unmonitored.',
		type: 'number',
		default: 3,
		min: 1
	},
	{
		key: 'failed.strikeDecayDays',
		env: 'JANITARR_FAILED_STRIKE_DECAY_DAYS',
		group: 'Failed downloads and imports',
		label: 'Strike decay (days)',
		description: 'Strikes reset if the last failure is older than this. 0 disables decay.',
		type: 'number',
		default: 30,
		min: 0
	},
	{
		key: 'failed.unmonitorAtThreshold',
		env: 'JANITARR_FAILED_UNMONITOR',
		group: 'Failed downloads and imports',
		label: 'Unmonitor at threshold',
		description: 'Unmonitor the item in Sonarr/Radarr when it reaches the strike limit.',
		type: 'boolean',
		default: true
	}
];

export const DEF_BY_KEY: ReadonlyMap<string, SettingDef> = new Map(DEFS.map((d) => [d.key, d]));
