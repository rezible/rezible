import { listOncallShiftsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context } from "runed";

export class OncallShiftsListController {
	paginatedShiftsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listOncallShiftsOptions({ query: pagination }),
	});

	query = $derived(this.paginatedShiftsQuery.query);
	total = $derived(this.query.data?.pagination.total);

	/** Render time for each shift's status, refreshed with each fetch. */
	now = $derived.by(() => {
		void this.query.dataUpdatedAt;
		return Date.now();
	});
}

const ctx = new Context<OncallShiftsListController>("OncallShiftsListController");
export const initOncallShiftsListController = () => ctx.set(new OncallShiftsListController());
export const useOncallShiftsListController = () => ctx.get();
