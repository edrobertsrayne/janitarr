<script lang="ts">
	import InstanceCard from '#lib/components/InstanceCard.svelte';
	import JobsPanel from '#lib/components/JobsPanel.svelte';
	import QueueTable from '#lib/components/QueueTable.svelte';
	import StrikesPanel from '#lib/components/StrikesPanel.svelte';
	import { status } from '#lib/status.svelte.js';
</script>

<h1 class="sr-only">Dashboard</h1>

{#if status.error}
	<p class="mb-4 rounded-lg border border-bad/40 bg-bad/10 px-3 py-2 text-sm text-bad">
		{status.error}
	</p>
{/if}

{#if status.data}
	{@const s = status.data}
	<div class="grid grid-cols-12 gap-4">
		{#each s.instances as instance (instance.app)}
			<div class="col-span-12 *:h-full md:col-span-6">
				<InstanceCard {instance} />
			</div>
		{/each}
		<div class="col-span-12 *:h-full">
			<QueueTable instances={s.instances} />
		</div>
		<div class="col-span-12 *:h-full lg:col-span-6">
			<JobsPanel jobs={s.jobs} now={s.now} onrun={status.runJob} />
		</div>
		<div class="col-span-12 *:h-full lg:col-span-6">
			<StrikesPanel
				strikes={s.strikes}
				limit={s.strikeLimit}
				now={s.now}
				onclear={status.clearStrikes}
			/>
		</div>
	</div>
{:else if !status.error}
	<p class="text-sm text-mute">Loading…</p>
{/if}
