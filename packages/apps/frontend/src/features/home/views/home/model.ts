import type { ActivityRecord, InboxItem } from "$lib/api";

export const inboxActions = {
	question: "Answer",
	annotation: "Add context",
	task: "Open task",
	maintenance: "Review",
} satisfies Record<InboxItem["attributes"]["kind"], string>;

export function activityHref(item: ActivityRecord) {
	const attrs = item.attributes;
	switch (attrs.recordKind) {
		case "incident-update":
			return attrs.incidentId ? `/incidents/${attrs.incidentId}#update-${attrs.recordId}` : undefined;
		case "situation-investigation":
			return attrs.situationId
				? `/situations/${attrs.situationId}#investigation-${attrs.recordId}`
				: undefined;
		case "inbox-item":
			return "/";
	}
}

export function timestamp(value?: string, now = Date.now()) {
	const date = value ? new Date(value) : undefined;
	if (!date || !Number.isFinite(date.getTime()) || date.getUTCFullYear() < 1900) {
		return {
			iso: undefined,
			full: "Time unavailable",
			clock: "—",
			relative: "Time unavailable",
			day: "Earlier",
		};
	}
	const seconds = (date.getTime() - now) / 1000;
	const [unit, divisor]: [Intl.RelativeTimeFormatUnit, number] =
		Math.abs(seconds) < 3600
			? ["minute", 60]
			: Math.abs(seconds) < 86400
				? ["hour", 3600]
				: ["day", 86400];
	return {
		iso: date.toISOString(),
		full: date.toLocaleString(),
		clock: date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
		relative: new Intl.RelativeTimeFormat(undefined, { numeric: "auto" }).format(
			Math.round(seconds / divisor),
			unit
		),
		day: date.toDateString() === new Date(now).toDateString() ? "Today" : "Earlier",
	};
}
