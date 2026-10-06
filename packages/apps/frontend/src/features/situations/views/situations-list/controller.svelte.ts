import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { listSituationsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import {
	situationFilterSchema,
	situationQueryFilters,
	situationSearchFilter,
	situationTab,
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
		// Rows from another tab must never show under the selected one while it loads.
		keepPreviousQueryData: false,
	});

	query = $derived(this.paginatedSituationsQuery.query);
	activeTab = $derived(situationTab(this.committed.tab));

	/** Render time for the long-running marker, refreshed with each fetch. */
	now = $derived.by(() => {
		void this.query.dataUpdatedAt;
		return Date.now();
	});

	private watchingPreview = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listSituationsOptions({
				query: {
					page: 1,
					pageSize: 1,
					stage: ["candidate"],
					muted: false,
					search: situationSearchFilter(this.committed.search),
				},
			}),
	});

	/** Undefined until the preview has loaded. */
	watchingCount = $derived(this.watchingPreview.query.data?.pagination.total);

	setFilters = (values: Partial<SituationFilters>) => this.params.update(values);
	setTab = (value: string) => this.setFilters({ tab: situationFilterSchema.shape.tab.parse(value) });
}

const ctx = new Context<SituationsListController>("SituationsListController");
export const initSituationsListController = () => ctx.set(new SituationsListController());
export const useSituationsListController = () => ctx.get();
