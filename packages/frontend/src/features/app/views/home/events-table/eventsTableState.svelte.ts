import { listEventsOptions, type ListEventsData, type EventAttributes } from "$lib/api";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { subMonths, subWeeks } from "date-fns";
import { watch } from "runed";
import { useUserOncallInformation } from "$lib/userOncall.svelte";
import { QueryPaginatorState } from "$lib/paginator.svelte";

export type DateRangeOption = { label: string, value: "shift" | "7d" | "30d" | "custom" };

export const dateRangeOptions: DateRangeOption[] = [
	{ label: "Last 7 Days", value: "7d" },
	{ label: "Last Month", value: "30d" },
	{ label: "Custom", value: "custom" },
];

const last7Days = () => ({ from: subWeeks(new Date(), 1), to: new Date(), periodType: "day" });
const lastMonth = () => ({ from: subMonths(new Date(), 1), to: new Date(), periodType: "day" });

export type EventKind = EventAttributes["kind"];

export type FilterOptions = {
	rosterId?: string;
	eventKinds?: EventKind[];
	annotated?: boolean;
};

export class EventsTableState {
	private queryClient = useQueryClient();
	private oncallInfo = useUserOncallInformation();

	activeShift = $derived(this.oncallInfo.activeShifts.at(0));

	defaultShiftDateRange = $derived(this.activeShift && {
		from: new Date(this.activeShift.attributes.startAt),
		to: new Date(this.activeShift.attributes.endAt),
		periodType: "day",
	});

	dateRangeOption = $state<DateRangeOption["value"]>("7d");
	customDateRangeValue = $state(last7Days());
	shiftDateRange = $derived(!!this.defaultShiftDateRange ? this.defaultShiftDateRange : this.customDateRangeValue)

	dateRange = $derived.by(() => {
		switch (this.dateRangeOption) {
			case "7d": return last7Days();
			case "30d": return lastMonth();
			case "shift": return this.shiftDateRange;
			case "custom": return this.customDateRangeValue;
		}
	});

	filters = $state<FilterOptions>({});
	defaultRosterId = $derived.by(() => {
		if (this.activeShift) return this.activeShift.attributes.roster.id;
		if (this.oncallInfo.rosterIds.length > 0) return this.oncallInfo.rosterIds.at(0);
	});

	paginator = new QueryPaginatorState();
	queryEnabled = $derived(!!this.oncallInfo && !!this.defaultRosterId);

	private listRosterEventsQueryData = $derived<ListEventsData["query"]>({
		from: this.dateRange.from?.toISOString(),
		to: this.dateRange.to?.toISOString(),
		// rosterId: this.filters.rosterId,
	});
	private listShiftEventsQueryData = $derived<ListEventsData["query"]>({ 
		from: this.dateRange.from?.toISOString(),
		to: this.dateRange.to?.toISOString(),
		// shiftId: this.activeShift?.id,
	});
	private listShiftEventsFinalQueryData = $derived(this.dateRangeOption === "shift" ? this.listShiftEventsQueryData : this.listRosterEventsQueryData);

	private listEventsQueryData = $derived<ListEventsData["query"]>({
		...this.listShiftEventsFinalQueryData,
		limit: this.paginator.limit,
		offset: this.paginator.offset,
		// withAnnotations: true,
	});
	private listEventsQueryOptions = $derived(listEventsOptions({ query: this.listEventsQueryData }));

	private listEventsQuery = createQuery(() => ({
		...this.listEventsQueryOptions,
		enabled: this.queryEnabled,
	}));
	private listEventsQueryDataResult = $derived(this.listEventsQuery.data);
	events = $derived(this.listEventsQueryDataResult?.data ?? []);

	invalidateQuery() {
		this.queryClient.invalidateQueries(this.listEventsQueryOptions);
	}

	loading = $derived(this.listEventsQuery.isLoading || !this.oncallInfo.loaded)

	constructor() {
		watch(() => this.activeShift, s => {
			this.dateRangeOption = !!s ? "shift" : this.dateRangeOption;
		});

		watch(() => this.defaultRosterId, id => {
			this.filters.rosterId = id;
		});

		this.paginator.watchQuery(this.listEventsQuery);
	};
}
