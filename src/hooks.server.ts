import { getApp } from '#lib/server/app.js';

/** adapter-node runs this at process start, so the scheduler runs without any request. */
export function init() {
	getApp().scheduler.start();
}
