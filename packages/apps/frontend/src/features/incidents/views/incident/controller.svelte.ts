import {
	createIncidentRetrospectiveMutation,
	getDocumentSessionOptions,
	getIncidentOptions,
	getRetrospectiveOptions,
} from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { page } from "$app/state";
import { toast } from "svelte-sonner";

export class IncidentViewController {
	slug = $state("");

	constructor(slugFn: Getter<string>) {
		this.slug = slugFn();
		watch(slugFn, (slug) => {
			this.slug = slug;
			this.startRetrospectiveMutation.reset();
		});
	}

	incidentQuery = createQuery(() => ({
		...getIncidentOptions({ path: { id: this.slug } }),
		enabled: !!this.slug,
	}));
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");

	situations = $derived(this.incident?.attributes.situations ?? []);

	timezone = $derived(getLocalTimeZone());

	incidentRetrospectiveId = $derived(this.incident?.attributes.retrospective?.id ?? "");
	retrospectiveQuery = createQuery(() => ({
		...getRetrospectiveOptions({ path: { id: this.incidentRetrospectiveId } }),
		enabled: !!this.incidentRetrospectiveId,
	}));
	retrospective = $derived(this.retrospectiveQuery.data?.data);
	retrospectiveId = $derived(this.retrospective?.id);

	canStartRetrospective = $derived(
		this.incident?.attributes.responseState === "resolved" && !this.incidentRetrospectiveId
	);
	startRetrospectiveMutation = createMutation(() => ({
		...createIncidentRetrospectiveMutation(),
		onSuccess: async (_response, variables) => {
			if (variables.path.id === this.incidentId) {
				await this.incidentQuery.refetch();
			}
		},
	}));
	startRetrospective = () => {
		if (!this.canStartRetrospective || this.startRetrospectiveMutation.isPending) {
			return;
		}
		this.startRetrospectiveMutation.mutate({ path: { id: this.incidentId } });
	};

	retrospectiveDocumentId = $derived(this.retrospective?.attributes.documentId);
	retrospectiveDocumentSessionQuery = createQuery(() => ({
		...getDocumentSessionOptions({ path: { id: this.retrospectiveDocumentId ?? "" } }),
		enabled: !!this.retrospectiveDocumentId,
	}));
	retrospectiveDocument = $derived(this.retrospectiveDocumentSessionQuery.data?.data);

	documentAccess = $derived(this.retrospectiveDocument?.access);

	systemAnalysisId = $derived(this.retrospective?.attributes.systemAnalysisId);

	share = async () => {
		try {
			await navigator.clipboard.writeText(page.url.href);
			toast.success("Incident link copied");
		} catch {
			toast.error("Unable to copy the incident link");
		}
	};
}

const ctx = new Context<IncidentViewController>("IncidentViewController");
export const initIncidentViewController = (slugFn: Getter<string>) =>
	ctx.set(new IncidentViewController(slugFn));
export const useIncidentView = () => ctx.get();
