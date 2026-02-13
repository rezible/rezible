import { z } from "zod";
import { CalendarDate, getLocalTimeZone, now, type DateTimeDuration } from "@internationalized/date";

const refineZonedDateTimeString = (dateStr: string) => {
	try {
		const [datePart, timezonePart] = dateStr.split("[");
		if (!datePart || !timezonePart?.endsWith("]")) return false;

		const isoDateTimeRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$/;
		return isoDateTimeRegex.test(datePart) &&
			/^[A-Za-z_/]+$/.test(timezonePart.slice(0, -1));
	} catch {
		return false;
	}
}

export const ZodZonedDateTime = z.string().
	refine(refineZonedDateTimeString, "Invalid ZonedDateTime format");

export const DayHours = [
	'12am', '1am', '2am', '3am', '4am', '5am', '6am',
	'7am', '8am', '9am', '10am', '11am',
	'12pm', '1pm', '2pm', '3pm', '4pm', '5pm',
	'6pm', '7pm', '8pm', '9pm', '10pm', '11pm'
];

export const makeDateRangeWindow = (duration: DateTimeDuration) => {
	return { 
		from: now(getLocalTimeZone()).subtract(duration).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: "day",
	}
}

export const makeCalendarDateString = (d: Date) => {
	return new CalendarDate(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()).toString()
}