import { Context, watch } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { listIncidentsOptions, getIncidentMetadataOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import {
	incidentActiveFilterCount,
	incidentFilterSchema,
	incidentQueryFilters,
	incidentStatusOptions,
	type IncidentFilters,
	type IncidentStatusFilter,
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
	activeFilterCount = $derived(incidentActiveFilterCount(this.filters));
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
	total = $derived(this.query.data?.pagination.total);

	private statusTotals = incidentStatusOptions.map((option) => this.statusTotalQuery(option.value));

	/** Status options with how many incidents each would show under the other filters. */
	statusOptions = $derived(
		incidentStatusOptions.map((option, index) => ({
			...option,
			count: this.statusTotals[index].data?.pagination.total,
		}))
	);

	severityOptions = $derived([
		{ value: "", label: "All severities" },
		...(this.metadataQuery.data?.data.severities ?? [])
			.toSorted(
				(a, b) =>
					a.attributes.rank - b.attributes.rank ||
					a.attributes.name.localeCompare(b.attributes.name)
			)
			.map((severity) => ({ value: severity.id, label: severity.attributes.name })),
	]);

	private statusTotalQuery(status: IncidentStatusFilter) {
		return createQuery(() =>
			listIncidentsOptions({
				query: { ...incidentQueryFilters({ ...this.committed, status }), page: 1, pageSize: 1 },
			})
		);
	}
	setFilters = (values: Partial<IncidentFilters>) => this.params.update(values);
	resetFilters = () => this.setFilters(incidentFilterSchema.parse({}));
}

const ctx = new Context<IncidentsListViewController>("IncidentsListViewController");
export const initIncidentsListViewController = () => ctx.set(new IncidentsListViewController());
export const useIncidentsListViewController = () => ctx.get();
