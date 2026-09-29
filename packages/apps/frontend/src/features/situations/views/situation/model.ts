import type { AlertEpisode, Event, InvestigationAttributes, SituationObservationGroup } from "$lib/api";

export function timestamp(value?: string) {
	const date = value ? new Date(value) : undefined;
	const usable = date && Number.isFinite(date.getTime()) && date.getUTCFullYear() > 1;
	return {
		value: usable ? date.getTime() : undefined,
		iso: usable ? date.toISOString() : undefined,
		label: usable ? date.toLocaleString(undefined, { timeZoneName: "short" }) : "Time unavailable",
	};
}

export type SourceRecord = {
	key: string;
	id: string;
	type: "Event" | "Alert episode";
	title: string;
	source: string;
	time: ReturnType<typeof timestamp>;
	timeLabel: string;
	content: string;
	fields: { label: string; value: string }[];
	links: { label: string; href: string }[];
	eventId?: string;
	definitionId?: string;
};

export type SourceTarget =
	| { kind: "direct"; record: SourceRecord; observationGroupTitle: string }
	| { kind: "knowledgeEvidence"; id: string };

function hasBinaryControls(value: string) {
	return [...value].some((character) => {
		const code = character.codePointAt(0)!;
		return (code < 32 && ![9, 10, 13].includes(code)) || (code >= 127 && code <= 159);
	});
}

function readablePayload(value: string) {
	let decoded = "";
	try {
		const bytes = Uint8Array.from(atob(value), (character) => character.charCodeAt(0));
		decoded = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
	} catch {
		return undefined;
	}

	if (hasBinaryControls(decoded)) {
		return undefined;
	}

	try {
		const parsed: unknown = JSON.parse(decoded);
		if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
			return { decoded, fields: parsed as Record<string, unknown> };
		}
		return { decoded, fields: undefined };
	} catch {
		return { decoded, fields: undefined };
	}
}

function readableField(fields: Record<string, unknown> | undefined, name: string) {
	const value = fields?.[name];
	if (typeof value !== "string" || !value.trim() || hasBinaryControls(value)) {
		return undefined;
	}
	return value.trim();
}

function safeExternalUrl(value: string) {
	try {
		const url = new URL(value);
		if (url.protocol === "https:" || url.protocol === "http:") {
			return url.href;
		}
	} catch {
		// Stored references are not always URLs.
	}
	return undefined;
}

function eventSourceRecord(record: Event): SourceRecord {
	const attrs = record.attributes;
	const payload = readablePayload(attrs.attributes);
	const fields = payload?.fields;
	const content =
		readableField(fields, "message") ??
		readableField(fields, "description") ??
		(fields ? JSON.stringify(fields, null, 2) : payload?.decoded) ??
		"";
	const links: SourceRecord["links"] = [];
	const references: [string, string][] = [
		["Provider event reference", attrs.providerEventRef],
		["Provider event source", attrs.providerEventSource],
		["Resource reference", attrs.resourceRef.resourceRef],
	];

	for (const [label, reference] of references) {
		const href = safeExternalUrl(reference);
		if (href) {
			links.push({ label, href });
		}
	}

	return {
		key: `event:${record.id}`,
		id: record.id,
		type: "Event",
		eventId: record.id,
		title: readableField(fields, "title") ?? readableField(fields, "message") ?? (attrs.kind || "Event"),
		source: attrs.resourceRef.provider || attrs.providerEventSource || "Source unavailable",
		time: timestamp(attrs.occurredAt),
		timeLabel: "Occurred",
		content,
		fields: [
			{ label: "Kind", value: attrs.kind },
			{ label: "Provider", value: attrs.resourceRef.provider },
			{ label: "Provider namespace", value: attrs.resourceRef.providerNamespace },
			{ label: "Resource reference", value: attrs.resourceRef.resourceRef },
			{ label: "Provider event source", value: attrs.providerEventSource },
			{ label: "Provider event reference", value: attrs.providerEventRef },
			{ label: "Integration ID", value: attrs.integrationId ?? "" },
			{ label: "Received", value: timestamp(attrs.receivedAt).label },
		].filter((field) => field.value),
		links,
	};
}

function episodeSourceRecord(record: AlertEpisode): SourceRecord {
	const attrs = record.attributes;
	return {
		key: `alert-episode:${record.id}`,
		id: record.id,
		type: "Alert episode",
		title: attrs.definition?.attributes.title || "Alert episode",
		source: "Alert episode",
		time: timestamp(attrs.startedAt),
		timeLabel: "Started",
		content: attrs.definition?.attributes.description || "",
		definitionId: attrs.definition?.id,
		fields: [
			{ label: "Status", value: attrs.status },
			{ label: "Last observed", value: timestamp(attrs.lastObservedAt).label },
			...(attrs.closedAt ? [{ label: "Closed", value: timestamp(attrs.closedAt).label }] : []),
			{ label: "Definition", value: attrs.definition?.attributes.definition ?? "" },
		].filter((field) => field.value),
		links: [],
	};
}

function isEventRecord(record: Event | AlertEpisode): record is Event {
	return "occurredAt" in record.attributes;
}

export function sourceRecord(record: Event | AlertEpisode): SourceRecord {
	if (isEventRecord(record)) {
		return eventSourceRecord(record);
	}
	return episodeSourceRecord(record);
}

export function observationGroups(groups: SituationObservationGroup[]) {
	return groups.map((group) => {
		const records = new Map<string, SourceRecord>();
		for (const record of [...group.attributes.events, ...group.attributes.alertEpisodes]) {
			const source = sourceRecord(record);
			records.set(source.key, source);
		}
		return {
			id: group.id,
			title: group.attributes.title,
			body: group.attributes.body,
			records: [...records.values()].sort((a, b) => {
				const aTime = a.time.value ?? Infinity;
				const bTime = b.time.value ?? Infinity;
				return (aTime === bTime ? 0 : aTime < bTime ? -1 : 1) || a.key.localeCompare(b.key);
			}),
		};
	});
}

export type InvestigationExecution = {
	label: string;
	message?: string;
	shouldPoll: boolean;
};

export function investigationExecution(attributes?: InvestigationAttributes): InvestigationExecution {
	if (attributes?.activeTurn.status === "queued") {
		return { label: "Queued", shouldPoll: true };
	}
	if (attributes?.activeTurn.status === "running") {
		return { label: "Running", shouldPoll: true };
	}

	switch (attributes?.latestTurn.status) {
		case "completed":
			return { label: "Completed", shouldPoll: !!attributes.hasPendingWork };
		case "failed":
			return {
				label: "Failed",
				message: "The investigation turn could not finish.",
				shouldPoll: !!attributes.hasPendingWork,
			};
		case "aborted":
			return {
				label: "Aborted",
				message: "The investigation turn was stopped.",
				shouldPoll: !!attributes.hasPendingWork,
			};
		default:
			return { label: "Waiting to start", shouldPoll: true };
	}
}

export function investigationRefreshInterval(attributes?: InvestigationAttributes) {
	return investigationExecution(attributes).shouldPoll ? 2000 : 30000;
}

export function reportExcerpt(text: string, limit = 400) {
	const characters = Array.from(text);
	if (characters.length <= limit) {
		return text;
	}
	return `${characters.slice(0, limit).join("")}…`;
}

export function isDefinitiveUnavailableError(error?: { status?: number } | null) {
	return [401, 403, 404].includes(error?.status ?? 0);
}
