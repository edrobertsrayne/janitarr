<script lang="ts">
	import './layout.css';
	import favicon from '#lib/assets/favicon.svg';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();

	const links = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/settings', label: 'Settings' }
	] as const;
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Janitarr</title>
</svelte:head>

<div class="min-h-screen bg-slate-950 text-slate-200">
	<header class="border-b border-slate-800 bg-slate-900">
		<nav class="mx-auto flex max-w-6xl items-center gap-6 px-4 py-3">
			<span class="text-lg font-semibold tracking-tight text-white">Janitarr</span>
			{#each links as link (link.href)}
				<a
					href={resolve(link.href)}
					class="text-sm {page.url.pathname === link.href
						? 'text-white'
						: 'text-slate-400 hover:text-slate-200'}">{link.label}</a
				>
			{/each}
		</nav>
	</header>
	<main class="mx-auto max-w-6xl px-4 py-6">
		{@render children()}
	</main>
</div>
