<script lang="ts">
	import { clock } from '#lib/format.js';
	import type { ActivityEntry } from '#lib/types.js';

	let { entries }: { entries: ActivityEntry[] } = $props();

	const tone = { info: 'text-mute', warn: 'text-warn', error: 'text-bad' };
</script>

<section class="rounded-[10px] border border-line bg-panel p-4">
	<h2 class="mb-3 text-xs font-semibold tracking-[0.08em] text-mute uppercase">Activity</h2>
	{#if entries.length === 0}
		<p class="py-2 text-sm text-mute">Nothing yet.</p>
	{:else}
		<ul class="font-mono text-xs">
			{#each entries as e (e.id)}
				<li class="flex gap-3 py-[3px] {tone[e.level]}">
					<span class="shrink-0 text-faint">{clock(e.ts)}</span>
					<span class="w-16 shrink-0 text-faint">{e.job}{e.app ? `/${e.app[0]}` : ''}</span>
					<span class="break-words">
						{#if e.dryRun}<span class="mr-1 rounded bg-[#16335f] px-1 text-[#9cc4ff]">dry-run</span
							>{/if}{e.message}
					</span>
				</li>
			{/each}
		</ul>
	{/if}
</section>
