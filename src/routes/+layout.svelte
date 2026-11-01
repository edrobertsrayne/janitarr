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
		{ href: '/activity', label: 'Activity' },
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
	<header class="z-10 border-b border-line bg-panel md:sticky md:top-0">
		<nav class="flex flex-wrap items-center gap-x-3 gap-y-2 px-4 py-3 md:gap-x-4 md:px-6">
			<span class="text-base font-bold">jani<span class="text-accent">tarr</span></span>
			{#each links as link (link.href)}
				<a
					href={resolve(link.href)}
					class="rounded-full border border-line px-3 py-0.5 text-xs pointer-coarse:px-3.5 pointer-coarse:py-1.5 pointer-coarse:text-sm {page
						.url.pathname === link.href
						? 'bg-panel2 text-ink'
						: 'text-mute hover:text-ink'}">{link.label}</a
				>
			{/each}
			<div class="hidden flex-1 md:block"></div>
			{#if status.data}
				<div class="flex w-full flex-wrap items-center gap-2 md:w-auto">
					<span
						class="rounded-full border border-line px-3 py-0.5 text-xs text-mute pointer-coarse:py-1.5"
					>
						<span
							class="mr-1.5 inline-block size-2 rounded-full {configured.length > 0 &&
							online === configured.length
								? 'bg-ok'
								: 'bg-bad'}"
						></span>{online}/{configured.length} instance{configured.length === 1 ? '' : 's'} online
					</span>
					{#if next}
						<span
							class="rounded-full border border-line px-3 py-0.5 text-xs text-mute pointer-coarse:py-1.5"
							>Next run: {next.name} {relative(next.nextRun, status.data.now)}</span
						>
					{/if}
					<button
						class="rounded-full border px-3 py-0.5 text-xs disabled:cursor-not-allowed disabled:opacity-60 pointer-coarse:py-1.5 {status
							.data.dryRun
							? 'border-[#2b5aa0] bg-[#16335f] text-[#9cc4ff]'
							: 'border-line text-mute hover:text-ink'}"
						disabled={status.busy || status.data.dryRunLocked}
						title={status.data.dryRunLocked ? 'Set by JANITARR_DRY_RUN' : 'Toggle dry-run mode'}
						onclick={() => status.toggleDryRun()}
						>Dry run: {status.data.dryRun ? 'on' : 'off'}</button
					>
				</div>
			{/if}
		</nav>
	</header>
	<main class="mx-auto max-w-[1400px] px-4 py-4 md:px-6 md:py-5">
		{@render children()}
	</main>
	<footer class="px-4 py-4 text-center text-xs text-mute md:px-6">
		{__GIT_SHA__} · built {__BUILD_DATE__.slice(0, 10)}
	</footer>
</div>
