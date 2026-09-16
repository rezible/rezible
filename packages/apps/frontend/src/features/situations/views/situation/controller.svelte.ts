import { SvelteSet } from "svelte/reactivity";
import { createQueries, createQuery } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { getIncidentOptions, getSituationOptions, type Situation } from "$lib/api";
import { Context, watch, type Getter } from "runed";
import { situationHref } from "../../lib/routes";

const idPath = (id?: string) => ({ id: id ?? "" });

const isInvestigationPending = (situation?: Situation) => {
	if (!situation) return false;
	const attrs = situation.attributes;
	const sitInv = attrs?.investigation;
	if (!sitInv) return true;
	if (sitInv.requestedRevision > sitInv.completedRevision) return true;
	if (attrs.evidenceRevision > sitInv.completedRevision) return true;
	const invAttrs = sitInv.investigation.attributes;
	const reportMissing = !!attrs?.investigation && !invAttrs?.report;
	return reportMissing;
}

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
		refetchInterval(q) {
			return (isInvestigationPending(q.state.data?.data)) ? 2000 : false
		},
	}));

	situation = $derived(this.situationQuery.data?.data);
	incidentIds = $derived([...new SvelteSet(this.situation?.attributes?.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	investigationPending = $derived(isInvestigationPending(this.situation));
	situationInvestigation = $derived(this.situation?.attributes.investigation);
	investigation = $derived(this.situationInvestigation?.investigation);
	investigationReport = $derived(this.investigation?.attributes?.report);
	investigationAnalysisId = $derived(this.investigation?.attributes?.analysisId);

	private id = $derived(this.situationId ?? "");
	investigationHref = $derived(situationHref(this.id, "investigation", page.url.search));
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
