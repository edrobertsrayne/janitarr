<script lang="ts">
	import { onMount } from 'svelte';
	import type { SettingValue, SettingView } from '#lib/types.js';

	let settings = $state<SettingView[]>([]);
	let edits = $state<Record<string, SettingValue>>({});
	let message = $state<{ ok: boolean; text: string } | null>(null);
	let saving = $state(false);

	const groups = $derived(
		[...new Set(settings.map((s) => s.group))].map((name) => ({
			name,
			items: settings.filter((s) => s.group === name)
		}))
	);
	const dirty = $derived(Object.keys(edits).length > 0);

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

<div class="mb-6 flex items-center justify-between">
	<h1 class="text-xl font-semibold text-white">Settings</h1>
	<button
		class="rounded bg-sky-600 px-4 py-1.5 text-sm text-white hover:bg-sky-500 disabled:opacity-40"
		disabled={!dirty || saving}
		onclick={save}>Save changes</button
	>
</div>

<p class="mb-6 text-sm text-slate-400">
	Values set through environment variables are locked here. Everything else is saved to the data
	volume and survives restarts.
</p>

{#if message}
	<p
		class="mb-4 rounded border px-3 py-2 text-sm {message.ok
			? 'border-emerald-900 bg-emerald-950/50 text-emerald-300'
			: 'border-red-900 bg-red-950/50 text-red-300'}"
	>
		{message.text}
	</p>
{/if}

<div class="space-y-6">
	{#each groups as group (group.name)}
		<section class="rounded-lg border border-slate-800 bg-slate-900">
			<h2 class="border-b border-slate-800 px-4 py-3 text-base font-semibold text-white">
				{group.name}
			</h2>
			<div class="divide-y divide-slate-800">
				{#each group.items as s (s.key)}
					{@const current = s.key in edits ? edits[s.key] : s.value}
					<div class="grid items-start gap-2 px-4 py-3 md:grid-cols-[1fr_16rem]">
						<div>
							<label for={s.key} class="text-sm text-slate-100">{s.label}</label>
							<p class="text-xs text-slate-400">{s.description}</p>
							<p class="mt-0.5 text-[11px] text-slate-500">
								<code>{s.env}</code>
								{#if s.locked}
									· <span class="text-amber-300">locked by environment</span>
								{:else if s.source === 'ui'}
									· saved from UI ·
									<button class="underline hover:text-slate-300" onclick={() => reset(s.key)}
										>reset to default</button
									>
								{/if}
							</p>
						</div>
						<div class="md:text-right">
							{#if s.type === 'boolean'}
								<input
									id={s.key}
									type="checkbox"
									class="h-5 w-5 accent-sky-500 disabled:opacity-50"
									checked={current as boolean}
									disabled={s.locked}
									onchange={(e) => edit(s, e.currentTarget.checked)}
								/>
							{:else if s.type === 'number'}
								<input
									id={s.key}
									type="number"
									class="w-full rounded border border-slate-700 bg-slate-950 px-2 py-1 text-sm disabled:opacity-50"
									value={current as number}
									disabled={s.locked}
									oninput={(e) => edit(s, e.currentTarget.valueAsNumber)}
								/>
							{:else}
								<input
									id={s.key}
									type={s.secret ? 'password' : 'text'}
									autocomplete="off"
									class="w-full rounded border border-slate-700 bg-slate-950 px-2 py-1 text-sm disabled:opacity-50"
									value={current as string}
									disabled={s.locked}
									oninput={(e) => edit(s, e.currentTarget.value)}
								/>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/each}
</div>
