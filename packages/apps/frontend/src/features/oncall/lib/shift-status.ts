import type { OncallShift } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiPulseLine from "remixicon-svelte/icons/pulse-line";
import RiTimeLine from "remixicon-svelte/icons/time-line";

/** Whether the shift is on now, still to come, or over, at `now`. */
export function shiftStatus(shift: OncallShift, now = Date.now()): StatusPresentation {
	const startMs = Date.parse(shift.attributes.startAt);
	const endMs = Date.parse(shift.attributes.endAt);

	if (now < startMs) {
		return { label: "Upcoming", tone: "neutral", icon: RiTimeLine };
	}
	if (now < endMs) {
		return { label: "On call", tone: "success", icon: RiPulseLine };
	}
	return { label: "Ended", tone: "neutral", icon: RiCheckboxCircleLine };
}
