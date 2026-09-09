import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { eventFilterSchema, filteredEventsOptions, type EventFilters } from "$features/events/lib/filters";

class EventsListController {
	constructor() {
		onMount(() => {
			return () => this.params.cleanup();
		});
	}

	private params = useSearchParams(eventFilterSchema, {
		debounce: 300,
		noScroll: true,
	});

	filters = $derived(eventFilterSchema.parse(this.params));
	private committed = $derived(eventFilterSchema.parse(Object.fromEntries(page.url.searchParams)));
	private filterKey = $derived(JSON.stringify(this.committed));

	paginatedEventsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			filteredEventsOptions(this.committed, { ...pagination, withProjection: true }),
		resetWhen: () => this.filterKey,
	});

	query = $derived(this.paginatedEventsQuery.query);
	setFilters = (values: Partial<EventFilters>) => this.params.update(values);
	resetFilters = () => this.setFilters(eventFilterSchema.parse({}));
}

const ctx = new Context<EventsListController>("EventsListController");
export const initEventsListController = () => ctx.set(new EventsListController());
export const useEventsListController = () => ctx.get();
