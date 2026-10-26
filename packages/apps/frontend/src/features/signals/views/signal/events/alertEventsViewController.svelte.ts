import { listEventsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { eventRowPresentation } from "$features/signals/lib/events";

export class AlertEventsViewController {
	paginatedEventsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listEventsOptions({ query: { ...pagination } }),
	});

	query = $derived(this.paginatedEventsQuery.query);

	rows = $derived.by(() => {
		const events = this.query.data?.data;
		if (!events) {
			return undefined;
		}
		return events.map(eventRowPresentation);
	});
}
