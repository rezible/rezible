import { useAlertViewState } from "$features/alert";
import { listEventsOptions, type ListEventsData, type Event, type EventAttributes } from "$lib/api";
import { QueryPaginatorState } from "$lib/paginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { getLocalTimeZone, now } from "@internationalized/date";

export type EventKind = EventAttributes["kind"];

const defaultDateRange = () => {
	return { 
		from: now(getLocalTimeZone()).subtract({days: 7}).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: "day",
	}
}

export class AlertEventsState {
	viewState = useAlertViewState();

	paginator = new QueryPaginatorState();

	rosterId = $state<string>();
	eventKind = $state<EventKind>();
	dateRange = $state(defaultDateRange());

	queryParams = $derived<ListEventsData["query"]>({
		// alertId: this.viewState.alertId,
		...this.paginator.queryParams,
	});
	query = createQuery(() => listEventsOptions({ query: this.queryParams }));

	constructor() {
		this.paginator.watchQuery(this.query);
	}
};
