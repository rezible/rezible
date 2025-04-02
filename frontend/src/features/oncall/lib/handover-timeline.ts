import type { OncallEvent, OncallShiftAnnotation } from "$lib/api";
import { getLocalTimeZone, parseAbsolute, ZonedDateTime } from "@internationalized/date";

import { differenceInMinutes } from "date-fns";

export type ShiftTimelineNode = {
	height: number;
	timestamp: ZonedDateTime;
	event: OncallEvent;
};

const getIntervalHeight = (e1: Date, e2: Date) => {
	const diff = differenceInMinutes(e1, e2);
	if (diff < 60) return 80;
	if (diff < 60 * 24) return 160;
	return 240;
};

export const createTimeline = (
	currIds: Set<string>,
	events?: OncallEvent[],
): ShiftTimelineNode[] => {
	if (!events) return [];

	let timeline: ShiftTimelineNode[] = [];

	const tz = getLocalTimeZone();

	let nextDate: ZonedDateTime | undefined = undefined;
	for (let i = 0; i < events.length; i++) {
		const event = events[i];
		if (currIds.has(event.id)) continue;
		// let node: TimelineNode = {
		// 	event,
		// 	height: 80
		// }
		// if (i < sorted.length - 1) {
		// 	const diff = differenceInMinutes(event.occurredAt, sorted[i + 1].timestamp);
		// 	node.height = getIntervalHeight(diff);
		// }
		// if (event.kind === "incident") href = `/incidents/${event.eventId}`;
		const timestamp = nextDate || parseAbsolute(event.timestamp, tz);
		let height = 80;
		if (i < events.length - 1) {
			nextDate = parseAbsolute(events[i + 1].timestamp, tz);
			height = getIntervalHeight(timestamp.toDate(), nextDate.toDate());
		}
		timeline.push({ event, height, timestamp });
	}

	return timeline;
};
