import {
	listIncidentsOptions,
	type Incident,
	type IncidentAttributes,
	type ListIncidentsData,
} from "$lib/api";
import { QueryPaginatorState } from "$lib/paginator.svelte";
import { createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";

export type IncidentStatus = IncidentAttributes["currentStatus"];
export type ArchiveScope = "active" | "archived" | "all";
export type VisibilityFilter = "all" | "public" | "private";
export type FilterOption = { value: string; label: string };
export type StatusFilterOption = { value: IncidentStatus; label: string };

export const incidentStatusOptions: StatusFilterOption[] = [
	{ value: "started", label: "Started" },
	{ value: "mitigated", label: "Mitigated" },
	{ value: "resolved", label: "Resolved" },
	{ value: "closed", label: "Closed" },
];

export const archiveScopeOptions: { value: ArchiveScope; label: string }[] = [
	{ value: "all", label: "All" },
	{ value: "active", label: "Active" },
	{ value: "archived", label: "Archived" },
];

export const visibilityOptions: { value: VisibilityFilter; label: string }[] = [
	{ value: "all", label: "Any" },
	{ value: "public", label: "Public" },
	{ value: "private", label: "Private" },
];

const normalize = (value: string | undefined) => value?.trim().toLowerCase() ?? "";

const uniqueOptions = (values: FilterOption[]) => {
	const optionMap = new Map<string, string>();
	for (const option of values) {
		if (option.value) optionMap.set(option.value, option.label);
	}
	return [...optionMap.entries()]
		.map(([value, label]) => ({ value, label }))
		.sort((a, b) => a.label.localeCompare(b.label));
};

const hasValueAndLabel = (option: {
	value: string | undefined;
	label: string | undefined;
}): option is FilterOption => !!option.value && !!option.label;

class IncidentsListViewController {
	paginator = new QueryPaginatorState();

	searchValue = $state<string>();
	archiveScope = $state<ArchiveScope>("active");
	statusFilter = $state<IncidentStatus | "all">("all");
	severityFilter = $state("all");
	typeFilter = $state("all");
	tagFilter = $state("all");
	visibilityFilter = $state<VisibilityFilter>("all");

	statusOptions = incidentStatusOptions;
	archiveScopeOptions = archiveScopeOptions;
	visibilityOptions = visibilityOptions;

	private archivedParam = $derived(
		this.archiveScope === "all" ? undefined : this.archiveScope === "archived"
	);

	private queryParams = $derived<ListIncidentsData["query"]>({
		search: this.searchValue,
		archived: this.archivedParam,
		...this.paginator.queryParams,
	});

	incidentsQuery = createQuery(() => listIncidentsOptions({ query: this.queryParams }));

	incidents = $derived(this.incidentsQuery.data?.data ?? []);

	// private incidentMetadataQuery = createQuery(() => )

	severityOptions = $derived(
		uniqueOptions(
			this.incidents
				.map((incident) => ({
					value: incident.attributes.severity?.id,
					label: incident.attributes.severity?.attributes.name,
				}))
				.filter(hasValueAndLabel)
		)
	);

	typeOptions = $derived(
		uniqueOptions(
			this.incidents
				.map((incident) => ({
					value: incident.attributes.type?.id,
					label: incident.attributes.type?.attributes.name,
				}))
				.filter(hasValueAndLabel)
		)
	);

	tagOptions = $derived(
		uniqueOptions(
			this.incidents
				.flatMap(
					(incident) =>
						incident.attributes.tags?.map((tag) => ({
							value: tag.id,
							label: tag.attributes.value,
						})) ?? []
				)
				.filter(hasValueAndLabel)
		)
	);

	filteredIncidents = $derived(this.incidents.filter((incident) => this.matchesFilters(incident)));

	activeFilterCount = $derived(
		[
			this.searchValue,
			this.archiveScope !== "active" ? this.archiveScope : undefined,
			this.statusFilter !== "all" ? this.statusFilter : undefined,
			this.severityFilter !== "all" ? this.severityFilter : undefined,
			this.typeFilter !== "all" ? this.typeFilter : undefined,
			this.tagFilter !== "all" ? this.tagFilter : undefined,
			this.visibilityFilter !== "all" ? this.visibilityFilter : undefined,
		].filter(Boolean).length
	);

	activeStatusLabel = $derived(
		this.statusOptions.find((option) => option.value === this.statusFilter)?.label
	);

	constructor() {
		this.paginator.watchQuery(this.incidentsQuery);
	}

	resetFilters = () => {
		this.searchValue = undefined;
		this.archiveScope = "active";
		this.statusFilter = "all";
		this.severityFilter = "all";
		this.typeFilter = "all";
		this.tagFilter = "all";
		this.visibilityFilter = "all";
	};

	private matchesFilters({attributes: attrs}: Incident) {
		const searchNeedle = normalize(this.searchValue);
		const matchesSearch =
			!searchNeedle ||
			normalize(attrs.title).includes(searchNeedle) ||
			normalize(attrs.summary).includes(searchNeedle) ||
			normalize(attrs.slug).includes(searchNeedle);
		const matchesStatus = this.statusFilter === "all" || attrs.currentStatus === this.statusFilter;
		const matchesSeverity = this.severityFilter === "all" || attrs.severity?.id === this.severityFilter;
		const matchesType = this.typeFilter === "all" || attrs.type?.id === this.typeFilter;
		const matchesTag = this.tagFilter === "all" || attrs.tags?.some((tag) => tag.id === this.tagFilter);
		const matchesVisibility =
			this.visibilityFilter === "all" ||
			(this.visibilityFilter === "private" ? attrs.private : !attrs.private);

		return (
			matchesSearch &&
			matchesStatus &&
			matchesSeverity &&
			matchesType &&
			matchesTag &&
			matchesVisibility
		);
	}
}

const ctx = new Context<IncidentsListViewController>("IncidentsListViewController");
export const initIncidentsListViewController = () => ctx.set(new IncidentsListViewController());
export const useIncidentsListView = () => ctx.get();