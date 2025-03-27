import { getLocalTimeZone, now, parseAbsolute, ZonedDateTime } from "@internationalized/date";
import { createQuery, queryOptions, useQueryClient, type QueryClient } from "@tanstack/svelte-query";
import { getOncallShiftOptions } from "$lib/api";
import type { ShiftEvent } from "$features/oncall/lib/utils";
import { v4 as uuidv4 } from "uuid";
import { differenceInCalendarDays } from "date-fns";

// TODO: Implement this properly
const makeFakeShiftEvent = (date: ZonedDateTime): ShiftEvent => {
	const isAlert = Math.random() > 0.25;
	const eventType = isAlert ? "alert" : "incident";
	const hour = Math.floor(Math.random() * 24);
	const minute = Math.floor(Math.random() * 60);
	const timestamp = date.copy().set({ hour, minute });
	let annotation: string | undefined = undefined;
	if (Math.random() > .8) annotation = "annotation";
	return { id: uuidv4(), timestamp, eventType, description: "description", annotation };
};

const createFakeShiftEvents = (start: ZonedDateTime, days: number): ShiftEvent[] => {
	let events: ShiftEvent[] = [];
	for (let day = 0; day < days; day++) {
		const dayDate = start.add({ days: day });
		const numDayEvents = Math.floor(Math.random() * 10);
		const dayEvents = Array.from({ length: numDayEvents }, () => makeFakeShiftEvent(dayDate));
		events = events.concat(dayEvents);
	}
	return events;
}

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

	// TODO: implement in backend
	const listOncallShiftEventsOptions = (params: {path: {id: string}}) => queryOptions({
		queryKey: ["shiftEvents", params.path.id],
		queryFn: async () => {
			const numDays = differenceInCalendarDays(shiftEnd.toDate(), shiftStart.toDate());
			return { data: createFakeShiftEvents(shiftStart, numDays) };
		}
	});

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