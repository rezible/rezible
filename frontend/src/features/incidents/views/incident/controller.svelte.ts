import { createRetrospectiveMutation, getIncidentOptions, getRetrospectiveOptions, type CreateRetrospectiveResponseBody } from "$lib/api";
import type { Getter } from "$src/lib/utils.svelte";
import type { IncidentViewRouteParam } from "$src/params/incidentView";
import { getLocalTimeZone } from "@internationalized/date";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { RetrospectiveCollaborationState } from "$features/incidents/lib/collaborationState.svelte";

export class IncidentViewController {
	queryClient = useQueryClient();

	incidentSlug = $state<string>(null!);
	viewRouteParam = $state<IncidentViewRouteParam>(null!);

	constructor(slugFn: Getter<string>, viewFn: Getter<IncidentViewRouteParam>) {
		this.incidentSlug = slugFn();
		watch(slugFn, slug => {this.incidentSlug = slug});
		watch(viewFn, view => {this.viewRouteParam = view});

		watch(() => this.retroNeedsCreating, create => {
			if (create) this.maybeCreateRetrospective();
		});
	}

	private incidentQueryOptions = $derived(getIncidentOptions({ path: { id: this.incidentSlug } }));
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

	collaboration = new RetrospectiveCollaborationState(() => (this.documentId));

	onRetrospectiveCreated(resp: CreateRetrospectiveResponseBody) {
		const id = resp.data.id;
		this.queryClient.setQueryData(getRetrospectiveOptions({ path: { id } }).queryKey, resp);
		this.queryClient.setQueryData(this.incidentQueryOptions.queryKey, body => {
			if (!body) return;
			body.data.attributes.retrospectiveId = id;
			return body;
		});
		this.queryClient.invalidateQueries(this.incidentQueryOptions);
	}

	retroNeedsCreating = $derived(!this.incRetroId && this.incidentQuery.isSuccess);
	private createRetrospectiveMut = createMutation(() => ({
		...createRetrospectiveMutation(),
		onSuccess: resp => {this.onRetrospectiveCreated(resp)},
	}));
	private maybeCreateRetrospective() {
		// TODO: allow configuring retrospective type
		this.createRetrospectiveMut.mutate({
			body: {
				attributes: {
					incidentId: this.incidentId,
					systemAnalysis: true,
				}
			}
		});
	}
}

const ctx = new Context<IncidentViewController>("IncidentViewController");
export const initIncidentViewController = (idFn: Getter<string>, viewFn: Getter<IncidentViewRouteParam>) => ctx.set(new IncidentViewController(idFn, viewFn));
export const useIncidentViewController = () => ctx.get();