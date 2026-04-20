import {
	getIncidentMetadataOptions,
	listIncidentsOptions,
	type Incident,
	type IncidentAttributes,
	type ListIncidentsData,
} from "$lib/api";
import { QueryPaginatorState } from "$lib/paginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

type FilterOption = {label: string; value: any};

type IncidentFilters = {
	search?: string;
	includeArchived?: boolean;
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
}

type MetadataOption = {
	id: string;
	attributes: { name: string } | { value: string };
}
const mapNamedMetadataOptions = (values?: MetadataOption[]): FilterOption[] => 
	(values ?? [])
	.map(({id, attributes: a}) => ({
		value: id, 
		label: ("name" in a ? a.name : a.value)
	}));

const getLabel = (opts: FilterOption[], val?: any) => {
	if (!val) return "Any";
	return opts.find(o => (o.value === val))?.label || "Any";
}

class IncidentsListViewController {
	paginator = new QueryPaginatorState();

	private incidentMetadataQuery = createQuery(() => getIncidentMetadataOptions());
	private incidentMetadata = $derived(this.incidentMetadataQuery.data?.data);

	severityOptions = $derived(mapNamedMetadataOptions(this.incidentMetadata?.severities));
	typeOptions = $derived(mapNamedMetadataOptions(this.incidentMetadata?.types));
	tagOptions = $derived(mapNamedMetadataOptions(this.incidentMetadata?.tags));

	filters = $state<IncidentFilters>({});
	statusFilterLabel = $derived(getLabel(incidentStatusOptions, this.filters.status));
	severityFilterLabel = $derived(getLabel(this.severityOptions, this.filters.severity));
	typeFilterLabel = $derived(getLabel(this.typeOptions, this.filters.type));
	tagFilterLabel = $derived(getLabel(this.tagOptions, this.filters.tag));

	activeFilterCount = $derived(getActiveFilterCount(this.filters));

	private queryParams = $derived<ListIncidentsData["query"]>({
		search: this.filters.search,
		archived: this.filters.includeArchived,
		...this.paginator.queryParams,
	});

	incidentsQuery = createQuery(() => listIncidentsOptions({ query: this.queryParams }));
	incidents = $derived(this.incidentsQuery.data?.data ?? []);

	constructor() {
		this.paginator.watchQuery(this.incidentsQuery);
	}

	resetFilters = () => {
		this.filters = {};
	};
}

const ctx = new Context<IncidentsListViewController>("IncidentsListViewController");
export const initIncidentsListViewController = () => ctx.set(new IncidentsListViewController());
export const useIncidentsListView = () => ctx.get();