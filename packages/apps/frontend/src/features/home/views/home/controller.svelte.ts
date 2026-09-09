import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { page } from "$app/state";
import { resolve } from "$app/paths";
import { z } from "zod";
import { createQuery } from "@tanstack/svelte-query";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { getIncidentMetadataOptions, listIncidentsOptions, listSituationsOptions } from "$lib/api";
import {
	incidentFilterSchema,
	incidentQueryFilters,
	type IncidentFilters,
} from "$features/incidents/lib/filters";
import {
	situationFilterSchema,
	situationQueryFilters,
	type SituationFilters,
} from "$features/situations/lib/filters";
import { eventFilterSchema, filteredEventsOptions, type EventFilters } from "$features/events/lib/filters";

export const homeParamsSchema = z.object({
	incidentSearch: incidentFilterSchema.shape.search,
	incidentStatus: incidentFilterSchema.shape.status,
	incidentSeverity: incidentFilterSchema.shape.severityId,
	situationSearch: situationFilterSchema.shape.search,
	situationStatus: situationFilterSchema.shape.status,
	eventKind: eventFilterSchema.shape.kind,
	eventTime: eventFilterSchema.shape.time,
});

class HomeController {
	constructor() {
		onMount(() => {
			return () => this.params.cleanup();
		});
	}

	private params = useSearchParams(homeParamsSchema, { debounce: 300, noScroll: true });
	private committed = $derived(homeParamsSchema.parse(Object.fromEntries(page.url.searchParams)));

	incidents = $derived({
		search: this.params.incidentSearch,
		status: this.params.incidentStatus,
		severityId: this.params.incidentSeverity,
	});

	situations = $derived({ search: this.params.situationSearch, status: this.params.situationStatus });
	events = $derived({ kind: this.params.eventKind, time: this.params.eventTime });

	metadataQuery = createQuery(() => getIncidentMetadataOptions());
	incidentsPaginatedQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) =>
			listIncidentsOptions({
				query: {
					...incidentQueryFilters({
						search: this.committed.incidentSearch,
						status: this.committed.incidentStatus,
						severityId: this.committed.incidentSeverity,
					}),
					...pagination,
					pageSize: 5,
				},
			}),
	});
	incidentsQuery = $derived(this.incidentsPaginatedQuery.query);

	situationsPaginatedQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) =>
			listSituationsOptions({
				query: {
					...situationQueryFilters({
						search: this.committed.situationSearch,
						status: this.committed.situationStatus,
					}),
					...pagination,
					pageSize: 5,
				},
			}),
	});
	situationsQuery = $derived(this.situationsPaginatedQuery.query);

	eventsPaginatedQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) =>
			filteredEventsOptions(
				{ kind: this.committed.eventKind, time: this.committed.eventTime },
				{ ...pagination, pageSize: 5 }
			),
	});
	eventsQuery = $derived(this.eventsPaginatedQuery.query);

	incidentsHref = $derived(
		resolve("/incidents") + "?" + new URLSearchParams({ ...this.incidents, page: "1" })
	);

	situationsHref = $derived(
		resolve("/situations") + "?" + new URLSearchParams({ ...this.situations, page: "1" })
	);

	eventsHref = $derived(resolve("/events") + "?" + new URLSearchParams({ ...this.events, page: "1" }));

	refreshing = $derived(
		this.incidentsQuery.isFetching || this.situationsQuery.isFetching || this.eventsQuery.isFetching
	);

	setIncidents = (values: Partial<IncidentFilters>) => {
		const next = { ...this.incidents, ...values };
		this.params.update({
			incidentSearch: next.search,
			incidentStatus: next.status,
			incidentSeverity: next.severityId,
		});
	};

	setSituations = (values: Partial<SituationFilters>) => {
		const next = { ...this.situations, ...values };
		this.params.update({ situationSearch: next.search, situationStatus: next.status });
	};

	setEvents = (values: Partial<EventFilters>) => {
		const next = { ...this.events, ...values };
		this.params.update({ eventKind: next.kind, eventTime: next.time });
	};
	resetIncidents = () => this.setIncidents(incidentFilterSchema.parse({}));
	resetSituations = () => this.setSituations(situationFilterSchema.parse({}));
	resetEvents = () => this.setEvents(eventFilterSchema.parse({}));

	refresh = () => {
		void this.incidentsQuery.refetch();
		void this.situationsQuery.refetch();
		void this.eventsQuery.refetch();
		void this.metadataQuery.refetch();
	};
}

const ctx = new Context<HomeController>("HomeController");
export const initHomeController = () => ctx.set(new HomeController());
export const useHomeController = () => ctx.get();
