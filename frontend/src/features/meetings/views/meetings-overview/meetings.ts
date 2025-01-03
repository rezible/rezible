import { z } from "zod";
import { addDays } from 'date-fns/addDays';
import type { CreateMeetingScheduleRequestBody, CreateMeetingSessionRequestBody, DateTimeAnchor } from "$lib/api";

export type Weekday = "mon" | "tue" | "wed" | "thu" | "fri" | "sat" | "sun";
const WEEKDAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"] as const;
export const weekdays: {value: Weekday; label: string}[] = [
	{value: "sun", label: "Sunday"},
	{value: "mon", label: "Monday"}, 
	{value: "tue", label: "Tuesday"}, 
	{value: "wed", label: "Wednesday"},
	{value: "thu", label: "Thursday"},
	{value: "fri", label: "Friday"},
	{value: "sat", label: "Saturday"},
];

export type CreateMeetingFormData = {
	name: string;
	session_title: string;
	description: string;
	start: DateTimeAnchor;
	duration_minutes: number;
	repeats: "once" | "daily" | "weekly" | "monthly";

	repetition_step: number;
	week_days: Set<Weekday>;
	monthly_on: "same_day" | "same_weekday";
	until_type: "indefinite" | "num_repetitions" | "date";
	until_date: Date;
	num_repetitions: number;
}

export const emptyForm: CreateMeetingFormData = {
	name: "",
	session_title: "",
	description: "",
	start: {date: new Date(), time: "09:00:00", timezone: Intl.DateTimeFormat().resolvedOptions().timeZone},
	duration_minutes: 30,
	repeats: "once",
	week_days: new Set<Weekday>(),
	repetition_step: 1,
	monthly_on: 'same_day',
	until_type: "indefinite",
	until_date: addDays(new Date(), 1),
	num_repetitions: 2,
}
export const getEmptyForm = () => structuredClone(emptyForm);

const meetingFormSchema = z.object({
	name: z.string().min(1),
	session_title: z.string(),
	description: z.string(),
	start: z.object({
		date: z.date(),
		time: z.string(),
		timezone: z.string(),
	}),
	duration_minutes: z.number(),
	repeats: z.enum(["once", "daily", "weekly", "monthly"]),

	repetition_step: z.number().min(1),
	week_days: z.set(z.enum(WEEKDAYS)),
	monthly_on: z.enum(["same_day", "same_weekday"]),
	until_type: z.enum(["indefinite", "num_repetitions", "date"]),
	until_date: z.date(),
	num_repetitions: z.number().min(1),
});

type TransformedScheduleFormData = {
	requestType: "schedule";
	body: CreateMeetingScheduleRequestBody;
}
type TransformedSessionFormData = {
	requestType: "session";
	body: CreateMeetingSessionRequestBody;
}
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
					users: []
				},
				document_template_id: undefined,
				starts_at: form.start,
				duration_minutes: form.duration_minutes,
			}
		}
		return {requestType: "session", body}
	}
	const body: CreateMeetingScheduleRequestBody = {
		attributes: {
			name: form.name,
			session_title: form.session_title,
			description: form.description,
			attendees: {
				private: false,
				teams: [],
				users: []
			},
			starts_at: form.start,
			duration_minutes: form.duration_minutes,
			repeats: form.repeats,
			repeat_monthly_on: form.monthly_on,
			num_repetitions: form.num_repetitions > 1 ? form.num_repetitions : undefined,
			repetition_step: form.repetition_step,
			until_date: !!form.until_date ? form.until_date : undefined,
		}
	}
	return {requestType: "schedule", body};
});