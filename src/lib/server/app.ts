import { join } from 'node:path';
import { Scheduler } from './scheduler';
import { Settings } from './settings';
import { Store } from './store';

export interface JanitarrApp {
	store: Store;
	settings: Settings;
	scheduler: Scheduler;
}

const KEY = Symbol.for('janitarr.app');
type Holder = typeof globalThis & { [KEY]?: JanitarrApp };

/** Process-wide singleton (survives dev-server module reloads). */
export function getApp(): JanitarrApp {
	const g = globalThis as Holder;
	if (!g[KEY]) {
		const dataDir = process.env.JANITARR_DATA_DIR || './data';
		const store = new Store(join(dataDir, 'janitarr.db'));
		const settings = new Settings(store);
		for (const msg of settings.envErrors) {
			console.warn(`[janitarr] ${msg}`);
			store.log({ job: 'system', app: null, level: 'warn', message: msg }, Date.now());
		}
		g[KEY] = { store, settings, scheduler: new Scheduler(store, settings) };
	}
	return g[KEY];
}
