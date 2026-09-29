import { SvelteSet } from "svelte/reactivity";
import { createQueries, createQuery } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { getIncidentOptions, getSituationOptions } from "$lib/api";
import { Context, watch, type Getter } from "runed";
import { isDefinitiveUnavailableError } from "./model";
import { situationHref } from "../../lib/routes";

const idPath = (id?: string) => ({ id: id ?? "" });

export class SituationController {
	situationId = $state<string>();
	private overviewRefreshInterval: Getter<number> = () => 30000;

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.situationId = id;
		});
	}

	setOverviewRefreshInterval = (getInterval: Getter<number>) => {
		this.overviewRefreshInterval = getInterval;
	};

	situationQuery = createQuery(() => ({
		...getSituationOptions({ path: idPath(this.situationId) }),
		enabled: !!this.situationId,
		refetchInterval: () => (this.overviewRefreshInterval()),
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	situation = $derived(
		isDefinitiveUnavailableError(this.situationQuery.error) ? undefined : this.situationQuery.data?.data
	);
	situationUnavailable = $derived(isDefinitiveUnavailableError(this.situationQuery.error));
	incidentIds = $derived([...new SvelteSet(this.situation?.attributes.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	situationInvestigation = $derived(this.situation?.attributes.investigation);
	investigationId = $derived(this.situationInvestigation?.investigation.id);
	investigationHref = $derived(situationHref(this.situationId ?? "", "investigation", page.url.search));
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
