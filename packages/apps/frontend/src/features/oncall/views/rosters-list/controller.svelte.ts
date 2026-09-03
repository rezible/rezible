import { listOncallRostersOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context } from "runed";

export class OncallRostersListController {
	search = $state<string>();

	paginatedRostersQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listOncallRostersOptions({
				query: { search: this.search, ...pagination },
			}),
		resetWhen: () => this.search,
	});

	query = $derived(this.paginatedRostersQuery.query);
}

const ctx = new Context<OncallRostersListController>("OncallRostersListController");
export const initOncallRostersListController = () => ctx.set(new OncallRostersListController());
export const useOncallRostersListController = () => ctx.get();
