import { getIncidentOptions, getRetrospectiveOptions } from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { initIncidentCollaborationController } from "./collaboration.svelte";
import { onMount } from "svelte";

export class IncidentViewController {
	queryClient = useQueryClient();

	collab = initIncidentCollaborationController();

	slug = $state<string>(null!);

	constructor(slugFn: Getter<string>) {
		onMount(() => {
			return () => {
				this.cleanup();
			}
		});

		watch(slugFn, slug => {
			this.slug = slug;
		});


		watch(() => this.documentId, documentId => { 
			this.collab.connect(documentId);
		});
	}

	private cleanup() {
		this.collab.cleanup();
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
export const initIncidentViewController = (slugFn: Getter<string>) => ctx.set(new IncidentViewController(slugFn));
export const useIncidentView = () => ctx.get();