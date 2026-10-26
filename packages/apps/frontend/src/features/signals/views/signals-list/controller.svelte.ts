import { Context } from "runed";
import { listAlertsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";

class SignalsListViewController {
	search = $state<string>();

	paginatedAlertsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listAlertsOptions({ query: { search: this.search, ...pagination } }),
		resetWhen: () => this.search,
	});

	query = $derived(this.paginatedAlertsQuery.query);
}

const ctx = new Context<SignalsListViewController>("SignalsListViewController");
export const initSignalsListViewController = () => ctx.set(new SignalsListViewController());
