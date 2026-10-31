import { error, json } from '@sveltejs/kit';
import { getApp } from '#lib/server/app.js';
import { SettingsError } from '#lib/server/settings.js';

export function GET() {
	return json(getApp().settings.view());
}

/** Body: `{ values: { "<key>": value | null } }`; null resets the key to its default. */
export async function PUT({ request }) {
	const body: unknown = await request.json().catch(() => null);
	const values =
		body && typeof body === 'object' && 'values' in body
			? (body as { values: unknown }).values
			: null;
	if (!values || typeof values !== 'object' || Array.isArray(values)) {
		error(400, 'Expected { values: { key: value } }');
	}
	const { settings } = getApp();
	try {
		settings.update(values as Record<string, unknown>);
	} catch (e) {
		if (e instanceof SettingsError) error(400, e.message);
		throw e;
	}
	return json(settings.view());
}
