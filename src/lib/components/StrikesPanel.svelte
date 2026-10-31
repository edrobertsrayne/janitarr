<script lang="ts">
	import { relative } from '#lib/format.js';
	import type { StrikeRow } from '#lib/types.js';

	let {
		strikes,
		now,
		onclear
	}: { strikes: StrikeRow[]; now: number; onclear: (s: StrikeRow) => void } = $props();
</script>

<section class="rounded-lg border border-slate-800 bg-slate-900">
	<h2 class="border-b border-slate-800 px-4 py-3 text-base font-semibold text-white">Strikes</h2>
	{#if strikes.length === 0}
		<p class="px-4 py-4 text-sm text-slate-400">No failures recorded.</p>
	{:else}
		<ul class="max-h-72 divide-y divide-slate-800 overflow-y-auto">
			{#each strikes as s (s.app + s.itemId)}
				<li class="flex items-center justify-between gap-2 px-4 py-2 text-sm">
					<div class="min-w-0">
						<div class="truncate text-slate-100">
							<span class="mr-1 text-[10px] text-slate-500 uppercase">{s.app}</span>{s.title}
						</div>
						<div class="text-xs text-slate-400">
							{s.count} strike{s.count === 1 ? '' : 's'} · {relative(s.lastTs, now)}
							{#if s.disabled}<span class="text-red-300"> · given up</span>{/if}
						</div>
					</div>
					<button
						class="shrink-0 rounded bg-slate-800 px-2 py-1 text-xs hover:bg-slate-700"
						onclick={() => onclear(s)}>Clear</button
					>
				</li>
			{/each}
		</ul>
	{/if}
</section>
