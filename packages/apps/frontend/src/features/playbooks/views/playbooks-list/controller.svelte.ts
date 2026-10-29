import { listPlaybooksOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context, Debounced } from "runed";

export class PlaybooksListController {
	search = $state("");
	private debouncedSearch = new Debounced(() => this.search, 300);
	private searchQuery = $derived(this.debouncedSearch.current.trim() || undefined);

	paginatedPlaybooksQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listPlaybooksOptions({
				query: { search: this.searchQuery, ...pagination },
			}),
		resetWhen: () => this.searchQuery,
	});

	query = $derived(this.paginatedPlaybooksQuery.query);
	total = $derived(this.query.data?.pagination.total);

	setSearch = (value: string) => {
		this.search = value;
	};
}

const ctx = new Context<PlaybooksListController>("PlaybooksListController");
export const initPlaybooksListController = () => ctx.set(new PlaybooksListController());
export const usePlaybooksListController = () => ctx.get();
