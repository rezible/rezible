import { Time, ZonedDateTime } from "@internationalized/date";
import { settings } from "$lib/settings.svelte";
import { hour12 } from "$lib/format.svelte";

export type Period = "AM" | "PM";
type InternalValue = {
	date: Date;
	time: Time;
	period: Period;
	timezone: string;
};
export const convertTime = (t: ZonedDateTime): InternalValue => {
	const d = t.copy();
	return {
		date: d.toDate(),
		time: new Time(d.hour, d.minute, d.second),
		period: d.hour >= 12 ? "PM" : "AM",
		timezone: d.timeZone,
	};
};

const pad = (n: number) => String(n).padStart(2, "0");

// const dayPeriod = PeriodType.Day;
// const calendarDateFormat = [
// 	DateToken.Month_long,
// 	DateToken.DayOfMonth_withOrdinal,
// 	DateToken.Year_numeric,
// ];
// const dayOfWeekFormat = DateToken.DayOfWeek_long;

// export const createFormatter = () => {
// 	const f = $derived(settings.format);

// 	const asTime = (time: Time): string => {
// 		const period = time.hour >= 12 ? "PM" : "AM";
// 		return `${pad(hour12(time.hour))}:${pad(time.minute)}${period}`;
// 	}

// 	const asCalendarDate = (d: Date): string => {
// 		return f(d, dayPeriod, { custom: calendarDateFormat });
// 	}

// 	const asWeekday = (d: Date): string => {
// 		return f(d, dayPeriod, { custom: dayOfWeekFormat })
// 	}

// 	return {
// 		asTime,
// 		asCalendarDate,
// 		asWeekday
// 	}
// }
// export const format = createFormatter();