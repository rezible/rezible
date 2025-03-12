import type { OncallShift } from "$src/lib/api";
import type { ZonedDateTime } from "@internationalized/date";
import { differenceInCalendarDays, differenceInMinutes, isFuture, isPast } from "date-fns";

export type ShiftStatus = "active" | "upcoming" | "finished";
export type ShiftTimeDetails = {
	start: Date;
	end: Date;
	progress: number;
	minutesLeft: number;

	status: ShiftStatus;
};

export const buildShiftTimeDetails = (shift: OncallShift): ShiftTimeDetails => {
	const attr = shift.attributes;
	const start = new Date(attr.startAt);
	const end = new Date(attr.endAt);
	const progress = (Date.now() - start.valueOf()) / (end.valueOf() - start.valueOf());
	const minutesLeft = differenceInMinutes(end, Date.now());
	const status: ShiftStatus = isPast(end) ? "finished" : isFuture(start) ? "upcoming" : "active";
	return { start, end, progress, minutesLeft, status };
};

export const formatShiftDates = (shift: OncallShift) => {
	const start = new Date(shift.attributes.startAt);
	const end = new Date(shift.attributes.endAt);
	const rosterName = shift.attributes.roster.attributes.name;
	return `${rosterName} - ${start.toDateString()} to ${end.toDateString()}`;
};

export type ShiftEvent = {
	timestamp: ZonedDateTime;
	eventType: "alert" | "incident";
};
const fakeEvent = (day: ZonedDateTime): ShiftEvent => {
	const isAlert = Math.random() > 0.25;
	const timestamp = day
		.copy()
		.set({ hour: Math.floor(Math.random() * 24), minute: Math.floor(Math.random() * 60) });
	return { timestamp, eventType: isAlert ? "alert" : "incident" };
};
const dayEvents = (start: ZonedDateTime, day: number): ShiftEvent[] => {
	const date = start.add({ days: day });
	const numEvents = Math.floor(Math.random() * 20);
	return Array.from({ length: numEvents }, (_, i) => fakeEvent(date));
};

export const makeFakeShiftEvents = (start: ZonedDateTime, end: ZonedDateTime) => {
	const shiftDays = differenceInCalendarDays(end.toDate(), start.toDate());
	let events: ShiftEvent[] = [];
	for (let day = 0; day < shiftDays; day++) {
		events = events.concat(dayEvents(start, day));
	}
	return events;
};

export type ShiftEventFilterKind = "alerts" | "nightAlerts" | "incidents";

export const shiftEventMatchesFilter = (event: ShiftEvent, kind: ShiftEventFilterKind) => {
	if ((kind === "alerts" || kind === "nightAlerts") && event.eventType !== "alert") return false;
	if (kind === "nightAlerts" && (event.timestamp.hour < 18 && event.timestamp.hour > 6)) return false;
	if (kind === "incidents" && event.eventType !== "incident") return false;
	return true;
}

const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
export const flatShiftEvents = (start: ZonedDateTime, end: ZonedDateTime, events: ShiftEvent[], kind?: ShiftEventFilterKind) => {
	const day1 = start.toDate();

	const numEvents = new Map<string, number>();
	events.forEach((event) => {
		if (!!kind && !shiftEventMatchesFilter(event, kind)) return;
		const day = differenceInCalendarDays(event.timestamp.toDate(), day1);
		const key = eventDayKey(day, event.timestamp.hour);
		numEvents.set(key, (numEvents.get(key) || 0) + 1);
	});

	const numDays = differenceInCalendarDays(end.toDate(), start.toDate());

	return Array.from({ length: numDays }).flatMap((_, day) => {
		return Array.from({ length: 24 }).map((_, hour) => [
			day,
			hour,
			numEvents.get(eventDayKey(day, hour)) || 0,
		]);
	});
};
