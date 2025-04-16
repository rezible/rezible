import { getLocalTimeZone, now, parseAbsolute } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { getOncallShiftOptions, listOncallEventsOptions } from "$lib/api";
import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
import { Context, watch } from "runed";

export const shiftIdCtx = new Context<string>("shiftId");

type Getter<T> = () => T;

export class ShiftViewState {
	shiftId = $state<string>();
	useShiftTimezone = $state(false);
	timezone = $derived(this.useShiftTimezone ? "" : getLocalTimeZone());

	shiftQueryOpts = $derived(getOncallShiftOptions({ path: { id: (this.shiftId ?? "") } }))
	shiftQuery = createQuery(() => ({ ...this.shiftQueryOpts, enabled: !!this.shiftId }))
	shift = $derived(this.shiftQuery.data?.data);

	shiftStart = $derived(this.shift && parseAbsolute(this.shift.attributes.startAt, this.timezone));
	shiftEnd = $derived(this.shift && parseAbsolute(this.shift.attributes.endAt, this.timezone));

	shiftEventsQueryOpts = $derived(listOncallEventsOptions({ query: { shiftId: (this.shiftId ?? "") } }));
	eventsQuery = createQuery(() => ({ ...this.shiftEventsQueryOpts, enabled: !!this.shift }))
	events = $derived(this.eventsQuery.data?.data);

	eventsFilter = $state<ShiftEventFilterKind>();
	filteredEvents = $derived.by(() => {
		if (!this.events) return [];
		if (!this.eventsFilter) return this.events;
		return this.events.filter(e => (!this.eventsFilter || shiftEventMatchesFilter(e, this.eventsFilter)));
	});

	constructor(idFn: Getter<string>) {
		watch(idFn, id => { this.shiftId = id });
		// this.shiftId = shiftId;
	}
}

export const shiftViewStateCtx = new Context<ShiftViewState>("shiftViewState");
