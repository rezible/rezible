import { getLocalTimeZone, parseAbsolute } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { getOncallShiftOptions, listOncallEventsOptions, type OncallShift } from "$lib/api";
import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
import { Context, watch } from "runed";
import { settings } from "$lib/settings.svelte";
import { PeriodType } from "@layerstack/utils";

export class ShiftViewState {
	shiftId = $state("");

	constructor(idFn: () => string) {
		watch(idFn, id => { this.shiftId = id });
	}

	useShiftTimezone = $state(false);
	timezone = $derived(this.useShiftTimezone ? "" : getLocalTimeZone());

	shiftQueryOpts = $derived(getOncallShiftOptions({ path: { id: (this.shiftId ?? "") } }))
	shiftQuery = createQuery(() => ({ ...this.shiftQueryOpts, enabled: !!this.shiftId }))
	shift = $derived(this.shiftQuery.data?.data);

	roster = $derived(this.shift?.attributes.roster);

	shiftStart = $derived(this.shift && parseAbsolute(this.shift.attributes.startAt, this.timezone));
	shiftEnd = $derived(this.shift && parseAbsolute(this.shift.attributes.endAt, this.timezone));

	shiftTitle = $derived.by(() => {
		if (!this.shiftStart || !this.shiftEnd || !this.roster) return "";
		const startFmt = settings.format(this.shiftStart.toDate(), PeriodType.Day);
		const endFmt = settings.format(this.shiftEnd.toDate(), PeriodType.Day);
		return `${this.roster.attributes.name} - ${startFmt} to ${endFmt}`;
	})

	shiftEventsQueryOpts = $derived(listOncallEventsOptions({ query: { shiftId: (this.shiftId ?? "") } }));
	eventsQuery = createQuery(() => ({ ...this.shiftEventsQueryOpts, enabled: !!this.shift }))
	events = $derived(this.eventsQuery.data?.data);

	eventsFilter = $state<ShiftEventFilterKind>();
	filteredEvents = $derived.by(() => {
		if (!this.events) return [];
		if (!this.eventsFilter) return this.events;
		return this.events.filter(e => (!this.eventsFilter || shiftEventMatchesFilter(e, this.eventsFilter)));
	});
}

export const shiftViewStateCtx = new Context<ShiftViewState>("shiftViewState");
