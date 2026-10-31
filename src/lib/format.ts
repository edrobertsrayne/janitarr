const UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

export function bytes(n: number): string {
	let v = n;
	let i = 0;
	while (v >= 1024 && i < UNITS.length - 1) {
		v /= 1024;
		i++;
	}
	return `${v >= 100 || i === 0 ? v.toFixed(0) : v.toFixed(1)} ${UNITS[i]}`;
}

/** "45s" / "5m" / "3h" / "2d". */
export function duration(ms: number): string {
	const s = Math.round(Math.abs(ms) / 1000);
	return s < 60
		? `${s}s`
		: s < 3600
			? `${Math.round(s / 60)}m`
			: s < 86400
				? `${Math.round(s / 3600)}h`
				: `${Math.round(s / 86400)}d`;
}

/** "5m ago" / "in 5m". */
export function relative(ts: number, now: number): string {
	const diff = ts - now;
	const text = duration(diff);
	return diff < 0 ? `${text} ago` : `in ${text}`;
}

export function clock(ts: number): string {
	return new Date(ts).toLocaleString(undefined, {
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit'
	});
}
