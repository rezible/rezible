import { getLocalTimeZone, now, parseAbsolute } from "@internationalized/date";
import { createQuery, useQueryClient, type QueryClient } from "@tanstack/svelte-query";
import { getOncallShiftOptions, listOncallShiftEventsOptions } from "$lib/api";

const makeShiftState = () => {
	let shiftId = $state<string>();
	let useShiftTimezone = $state(false);
	const timezone = $derived(useShiftTimezone ? "" : getLocalTimeZone());

	let queryClient = $state<QueryClient>();

	const shiftQueryOpts = $derived(getOncallShiftOptions({ path: { id: (shiftId ?? "") } }))
	const makeShiftQuery = () => createQuery(() => ({...shiftQueryOpts, enabled: !!shiftId}));
	let shiftQuery = $state<ReturnType<typeof makeShiftQuery>>();

	const shift = $derived(shiftQuery?.data?.data);
	const shiftStart = $derived(shift ? parseAbsolute(shift.attributes.startAt, timezone) : now(timezone));
	const shiftEnd = $derived(shift ? parseAbsolute(shift.attributes.endAt, timezone) : now(timezone));

	const shiftEventsQueryOpts = $derived(listOncallShiftEventsOptions({ path: { id: (shiftId ?? "") } }))
	const makeShiftEventsQuery = () => createQuery(() => ({...shiftEventsQueryOpts, enabled: !!shift}));
	let eventsQuery = $state<ReturnType<typeof makeShiftEventsQuery>>();
	const shiftEvents = $derived(eventsQuery?.data?.data ?? []);

	const setUseShiftTimezone = (b: boolean) => {
		useShiftTimezone = b;
	}

	const setup = (id: string) => {
		shiftId = id;
		queryClient = useQueryClient();
		shiftQuery = makeShiftQuery();
		eventsQuery = makeShiftEventsQuery();
	}

	return {
		setup,
		setUseShiftTimezone,
		get usingShiftTimezone() { return useShiftTimezone },
		get shiftId() { return shiftId },
		get shift() { return shift },
		get shiftEvents() { return shiftEvents },
		get shiftStart() { return shiftStart },
		get shiftEnd() { return shiftEnd },
	}
}

export const shiftState = makeShiftState();