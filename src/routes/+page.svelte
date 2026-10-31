<script lang="ts">
	import { onMount } from 'svelte';
	import ActivityLog from '#lib/components/ActivityLog.svelte';
	import InstanceCard from '#lib/components/InstanceCard.svelte';
	import JobsPanel from '#lib/components/JobsPanel.svelte';
	import QueueTable from '#lib/components/QueueTable.svelte';
	import StrikesPanel from '#lib/components/StrikesPanel.svelte';
	import type { JobName, StatusPayload, StrikeRow } from '#lib/types.js';

	const POLL_MS = 10_000;

	let status = $state<StatusPayload | null>(null);
	let error = $state<string | null>(null);
	let busy = $state(false);

	async function refresh() {
		try {
			const res = await fetch('/api/status');
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			status = await res.json();
			error = null;
		} catch (e) {
			error = `Could not load status: ${(e as Error).message}`;
		}
	}

	async function post(url: string, body: unknown = {}) {
		// JSON content type keeps SvelteKit's CSRF check from treating this as a form post.
		const res = await fetch(url, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		if (!res.ok) error = (await res.json().catch(() => null))?.message ?? `HTTP ${res.status}`;
		await refresh();
	}

	const runJob = (job: JobName) => post(`/api/run/${job}`);
	const clearStrikes = (s: StrikeRow) =>
		post('/api/strikes/clear', { app: s.app, itemId: s.itemId });

	async function toggleDryRun() {
		if (!status) return;
		busy = true;
		const res = await fetch('/api/settings', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ values: { 'general.dryRun': !status.dryRun } })
		});
		if (!res.ok) error = (await res.json().catch(() => null))?.message ?? `HTTP ${res.status}`;
		busy = false;
		await refresh();
	}

	onMount(() => {
		refresh();
		const timer = setInterval(refresh, POLL_MS);
		return () => clearInterval(timer);
	});
</script>

<div class="mb-6 flex items-center justify-between">
	<h1 class="text-xl font-semibold text-white">Dashboard</h1>
	{#if status}
		<button
			class="rounded px-3 py-1 text-sm disabled:opacity-50 {status.dryRun
				? 'bg-sky-900/60 text-sky-200'
				: 'bg-slate-800 text-slate-300'}"
			disabled={busy || status.dryRunLocked}
			title={status.dryRunLocked ? 'Set by JANITARR_DRY_RUN' : 'Toggle dry-run mode'}
			onclick={toggleDryRun}>Dry run: {status.dryRun ? 'on' : 'off'}</button
		>
	{/if}
</div>

{#if error}
	<p class="mb-4 rounded border border-red-900 bg-red-950/50 px-3 py-2 text-sm text-red-300">
		{error}
	</p>
{/if}

{#if status}
	<div class="space-y-6">
		<div class="grid gap-6 md:grid-cols-2">
			{#each status.instances as instance (instance.app)}
				<InstanceCard {instance} />
			{/each}
		</div>
		<QueueTable instances={status.instances} />
		<div class="grid gap-6 lg:grid-cols-2">
			<JobsPanel jobs={status.jobs} now={status.now} onrun={runJob} />
			<StrikesPanel strikes={status.strikes} now={status.now} onclear={clearStrikes} />
		</div>
		<ActivityLog entries={status.activity} />
	</div>
{:else if !error}
	<p class="text-sm text-slate-400">Loading…</p>
{/if}
