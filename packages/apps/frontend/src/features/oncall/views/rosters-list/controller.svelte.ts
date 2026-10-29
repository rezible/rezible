import { listOncallRostersOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context, Debounced } from "runed";

export class OncallRostersListController {
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, 300);
	private searchQuery = $derived(this.debouncedSearch.current.trim() || undefined);

	paginatedRostersQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listOncallRostersOptions({
				query: { search: this.searchQuery, ...pagination },
			}),
		resetWhen: () => this.searchQuery,
	});

	query = $derived(this.paginatedRostersQuery.query);
	total = $derived(this.query.data?.pagination.total);

	setSearch = (value: string) => {
		this.search = value;
	};
}

const ctx = new Context<OncallRostersListController>("OncallRostersListController");
export const initOncallRostersListController = () => ctx.set(new OncallRostersListController());
export const useOncallRostersListController = () => ctx.get();
