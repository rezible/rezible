import type { Component } from "svelte";
import { resolve } from "$app/paths";
import type { ActivityRecord, InboxItem, Incident, Situation } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { formatDuration, formatTime } from "$lib/time";
import { investigationHref } from "$features/situations/lib/routes";
import { situationStatus } from "$features/situations/lib/status";
import { incidentServiceImpacts } from "$features/incidents/lib/impacts";
import type { SituationConclusion } from "$features/situations/lib/model";
import { incidentResponseStatus, incidentSeverityStatus } from "$features/incidents/lib/status";

import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";
import RiChatQuoteLine from "remixicon-svelte/icons/chat-quote-line";
import RiFileTextLine from "remixicon-svelte/icons/file-text-line";
import RiInboxLine from "remixicon-svelte/icons/inbox-line";
import RiQuestionLine from "remixicon-svelte/icons/question-line";
import RiSearchLine from "remixicon-svelte/icons/search-line";
import RiTaskLine from "remixicon-svelte/icons/task-line";
import RiTimeLine from "remixicon-svelte/icons/time-line";

type InboxKind = InboxItem["attributes"]["kind"];
type ActivityKind = ActivityRecord["attributes"]["recordKind"];

export type HomeInboxRow = {
	id: string;
	icon: Component;
	title: string;
	reason: string;
	context: string | undefined;
	href: string | undefined;
	actionLabel: string | undefined;
	occurredAt: string;
	due: StatusPresentation | undefined;
};

export type HomeIncidentRow = {
	id: string;
	title: string;
	summary: string;
	href: string;
	severity: StatusPresentation;
	response: StatusPresentation;
	accent: "danger" | "warning" | "none";
	openedAt: string;
	openFor: string | undefined;
	updatedAt: string;
	services: string[];
	moreServices: number;
};

export type HomeSituationRow = {
	id: string;
	title: string;
	summary: string;
	href: string;
	status: StatusPresentation;
	openedAt: string;
	sourceCount: number;
	linkedIncidentCount: number;
	conclusion: SituationConclusion;
};

export type HomeActivityRow = {
	id: string;
	icon: Component;
	title: string;
	explanation: string;
	href: string | undefined;
	occurredAt: string;
	day: "Today" | "Earlier";
};

const inboxActions = {
	question: "Answer",
	annotation: "Add context",
	task: "Open task",
	maintenance: "Review",
} satisfies Record<InboxKind, string>;

const inboxIcons = {
	question: RiQuestionLine,
	annotation: RiChatQuoteLine,
	task: RiTaskLine,
	maintenance: RiFileTextLine,
} satisfies Record<InboxKind, Component>;

const activityIcons = {
	"incident-update": RiAlarmWarningLine,
	"situation-investigation": RiSearchLine,
	"inbox-item": RiInboxLine,
} satisfies Record<ActivityKind, Component>;

const MAX_SERVICES = 3;

function incidentHref(slugOrId: string, view?: "report") {
	return resolve("/incidents/[slug]/[[view=incidentView]]", { slug: slugOrId, view });
}

export function inboxItemHref(item: InboxItem): string | undefined {
	const attributes = item.attributes;

	if (attributes.targetKind === "task" && attributes.targetId) {
		return resolve("/tasks/[id]", { id: attributes.targetId });
	}
	if (attributes.targetKind === "normalized-event" && attributes.targetId) {
		return resolve("/events/[id]", { id: attributes.targetId });
	}
	if (attributes.targetKind === "discussion-thread" && attributes.incidentId) {
		return incidentHref(attributes.incidentId, "report");
	}
	return undefined;
}

export function activityHref(item: ActivityRecord): string | undefined {
	const attributes = item.attributes;

	switch (attributes.recordKind) {
		case "incident-update":
			if (attributes.incidentId) {
				return incidentHref(attributes.incidentId);
			}
			return undefined;
		case "situation-investigation":
			if (attributes.situationId) {
				return investigationHref(attributes.situationId);
			}
			return undefined;
		default:
			return undefined;
	}
}

