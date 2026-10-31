import type { JobName, StatusPayload, StrikeRow } from '#lib/types.js';

const POLL_MS = 10_000;

async function errorOf(res: Response): Promise<string> {
	return (await res.json().catch(() => null))?.message ?? `HTTP ${res.status}`;
}

/** Polled dashboard status shared by the layout (top bar) and the dashboard page. */
class StatusStore {
	data = $state<StatusPayload | null>(null);
	error = $state<string | null>(null);
	busy = $state(false);

	async refresh() {
		try {
			const res = await fetch('/api/status');
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			this.data = await res.json();
			this.error = null;
		} catch (e) {
			this.error = `Could not load status: ${(e as Error).message}`;
		}
	}

	/** Polls until the returned function is called. */
	start(): () => void {
		this.refresh();
		const timer = setInterval(() => this.refresh(), POLL_MS);
		return () => clearInterval(timer);
	}

	private async post(url: string, body: unknown = {}) {
		// JSON content type keeps SvelteKit's CSRF check from treating this as a form post.
		const res = await fetch(url, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(body)
		});
		if (!res.ok) this.error = await errorOf(res);
		await this.refresh();
	}

	runJob = (job: JobName) => this.post(`/api/run/${job}`);
	clearStrikes = (s: StrikeRow) =>
		this.post('/api/strikes/clear', { app: s.app, itemId: s.itemId });

	async toggleDryRun() {
		if (!this.data) return;
		this.busy = true;
		const res = await fetch('/api/settings', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ values: { 'general.dryRun': !this.data.dryRun } })
		});
		if (!res.ok) this.error = await errorOf(res);
		this.busy = false;
		await this.refresh();
	}
}

export const status = new StatusStore();
