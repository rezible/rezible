import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { type ListEventsData, type EventAttributes, listEventsOptions } from "$lib/api";
import { subMonths, subWeeks } from "date-fns";
import { QueryPaginatorState } from "$src/lib/paginator.svelte";

export type DateRangeOption = { label: string, value: "shift" | "7d" | "30d" | "custom" };

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
			case "7d": return last7Days();
			case "30d": return lastMonth();
			case "custom": return this.customDateRangeValue;
		}
	});

	eventKinds = $state<EventKind[]>();

	queryData = $derived<ListEventsData["query"]>({
		// from: this.dateRange.from?.toISOString(),
		// to: this.dateRange.to?.toISOString(),
		withProjections: true,
	});
	queryEnabled = $derived(true);
};

export class EventsListController {
	filters = new EventsListFiltersState();
	paginator = new QueryPaginatorState();
	private queryOptions = $derived(listEventsOptions({ 
		query: {
			...this.filters.queryData,
			limit: this.paginator.limit,
			offset: this.paginator.offset,
		}
	}));
	query = createQuery(() => ({
		...this.queryOptions,
		enabled: this.filters.queryEnabled,
	}));

    constructor() {
	    this.paginator.watchQuery(this.query);
    }

	events = $derived(this.query.data?.data ?? []);
}

const ctx = new Context<EventsListController>("EventsListController");
export const initEventsListController = () => ctx.set(new EventsListController());
export const useEventsListController = () => ctx.get();
