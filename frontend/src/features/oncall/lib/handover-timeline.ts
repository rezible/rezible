import type {
	OncallAlert,
	Incident,
	OncallShiftAnnotation,
	CreateOncallShiftAnnotationRequestAttributes,
} from "$lib/api";
import { getLocalTimeZone, parseAbsolute, parseAbsoluteToLocal, ZonedDateTime } from "@internationalized/date";
import { mdiChat, mdiCircleMedium, mdiFire, mdiPhoneAlert, mdiSlack } from "@mdi/js";
import { differenceInMinutes } from "date-fns";

export const eventKindIcons: Record<ShiftEventKind, string> = {
	["incident"]: mdiFire,
	["alert"]: mdiPhoneAlert,
	["ping"]: mdiSlack,
	["toil"]: mdiCircleMedium,
};

export type ShiftTimelineNode = {
	height: number;
	event: ShiftTimelineEvent;
};

export type ShiftEventKind = CreateOncallShiftAnnotationRequestAttributes["eventKind"];

export type ShiftTimelineEvent = {
	eventId: string;
	kind: ShiftEventKind;
	title: string;
	description?: string;
	occurredAt: ZonedDateTime;
	notes?: string;
};

type MergedEvent = {
	timestamp: ZonedDateTime;
	incident?: Incident;
	alert?: OncallAlert;
	annotation?: OncallShiftAnnotation;
};

const convertMergedEvent = (e: MergedEvent): ShiftTimelineEvent => {
	if (e.incident) {
		const attr = e.incident.attributes;
		return {
			eventId: e.incident.id,
			kind: "incident",
			title: attr.title,
			description: attr.summary,
			occurredAt: e.timestamp,
		};
	}
	if (e.alert) {
		return {
			eventId: e.alert.id,
			kind: "alert",
			title: e.alert.attributes.title,
			occurredAt: e.timestamp,
		};
	}
	if (e.annotation) {
		const attr = e.annotation.attributes;
		return {
			eventId: e.annotation.id,
			kind: "toil",
			title: "annotation title",
			occurredAt: e.timestamp,
		};
	}
	throw new Error("invalid event type");
};

const getIntervalHeight = (e1: MergedEvent, e2: MergedEvent) => {
	const diff = differenceInMinutes(e1.timestamp.toDate(), e2.timestamp.toDate());
	if (diff < 60) return 80;
	if (diff < 60 * 24) return 160;
	return 240;
};

export const createTimeline = (
	currIds: Set<string>,
	incidents?: Incident[],
	alerts?: OncallAlert[]
): ShiftTimelineNode[] => {
	if (!incidents || !alerts) return [];

	const timeline: ShiftTimelineNode[] = [];

	const tz = getLocalTimeZone();

	const merged: MergedEvent[] = [];
	incidents.forEach(incident => {
		merged.push({ timestamp: parseAbsolute(incident.attributes.openedAt, tz), incident })
	});
	alerts.forEach(alert => {
		merged.push({ timestamp: parseAbsolute(alert.attributes.occurredAt, tz), alert })
	});
	// annotations.forEach(annotation => merged.push({timestamp: Date.parse(annotation.attributes.occurredAt), annotation}));

	const sorted = merged.toSorted((a, b) => (a.timestamp < b.timestamp ? 1 : -1));

	for (let i = 0; i < sorted.length; i++) {
		const event = convertMergedEvent(sorted[i]);
		if (currIds.has(event.eventId)) continue;
		// let node: TimelineNode = {
		// 	event,
		// 	height: 80
		// }
		// if (i < sorted.length - 1) {
		// 	const diff = differenceInMinutes(event.occurredAt, sorted[i + 1].timestamp);
		// 	node.height = getIntervalHeight(diff);
		// }
		// if (event.kind === "incident") href = `/incidents/${event.eventId}`;
		let height = 80;
		if (i < sorted.length - 1) height = getIntervalHeight(sorted[i], sorted[i + 1]);
		timeline.push({ event, height });
	}

	return timeline;
};
