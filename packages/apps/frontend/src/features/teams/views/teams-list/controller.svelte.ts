import { listTeamsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context } from "runed";

export class TeamsListController {
	search = $state<string>();

	paginatedTeamsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listTeamsOptions({
				query: { search: this.search, ...pagination },
			}),
		resetWhen: () => this.search,
	});

	query = $derived(this.paginatedTeamsQuery.query);
}

const ctx = new Context<TeamsListController>("TeamsListController");
export const initTeamsListController = () => ctx.set(new TeamsListController());
export const useTeamsListController = () => ctx.get();
