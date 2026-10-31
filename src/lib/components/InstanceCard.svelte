<script lang="ts">
	import { resolve } from '$app/paths';
	import { bytes } from '#lib/format.js';
	import type { InstanceStatus } from '#lib/types.js';

	let { instance }: { instance: InstanceStatus } = $props();

	const name = $derived(instance.app === 'sonarr' ? 'Sonarr' : 'Radarr');
	const stats = $derived(
		instance.counts
			? [
					{ label: 'Missing', value: instance.counts.missing, warn: false },
					{ label: 'Cutoff unmet', value: instance.counts.cutoff, warn: false },
					{ label: 'In queue', value: instance.counts.queue, warn: false },
					{ label: 'Stalled', value: instance.counts.stalled, warn: instance.counts.stalled > 0 },
					{ label: 'Failed', value: instance.counts.failed, warn: instance.counts.failed > 0 }
				]
			: []
	);
</script>

<section class="rounded-lg border border-slate-800 bg-slate-900 p-4">
	<div class="flex items-center justify-between">
		<h2 class="text-base font-semibold text-white">{name}</h2>
		{#if !instance.configured}
			<span class="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-400">not configured</span>
		{:else if instance.online}
			<span class="rounded bg-emerald-900/60 px-2 py-0.5 text-xs text-emerald-300"
				>online · v{instance.version}</span
			>
		{:else}
			<span class="rounded bg-red-900/60 px-2 py-0.5 text-xs text-red-300">unreachable</span>
		{/if}
	</div>

	{#if !instance.configured}
		<p class="mt-3 text-sm text-slate-400">
			Set the URL and API key in <a href={resolve('/settings')} class="underline">Settings</a>.
		</p>
	{:else if instance.error}
		<p class="mt-3 text-sm break-words text-red-300">{instance.error}</p>
	{/if}

	{#if stats.length}
		<dl class="mt-4 grid grid-cols-5 gap-2 text-center">
			{#each stats as s (s.label)}
				<div class="rounded bg-slate-800/60 px-1 py-2">
					<dd class="text-xl font-semibold {s.warn ? 'text-amber-300' : 'text-white'}">
						{s.value}
					</dd>
					<dt class="text-[11px] text-slate-400">{s.label}</dt>
				</div>
			{/each}
		</dl>
	{/if}

	{#if instance.health.length}
		<ul class="mt-4 space-y-1 text-sm">
			{#each instance.health as h, i (i)}
				<li class={h.type === 'error' ? 'text-red-300' : 'text-amber-300'}>
					{h.type === 'error' ? '✖' : '⚠'}
					{h.message}
				</li>
			{/each}
		</ul>
	{/if}

	{#if instance.disks.length}
		<ul class="mt-4 space-y-2">
			{#each instance.disks as d (d.path)}
				{@const used = d.total > 0 ? 1 - d.free / d.total : 0}
				<li class="text-xs text-slate-400">
					<div class="flex justify-between">
						<span class="truncate">{d.path}</span>
						<span>{bytes(d.free)} free of {bytes(d.total)}</span>
					</div>
					<div class="mt-1 h-1.5 rounded bg-slate-800">
						<div
							class="h-1.5 rounded {used > 0.9 ? 'bg-red-500' : 'bg-sky-500'}"
							style="width: {Math.round(used * 100)}%"
						></div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>
