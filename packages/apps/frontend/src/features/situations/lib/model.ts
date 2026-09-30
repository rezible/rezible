import type {
	AlertEpisode,
	Event,
	InvestigationReport,
	InvestigationReportAttributes,
	SituationObservationGroup,
} from "$lib/api";
import { formatTime, type FormattedTime } from "$lib/time";
import { safeExternalUrl } from "$lib/utils";
import { markdownSummary } from "$components/rich-text-view/markdown-summary";
import type { Component } from "svelte";
import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";
import RiPulseLine from "remixicon-svelte/icons/pulse-line";

export type SourceRecord = {
	key: string;
	id: string;
	type: "Event" | "Alert episode";
	title: string;
	source: string;
	time: FormattedTime;
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
		time: formatTime(attributes.occurredAt),
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
			{ label: "Received", value: formatTime(attributes.receivedAt).absolute },
		].filter((field) => field.value),
		links,
	};
}

function episodeSourceRecord(record: AlertEpisode): SourceRecord {
	const attributes = record.attributes;
	const fields = [
		{ label: "Status", value: attributes.status },
		{ label: "Last observed", value: formatTime(attributes.lastObservedAt).absolute },
	];
	if (attributes.closedAt) {
		fields.push({ label: "Closed", value: formatTime(attributes.closedAt).absolute });
	}
	fields.push({ label: "Definition", value: attributes.definition?.attributes.definition ?? "" });
	return {
		key: `alert-episode:${record.id}`,
		id: record.id,
		type: "Alert episode",
		title: attributes.definition?.attributes.title || "Alert episode",
		source: "Alert episode",
		time: formatTime(attributes.startedAt),
		timeLabel: "Started",
		content: attributes.definition?.attributes.description || "",
		definitionId: attributes.definition?.id,
		fields: fields.filter((field) => field.value),
		links: [],
	};
}

/** "Alert episode" or "Event · datadog"; never repeats the type. */
export function sourceMeta(record: SourceRecord): string {
	if (record.source && record.source !== record.type) {
		return `${record.type} · ${record.source}`;
	}
	return record.type;
}

export function sourceIcon(record: SourceRecord): Component {
	if (record.type === "Event") {
		return RiPulseLine;
	}
	return RiAlarmWarningLine;
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
	const firstTime = first.time.epochMs ?? Infinity;
	const secondTime = second.time.epochMs ?? Infinity;
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

export const SITUATION_POLL_INTERVAL_MS = 30_000;

/** One- or two-sentence conclusion for lists, cards and headers. */
export function reportSummary(report: InvestigationReportAttributes): string | undefined {
	const summary = report.summary.trim();
	if (summary) {
		return summary;
	}
	return markdownSummary(report.text);
}

export type ReportState =
	| { kind: "loading" }
	| { kind: "none" }
	| { kind: "unavailable" }
	| { kind: "error" }
	| {
			kind: "published";
			attributes: InvestigationReportAttributes;
			summary: string | undefined;
			provisional: boolean;
			refreshFailed: boolean;
	  };

/** A report is provisional while the turn that published it has not finished. */
export function isProvisionalReport(report: InvestigationReportAttributes) {
	return report.provisional || report.turnStatus === "running";
}

export type SituationConclusion = { kind: "loading" } | { kind: "none" } | { kind: "text"; text: string };

type ReportQueryState = {
	isPending: boolean;
	isError: boolean;
	data?: { data: InvestigationReport };
};

/** Maps a report query (or its absence, when there is no investigation) to a conclusion. */
export function situationConclusion(reportQuery: ReportQueryState | undefined): SituationConclusion {
	if (!reportQuery) {
		return { kind: "none" };
	}

	const report = reportQuery.data?.data;
	if (report) {
		const text = reportSummary(report.attributes);
		if (text) {
			return { kind: "text", text };
		}
		return { kind: "none" };
	}

	if (reportQuery.isPending) {
		return { kind: "loading" };
	}

	return { kind: "none" };
}

export function isDefinitiveUnavailableError(error?: { status?: number } | null) {
	return [401, 403, 404].includes(error?.status ?? 0);
}
