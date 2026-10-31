<script lang="ts">
	import { onMount } from 'svelte';
	import { status } from '#lib/status.svelte.js';
	import type { SettingValue, SettingView } from '#lib/types.js';

	let settings = $state<SettingView[]>([]);
	let edits = $state<Record<string, SettingValue>>({});
	let message = $state<{ ok: boolean; text: string } | null>(null);
	let saving = $state(false);
	let tests = $state<Record<string, { busy: boolean; ok?: boolean; text?: string }>>({});

	async function testConnection(app: string) {
		const val = (k: string) => (k in edits ? edits[k] : settings.find((s) => s.key === k)?.value);
		tests[app] = { busy: true };
		try {
			const res = await fetch('/api/settings/test', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ app, url: val(`${app}.url`), apiKey: val(`${app}.apiKey`) })
			});
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const r = await res.json();
			tests[app] = { busy: false, ok: r.ok, text: r.message };
		} catch (e) {
			tests[app] = { busy: false, ok: false, text: (e as Error).message };
		}
	}

	const groups = $derived(
		[...new Set(settings.map((s) => s.group))].map((name) => {
			const items = settings.filter((s) => s.group === name);
			return {
				name,
				slug: name.toLowerCase().replace(/\W+/g, '-'),
				items,
				dirty: items.filter((s) => s.key in edits).length,
				locked: items.some((s) => s.locked)
			};
		})
	);
	const dirtyCount = $derived(Object.keys(edits).length);
	let active = $state('');

	function spy() {
		const y = window.scrollY + 120;
		let cur = groups[0]?.slug ?? '';
		for (const g of groups) {
			const el = document.getElementById(`g-${g.slug}`);
			if (el && el.getBoundingClientRect().top + window.scrollY <= y) cur = g.slug;
		}
		active = cur;
	}

	function load(next: SettingView[]) {
		settings = next;
		edits = {};
	}

	async function send(values: Record<string, SettingValue | null>, okText: string) {
		saving = true;
		message = null;
		try {
			const res = await fetch('/api/settings', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ values })
			});
			if (!res.ok)
				throw new Error((await res.json().catch(() => null))?.message ?? `HTTP ${res.status}`);
			load(await res.json());
			status.refresh();
			message = { ok: true, text: okText };
		} catch (e) {
			message = { ok: false, text: (e as Error).message };
		} finally {
			saving = false;
		}
	}

	const save = () => send({ ...edits }, 'Saved. Changes apply on the next job run.');
	const reset = (key: string) => send({ [key]: null }, 'Reset to default.');

	function edit(s: SettingView, value: SettingValue) {
		if (value === s.value) delete edits[s.key];
		else edits[s.key] = value;
	}

	onMount(async () => {
		const res = await fetch('/api/settings');
		if (res.ok) load(await res.json());
		else message = { ok: false, text: `Could not load settings: HTTP ${res.status}` };
	});
</script>

<svelte:window onscroll={spy} />

