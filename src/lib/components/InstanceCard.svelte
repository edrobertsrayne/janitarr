<script lang="ts">
	import { resolve } from '$app/paths';
	import { bytes } from '#lib/format.js';
	import type { InstanceStatus } from '#lib/types.js';

	let { instance }: { instance: InstanceStatus } = $props();

	const name = $derived(instance.app === 'sonarr' ? 'Sonarr' : 'Radarr');
	const stats = $derived(
		instance.counts
			? [
					{ label: 'missing', value: instance.counts.missing, tone: '' },
					{ label: 'cutoff', value: instance.counts.cutoff, tone: '' },
					{ label: 'queue', value: instance.counts.queue, tone: '' },
					{
						label: 'stalled',
						value: instance.counts.stalled,
						tone: instance.counts.stalled > 0 ? 'text-warn' : ''
					},
					{
						label: 'failed',
						value: instance.counts.failed,
						tone: instance.counts.failed > 0 ? 'text-bad' : ''
					}
				]
			: []
	);
</script>

<section class="@container rounded-[10px] border border-line bg-panel p-4">
	<div class="flex items-center gap-2">
		<h2 class="text-xs font-semibold tracking-[0.08em] text-mute uppercase">{name}</h2>
		{#if !instance.configured}
			<span class="text-xs text-mute">not configured</span>
		{:else if instance.online}
			<span class="text-xs text-mute">
				v{instance.version} ·
				<span
					class="mr-0.5 inline-block size-2 rounded-full {instance.health.length
						? 'bg-warn'
						: 'bg-ok'}"
				></span>
				{instance.health.length ? 'health warning' : 'online'}
			</span>
		{:else}
			<span class="text-xs text-bad"
				><span class="mr-0.5 inline-block size-2 rounded-full bg-bad"></span> unreachable</span
			>
		{/if}
	</div>

	{#if !instance.configured}
		<p class="mt-3 text-sm text-mute">
			Set the URL and API key in <a href={resolve('/settings')} class="text-accent underline"
				>Settings</a
			>.
		</p>
	{:else if instance.error}
		<p class="mt-3 text-sm break-words text-bad">{instance.error}</p>
	{/if}

	{#if stats.length}
		<dl class="my-3 grid grid-cols-3 gap-2 @md:grid-cols-5">
			{#each stats as s (s.label)}
				<div class="rounded-lg bg-panel2 px-2.5 py-2">
					<dd class="text-[22px] leading-tight font-semibold {s.tone}">{s.value}</dd>
					<dt class="text-xs text-mute">{s.label}</dt>
				</div>
			{/each}
		</dl>
	{/if}

	{#if instance.health.length}
		<ul class="mb-3 space-y-1 text-sm">
			{#each instance.health as h, i (i)}
				<li class={h.type === 'error' ? 'text-bad' : 'text-warn'}>
					{h.type === 'error' ? '✖' : '⚠'}
					{h.message}
				</li>
			{/each}
		</ul>
	{/if}

	{#if instance.disks.length}
		<ul class="space-y-2">
			{#each instance.disks as d (d.path)}
				{@const used = d.total > 0 ? 1 - d.free / d.total : 0}
				<li class="text-xs {used > 0.9 ? 'text-warn' : 'text-mute'}">
					<div class="h-1.5 overflow-hidden rounded-[3px] bg-panel2">
						<div
							class="h-full {used > 0.9 ? 'bg-warn' : 'bg-accent'}"
							style="width: {Math.round(used * 100)}%"
						></div>
					</div>
					<div class="mt-1 flex justify-between gap-2">
						<span class="truncate">Disk {d.path}</span>
						<span class="shrink-0"
							>{bytes(d.free)} free of {bytes(d.total)}{used > 0.9 ? ' — low' : ''}</span
						>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>
