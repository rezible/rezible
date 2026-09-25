import { SvelteSet } from "svelte/reactivity";
import { createQueries, createQuery } from "@tanstack/svelte-query";
import { page } from "$app/state";
import {
	getIncidentOptions,
	getInvestigationOptions,
	getInvestigationReportOptions,
	getSituationOptions,
} from "$lib/api";
import { Context, watch, type Getter } from "runed";
import { situationHref } from "../../lib/routes";

const idPath = (id?: string) => ({ id: id ?? "" });

export class SituationController {
	situationId = $state<string>();

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.situationId = id;
		});
	}

	situationQuery = createQuery(() => ({
		...getSituationOptions({ path: idPath(this.situationId) }),
		enabled: !!this.situationId,
	}));

	situation = $derived(this.situationQuery.data?.data);
	incidentIds = $derived([...new SvelteSet(this.situation?.attributes?.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	situationInvestigation = $derived(this.situation?.attributes.investigation);
	investigationId = $derived(this.situationInvestigation?.investigation.id);
	investigationQuery = createQuery(() => ({
		...getInvestigationOptions({ path: idPath(this.investigationId) }),
		enabled: !!this.investigationId,
		refetchInterval(query) {
			return ["queued", "running"].includes(query.state.data?.data.attributes.latestTurnStatus ?? "")
				? 2000
				: 30000;
		},
	}));
	investigation = $derived(this.investigationQuery.data?.data);
	investigationReportQuery = createQuery(() => ({
		...getInvestigationReportOptions({ path: idPath(this.investigationId) }),
		enabled: !!this.investigationId,
		refetchInterval: ["queued", "running"].includes(this.investigation?.attributes.latestTurnStatus ?? "")
			? 2000
			: 30000,
		retry: (failureCount, error) => error.status !== 404 && failureCount < 3,
	}));
	investigationReport = $derived(this.investigationReportQuery.data?.data);

	private id = $derived(this.situationId ?? "");
	investigationHref = $derived(situationHref(this.id, "investigation", page.url.search));
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
