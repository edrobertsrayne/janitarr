import { error, json } from '@sveltejs/kit';
import { getApp } from '#lib/server/app.js';
import { JOBS, type JobName } from '#lib/types.js';

export function POST({ params }) {
	const job = JOBS.find((j) => j === params.job);
	if (!job) error(404, `Unknown job: ${params.job}`);
	const { scheduler } = getApp();
	if (scheduler.states().find((s) => s.name === job)?.running)
		error(409, `${job} is already running`);
	// Fire and forget: the dashboard polls for the result.
	void scheduler.run(job as JobName);
	return json({ started: job }, { status: 202 });
}
