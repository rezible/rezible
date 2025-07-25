import { getIncidentOptions, getRetrospectiveForIncidentOptions, type Retrospective } from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class IncidentViewState {
	private incidentIdParam = $state<string>(null!);

	constructor(idParamFn: () => string) {
		this.incidentIdParam = idParamFn();
		watch(idParamFn, id => {this.incidentIdParam = id});
	}

	private incidentQuery = createQuery(() => getIncidentOptions({ path: { id: this.incidentIdParam } }));
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");

	// TODO: get from incident?
	timezone = $derived(getLocalTimeZone());

	private retroQuery = createQuery(() => ({
		...getRetrospectiveForIncidentOptions({ path: { id: this.incidentId } }),
		enabled: !!this.incidentId,
	}));
	retrospective = $derived(this.retroQuery.data?.data);
	retrospectiveId = $derived(this.retrospective?.id);
	systemAnalysisId = $derived(this.retrospective?.attributes.systemAnalysisId);
}

const incidentViewCtx = new Context<IncidentViewState>("incidentView");
export const setIncidentViewState = (s: IncidentViewState) => incidentViewCtx.set(s);
export const useIncidentViewState = () => incidentViewCtx.get();