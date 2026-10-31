<script lang="ts">
	import { relative } from '#lib/format.js';
	import type { JobName, JobState } from '#lib/types.js';

	let { jobs, now, onrun }: { jobs: JobState[]; now: number; onrun: (job: JobName) => void } =
		$props();

	const titles: Record<JobName, string> = {
		search: 'Search missing / cutoff',
		stalled: 'Remove stalled',
		failed: 'Clear failed + strikes'
	};
</script>

<section class="rounded-lg border border-slate-800 bg-slate-900">
	<h2 class="border-b border-slate-800 px-4 py-3 text-base font-semibold text-white">Jobs</h2>
	<ul class="divide-y divide-slate-800">
		{#each jobs as job (job.name)}
			<li class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm">
				<div>
					<div class="text-slate-100">
						{titles[job.name]}
						{#if !job.enabled}<span class="ml-1 rounded bg-slate-800 px-1.5 text-xs text-slate-400"
								>disabled</span
							>{/if}
					</div>
					<div class="text-xs text-slate-400">
						every {job.intervalMinutes}m ·
						{#if job.running}
							running…
						{:else}
							last {job.lastEnd ? relative(job.lastEnd, now) : 'never'}{job.enabled
								? ` · next ${relative(job.nextRun, now)}`
								: ''}
						{/if}
					</div>
					{#if job.lastError}<div class="text-xs text-red-300">{job.lastError}</div>{/if}
				</div>
				<button
					class="rounded bg-slate-800 px-3 py-1 text-xs text-slate-100 hover:bg-slate-700 disabled:opacity-50"
					disabled={job.running}
					onclick={() => onrun(job.name)}>Run now</button
				>
			</li>
		{/each}
	</ul>
</section>
