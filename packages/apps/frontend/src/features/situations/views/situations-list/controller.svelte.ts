import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { listSituationsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import {
	situationFilterSchema,
	situationQueryFilters,
	situationSearchFilter,
	situationStatusOption,
	situationStatusOptions,
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
		// Rows from another status must never show under the selected one while it loads.
		keepPreviousQueryData: false,
	});

	query = $derived(this.paginatedSituationsQuery.query);
	activeStatus = $derived(situationStatusOption(this.committed.status));

	/** Render time for the long-running marker, refreshed with each fetch. */
	now = $derived.by(() => {
		void this.query.dataUpdatedAt;
		return Date.now();
	});

	private statusTotals = situationStatusOptions.map((option) =>
		createQuery(() =>
			listSituationsOptions({
				query: {
					...option.query,
					search: situationSearchFilter(this.committed.search),
					page: 1,
					pageSize: 1,
				},
			})
		)
	);

	/** Status options with how many situations each would show for the current search. */
	statusOptions = $derived(
		situationStatusOptions.map((option, index) => ({
			value: option.value,
			label: option.label,
			count: this.statusTotals[index].data?.pagination.total,
		}))
	);

	total = $derived(this.query.data?.pagination.total);
	activeFilterCount = $derived(situationSearchFilter(this.filters.search) ? 1 : 0);

	setFilters = (values: Partial<SituationFilters>) => this.params.update(values);
	clearSearch = () => this.setFilters({ search: "" });
}

const ctx = new Context<SituationsListController>("SituationsListController");
export const initSituationsListController = () => ctx.set(new SituationsListController());
export const useSituationsListController = () => ctx.get();
