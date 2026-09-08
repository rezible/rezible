import { useAlertViewController } from "$features/signals/views/signal";
import { listEventsOptions, type ListEventsData, type EventAttributes } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { getLocalTimeZone, now } from "@internationalized/date";

export type EventKind = EventAttributes["kind"];

const defaultDateRange = () => {
	return {
		from: now(getLocalTimeZone()).subtract({ days: 7 }).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: "day",
	};
};

export class AlertEventsViewController {
	view = useAlertViewController();

	rosterId = $state<string>();
	eventKind = $state<EventKind>();
	dateRange = $state(defaultDateRange());

	queryParams = $derived<ListEventsData["query"]>({
		// alertId: this.viewState.alertId,
	});
	paginatedEventsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listEventsOptions({
				query: {
					...this.queryParams,
					...pagination,
				},
			}),
	});
}
