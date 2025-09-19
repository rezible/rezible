import { getLocalTimeZone, parseAbsolute } from "@internationalized/date";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { AnnotationDialogState, setAnnotationDialogState } from "$src/components/events/annotation-dialog/dialogState.svelte";
import { getAdjacentOncallShiftsOptions, getOncallShiftOptions, listEventAnnotationsOptions, listEventsOptions, type EventAnnotation } from "$lib/api";
import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$features/oncall-shift/lib/utils";
import { Context, watch } from "runed";
import { settings } from "$lib/settings.svelte";
import { PeriodType } from "@layerstack/utils";
import type { Getter } from "$src/lib/utils.svelte";

class OncallShiftViewState {
	private queryClient = useQueryClient();
	shiftId = $state<string>(null!);

	constructor(idFn: () => string) {
		this.shiftId = idFn();
		watch(idFn, id => { this.shiftId = id });

		setAnnotationDialogState(new AnnotationDialogState({
			onClosed: (updated?: EventAnnotation) => {this.onAnnotationDialogUpdated(updated)},
		}));
	}

	useShiftTimezone = $state(false);
	timezone = $derived(this.useShiftTimezone ? "" : getLocalTimeZone());

	private shiftQuery = createQuery(() => getOncallShiftOptions({ path: { id: this.shiftId } }))
	shift = $derived(this.shiftQuery.data?.data);
	roster = $derived(this.shift?.attributes.roster);

	shiftStart = $derived(this.shift && parseAbsolute(this.shift.attributes.startAt, this.timezone));
	shiftEnd = $derived(this.shift && parseAbsolute(this.shift.attributes.endAt, this.timezone));

	shiftTitle = $derived.by(() => {
		if (!this.shiftStart || !this.shiftEnd || !this.roster) return "";
		const startFmt = settings.format(this.shiftStart.toDate(), PeriodType.Day);
		const endFmt = settings.format(this.shiftEnd.toDate(), PeriodType.Day);
		return `${this.roster.attributes.name} - ${startFmt} to ${endFmt}`;
	});

	private adjacentShiftsQuery = createQuery(() => getAdjacentOncallShiftsOptions({ path: { id: this.shiftId }}));
	nextShift = $derived(this.adjacentShiftsQuery.data?.data.next);
	previousShift = $derived(this.adjacentShiftsQuery.data?.data.previous);

	private eventsQueryOptions = $derived(listEventsOptions({ query: {
		// TODO
		// shiftId: this.shiftId,
		// withAnnotations: true,
	}}));
	eventsQuery = createQuery(() => (this.eventsQueryOptions));
	events = $derived(this.eventsQuery.data?.data);

	eventsFilter = $state<ShiftEventFilterKind>();
	filteredEvents = $derived.by(() => {
		if (!this.events) return [];
		if (!this.eventsFilter) return this.events;
		return this.events.filter(e => (!this.eventsFilter || shiftEventMatchesFilter(e, this.eventsFilter)));
	});

	onAnnotationDialogUpdated(updated?: EventAnnotation) {
		if (!updated) return;
		this.queryClient.invalidateQueries(this.eventsQueryOptions);
		this.queryClient.invalidateQueries(listEventAnnotationsOptions({ query: { 
			// TODO
			// shiftId: this.shiftId, 
			withEvents: true 
		} }));
	}
}

const ctx = new Context<OncallShiftViewState>("oncallShiftViewState");
export const setOncallShiftViewState = (idFn: Getter<string>) => ctx.set(new OncallShiftViewState(idFn));
export const useOncallShiftViewState = () => ctx.get();
