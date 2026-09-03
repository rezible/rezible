import { listPlaybooksOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context } from "runed";

export class PlaybooksListController {
	search = $state<string>();

	paginatedPlaybooksQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listPlaybooksOptions({
				query: { search: this.search, ...pagination },
			}),
		resetWhen: () => this.search,
	});

	query = $derived(this.paginatedPlaybooksQuery.query);
}

const ctx = new Context<PlaybooksListController>("PlaybooksListController");
export const initPlaybooksListController = () => ctx.set(new PlaybooksListController());
export const usePlaybooksListController = () => ctx.get();
