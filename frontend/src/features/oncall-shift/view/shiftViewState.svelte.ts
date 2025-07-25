import { getLocalTimeZone, parseAbsolute } from "@internationalized/date";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { AnnotationDialogState, setAnnotationDialogState } from "$components/oncall-events/annotation-dialog/dialogState.svelte";
import { getOncallShiftOptions, listOncallAnnotationsOptions, type ListOncallEventsData, listOncallEventsOptions, type OncallAnnotation } from "$lib/api";
import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$src/features/oncall-shift/lib/utils";
import { Context, watch } from "runed";
import { settings } from "$lib/settings.svelte";
import { PeriodType } from "@layerstack/utils";

export class ShiftViewState {
	private queryClient = useQueryClient();
	shiftId = $state<string>(null!);

	constructor(idFn: () => string) {
		this.shiftId = idFn();
		watch(idFn, id => { this.shiftId = id });

		setAnnotationDialogState(new AnnotationDialogState({
			onClosed: (updated?: OncallAnnotation) => {this.onAnnotationDialogUpdated(updated)},
		}));
	}

	useShiftTimezone = $state(false);
	timezone = $derived(this.useShiftTimezone ? "" : getLocalTimeZone());

	shiftQuery = createQuery(() => getOncallShiftOptions({ path: { id: this.shiftId } }))
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

	private eventsQueryOptions = $derived(listOncallEventsOptions({ query: {
		shiftId: this.shiftId,
		withAnnotations: true,
	}}));
	eventsQuery = createQuery(() => (this.eventsQueryOptions));
	events = $derived(this.eventsQuery.data?.data);

	eventsFilter = $state<ShiftEventFilterKind>();
	filteredEvents = $derived.by(() => {
		if (!this.events) return [];
		if (!this.eventsFilter) return this.events;
		return this.events.filter(e => (!this.eventsFilter || shiftEventMatchesFilter(e, this.eventsFilter)));
	});

	onAnnotationDialogUpdated(updated?: OncallAnnotation) {
		if (!updated) return;
		this.queryClient.invalidateQueries(this.eventsQueryOptions);
		this.queryClient.invalidateQueries(listOncallAnnotationsOptions({ query: { shiftId: this.shiftId, withEvents: true } }));
	}
}

const shiftViewStateCtx = new Context<ShiftViewState>("shiftViewState");
export const setShiftViewState = (s: ShiftViewState) => shiftViewStateCtx.set(s);
export const useShiftViewState = () => shiftViewStateCtx.get();
