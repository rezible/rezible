import type { OncallEvent, OncallShift } from "$lib/api";
import { settings } from "$lib/settings.svelte";
import { PeriodType } from "@layerstack/utils";

export type ShiftEventFilterKind = "alerts" | "nightAlerts" | "incidents";

export const isBusinessHours = (hour: number) => {
	return hour >= 9 && hour < 18; // 9am to 5pm
};

export const isNightHours = (hour: number) => {
	return hour >= 22 || hour < 6; // 10pm to 6am
};

export const shiftEventMatchesFilter = (event: OncallEvent, kind: ShiftEventFilterKind) => {
	const attrs = event.attributes;
	if ((kind === "alerts" || kind === "nightAlerts") && attrs.kind !== "alert") return false;
	if (kind === "incidents" && attrs.kind !== "incident") return false;
	const hour = new Date(attrs.timestamp).getHours(); // TODO: check this
	if (kind === "nightAlerts" && (hour < 18 && hour > 6)) return false;
	return true;
}



