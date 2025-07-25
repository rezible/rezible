import { z } from "zod";
import type {
	CreateMeetingScheduleRequestBody,
	CreateMeetingSessionRequestBody,
} from "$lib/api";
import { getLocalTimeZone, now, type ZonedDateTime } from "@internationalized/date";
import { ZodZonedDateTime } from "$lib/utils.svelte";
import { WeekdaysShort, type Weekday } from "$lib/scheduling";

export type CreateMeetingFormData = {
	name: string;
	sessionTitle: string;
	description: string;
	start: ZonedDateTime;
	durationMinutes: number;
	repeats: "once" | "daily" | "weekly" | "monthly";
	repetitionStep: number;
	weekDays: Set<Weekday>;
	monthlyOn: "same_day" | "same_weekday";
	untilType: "indefinite" | "num_repetitions" | "date";
	untilDate: ZonedDateTime;
	numRepetitions: number;
};

export const getEmptyForm = (): CreateMeetingFormData => {
	const curTime = now(getLocalTimeZone());
	return {
		name: "",
		sessionTitle: "",
		description: "",
		start: curTime,
		durationMinutes: 30,
		repeats: "once",
		weekDays: new Set<Weekday>(),
		repetitionStep: 1,
		monthlyOn: "same_day",
		untilType: "indefinite",
		untilDate: curTime.add({days: 1}),
		numRepetitions: 2,
	}
};

const meetingFormSchema = z.object({
	name: z.string().min(1),
	sessionTitle: z.string(),
	description: z.string(),
	start: ZodZonedDateTime,
	durationMinutes: z.number(),
	repeats: z.enum(["once", "daily", "weekly", "monthly"]),

	repetitionStep: z.number().min(1),
	weekDays: z.set(z.enum(WeekdaysShort)),
	monthlyOn: z.enum(["same_day", "same_weekday"]),
	untilType: z.enum(["indefinite", "num_repetitions", "date"]),
	untilDate: ZodZonedDateTime,
	numRepetitions: z.number().min(1),
});

type TransformedScheduleFormData = {
	requestType: "schedule";
	body: CreateMeetingScheduleRequestBody;
};
type TransformedSessionFormData = {
	requestType: "session";
	body: CreateMeetingSessionRequestBody;
};
export type TransformedFormData = TransformedScheduleFormData | TransformedSessionFormData;

export const CreateMeetingFormSchema = meetingFormSchema.transform((form, ctx): TransformedFormData => {
	if (form.repeats === "once") {
		const body: CreateMeetingSessionRequestBody = {
			attributes: {
				title: form.name,
				description: form.description,
				attendees: {
					private: false,
					teams: [],
					users: [],
				},
				documentTemplateId: undefined,
				startsAt: form.start,
				durationMinutes: form.durationMinutes,
			},
		};
		return { requestType: "session", body };
	}
	const body: CreateMeetingScheduleRequestBody = {
		attributes: {
			name: form.name,
			sessionTitle: form.sessionTitle,
			description: form.description,
			attendees: {
				private: false,
				teams: [],
				users: [],
			},
			startsAt: form.start,
			durationMinutes: form.durationMinutes,
			repeats: form.repeats,
			repeatMonthlyOn: form.monthlyOn,
			numRepetitions: form.numRepetitions > 1 ? form.numRepetitions : undefined,
			repetitionStep: form.repetitionStep,
			untilDate: !!form.untilDate ? form.untilDate : undefined,
		},
	};
	return { requestType: "schedule", body };
});
