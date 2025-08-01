import { getIncidentOptions, getRetrospectiveForIncidentOptions, type Retrospective } from "$lib/api";
import type { Getter } from "$src/lib/utils.svelte";
import type { IncidentViewRouteParam } from "$src/params/incidentView";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

type StateParams = {slug: string, viewRouteParam: IncidentViewRouteParam};

export class IncidentViewState {
	incidentSlug = $state<string>(null!);
	viewRouteParam = $state<IncidentViewRouteParam>(null!);

	private setParams({slug, viewRouteParam}: StateParams) {
		this.incidentSlug = slug;
		this.viewRouteParam = viewRouteParam;
	}

	constructor(paramsFn: Getter<StateParams>) {
		this.setParams(paramsFn());

		watch(paramsFn, p => {this.setParams(p)});
	}

	private incidentQuery = createQuery(() => getIncidentOptions({ path: { id: this.incidentSlug } }));
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
export const setIncidentViewState = (paramsFn: Getter<StateParams>) => incidentViewCtx.set(new IncidentViewState(paramsFn));
export const useIncidentViewState = () => incidentViewCtx.get();