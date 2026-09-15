import {
	getDocumentSessionOptions,
	getIncidentOptions,
	getRetrospectiveOptions,
	getSituationOptions,
} from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createQueries, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { initIncidentCollaborationController } from "./collaboration.svelte";

export class IncidentViewController {
	queryClient = useQueryClient();

	slug = $state<string>(null!);

	constructor(slugFn: Getter<string>) {
		watch(slugFn, (slug) => {
			this.slug = slug;
		});
	}

	incidentQuery = createQuery(() => getIncidentOptions({ path: { id: this.slug } }));
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");

	linkedSituationIds = $derived(this.incident?.attributes.linkedSituationIds ?? []);
	situationsQuery = createQueries(() => ({
		queries: this.linkedSituationIds.map((id) => getSituationOptions({ path: { id } })),
	}));

	// TODO: get from incident?
	timezone = $derived(getLocalTimeZone());

	private incidentRetrospectiveId = $derived(this.incident?.attributes.retrospectiveId || "");
	retrospectiveQuery = createQuery(() => ({
		...getRetrospectiveOptions({ path: { id: this.incidentRetrospectiveId } }),
		enabled: !!this.incidentRetrospectiveId,
	}));;
	retrospective = $derived(this.retrospectiveQuery.data?.data);
	retrospectiveId = $derived(this.retrospective?.id);

	retrospectiveDocumentId = $derived(this.retrospective?.attributes.documentId);
	retrospectiveDocumentSessionQuery = createQuery(() => ({
		...getDocumentSessionOptions({ path: { id: this.retrospectiveDocumentId ?? "" } }),
		enabled: !!this.retrospectiveDocumentId,
	}));
	retrospectiveDocument = $derived(this.retrospectiveDocumentSessionQuery.data?.data);

	documentAccess = $derived(this.retrospectiveDocument?.access);
	
	systemAnalysisId = $derived(this.retrospective?.attributes.systemAnalysisId);
}

const ctx = new Context<IncidentViewController>("IncidentViewController");
export const initIncidentViewController = (slugFn: Getter<string>) => ctx.set(new IncidentViewController(slugFn));
export const useIncidentView = () => ctx.get();
