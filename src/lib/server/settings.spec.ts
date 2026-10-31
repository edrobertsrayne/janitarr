import { beforeEach, describe, expect, it } from 'vitest';
import { SECRET_MASK, Settings, SettingsError } from './settings.js';
import { Store } from './store.js';

let store: Store;
beforeEach(() => {
	store = new Store(':memory:');
});

describe('settings resolution', () => {
	it('falls back to defaults', () => {
		const cfg = new Settings(store, {}).config();
		expect(cfg.search.maxPerCycle).toBe(10);
		expect(cfg.general.dryRun).toBe(false);
	});

	it('lets the UI change keys that env does not set', () => {
		const s = new Settings(store, {});
		s.update({ 'search.maxPerCycle': '25', 'general.dryRun': true });
		const cfg = s.config();
		expect(cfg.search.maxPerCycle).toBe(25);
		expect(cfg.general.dryRun).toBe(true);
	});

	it('env wins, is reported as locked, and rejects UI edits', () => {
		const s = new Settings(store, { JANITARR_SEARCH_MAX_PER_CYCLE: '7' });
		expect(s.config().search.maxPerCycle).toBe(7);
		const row = s.view().find((v) => v.key === 'search.maxPerCycle');
		expect(row).toMatchObject({ locked: true, source: 'env', value: 7 });
		expect(() => s.update({ 'search.maxPerCycle': 9 })).toThrow(/JANITARR_SEARCH_MAX_PER_CYCLE/);
	});

	it('persists UI edits across restarts and reset returns to the default', () => {
		new Settings(store, {}).update({ 'search.intervalMinutes': 15 });
		const s = new Settings(store, {});
		expect(s.config().search.intervalMinutes).toBe(15);
		s.update({ 'search.intervalMinutes': null });
		expect(s.config().search.intervalMinutes).toBe(60);
	});

	it('applies nothing when any key in an update is invalid', () => {
		const s = new Settings(store, {});
		expect(() => s.update({ 'search.maxPerCycle': 20, 'search.intervalMinutes': 0 })).toThrow(
			SettingsError
		);
		expect(s.config().search.maxPerCycle).toBe(10);
	});

	it('validates types, unknown keys and URLs', () => {
		const s = new Settings(store, {});
		expect(() => s.update({ 'general.dryRun': 'yes' })).toThrow(SettingsError);
		expect(() => s.update({ 'nope.nothing': 1 })).toThrow(/Unknown/);
		expect(() => s.update({ 'sonarr.url': 'sonarr:8989' })).toThrow(SettingsError);
		expect(() => s.update({ 'sonarr.url': 'http://sonarr:8989' })).not.toThrow();
	});

	it('ignores unusable env values and reports them', () => {
		const s = new Settings(store, {
			JANITARR_SEARCH_MAX_PER_CYCLE: 'lots',
			JANITARR_DRY_RUN: 'maybe'
		});
		expect(s.envErrors).toHaveLength(2);
		expect(s.config().search.maxPerCycle).toBe(10);
		expect(s.isLocked('search.maxPerCycle')).toBe(false);
	});

	it('never exposes secrets and keeps them when the masked value is submitted back', () => {
		const s = new Settings(store, { SONARR_API_KEY: 'abc123' });
		expect(JSON.stringify(s.view())).not.toContain('abc123');
		expect(s.view().find((v) => v.key === 'sonarr.apiKey')?.value).toBe(SECRET_MASK);

		const t = new Settings(store, {});
		t.update({ 'radarr.apiKey': 'k1' });
		t.update({ 'radarr.apiKey': SECRET_MASK });
		expect(t.config().radarr.apiKey).toBe('k1');
	});
});
