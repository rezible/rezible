import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { listSituationsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import {
	situationFilterSchema,
	situationQueryFilters,
	type SituationFilters,
} from "$features/situations/lib/filters";

class SituationsListController {
	constructor() {
		onMount(() => {
			return () => this.params.cleanup();
		});
	}

	private params = useSearchParams(situationFilterSchema, {
		debounce: 300,
		noScroll: true,
	});

	filters = $derived(situationFilterSchema.parse(this.params));
	private committed = $derived(situationFilterSchema.parse(Object.fromEntries(page.url.searchParams)));
	private filterKey = $derived(JSON.stringify(this.committed));

	paginatedSituationsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listSituationsOptions({ query: { ...situationQueryFilters(this.committed), ...pagination } }),
		resetWhen: () => this.filterKey,
	});

	query = $derived(this.paginatedSituationsQuery.query);
	setFilters = (values: Partial<SituationFilters>) => this.params.update(values);
	resetFilters = () => this.setFilters(situationFilterSchema.parse({}));
}

const ctx = new Context<SituationsListController>("SituationsListController");
export const initSituationsListController = () => ctx.set(new SituationsListController());
export const useSituationsListController = () => ctx.get();
