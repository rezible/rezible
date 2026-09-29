import type { AlertEpisode, Event, InvestigationAttributes, SituationObservationGroup } from "$lib/api";

export function timestamp(value?: string) {
	if (!value) {
		return { value: undefined, iso: undefined, label: "Time unavailable" };
	}
	const date = new Date(value);
	if (!Number.isFinite(date.getTime()) || date.getUTCFullYear() <= 1) {
		return { value: undefined, iso: undefined, label: "Time unavailable" };
	}
	return {
		value: date.getTime(),
		iso: date.toISOString(),
		label: date.toLocaleString(undefined, { timeZoneName: "short" }),
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

function eventContent(payload: ReturnType<typeof readablePayload>) {
	const message = readableField(payload?.fields, "message");
	if (message) {
		return message;
	}
	const description = readableField(payload?.fields, "description");
	if (description) {
		return description;
	}
	if (payload?.fields) {
		return JSON.stringify(payload.fields, null, 2);
	}
	return payload?.decoded ?? "";
}

function eventTitle(fields: Record<string, unknown> | undefined, kind: string) {
	const title = readableField(fields, "title");
	if (title) {
		return title;
	}
	const message = readableField(fields, "message");
	if (message) {
		return message;
	}
	return kind || "Event";
}

function eventSourceRecord(record: Event): SourceRecord {
	const attributes = record.attributes;
	const payload = readablePayload(attributes.attributes);
	const fields = payload?.fields;
	const content = eventContent(payload);
	const title = eventTitle(fields, attributes.kind);
	const links: SourceRecord["links"] = [];
	const references: [string, string][] = [
		["Provider event reference", attributes.providerEventRef],
		["Provider event source", attributes.providerEventSource],
		["Resource reference", attributes.resourceRef.resourceRef],
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
		title,
		source: attributes.resourceRef.provider || attributes.providerEventSource || "Source unavailable",
		time: timestamp(attributes.occurredAt),
		timeLabel: "Occurred",
		content,
		fields: [
			{ label: "Kind", value: attributes.kind },
			{ label: "Provider", value: attributes.resourceRef.provider },
			{ label: "Provider namespace", value: attributes.resourceRef.providerNamespace },
			{ label: "Resource reference", value: attributes.resourceRef.resourceRef },
			{ label: "Provider event source", value: attributes.providerEventSource },
			{ label: "Provider event reference", value: attributes.providerEventRef },
			{ label: "Integration ID", value: attributes.integrationId ?? "" },
			{ label: "Received", value: timestamp(attributes.receivedAt).label },
		].filter((field) => field.value),
		links,
	};
}

function episodeSourceRecord(record: AlertEpisode): SourceRecord {
	const attributes = record.attributes;
	const fields = [
		{ label: "Status", value: attributes.status },
		{ label: "Last observed", value: timestamp(attributes.lastObservedAt).label },
	];
	if (attributes.closedAt) {
		fields.push({ label: "Closed", value: timestamp(attributes.closedAt).label });
	}
	fields.push({ label: "Definition", value: attributes.definition?.attributes.definition ?? "" });
	return {
		key: `alert-episode:${record.id}`,
		id: record.id,
		type: "Alert episode",
		title: attributes.definition?.attributes.title || "Alert episode",
		source: "Alert episode",
		time: timestamp(attributes.startedAt),
		timeLabel: "Started",
		content: attributes.definition?.attributes.description || "",
		definitionId: attributes.definition?.id,
		fields: fields.filter((field) => field.value),
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

function compareSources(first: SourceRecord, second: SourceRecord) {
	const firstTime = first.time.value ?? Infinity;
	const secondTime = second.time.value ?? Infinity;
	if (firstTime < secondTime) {
		return -1;
	}
	if (firstTime > secondTime) {
		return 1;
	}
	return first.key.localeCompare(second.key);
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
			records: [...records.values()].sort(compareSources),
		};
	});
}

export type InvestigationExecution = {
	label: string;
	message?: string;
};

export function investigationExecution(attributes?: InvestigationAttributes): InvestigationExecution {
	if (attributes?.activeTurn?.status === "queued") {
		return { label: "Queued" };
	}
	if (attributes?.activeTurn?.status === "running") {
		return { label: "Running" };
	}

	switch (attributes?.latestTurn?.status) {
		case "completed":
			return { label: "Completed" };
		case "failed":
			return {
				label: "Failed",
				message: "The investigation turn could not finish.",
			};
		case "aborted":
			return {
				label: "Aborted",
				message: "The investigation turn was stopped.",
			};
		default:
			return { label: "Waiting to start" };
	}
}

export const SITUATION_POLL_INTERVAL_MS = 30_000;

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
