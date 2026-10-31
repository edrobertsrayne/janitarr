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

<section class="rounded-[10px] border border-line bg-panel p-4">
	<h2 class="mb-3 text-xs font-semibold tracking-[0.08em] text-mute uppercase">
		Queue <span class="font-normal tracking-normal normal-case">({rows.length})</span>
	</h2>
	{#if rows.length === 0}
		<p class="py-4 text-sm text-mute">Nothing downloading.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-left text-sm">
				<thead class="text-xs text-mute">
					<tr>
						<th class="px-2 py-1.5 font-medium">Item</th>
						<th class="px-2 py-1.5 font-medium">State</th>
						<th class="w-44 px-2 py-1.5 font-medium">Progress</th>
						<th class="px-2 py-1.5 font-medium">Size</th>
						<th class="px-2 py-1.5 font-medium">Strikes</th>
					</tr>
				</thead>
				<tbody>
					{#each rows as q (q.app + q.id)}
						<tr class="border-t border-line align-top">
							<td class="max-w-md px-2 py-2">
								<div>
									<span class="mr-1 text-[10px] text-faint uppercase">{q.app}</span>{q.title}
								</div>
								<div class="truncate text-xs text-mute" title={q.release}>{q.release}</div>
								{#if q.message && q.flag}
									<div class="mt-0.5 text-xs text-warn/80">{q.message}</div>
								{/if}
							</td>
							<td class="px-2 py-2 whitespace-nowrap">
								{#if q.flag === 'failed'}
									<span class="rounded bg-[#4a1822] px-1.5 py-px text-[11px] text-bad">failed</span>
								{:else if q.flag === 'stalled'}
									<span class="rounded bg-[#4a3410] px-1.5 py-px text-[11px] text-warn"
										>stalled</span
									>
								{:else}
									<span class="text-xs text-mute">{q.state ?? q.status}</span>
								{/if}
							</td>
							<td class="px-2 py-2">
								<div class="h-1.5 overflow-hidden rounded-[3px] bg-panel2">
									<div
										class="h-full {q.flag === 'failed'
											? 'bg-bad'
											: q.flag === 'stalled'
												? 'bg-warn'
												: 'bg-accent'}"
										style="width: {Math.round(q.progress * 100)}%"
									></div>
								</div>
								<div class="mt-1 text-xs text-mute">
									{Math.round(q.progress * 100)}%{q.timeleft ? ` · ${q.timeleft}` : ''}
								</div>
							</td>
							<td class="px-2 py-2 whitespace-nowrap text-mute">{bytes(q.size)}</td>
							<td class="px-2 py-2 {q.strikes > 0 ? 'text-warn' : 'text-faint'}">{q.strikes}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
