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
	const start = new Date(attr.start_at);
	const end = new Date(attr.end_at);
	const progress =
		(Date.now() - start.valueOf()) / (end.valueOf() - start.valueOf());
	const minutesLeft = differenceInMinutes(end, Date.now());
	const status: ShiftStatus = isPast(end)
		? "finished"
		: isFuture(start)
			? "upcoming"
			: "active";
	return { start, end, progress, minutesLeft, status };
};
