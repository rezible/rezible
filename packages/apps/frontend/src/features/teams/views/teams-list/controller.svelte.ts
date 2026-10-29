import { listTeamsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context, Debounced } from "runed";

export class TeamsListController {
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, 300);
	private searchQuery = $derived(this.debouncedSearch.current.trim() || undefined);

	paginatedTeamsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listTeamsOptions({
				query: { search: this.searchQuery, ...pagination },
			}),
		resetWhen: () => this.searchQuery,
	});

	query = $derived(this.paginatedTeamsQuery.query);
	total = $derived(this.query.data?.pagination.total);

	setSearch = (value: string) => {
		this.search = value;
	};
}

const ctx = new Context<TeamsListController>("TeamsListController");
export const initTeamsListController = () => ctx.set(new TeamsListController());
export const useTeamsListController = () => ctx.get();
