import type { IncidentMilestone, IncidentMilestoneAttributes } from "$src/lib/api";
import { parseAbsolute, parseAbsoluteToLocal, type ZonedDateTime } from "@internationalized/date";
import { mdiAlertDecagram, mdiAccountAlert, mdiAccountEye, mdiFireExtinguisher, mdiTimelineClock } from "@mdi/js";

const kindOrder = ["impact", "detection", "investigation", "mitigation", "resolution"] as const;

type IncidentMilestoneKind = IncidentMilestoneAttributes["kind"];
export const getIconForIncidentMilestoneKind = (kind: IncidentMilestoneKind) => {
	switch (kind) {
		case "impact": return mdiAlertDecagram;
		case "detection": return mdiAccountAlert;
		case "investigation": return mdiAccountEye;
		case "mitigation": return mdiFireExtinguisher;
		case "resolution": return mdiTimelineClock;
	}
}

export const orderedMilestones = (ms: IncidentMilestone[]): IncidentMilestone[] => {
	return ms.toSorted((a, b) => {
		const aKindIdx = kindOrder.indexOf(a.attributes.kind);
		const bKindIdx = kindOrder.indexOf(b.attributes.kind);
		if (aKindIdx !== bKindIdx) {
			return aKindIdx - bKindIdx;
		}
		return parseAbsoluteToLocal(a.attributes.timestamp).compare(parseAbsoluteToLocal(b.attributes.timestamp))
	})
}

export const getPreviousOrderedMilestone = (kind: IncidentMilestoneKind, others: IncidentMilestone[], tz: string) => {
	const kindIndex = kindOrder.indexOf(kind);
	if (kindIndex === 0) return null;

	let earliestIdx = -1;
	let earliest: ZonedDateTime | undefined = undefined;
	for (const milestone of others) {
		const idx = kindOrder.indexOf(milestone.attributes.kind);
		if (kindIndex > idx) {
			const time = parseAbsolute(milestone.attributes.timestamp, tz);
			if (!earliest || time.compare(earliest) < 0) {
				earliest = time;
				earliestIdx = idx;
			}
		}
	}

	return earliestIdx >= 0 ? others[earliestIdx] : undefined;
}

export const getNextOrderedMilestone = (kind: IncidentMilestoneKind, others: IncidentMilestone[], tz: string) => {
	const kindIndex = kindOrder.indexOf(kind);
	if (kindIndex === kindOrder.length - 1) return undefined;

	let latestIdx = -1;
	let latest: ZonedDateTime | undefined = undefined;
	for (const milestone of others) {
		const idx = kindOrder.indexOf(milestone.attributes.kind);
		if (kindIndex < idx) {
			const time = parseAbsolute(milestone.attributes.timestamp, tz);
			if (!latest || time.compare(latest) > 0) {
				latest = time;
				latestIdx = idx;
			}
		}
	}
	return latestIdx >= 0 ? others[latestIdx] : undefined;
}

// check if the time is valid for the milestone kind in the context of the other milestones
export const isNewMilestoneTimeValid = (kind: IncidentMilestoneKind, time: ZonedDateTime, others: IncidentMilestone[]) => {
	const tz = time.timeZone;
	const prevMs = getPreviousOrderedMilestone(kind, others, tz);
	if (prevMs && time.compare(parseAbsolute(prevMs.attributes.timestamp, tz)) < 0) return false;

	const nextMs = getNextOrderedMilestone(kind, others, tz);
	if (nextMs && time.compare(parseAbsolute(nextMs.attributes.timestamp, tz)) > 0) return false;

	return true;
}