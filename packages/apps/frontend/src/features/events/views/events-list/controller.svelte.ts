import { Context } from "runed";
import { type ListEventsData, type EventAttributes, listEventsOptions } from "$lib/api";
import { subMonths, subWeeks } from "date-fns";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";

export type DateRangeOption = { label: string; value: "shift" | "7d" | "30d" | "custom" };

const last7Days = () => ({ from: subWeeks(new Date(), 1), to: new Date(), periodType: "day" });
const lastMonth = () => ({ from: subMonths(new Date(), 1), to: new Date(), periodType: "day" });

export type EventKind = EventAttributes["kind"];

export type FilterOptions = {
	rosterId?: string;
	eventKinds?: EventKind[];
	annotated?: boolean;
};

export class EventsListFiltersState {
	dateRangeOption = $state<DateRangeOption["value"]>("7d");
	customDateRangeValue = $state(last7Days());

	dateRange = $derived.by(() => {
		switch (this.dateRangeOption) {
			case "7d":
				return last7Days();
			case "30d":
				return lastMonth();
			case "custom":
				return this.customDateRangeValue;
		}
	});

	eventKinds = $state<EventKind[]>();

	queryData = $derived<ListEventsData["query"]>({
		// from: this.dateRange.from?.toISOString(),
		// to: this.dateRange.to?.toISOString(),
		withProjection: true,
	});
	queryEnabled = $derived(true);
}

export class EventsListController {
	filters = new EventsListFiltersState();

	paginatedEventsQuery = createPaginatedQuery({
		queryOptions: (pagination) => ({
			...listEventsOptions({ query: {...this.filters.queryData, ...pagination} }),
			enabled: this.filters.queryEnabled,
		}),
		resetWhen: () => [
			$state.snapshot(this.filters.queryData),
		],
	});

	private query = $derived(this.paginatedEventsQuery.query);
	events = $derived(this.query.data?.data ?? []);
	isLoading = $derived(this.query.isLoading);
}

const ctx = new Context<EventsListController>("EventsListController");
export const initEventsListController = () => ctx.set(new EventsListController());
export const useEventsListController = () => ctx.get();
