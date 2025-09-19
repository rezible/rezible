import { useAlertViewState } from "$features/alert";
import { listEventsOptions, type ListEventsData, type Event, type EventAttributes } from "$lib/api";
import { QueryPaginatorState } from "$lib/paginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { type DateRange as DateRangeType } from '@layerstack/utils/dateRange';
import { getLocalTimeZone, now } from "@internationalized/date";
import { PeriodType } from "@layerstack/utils";

export type EventKind = EventAttributes["kind"];

const defaultDateRange = (): DateRangeType => {
	return { 
		from: now(getLocalTimeZone()).subtract({days: 7}).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: PeriodType.Day,
	}
}

export class AlertEventsState {
	viewState = useAlertViewState();

	paginator = new QueryPaginatorState();

	rosterId = $state<string>();
	eventKind = $state<EventKind>();
	dateRange = $state<DateRangeType>(defaultDateRange());

	queryParams = $derived<ListEventsData["query"]>({
		// alertId: this.viewState.alertId,
		...this.paginator.queryParams,
	});
	query = createQuery(() => listEventsOptions({ query: this.queryParams }));

	constructor() {
		this.paginator.watchQuery(this.query);
	}
};
