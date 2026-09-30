export type FormattedTime = {
	/** Undefined when the input is missing or invalid. */
	epochMs?: number;
	iso?: string;
	/** Locale-formatted date and time with a short timezone, or "Time unavailable". */
	absolute: string;
	/** Hours and minutes, or "—". */
	clock: string;
	/** "just now", "5 minutes ago", an absolute date beyond 30 days, or "Time unavailable". */
	relative: string;
};

const SECOND_MS = 1000;
const MINUTE_MS = 60 * SECOND_MS;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;

const unavailableTime: FormattedTime = {
	absolute: "Time unavailable",
	clock: "—",
	relative: "Time unavailable",
};

function parseTime(value: string | null | undefined) {
	if (!value) {
		return undefined;
	}

	const date = new Date(value);
	if (!Number.isFinite(date.getTime()) || date.getUTCFullYear() < 1900) {
		return undefined;
	}

	return date;
}

function relativeLabel(date: Date, now: number) {
	const delta = date.getTime() - now;
	const distance = Math.abs(delta);

	if (distance < 45 * SECOND_MS) {
		return "just now";
	}

	if (distance >= 30 * DAY_MS) {
		return new Intl.DateTimeFormat(undefined, { dateStyle: "medium" }).format(date);
	}

	const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
	if (distance < HOUR_MS) {
		return formatter.format(Math.round(delta / MINUTE_MS), "minute");
	}
	if (distance < DAY_MS) {
		return formatter.format(Math.round(delta / HOUR_MS), "hour");
	}
	return formatter.format(Math.round(delta / DAY_MS), "day");
}

export function isValidTime(value: string | null | undefined): value is string {
	return parseTime(value) !== undefined;
}

export function formatTime(value: string | null | undefined, now = Date.now()): FormattedTime {
	const date = parseTime(value);
	if (!date) {
		return unavailableTime;
	}

	const absolute = new Intl.DateTimeFormat(undefined, {
		year: "numeric",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		timeZoneName: "short",
	}).format(date);

	return {
		epochMs: date.getTime(),
		iso: date.toISOString(),
		absolute,
		clock: date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
		relative: relativeLabel(date, now),
	};
}

/** "45 s", "23 min", "4 h 5 min", "3 d 2 h". Undefined when start is invalid or end < start. End defaults to now. */
export function formatDuration(
	start: string | null | undefined,
	end?: string | null,
	now = Date.now()
): string | undefined {
	const startDate = parseTime(start);
	if (!startDate) {
		return undefined;
	}

	let endMs = now;
	if (end) {
		const endDate = parseTime(end);
		if (!endDate) {
			return undefined;
		}
		endMs = endDate.getTime();
	}

	const duration = endMs - startDate.getTime();
	if (duration < 0) {
		return undefined;
	}

	if (duration < MINUTE_MS) {
		return `${Math.floor(duration / SECOND_MS)} s`;
	}
	if (duration < HOUR_MS) {
		return `${Math.floor(duration / MINUTE_MS)} min`;
	}
	if (duration < DAY_MS) {
		const hours = Math.floor(duration / HOUR_MS);
		const minutes = Math.floor((duration % HOUR_MS) / MINUTE_MS);
		if (minutes === 0) {
			return `${hours} h`;
		}
		return `${hours} h ${minutes} min`;
	}

	const days = Math.floor(duration / DAY_MS);
	const hours = Math.floor((duration % DAY_MS) / HOUR_MS);
	if (hours === 0) {
		return `${days} d`;
	}
	return `${days} d ${hours} h`;
}
