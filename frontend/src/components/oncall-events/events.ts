import { mdiSleepOff, mdiWeatherSunset, mdiClockOutline, mdiChatQuestion, mdiFire, mdiPhoneAlert } from "@mdi/js";
import { isBusinessHours, isNightHours } from "$src/features/oncall-shift/lib/utils";

export const getEventTimeIcon = (date: Date) => {
	const isOutsideBusinessHours = !isBusinessHours(date.getHours());
	const isNightTime = isNightHours(date.getHours());
	if (isNightTime) return {tooltip: "Night hours (10pm-6am)", icon: mdiSleepOff, color: "text-danger-600"};
	if (isOutsideBusinessHours) return {tooltip: "Outside business hours (9am-5pm)", icon: mdiWeatherSunset, color: "text-warning-500"};
	return {tooltip: "", icon: mdiClockOutline, color: "text-surface-content/70"};
}

export const getEventKindIcon = (kind: string) => {
	switch (kind) {
		case "incident": return {icon: mdiFire, color: "text-danger-900/50"};
		case "alert": return {icon: mdiPhoneAlert, color: "text-warning-700/50"};
		default: return {icon: mdiChatQuestion, color: "text-surface-content/40"};
		}
}