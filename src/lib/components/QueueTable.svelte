<script lang="ts">
	import { bytes } from '#lib/format.js';
	import type { InstanceStatus } from '#lib/types.js';

	let { instances }: { instances: InstanceStatus[] } = $props();

	const rows = $derived(
		instances
			.flatMap((i) => i.queue.map((q) => ({ ...q, app: i.app })))
			.sort((a, b) => {
				const weight = (f: string | null) => (f === 'failed' ? 0 : f === 'stalled' ? 1 : 2);
				return weight(a.flag) - weight(b.flag) || a.title.localeCompare(b.title);
			})
	);
</script>

<section class="rounded-lg border border-slate-800 bg-slate-900">
	<h2 class="border-b border-slate-800 px-4 py-3 text-base font-semibold text-white">
		Queue <span class="text-sm font-normal text-slate-400">({rows.length})</span>
	</h2>
	{#if rows.length === 0}
		<p class="px-4 py-6 text-sm text-slate-400">Nothing downloading.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-left text-sm">
				<thead class="text-xs text-slate-400 uppercase">
					<tr>
						<th class="px-4 py-2">Item</th>
						<th class="px-2 py-2">State</th>
						<th class="w-48 px-2 py-2">Progress</th>
						<th class="px-2 py-2">Size</th>
						<th class="px-4 py-2">Strikes</th>
					</tr>
				</thead>
				<tbody>
					{#each rows as q (q.app + q.id)}
						<tr class="border-t border-slate-800 align-top">
							<td class="max-w-md px-4 py-2">
								<div class="text-slate-100">
									<span class="mr-1 text-[10px] text-slate-500 uppercase">{q.app}</span>{q.title}
								</div>
								<div class="truncate text-xs text-slate-500" title={q.release}>{q.release}</div>
								{#if q.message && q.flag}
									<div class="mt-0.5 text-xs text-amber-300/80">{q.message}</div>
								{/if}
							</td>
							<td class="px-2 py-2 whitespace-nowrap">
								{#if q.flag === 'failed'}
									<span class="rounded bg-red-900/60 px-1.5 py-0.5 text-xs text-red-300"
										>failed</span
									>
								{:else if q.flag === 'stalled'}
									<span class="rounded bg-amber-900/60 px-1.5 py-0.5 text-xs text-amber-300"
										>stalled</span
									>
								{:else}
									<span class="text-xs text-slate-400">{q.state ?? q.status}</span>
								{/if}
							</td>
							<td class="px-2 py-2">
								<div class="h-1.5 rounded bg-slate-800">
									<div
										class="h-1.5 rounded bg-sky-500"
										style="width: {Math.round(q.progress * 100)}%"
									></div>
								</div>
								<div class="mt-1 text-xs text-slate-400">
									{Math.round(q.progress * 100)}%{q.timeleft ? ` · ${q.timeleft}` : ''}
								</div>
							</td>
							<td class="px-2 py-2 whitespace-nowrap text-slate-400">{bytes(q.size)}</td>
							<td class="px-4 py-2 {q.strikes > 0 ? 'text-amber-300' : 'text-slate-500'}"
								>{q.strikes}</td
							>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
