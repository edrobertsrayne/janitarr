<script lang="ts">
	import './layout.css';
	import { onMount } from 'svelte';
	import favicon from '#lib/assets/favicon.svg';
	import { relative } from '#lib/format.js';
	import { status } from '#lib/status.svelte.js';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const links = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/settings', label: 'Settings' }
	] as const;

	const configured = $derived(status.data?.instances.filter((i) => i.configured) ?? []);
	const online = $derived(configured.filter((i) => i.online).length);
	const next = $derived.by(() => {
		const jobs = status.data?.jobs.filter((j) => j.enabled && !j.running) ?? [];
		return jobs.length ? jobs.reduce((a, b) => (b.nextRun < a.nextRun ? b : a)) : null;
	});

	onMount(() => status.start());
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Janitarr</title>
</svelte:head>

<div class="min-h-screen">
	<header class="sticky top-0 z-10 border-b border-line bg-panel">
		<nav class="flex flex-wrap items-center gap-x-4 gap-y-2 px-6 py-3">
			<span class="text-base font-bold">jani<span class="text-accent">tarr</span></span>
			{#each links as link (link.href)}
				<a
					href={resolve(link.href)}
					class="rounded-full border border-line px-3 py-0.5 text-xs {page.url.pathname ===
					link.href
						? 'bg-panel2 text-ink'
						: 'text-mute hover:text-ink'}">{link.label}</a
				>
			{/each}
			<div class="flex-1"></div>
			{#if status.data}
				<span class="rounded-full border border-line px-3 py-0.5 text-xs text-mute">
					<span
						class="mr-1.5 inline-block size-2 rounded-full {configured.length > 0 &&
						online === configured.length
							? 'bg-ok'
							: 'bg-bad'}"
					></span>{online}/{configured.length} instance{configured.length === 1 ? '' : 's'} online
				</span>
				{#if next}
					<span class="rounded-full border border-line px-3 py-0.5 text-xs text-mute"
						>Next run: {next.name} {relative(next.nextRun, status.data.now)}</span
					>
				{/if}
				<button
					class="rounded-full border px-3 py-0.5 text-xs disabled:cursor-not-allowed disabled:opacity-60 {status
						.data.dryRun
						? 'border-[#2b5aa0] bg-[#16335f] text-[#9cc4ff]'
						: 'border-line text-mute hover:text-ink'}"
					disabled={status.busy || status.data.dryRunLocked}
					title={status.data.dryRunLocked ? 'Set by JANITARR_DRY_RUN' : 'Toggle dry-run mode'}
					onclick={() => status.toggleDryRun()}>Dry run: {status.data.dryRun ? 'on' : 'off'}</button
				>
			{/if}
		</nav>
	</header>
	<main class="mx-auto max-w-[1400px] px-6 py-5">
		{@render children()}
	</main>
</div>
