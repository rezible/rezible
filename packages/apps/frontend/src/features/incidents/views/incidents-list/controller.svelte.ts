import { Context, watch } from "runed";
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
import { useIncidentCreateDialog } from "$features/incidents/components/create-incident-dialog";

class IncidentsListViewController {
	private incidentDialog = useIncidentCreateDialog();
	private declareRequested = $derived(page.url.searchParams.get("declare") === "incident");

	constructor() {
		onMount(() => {
			return () => this.params.cleanup();
		});
		// ?declare=incident opens the declare dialog, empty, as the page's button does. The parameter is cleared
		// straight away, so going back doesn't reopen it.
		watch(
			() => this.declareRequested,
			(requested) => {
				if (!requested) {
					return;
				}
				if (this.incidentDialog.enabled) {
					this.incidentDialog.openCreate();
				}
				this.incidentDialog.clearDeclareParam();
			}
		);
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
