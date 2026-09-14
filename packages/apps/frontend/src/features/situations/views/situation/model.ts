import type { AlertEpisode, Event, SituationInvestigation, SituationObservationGroup } from "$lib/api";

export function investigationSearch(search: string, investigationId: string) {
	const params = new URLSearchParams(search);
	params.set("investigation", investigationId);
	return `?${params}`;
}

export function timestamp(value?: string) {
	const date = value ? new Date(value) : undefined;
	const usable = date && Number.isFinite(date.getTime()) && date.getUTCFullYear() > 1;
	return {
		value: usable ? date.getTime() : undefined,
		iso: usable ? date.toISOString() : undefined,
		label: usable ? date.toLocaleString(undefined, { timeZoneName: "short" }) : "Time unavailable",
	};
}

export function latestInvestigation(items: SituationInvestigation[], reportOnly = false) {
	return items
		.filter((item) => !reportOnly || item.attributes.report)
		.sort((a, b) => {
			const aTime = timestamp(a.attributes.updatedAt).value ?? -Infinity;
			const bTime = timestamp(b.attributes.updatedAt).value ?? -Infinity;
			return (aTime === bTime ? 0 : aTime > bTime ? -1 : 1) || a.id.localeCompare(b.id);
		})[0];
}

export function selectInvestigation(
	explicitId: string | null,
	retainedId: string | undefined,
	items: SituationInvestigation[]
) {
	return explicitId || retainedId || latestInvestigation(items, true)?.id || latestInvestigation(items)?.id;
}

export function evidenceChanged(revision: number, investigation?: SituationInvestigation) {
	return !!investigation?.attributes.report && revision > investigation.attributes.completedRevision;
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

export function sourceRecord(record: Event | AlertEpisode): SourceRecord {
	const attrs = record.attributes;
	if ("occurredAt" in attrs) {
		let content = "";
		try {
			const bytes = Uint8Array.from(atob(attrs.attributes), (character) => character.charCodeAt(0));
			const decoded = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
			// Preserve readable source text, including tabs/newlines, but never display binary controls.
			if (
				![...decoded].some((character) => {
					const code = character.codePointAt(0)!;
					return (code < 32 && ![9, 10, 13].includes(code)) || (code >= 127 && code <= 159);
				})
			)
				content = decoded;
		} catch {
			// Invalid base64 or non-UTF-8 content uses the existing unavailable-content presentation.
		}
		const links: SourceRecord["links"] = [];
		for (const [label, value] of [
			["Provider event reference", attrs.providerEventRef],
			["Provider event source", attrs.providerEventSource],
			["Resource reference", attrs.resourceRef.resourceRef],
		]) {
			try {
				const url = new URL(value);
				if (["https:", "http:"].includes(url.protocol)) links.push({ label, href: url.href });
			} catch {
				/* References are not necessarily URLs. */
			}
		}
		return {
			key: `event:${record.id}`,
			id: record.id,
			type: "Event",
			eventId: record.id,
			title: attrs.kind || "Event",
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
			{ label: "Definition ID", value: attrs.definition?.id ?? "" },
			{ label: "Definition", value: attrs.definition?.attributes.definition ?? "" },
		].filter((field) => field.value),
		links: [],
	};
}

export function observationGroups(groups: SituationObservationGroup[]) {
	const keys = new Set<string>();
	const items = groups.map((group) => {
		const records = new Map<string, SourceRecord>();
		for (const record of [...group.attributes.events, ...group.attributes.alertEpisodes]) {
			const source = sourceRecord(record);
			records.set(source.key, source);
			keys.add(source.key);
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
	return { items, sourceCount: keys.size };
}
