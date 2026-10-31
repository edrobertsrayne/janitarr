<script lang="ts">
	import { clock } from '#lib/format.js';
	import type { ActivityEntry } from '#lib/types.js';

	let { entries }: { entries: ActivityEntry[] } = $props();

	const tone = { info: 'text-slate-300', warn: 'text-amber-300', error: 'text-red-300' };
</script>

<section class="rounded-lg border border-slate-800 bg-slate-900">
	<h2 class="border-b border-slate-800 px-4 py-3 text-base font-semibold text-white">Activity</h2>
	{#if entries.length === 0}
		<p class="px-4 py-4 text-sm text-slate-400">Nothing yet.</p>
	{:else}
		<ul class="max-h-96 divide-y divide-slate-800 overflow-y-auto font-mono text-xs">
			{#each entries as e (e.id)}
				<li class="flex gap-3 px-4 py-1.5 {tone[e.level]}">
					<span class="shrink-0 text-slate-500">{clock(e.ts)}</span>
					<span class="w-14 shrink-0 text-slate-500">{e.job}{e.app ? `/${e.app[0]}` : ''}</span>
					<span class="break-words">
						{#if e.dryRun}<span class="mr-1 rounded bg-sky-900/60 px-1 text-sky-300">dry-run</span
							>{/if}{e.message}
					</span>
				</li>
			{/each}
		</ul>
	{/if}
</section>
