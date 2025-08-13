import { useAlertViewState } from "$features/alert";
import { listIncidentsOptions, type ListIncidentsData } from "$lib/api";
import { QueryPaginatorState } from "$lib/paginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { type DateRange as DateRangeType } from '@layerstack/utils/dateRange';
import { getLocalTimeZone, now } from "@internationalized/date";
import { PeriodType } from "@layerstack/utils";

const defaultDateRange = (): DateRangeType => {
	return { 
		from: now(getLocalTimeZone()).subtract({days: 7}).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: PeriodType.Day,
	}
}

export class AlertIncidentsState {
	viewState = useAlertViewState();

	paginator = new QueryPaginatorState();

	rosterId = $state<string>();
	dateRange = $state<DateRangeType>(defaultDateRange());

	queryParams = $derived<ListIncidentsData["query"]>({
		// alertId: this.viewState.alertId,
		...this.paginator.queryParams,
	});
	query = createQuery(() => listIncidentsOptions({ query: this.queryParams }));

	constructor() {
		this.paginator.watchQuery(this.query);
	}
};
