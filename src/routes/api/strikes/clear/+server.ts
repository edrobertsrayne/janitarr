import { error, json } from '@sveltejs/kit';
import { getApp } from '#lib/server/app.js';
import { APPS } from '#lib/types.js';

/** Body: `{ app, itemId }`. Forgets strikes so the item is searched again. */
export async function POST({ request }) {
	const body = (await request.json().catch(() => null)) as {
		app?: unknown;
		itemId?: unknown;
	} | null;
	const app = APPS.find((a) => a === body?.app);
	if (!app || typeof body?.itemId !== 'number') error(400, 'Expected { app, itemId }');
	getApp().store.clearStrikes(app, body.itemId);
	return json({ ok: true });
}
