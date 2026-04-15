import { getIncidentOptions, getRetrospectiveOptions } from "$lib/api";
import type { IncidentViewRouteParam } from "$params/incidentView";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";

export type IncidentViewParams = {
	slug: string;
	routeParam: IncidentViewRouteParam;
}

export class IncidentViewController {
	queryClient = useQueryClient();

	slug = $state<string>(null!);
	routeParam = $state<IncidentViewRouteParam>(null!);

	constructor(paramsFn: Getter<IncidentViewParams>) {
		watch(paramsFn, ({slug, routeParam}) => {
			this.slug = slug;
			this.routeParam = routeParam;
		});
	}

	private incidentQueryOptions = $derived(getIncidentOptions({ path: { id: this.slug } }));
	private incidentQuery = createQuery(() => this.incidentQueryOptions);
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");

	// TODO: get from incident?
	timezone = $derived(getLocalTimeZone());

	private incRetroId = $derived(this.incident?.attributes.retrospectiveId || "");
	private retroQueryOptions = $derived(getRetrospectiveOptions({ path: { id: this.incRetroId } }));
	private retroQuery = createQuery(() => ({
		...this.retroQueryOptions,
		enabled: !!this.incRetroId,
	}));
	retrospective = $derived(this.retroQuery.data?.data);
	retrospectiveId = $derived(this.retrospective?.id);
	documentId = $derived(this.retrospective?.attributes.documentId);
	systemAnalysisId = $derived(this.retrospective?.attributes.systemAnalysisId);
}

const ctx = new Context<IncidentViewController>("IncidentViewController");
export const initIncidentViewController = (paramsFn: Getter<IncidentViewParams>) => ctx.set(new IncidentViewController(paramsFn));
export const useIncidentView = () => ctx.get();