function dueStatus(dueAt: string | undefined, now: number): StatusPresentation | undefined {
	const due = formatTime(dueAt, now);
	if (due.epochMs === undefined) {
		return undefined;
	}
	if (due.epochMs < now) {
		return { label: "Overdue", tone: "danger", icon: RiTimeLine, description: due.absolute };
	}
	return { label: `Due ${due.relative}`, tone: "neutral", icon: RiTimeLine };
}

export function inboxRow(item: InboxItem, now = Date.now()): HomeInboxRow {
	const attributes = item.attributes;
	const href = inboxItemHref(item);
	const context = attributes.context.trim();

	let actionLabel: string | undefined;
	if (href) {
		actionLabel = inboxActions[attributes.kind];
	}

	return {
		id: item.id,
		icon: inboxIcons[attributes.kind],
		title: attributes.title,
		reason: attributes.reason,
		context: context || undefined,
		href,
		actionLabel,
		occurredAt: attributes.occurredAt,
		due: dueStatus(attributes.dueAt, now),
	};
}

function serviceNames(incident: Incident) {
	const names = incidentServiceImpacts(incident.attributes).map((service) => service.name);
	return [...new Set(names)];
}

function severityAccent(incident: Incident): HomeIncidentRow["accent"] {
	switch (incidentSeverityStatus(incident.attributes.severity).tone) {
		case "danger":
			return "danger";
		case "warning":
			return "warning";
		default:
			return "none";
	}
}

export function incidentRow(incident: Incident): HomeIncidentRow {
	const attributes = incident.attributes;
	const names = serviceNames(incident);

	return {
		id: incident.id,
		title: attributes.title,
		summary: attributes.summary.trim(),
		href: incidentHref(attributes.slug),
		severity: incidentSeverityStatus(attributes.severity),
		response: incidentResponseStatus(attributes.responseState),
		accent: severityAccent(incident),
		openedAt: attributes.openedAt,
		openFor: formatDuration(attributes.openedAt, attributes.resolvedAt),
		updatedAt: attributes.updatedAt,
		services: names.slice(0, MAX_SERVICES),
		moreServices: Math.max(0, names.length - MAX_SERVICES),
	};
}

function sourceCount(situation: Situation) {
	const keys = new Set<string>();
	for (const group of situation.attributes.observationGroups) {
		for (const event of group.attributes.events) {
			keys.add(`event:${event.id}`);
		}
		for (const episode of group.attributes.alertEpisodes) {
			keys.add(`alert-episode:${episode.id}`);
		}
	}
	return keys.size;
}

export function situationRow(situation: Situation, conclusion: SituationConclusion): HomeSituationRow {
	const attributes = situation.attributes;
	return {
		id: situation.id,
		title: attributes.title,
		summary: attributes.summary.trim(),
		href: resolve("/situations/[id]/[[view=situationView]]", { id: situation.id }),
		status: situationStatus(attributes),
		openedAt: attributes.openedAt,
		sourceCount: sourceCount(situation),
		linkedIncidentCount: attributes.linkedIncidentIds.length,
		conclusion,
	};
}

function isSameLocalDay(value: string, now: number) {
	const epochMs = formatTime(value, now).epochMs;
	if (epochMs === undefined) {
		return false;
	}
	return new Date(epochMs).toDateString() === new Date(now).toDateString();
}

export function activityRow(item: ActivityRecord, now = Date.now()): HomeActivityRow {
	const attributes = item.attributes;
	let day: HomeActivityRow["day"] = "Earlier";
	if (isSameLocalDay(attributes.occurredAt, now)) {
		day = "Today";
	}

	return {
		id: item.id,
		icon: activityIcons[attributes.recordKind],
		title: attributes.title,
		explanation: attributes.explanation,
		href: activityHref(item),
		occurredAt: attributes.occurredAt,
		day,
	};
}
