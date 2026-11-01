<script lang="ts">
	import { relative } from '#lib/format.js';
	import type { StrikeRow } from '#lib/types.js';

	let {
		strikes,
		limit,
		now,
		onclear
	}: { strikes: StrikeRow[]; limit: number; now: number; onclear: (s: StrikeRow) => void } =
		$props();
</script>

<section class="rounded-[10px] border border-line bg-panel p-4">
	<h2 class="mb-3 text-xs font-semibold tracking-[0.08em] text-mute uppercase">Strikes</h2>
	{#if strikes.length === 0}
		<p class="py-2 text-sm text-mute">No failures recorded.</p>
	{:else}
		<ul class="max-h-72 overflow-y-auto">
			{#each strikes as s (s.app + s.itemId)}
				<li class="flex items-center gap-3 border-t border-line py-2.5 text-sm first:border-t-0">
					<span
						class="flex shrink-0 gap-1"
						title="{s.count} of {limit} strike{limit === 1 ? '' : 's'}"
					>
						{#each Array.from({ length: Math.max(limit, s.count) }, (_, n) => n) as i (i)}
							<i class="size-2.5 rounded-xs {i < s.count ? 'bg-bad' : 'bg-line'}"></i>
						{/each}
					</span>
					<div class="min-w-0 flex-1">
						<div class="truncate font-semibold">
							<span class="mr-1 text-[10px] font-normal text-faint uppercase">{s.app}</span
							>{s.title}
						</div>
						<div class="text-xs text-mute">
							{s.disabled ? 'given up' : `last ${relative(s.lastTs, now)}`}
							{#if s.disabled}<span class="text-bad"> · skipped until cleared</span>{/if}
						</div>
					</div>
					<button
						class="shrink-0 rounded-md border border-line bg-panel2 px-3 py-1 text-xs hover:border-accent pointer-coarse:px-4 pointer-coarse:py-2 pointer-coarse:text-sm"
						onclick={() => onclear(s)}>Clear</button
					>
				</li>
			{/each}
		</ul>
	{/if}
</section>
