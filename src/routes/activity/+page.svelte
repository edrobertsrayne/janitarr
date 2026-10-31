<script lang="ts">
	import ActivityLog from '#lib/components/ActivityLog.svelte';
	import { status } from '#lib/status.svelte.js';
</script>

<h1 class="sr-only">Activity</h1>

{#if status.error}
	<p class="mb-4 rounded-lg border border-bad/40 bg-bad/10 px-3 py-2 text-sm text-bad">
		{status.error}
	</p>
{/if}

{#if status.data}
	<ActivityLog entries={status.data.activity} />
{:else if !status.error}
	<p class="text-sm text-mute">Loading…</p>
{/if}