<div class="grid gap-6 md:grid-cols-[200px_1fr]">
	<nav class="sticky top-20 hidden flex-col gap-0.5 self-start md:flex">
		{#each groups as g (g.slug)}
			<a
				href="#g-{g.slug}"
				class="flex items-center justify-between gap-2 rounded-md px-3 py-1.5 text-sm {(active ||
					groups[0]?.slug) === g.slug
					? 'bg-panel2 text-ink'
					: 'text-mute hover:bg-panel'}"
			>
				<span class="truncate">{g.name}</span>
				{#if g.dirty}
					<span class="rounded-full bg-warn px-1.5 text-[11px] text-[#1a1200]">{g.dirty}</span>
				{/if}
			</a>
		{/each}
	</nav>

	<div class="pb-24">
		<h1 class="text-xl font-semibold">Settings</h1>
		<p class="mt-1 mb-5 text-sm text-mute">
			Values set through environment variables are locked. Everything else is saved to the data
			volume and survives restarts. Changes apply on the next job run.
		</p>

		{#if message}
			<p
				class="mb-4 rounded-lg border px-3 py-2 text-sm {message.ok
					? 'border-ok/40 bg-ok/10 text-ok'
					: 'border-bad/40 bg-bad/10 text-bad'}"
			>
				{message.text}
			</p>
		{/if}

		{#each groups as group (group.slug)}
			<section
				id="g-{group.slug}"
				class="mb-4 scroll-mt-20 rounded-[10px] border border-line bg-panel"
			>
				<h2
					class="flex items-center justify-between border-b border-line px-4 py-3 text-xs font-semibold tracking-[0.08em] text-mute uppercase"
				>
					{group.name}
					{#if group.locked}
						<span class="text-xs font-normal tracking-normal text-warn normal-case"
							>some values locked by environment</span
						>
					{/if}
				</h2>
				{#each group.items as s (s.key)}
					{@const current = s.key in edits ? edits[s.key] : s.value}
					<div
						class="grid items-center gap-2 border-t border-line px-4 py-3 first-of-type:border-t-0 md:grid-cols-[1fr_15rem] md:gap-4 {s.key in
						edits
							? 'bg-[#1a2748] shadow-[inset_3px_0_var(--color-accent)]'
							: ''}"
					>
						<div>
							<label for={s.key} class="text-sm font-medium">{s.label}</label>
							<p class="text-xs text-mute">{s.description}</p>
							<p class="mt-1 font-mono text-[11px] text-faint">
								{s.env}
								{#if s.locked}
									· <span class="text-warn">🔒 locked by environment</span>
								{:else if s.source === 'ui'}
									· saved from UI ·
									<button class="text-accent underline" onclick={() => reset(s.key)}
										>reset to default</button
									>
								{/if}
							</p>
						</div>
						<div class="flex md:justify-end">
							{#if s.type === 'boolean'}
								<input
									id={s.key}
									type="checkbox"
									class="relative h-[22px] w-[38px] shrink-0 cursor-pointer appearance-none rounded-full border border-line bg-panel2 transition after:absolute after:top-0.5 after:left-0.5 after:size-4 after:rounded-full after:bg-mute after:transition-all checked:border-accent checked:bg-[#244a8a] checked:after:left-[18px] checked:after:bg-white disabled:cursor-not-allowed disabled:opacity-45"
									checked={current as boolean}
									disabled={s.locked}
									onchange={(e) => edit(s, e.currentTarget.checked)}
								/>
							{:else if s.type === 'number'}
								<input
									id={s.key}
									type="number"
									class="w-full rounded-md border border-line bg-bg px-2.5 py-1.5 text-sm focus:outline-2 focus:-outline-offset-1 focus:outline-accent disabled:cursor-not-allowed disabled:opacity-50"
									value={current as number}
									disabled={s.locked}
									oninput={(e) => edit(s, e.currentTarget.valueAsNumber)}
								/>
							{:else}
								<input
									id={s.key}
									type={s.secret ? 'password' : 'text'}
									autocomplete="off"
									class="w-full rounded-md border border-line bg-bg px-2.5 py-1.5 text-sm focus:outline-2 focus:-outline-offset-1 focus:outline-accent disabled:cursor-not-allowed disabled:opacity-50"
									value={current as string}
									disabled={s.locked}
									oninput={(e) => edit(s, e.currentTarget.value)}
								/>
							{/if}
						</div>
					</div>
				{/each}
				{#if group.name === 'Sonarr' || group.name === 'Radarr'}
					{@const app = group.name.toLowerCase()}
					{@const t = tests[app]}
					<div class="flex items-center gap-3 border-t border-line px-4 py-3">
						<button
							class="rounded-md border border-line bg-panel2 px-3 py-1 text-sm hover:border-accent disabled:opacity-40"
							disabled={t?.busy}
							onclick={() => testConnection(app)}>{t?.busy ? 'Testing…' : 'Test connection'}</button
						>
						{#if t?.text}
							<span class="text-sm {t.ok ? 'text-ok' : 'text-bad'}">{t.text}</span>
						{/if}
					</div>
				{/if}
			</section>
		{/each}
	</div>
</div>

<div
	class="fixed inset-x-0 bottom-0 z-10 flex items-center justify-center gap-3 border-t border-line bg-panel px-6 py-3 transition-transform {dirtyCount >
	0
		? ''
		: 'translate-y-full'}"
	inert={dirtyCount === 0}
>
	<span class="text-sm text-warn">{dirtyCount} unsaved change{dirtyCount === 1 ? '' : 's'}</span>
	<button
		class="rounded-md border border-line bg-panel2 px-3 py-1.5 text-sm hover:border-accent"
		disabled={saving}
		onclick={() => (edits = {})}>Discard</button
	>
	<button
		class="rounded-md bg-accent px-4 py-1.5 text-sm font-semibold text-[#06122b] disabled:opacity-40"
		disabled={saving}
		onclick={save}>Save changes</button
	>
</div>
