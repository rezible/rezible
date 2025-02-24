import type { OncallShift } from "$src/lib/api";
import { differenceInMinutes, isFuture, isPast } from "date-fns";

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