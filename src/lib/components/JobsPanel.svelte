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

	const dot = (job: JobState) =>
		job.lastError ? 'bg-bad' : job.running ? 'bg-accent' : job.enabled ? 'bg-ok' : 'bg-line';
</script>

<section class="rounded-[10px] border border-line bg-panel p-4">
	<h2 class="mb-3 text-xs font-semibold tracking-[0.08em] text-mute uppercase">Jobs</h2>
	<ul>
		{#each jobs as job (job.name)}
			<li class="flex items-center gap-3 border-t border-line py-2.5 text-sm first:border-t-0">
				<span class="size-2 shrink-0 rounded-full {dot(job)}"></span>
				<div class="min-w-0 flex-1">
					<div class="font-semibold">
						{titles[job.name]}
						{#if !job.enabled}<span class="ml-1 text-xs font-normal text-mute">disabled</span>{/if}
					</div>
					<div class="text-xs text-mute">
						every {job.intervalMinutes}m ·
						{#if job.running}
							running…
						{:else}
							last {job.lastEnd ? relative(job.lastEnd, now) : 'never'}{job.enabled
								? ` · next ${relative(job.nextRun, now)}`
								: ''}
						{/if}
					</div>
					{#if job.lastError}<div class="text-xs text-bad">{job.lastError}</div>{/if}
				</div>
				<button
					class="rounded-md border border-line bg-panel2 px-3 py-1 text-xs hover:border-accent disabled:opacity-50"
					disabled={job.running}
					onclick={() => onrun(job.name)}>Run</button
				>
			</li>
		{/each}
	</ul>
</section>
