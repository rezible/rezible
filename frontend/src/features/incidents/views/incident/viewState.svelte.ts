import { getIncidentOptions, getRetrospectiveForIncidentOptions, type Retrospective } from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

export class IncidentViewState {
	incidentIdParam = $state("");

	constructor(idFn: () => string) {
		watch(idFn, id => {this.incidentIdParam = id});
	}

	private incidentQuery = createQuery(() => ({
		...getIncidentOptions({ path: { id: this.incidentIdParam } }),
		enabled: !!this.incidentIdParam,
	}));
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");

	// TODO: derive from incident?
	timezone = $derived(getLocalTimeZone());

	private retroQuery = createQuery(() => ({
		...getRetrospectiveForIncidentOptions({ path: { id: this.incidentId } }),
		enabled: !!this.incidentId,
	}));
	retrospective = $derived(this.retroQuery.data?.data);
	retrospectiveId = $derived(this.retrospective?.id);

	systemAnalysisId = $derived(this.retrospective?.attributes.systemAnalysisId);

	// TODO: properly check response
	retrospectiveNeedsCreating = $derived(this.incidentQuery.isSuccess && this.retroQuery.isError);
	createRetrospectiveDialogOpen = $state(false);

	onRetrospectiveCreated(retro: Retrospective) {
		this.retroQuery.refetch();
		this.createRetrospectiveDialogOpen = false;
	}
}

const incidentViewCtx = new Context<IncidentViewState>("incidentView");
export const setIncidentViewState = (s: IncidentViewState) => incidentViewCtx.set(s);
export const useIncidentViewState = () => incidentViewCtx.get();