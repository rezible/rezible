
export const WeekdaysShort = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"] as const;
export type Weekday = "mon" | "tue" | "wed" | "thu" | "fri" | "sat" | "sun";
export const Weekdays: { value: Weekday; label: string }[] = [
	{ value: "sun", label: "Sunday" },
	{ value: "mon", label: "Monday" },
	{ value: "tue", label: "Tuesday" },
	{ value: "wed", label: "Wednesday" },
	{ value: "thu", label: "Thursday" },
	{ value: "fri", label: "Friday" },
	{ value: "sat", label: "Saturday" },
];