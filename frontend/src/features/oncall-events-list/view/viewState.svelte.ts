import { listOncallEventsOptions, type ListOncallEventsData, type OncallEventAttributes } from "$lib/api";
import { useUserOncallInformation } from "$lib/userOncall.svelte";
import { PeriodType } from "@layerstack/utils";
import type { DateRange as DateRangeType } from "@layerstack/utils/dateRange";
import { subMonths, subWeeks } from "date-fns";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { QueryPaginatorState } from "$lib/paginator.svelte";

export type DateRangeOption = { label: string, value: "shift" | "7d" | "30d" | "custom" };

const last7Days = () => ({ from: subWeeks(new Date(), 1), to: new Date(), periodType: PeriodType.Day });
const lastMonth = () => ({ from: subMonths(new Date(), 1), to: new Date(), periodType: PeriodType.Day });

export type EventKind = OncallEventAttributes["kind"];

export type FilterOptions = {
	rosterId?: string;
	eventKinds?: EventKind[];
	annotated?: boolean;
};


export class EventsListViewState {
	private queryClient = useQueryClient();
	private oncallInfo = useUserOncallInformation();

	activeShift = $derived(this.oncallInfo.activeShifts.at(0));

	defaultShiftDateRange = $derived(this.activeShift && {
		from: new Date(this.activeShift.attributes.startAt),
		to: new Date(this.activeShift.attributes.endAt),
		periodType: PeriodType.Day,
	});

	dateRangeOption = $state<DateRangeOption["value"]>("7d");
	customDateRangeValue = $state<DateRangeType>(last7Days());
	shiftDateRange = $derived(!!this.defaultShiftDateRange ? this.defaultShiftDateRange : this.customDateRangeValue)

	dateRange = $derived.by(() => {
		switch (this.dateRangeOption) {
			case "7d": return last7Days();
			case "30d": return lastMonth();
			case "shift": return this.shiftDateRange;
			case "custom": return this.customDateRangeValue;
		}
	});
	
	defaultRosterId = $derived.by(() => {
		if (this.activeShift) return this.activeShift.attributes.roster.id;
		if (this.oncallInfo.rosterIds.length > 0) return this.oncallInfo.rosterIds.at(0);
	});

	paginator = new QueryPaginatorState();
	queryEnabled = $derived(!!this.oncallInfo && !!this.defaultRosterId);

	filterEventKinds = $state<EventKind[]>();
	filterAnnotation = $state<boolean>();
	filterRosterId = $state<string>();

	private listRosterEventsQueryData = $derived<ListOncallEventsData["query"]>({
		from: this.dateRange.from?.toISOString(),
		to: this.dateRange.to?.toISOString(),
		rosterId: this.filterRosterId,
	});
	private listShiftEventsQueryData = $derived<ListOncallEventsData["query"]>({ shiftId: this.activeShift?.id });
	private listShiftEventsFinalQueryData = $derived(this.dateRangeOption === "shift" ? this.listShiftEventsQueryData : this.listRosterEventsQueryData);

	private listEventsQueryData = $derived<ListOncallEventsData["query"]>({
		...this.listShiftEventsFinalQueryData,
		limit: this.paginator.limit,
		offset: this.paginator.offset,
		withAnnotations: true,
	})
	private listEventsQueryOptions = $derived(listOncallEventsOptions({ query: this.listEventsQueryData }));

	private listEventsQuery = createQuery(() => ({
		...this.listEventsQueryOptions,
		enabled: this.queryEnabled,
	}));
	private listEventsQueryDataResult = $derived(this.listEventsQuery.data);
	events = $derived(this.listEventsQueryDataResult?.data ?? []);

	invalidateQuery() {
		this.queryClient.invalidateQueries(this.listEventsQueryOptions);
	}

	loading = $derived(this.listEventsQuery.isLoading || !this.oncallInfo.loaded);

	constructor() {
		this.paginator.watchQuery(this.listEventsQuery);
	};
}

export const eventsListViewStateCtx = new Context<EventsListViewState>("eventsListView");