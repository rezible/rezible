import type { OncallEvent, OncallShift } from "$lib/api";
import { settings } from "$lib/settings.svelte";
import { PeriodType } from "@layerstack/utils";

export const formatShiftDates = (shift: OncallShift) => {
	const startFmt = settings.format(new Date(shift.attributes.startAt), PeriodType.Day);
	const endFmt = settings.format(new Date(shift.attributes.endAt), PeriodType.Day);
	const rosterName = shift.attributes.roster.attributes.name;
	return `${rosterName} - ${startFmt} to ${endFmt}`;
};

export type ShiftEventFilterKind = "alerts" | "nightAlerts" | "incidents";

export const isBusinessHours = (hour: number) => {
	return hour >= 9 && hour < 18; // 9am to 5pm
};

export const isNightHours = (hour: number) => {
	return hour >= 22 || hour < 6; // 10pm to 6am
};

export const shiftEventMatchesFilter = (event: OncallEvent, kind: ShiftEventFilterKind) => {
	if ((kind === "alerts" || kind === "nightAlerts") && event.kind !== "alert") return false;
	if (kind === "incidents" && event.kind !== "incident") return false;
	const hour = new Date(event.timestamp).getHours(); // TODO: check this
	if (kind === "nightAlerts" && (hour < 18 && hour > 6)) return false;
	return true;
}

export const getHourLabel = (hour: number): string => {
	const ampm = hour >= 12 ? 'PM' : 'AM';
	const displayHour = hour % 12 || 12;
	return `${displayHour}${ampm}`;
};


