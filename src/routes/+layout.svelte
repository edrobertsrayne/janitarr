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

{#snippet icon(href: string)}
	<svg
		class="size-6"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.8"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		{#if href === '/'}
			<rect x="3" y="3" width="7" height="9" rx="1.5" />
			<rect x="14" y="3" width="7" height="5" rx="1.5" />
			<rect x="14" y="12" width="7" height="9" rx="1.5" />
			<rect x="3" y="16" width="7" height="5" rx="1.5" />
		{:else if href === '/activity'}
			<path d="M3 12h4l3-8 4 16 3-8h4" />
		{:else}
			<path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12" />
			<circle cx="16" cy="6" r="2" />
			<circle cx="10" cy="12" r="2" />
			<circle cx="18" cy="18" r="2" />
		{/if}
	</svg>
{/snippet}

<div class="min-h-screen pb-[calc(4rem+env(safe-area-inset-bottom))] md:pb-0">
	<header
		class="sticky top-0 z-10 border-b border-line bg-panel/95 pt-[env(safe-area-inset-top)] backdrop-blur"
	>
		<div class="flex items-center gap-3 px-4 py-2.5 md:gap-4 md:px-6 md:py-3">
			<a href={resolve('/')} class="text-lg font-bold" aria-label="Janitarr dashboard"
				>jani<span class="text-accent">tarr</span></a
			>
			<nav class="hidden items-center gap-4 md:flex" aria-label="Main">
				{#each links as link (link.href)}
					<a
						href={resolve(link.href)}
						aria-current={page.url.pathname === link.href ? 'page' : undefined}
						class="rounded-full border border-line px-3 py-0.5 text-sm {page.url.pathname ===
						link.href
							? 'bg-panel2 text-ink'
							: 'text-mute hover:text-ink'}">{link.label}</a
					>
				{/each}
			</nav>
			<div class="flex-1"></div>
			{#if status.data}
				<div class="flex items-center gap-2">
					<span
						class="flex items-center rounded-full border border-line px-2.5 py-1 text-sm text-mute md:px-3 md:py-0.5"
						title="{online}/{configured.length} instances online"
					>
						<span
							class="mr-1.5 inline-block size-2 rounded-full {configured.length > 0 &&
							online === configured.length
								? 'bg-ok'
								: 'bg-bad'}"
						></span>{online}/{configured.length}<span class="ml-1 hidden sm:inline"
							>instance{configured.length === 1 ? '' : 's'} online</span
						>
					</span>
					{#if next}
						<span
							class="hidden rounded-full border border-line px-3 py-0.5 text-sm text-mute md:inline"
							>Next run: {next.name} {relative(next.nextRun, status.data.now)}</span
						>
					{/if}
					<button
						class="rounded-full border px-2.5 py-1 text-sm disabled:cursor-not-allowed disabled:opacity-60 md:px-3 md:py-0.5 {status
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
		</div>
	</header>
	<main class="mx-auto max-w-[1400px] px-4 py-4 md:px-6 md:py-5">
		{@render children()}
	</main>
	<footer class="px-4 py-4 text-center text-xs text-mute md:px-6">
		{__GIT_SHA__} · built {__BUILD_DATE__.slice(0, 10)}
	</footer>
	<nav
		class="fixed inset-x-0 bottom-0 z-10 flex border-t border-line bg-panel/95 pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden"
		aria-label="Main"
	>
		{#each links as link (link.href)}
			{@const active = page.url.pathname === link.href}
			<a
				href={resolve(link.href)}
				aria-current={active ? 'page' : undefined}
				class="flex h-16 flex-1 flex-col items-center justify-center gap-1 text-xs font-medium {active
					? 'text-accent'
					: 'text-mute active:text-ink'}"
			>
				{@render icon(link.href)}
				{link.label}
			</a>
		{/each}
	</nav>
</div>
