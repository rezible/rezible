import { useAlertViewController } from "$features/alerts/views/alert";
import { listAlertIncidentLinksOptions, type ListIncidentsData } from "$lib/api";
import { QueryPaginatorState } from "$lib/paginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { getLocalTimeZone, now } from "@internationalized/date";

const defaultDateRange = () => {
	return { 
		from: now(getLocalTimeZone()).subtract({days: 7}).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: "day",
	}
}

export class AlertIncidentsViewController {
	view = useAlertViewController();

	paginator = new QueryPaginatorState();

	rosterId = $state<string>();
	dateRange = $state(defaultDateRange());

	queryParams = $derived<ListIncidentsData["query"]>({
		...this.paginator.queryParams,
	});
	query = createQuery(() => listAlertIncidentLinksOptions({ path: {id: this.view.alertId}, query: this.queryParams }));

	constructor() {
		this.paginator.watchQuery(this.query);
	}
};
