import { z } from "zod";
import { addDays } from "date-fns/addDays";
import type {
	CreateMeetingScheduleRequestBody,
	CreateMeetingSessionRequestBody,
	DateTimeAnchor,
} from "$lib/api";

export type Weekday = "mon" | "tue" | "wed" | "thu" | "fri" | "sat" | "sun";
const WEEKDAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"] as const;
export const weekdays: { value: Weekday; label: string }[] = [
	{ value: "sun", label: "Sunday" },
	{ value: "mon", label: "Monday" },
	{ value: "tue", label: "Tuesday" },
	{ value: "wed", label: "Wednesday" },
	{ value: "thu", label: "Thursday" },
	{ value: "fri", label: "Friday" },
	{ value: "sat", label: "Saturday" },
];

export type CreateMeetingFormData = {
	name: string;
	sessionTitle: string;
	description: string;
	start: DateTimeAnchor;
	durationMinutes: number;
	repeats: "once" | "daily" | "weekly" | "monthly";
	repetitionStep: number;
	weekDays: Set<Weekday>;
	monthlyOn: "same_day" | "same_weekday";
	untilType: "indefinite" | "num_repetitions" | "date";
	untilDate: Date;
	numRepetitions: number;
};

export const emptyForm: CreateMeetingFormData = {
	name: "",
	sessionTitle: "",
	description: "",
	start: {
		date: new Date(),
		time: "09:00:00",
		timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
	},
	durationMinutes: 30,
	repeats: "once",
	weekDays: new Set<Weekday>(),
	repetitionStep: 1,
	monthlyOn: "same_day",
	untilType: "indefinite",
	untilDate: addDays(new Date(), 1),
	numRepetitions: 2,
};
export const getEmptyForm = () => structuredClone(emptyForm);

const meetingFormSchema = z.object({
	name: z.string().min(1),
	sessionTitle: z.string(),
	description: z.string(),
	start: z.object({
		date: z.date(),
		time: z.string(),
		timezone: z.string(),
	}),
	durationMinutes: z.number(),
	repeats: z.enum(["once", "daily", "weekly", "monthly"]),

	repetitionStep: z.number().min(1),
	weekDays: z.set(z.enum(WEEKDAYS)),
	monthlyOn: z.enum(["same_day", "same_weekday"]),
	untilType: z.enum(["indefinite", "num_repetitions", "date"]),
	untilDate: z.date(),
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
