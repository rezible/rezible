import {
	getIncidentMetadataOptions,
	listIncidentsOptions,
	type IncidentAttributes,
} from "$lib/api";
import { QueryPaginator, createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { createQuery, keepPreviousData } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

type FilterOption = { label: string; value: any };

type IncidentFilters = {
	search?: string;
	status?: IncidentAttributes["currentStatus"];
	severity?: string;
	type?: string;
	tag?: string;
};

export const incidentStatusOptions: FilterOption[] = [
	{ label: "Started", value: "started" },
	{ label: "Mitigated", value: "mitigated" },
	{ label: "Resolved", value: "resolved" },
	{ label: "Closed", value: "closed" },
];

const getActiveFilterCount = (f: IncidentFilters) => {
	let count = 0;
	if (!!f.search) count++;
	// TODO
	return count;
};

type MetadataOption = {
	id: string;
	attributes: { name: string } | { value: string };
};
const mapNamedMetadataOptions = (values?: MetadataOption[]): FilterOption[] =>
	(values ?? []).map(({ id, attributes: a }) => ({
		value: id,
		label: "name" in a ? a.name : a.value,
	}));

const getLabel = (opts: FilterOption[], val?: any) => {
	if (!val) return "Any";
	return opts.find((o) => o.value === val)?.label || "Any";
};

class IncidentsListViewController {
	filters = $state<IncidentFilters>({});

	private incidentMetadataQuery = createQuery(() => getIncidentMetadataOptions());
	private incidentMetadata = $derived(this.incidentMetadataQuery.data?.data);
	
	statusFilterLabel = $derived(getLabel(incidentStatusOptions, this.filters.status));
	
	severityOptions = $derived(mapNamedMetadataOptions(this.incidentMetadata?.severities));
	severityFilterLabel = $derived(getLabel(this.severityOptions, this.filters.severity));
	
	typeOptions = $derived(mapNamedMetadataOptions(this.incidentMetadata?.types));
	typeFilterLabel = $derived(getLabel(this.typeOptions, this.filters.type));
	
	tagOptions = $derived(mapNamedMetadataOptions(this.incidentMetadata?.tags));
	tagFilterLabel = $derived(getLabel(this.tagOptions, this.filters.tag));

	activeFilterCount = $derived(getActiveFilterCount(this.filters));

	paginatedIncidentsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listIncidentsOptions({ 
			query: {
				search: this.filters.search,
				...pagination,
			},
		}),
		resetWhen: () => $state.snapshot(this.filters),
	});
	incidentsQuery = $derived(this.paginatedIncidentsQuery.query);
	incidents = $derived(this.incidentsQuery.data?.data ?? []);

	resetFilters = () => {
		this.filters = {};
	};
}

const ctx = new Context<IncidentsListViewController>("IncidentsListViewController");
export const initIncidentsListViewController = () => ctx.set(new IncidentsListViewController());
export const useIncidentsListView = () => ctx.get();
