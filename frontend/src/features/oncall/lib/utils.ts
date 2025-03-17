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
	id: string;
	timestamp: ZonedDateTime;
	eventType: "alert" | "incident";
	title?: string;
	description?: string;
	severity?: "critical" | "high" | "medium" | "low";
	status?: "active" | "resolved" | "acknowledged";
	source?: string;
	notes?: string;
};

export type ShiftEventFilterKind = "alerts" | "nightAlerts" | "incidents";

export const isBusinessHours = (hour: number) => {
	return hour >= 9 && hour < 17; // 9am to 5pm
};

export const isNightHours = (hour: number) => {
	return hour >= 22 || hour < 6; // 10pm to 6am
};

export const shiftEventMatchesFilter = (event: ShiftEvent, kind: ShiftEventFilterKind) => {
	if ((kind === "alerts" || kind === "nightAlerts") && event.eventType !== "alert") return false;
	if (kind === "nightAlerts" && (event.timestamp.hour < 18 && event.timestamp.hour > 6)) return false;
	if (kind === "incidents" && event.eventType !== "incident") return false;
	return true;
}

