import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { listIncidentsOptions, getIncidentMetadataOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import {
	incidentFilterSchema,
	incidentQueryFilters,
	type IncidentFilters,
} from "$features/incidents/lib/filters";
import { createQuery } from "@tanstack/svelte-query";

class IncidentsListViewController {
	constructor() {
		onMount(() => {
			return () => this.params.cleanup();
		});
	}

	private params = useSearchParams(incidentFilterSchema, {
		debounce: 300,
		noScroll: true,
	});

	filters = $derived(incidentFilterSchema.parse(this.params));
	private committed = $derived(incidentFilterSchema.parse(Object.fromEntries(page.url.searchParams)));
	private filterKey = $derived(JSON.stringify(this.committed));

	metadataQuery = createQuery(() => getIncidentMetadataOptions());
	paginatedIncidentsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listIncidentsOptions({
				query: { ...incidentQueryFilters(this.committed), ...pagination },
			}),
		resetWhen: () => this.filterKey,
	});

	query = $derived(this.paginatedIncidentsQuery.query);
	setFilters = (values: Partial<IncidentFilters>) => this.params.update(values);
	resetFilters = () => this.setFilters(incidentFilterSchema.parse({}));
}

const ctx = new Context<IncidentsListViewController>("IncidentsListViewController");
export const initIncidentsListViewController = () => ctx.set(new IncidentsListViewController());
export const useIncidentsListViewController = () => ctx.get();
