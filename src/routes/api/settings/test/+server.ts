import { error, json } from '@sveltejs/kit';
import { getApp } from '#lib/server/app.js';
import { HttpArrClient } from '#lib/server/arr.js';
import { SECRET_MASK } from '#lib/server/settings.js';

/**
 * Body: `{ app: "sonarr" | "radarr", url?: string, apiKey?: string }`.
 * Omitted (or masked) fields fall back to the currently effective value, so
 * unsaved edits can be tested without saving. Always answers 200 with
 * `{ ok, message }`; connection failures are reported, not thrown.
 */
export async function POST({ request }) {
	const body: unknown = await request.json().catch(() => null);
	const b = body && typeof body === 'object' ? (body as Record<string, unknown>) : {};
	if (b.app !== 'sonarr' && b.app !== 'radarr') error(400, 'app must be "sonarr" or "radarr"');

	const saved = getApp().settings.config()[b.app];
	const url = typeof b.url === 'string' && b.url.trim() !== '' ? b.url.trim() : saved.url;
	const apiKey =
		typeof b.apiKey === 'string' && b.apiKey !== '' && b.apiKey !== SECRET_MASK
			? b.apiKey
			: saved.apiKey;
	if (!url || !apiKey) return json({ ok: false, message: 'URL and API key are both required.' });
	let parsed: URL;
	try {
		parsed = new URL(url);
	} catch {
		return json({ ok: false, message: 'URL is not valid (include http:// or https://).' });
	}
	if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
		return json({ ok: false, message: 'URL must start with http:// or https://' });
	}

	try {
		const { version } = await new HttpArrClient(b.app, url, apiKey).status();
		return json({ ok: true, message: `Connected. Version ${version}.` });
	} catch (e) {
		let code = '';
		if (e instanceof Error && e.cause && typeof e.cause === 'object' && 'code' in e.cause) {
			code = ` (${String(e.cause.code)})`;
		}
		const msg = (e instanceof Error ? e.message : String(e)) + code;
		return json({
			ok: false,
			message: /HTTP 401/.test(msg) ? 'Server reachable but the API key was rejected.' : msg
		});
	}
}
