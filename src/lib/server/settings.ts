import type { SettingSource, SettingValue, SettingView } from '#lib/types.js';
import { DEFS, DEF_BY_KEY, type Config, type SettingDef } from './settings-defs';
import type { Store } from './store';

export const SECRET_MASK = '********';

export class SettingsError extends Error {}

function parseEnv(def: SettingDef, raw: string): SettingValue {
	const v = raw.trim();
	switch (def.type) {
		case 'boolean': {
			const l = v.toLowerCase();
			if (['true', '1', 'yes', 'on'].includes(l)) return true;
			if (['false', '0', 'no', 'off'].includes(l)) return false;
			throw new SettingsError(`${def.env}: expected a boolean, got "${raw}"`);
		}
		case 'number': {
			const n = Number(v);
			if (v === '' || !Number.isFinite(n)) {
				throw new SettingsError(`${def.env}: expected a number, got "${raw}"`);
			}
			return checkNumber(def, n);
		}
		case 'string':
			return checkString(def, v);
	}
}

function checkNumber(def: SettingDef, n: number): number {
	if (def.min !== undefined && n < def.min) {
		throw new SettingsError(`${def.label}: must be at least ${def.min}`);
	}
	return n;
}

function checkString(def: SettingDef, s: string): string {
	if (def.key.endsWith('.url') && s !== '') {
		let u: URL;
		try {
			u = new URL(s);
		} catch {
			throw new SettingsError(`${def.label}: not a valid URL`);
		}
		if (u.protocol !== 'http:' && u.protocol !== 'https:') {
			throw new SettingsError(`${def.label}: must start with http:// or https://`);
		}
	}
	return s;
}

/** Validate and normalise a value submitted via the UI. */
function coerce(def: SettingDef, input: unknown): SettingValue {
	switch (def.type) {
		case 'boolean':
			if (typeof input !== 'boolean') throw new SettingsError(`${def.label}: expected a boolean`);
			return input;
		case 'number': {
			const n = typeof input === 'string' && input.trim() !== '' ? Number(input) : input;
			if (typeof n !== 'number' || !Number.isFinite(n)) {
				throw new SettingsError(`${def.label}: expected a number`);
			}
			return checkNumber(def, n);
		}
		case 'string':
			if (typeof input !== 'string') throw new SettingsError(`${def.label}: expected a string`);
			return checkString(def, input.trim());
	}
}

/**
 * Resolution order per key: environment variable (locks the key) → value saved
 * from the UI → built-in default.
 */
export class Settings {
	private env: Map<string, SettingValue>;
	/** Environment variables that were set but unusable. Reported once at startup. */
	readonly envErrors: string[] = [];

	constructor(
		private store: Store,
		environ: Record<string, string | undefined> = process.env
	) {
		this.env = new Map();
		for (const def of DEFS) {
			const raw = environ[def.env];
			if (raw === undefined || raw.trim() === '') continue;
			try {
				this.env.set(def.key, parseEnv(def, raw));
			} catch (e) {
				this.envErrors.push(`${(e as Error).message} (ignored)`);
			}
		}
	}

	isLocked(key: string): boolean {
		return this.env.has(key);
	}

	private resolveOne(
		def: SettingDef,
		overrides: Map<string, string>
	): { value: SettingValue; source: SettingSource } {
		const fromEnv = this.env.get(def.key);
		if (fromEnv !== undefined) return { value: fromEnv, source: 'env' };
		const saved = overrides.get(def.key);
		if (saved !== undefined) {
			try {
				return { value: coerce(def, JSON.parse(saved)), source: 'ui' };
			} catch {
				// Corrupt or no-longer-valid saved value: fall through to the default.
			}
		}
		return { value: def.default, source: 'default' };
	}

	config(): Config {
		const overrides = this.store.overrides();
		const out: Record<string, Record<string, SettingValue>> = {};
		for (const def of DEFS) {
			const [group, name] = def.key.split('.');
			(out[group] ??= {})[name] = this.resolveOne(def, overrides).value;
		}
		return out as unknown as Config;
	}

	view(): SettingView[] {
		const overrides = this.store.overrides();
		return DEFS.map((def) => {
			const { value, source } = this.resolveOne(def, overrides);
			return {
				key: def.key,
				env: def.env,
				group: def.group,
				label: def.label,
				description: def.description,
				type: def.type,
				secret: def.secret === true,
				value: def.secret && value !== '' ? SECRET_MASK : value,
				default: def.default,
				source,
				locked: source === 'env'
			};
		});
	}

	/**
	 * Apply UI edits. `null` resets a key to its default. All keys are validated
	 * before any is written.
	 */
	update(values: Record<string, unknown>): void {
		const writes: [string, SettingValue | null][] = [];
		for (const [key, input] of Object.entries(values)) {
			const def = DEF_BY_KEY.get(key);
			if (!def) throw new SettingsError(`Unknown setting: ${key}`);
			if (this.isLocked(key))
				throw new SettingsError(`${def.label} is set by ${def.env} and cannot be changed here`);
			if (input === null) {
				writes.push([key, null]);
				continue;
			}
			if (def.secret && input === SECRET_MASK) continue; // unchanged
			writes.push([key, coerce(def, input)]);
		}
		for (const [key, v] of writes) {
			if (v === null) this.store.deleteOverride(key);
			else this.store.setOverride(key, JSON.stringify(v));
		}
	}
}
