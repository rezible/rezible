import type {
	AlertEpisode,
	AlertInstanceAttributes,
	InvestigationReport,
	InvestigationReportAttributes,
	SituationObservation,
	SituationObservationGroup,
} from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { attentionStatus } from "$features/signals/lib/attention";
import { formatTime, type FormattedTime } from "$lib/time";
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
	/** A watch-only or join-only definition's signal attention. */
	attention?: StatusPresentation;
};

export type SourceTarget =
	| { kind: "direct"; record: SourceRecord; observationGroupTitle: string }
	| { kind: "knowledgeEvidence"; id: string };

function windowEndLabel(reason: AlertInstanceAttributes["endReason"]) {
	switch (reason) {
		case "resolved":
			return "Resolved";
		case "superseded":
			return "Superseded";
		case "timeout":
			return "Timed out without a notification";
		default:
			return "Ended";
	}
}

function windowDescription(window: AlertInstanceAttributes) {
	const fired = `Fired ${formatTime(window.firedAt).absolute}`;
	if (!window.endedAt) {
		return `${fired} · Still firing`;
	}
	return `${fired} · ${windowEndLabel(window.endReason)} ${formatTime(window.endedAt).absolute}`;
}

/** An observation's record; its alert episode, when loaded, adds its definition and firing windows. */
function observationSourceRecord(
	observation: SituationObservation,
	groupTitle: string,
	episode?: AlertEpisode
): SourceRecord {
	const record: SourceRecord = {
		key: `observation:${observation.id}`,
		id: observation.alertEpisodeId ?? observation.id,
		type: "Alert episode",
		title: groupTitle,
		source: "Alert episode",
		time: formatTime(observation.attachedAt),
		timeLabel: "Attached",
		content: "",
		fields: [],
		links: [],
	};
	if (observation.matchExplanation) {
		record.fields.push({ label: "Attached because", value: observation.matchExplanation });
	}
	if (!episode) {
		return record;
	}

	const attributes = episode.attributes;
	const definition = attributes.definition;
	if (definition) {
		record.title = definition.attributes.title;
		record.definitionId = definition.id;
		record.attention = attentionStatus(definition.attributes.situationSignalAttention);
	}
	record.time = formatTime(attributes.startedAt);
	record.timeLabel = "Started";
	record.content = attributes.instances.at(-1)?.attributes.summary ?? "";
	record.fields.push({ label: "Severity", value: attributes.highestSeverity });
	attributes.instances.forEach((instance, index) => {
		record.fields.push({ label: `Window ${index + 1}`, value: windowDescription(instance.attributes) });
	});
	return record;
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

const UNGROUPED_OBSERVATIONS_TITLE = "Ungrouped observations";

/**
 * Groups with their records, enriched by the situation's alert episodes keyed by ID. Ungrouped observations
 * come first, as a group of their own.
 */
export function observationGroups(
	observations: SituationObservation[],
	groups: SituationObservationGroup[],
	episodesById: Map<string, AlertEpisode>
) {
	const records = (groupId: string | undefined, title: string) =>
		observations
			.filter((observation) => observation.observationGroupId === groupId)
			.map((observation) =>
				observationSourceRecord(
					observation,
					title,
					observation.alertEpisodeId ? episodesById.get(observation.alertEpisodeId) : undefined
				)
			);
	const grouped = groups.map((group) => ({
		id: group.id,
		title: group.attributes.title,
		body: group.attributes.body,
		records: records(group.id, group.attributes.title),
	}));
	const ungrouped = records(undefined, UNGROUPED_OBSERVATIONS_TITLE);
	if (!ungrouped.length) {
		return grouped;
	}
	return [
		{ id: "ungrouped", title: UNGROUPED_OBSERVATIONS_TITLE, body: undefined, records: ungrouped },
		...grouped,
	];
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
	/** A null report means none has been published. */
	data?: { data: InvestigationReport | null };
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
