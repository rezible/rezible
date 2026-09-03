import { listAgentSessionsOptions } from "$lib/api";
import { Context } from "runed";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";

export class SessionsListController {
	paginatedSessionsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listAgentSessionsOptions({ 
			query: {
				...pagination,
			},
		}),
	});
	sessions = $derived(this.paginatedSessionsQuery.query.data?.data ?? []);
}
const context = new Context<SessionsListController>("SessionsListController");
export const initSessionsListController = () => context.set(new SessionsListController());
