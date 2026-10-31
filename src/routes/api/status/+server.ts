import { json } from '@sveltejs/kit';
import { getApp } from '#lib/server/app.js';
import { buildStatus } from '#lib/server/status.js';

export async function GET() {
	const { store, settings, scheduler } = getApp();
	return json(await buildStatus(store, settings, scheduler));
}
