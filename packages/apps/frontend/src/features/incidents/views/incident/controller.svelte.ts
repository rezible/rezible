import { closedReviewLine } from "$features/incidents/lib/review";
import {
	getDocumentSessionOptions,
	getIncidentOptions,
	getIncidentQueryKey,
	getRetrospectiveOptions,
	type GetIncidentResponse,
} from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { page } from "$app/state";
import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import { referenceUrl } from "$lib/reference-url";
import { toast } from "svelte-sonner";
import type { PageRelatedLink } from "$lib/app-shell.svelte";
import { convertIncidentViewParam } from "$src/params/incidentView";
import RiArticleLine from "remixicon-svelte/icons/article-line";

export class IncidentViewController {
	private queryClient = useQueryClient();

	/** The URL value: the incident's reference, or another value the API resolves, such as its UUID. */
	ref = $state("");

	constructor(refFn: Getter<string>) {
		this.ref = refFn();
		watch(refFn, (ref) => {
			this.ref = ref;
		});
		watch(
			() => this.incidentQuery.data,
			(response) => {
				if (response) {
					this.useReferenceUrl(response);
				}
			}
		);
	}

	incidentQuery = createQuery(() => ({
		...getIncidentOptions({ path: { id: this.ref } }),
		enabled: !!this.ref,
	}));
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");
	reference = $derived(this.incident?.attributes.reference);

	/** Replaces a URL that names the incident another way with its reference URL, caching the response first. */
	private useReferenceUrl(response: GetIncidentResponse) {
		const reference = response.data.attributes.reference;
		if (reference === this.ref) {
			return;
		}
		this.queryClient.setQueryData(getIncidentQueryKey({ path: { id: reference } }), response);
		void goto(referenceUrl(page.url, this.ref, reference), {
			replaceState: true,
			noScroll: true,
			keepFocus: true,
		});
	}

	situations = $derived(this.incident?.attributes.situations ?? []);
	relatedLinks = $derived<PageRelatedLink[]>(
		this.situations.map((situation) => ({
			key: situation.id,
			kind: "Situation",
			icon: RiArticleLine,
			label: situation.title,
			path: resolve("/situations/[ref]/[[view=situationView]]", { ref: situation.id }),
		}))
	);

	activeView = $derived(convertIncidentViewParam(page.params.view));

	timezone = $derived(getLocalTimeZone());

	incidentRetrospectiveId = $derived(this.incident?.attributes.retrospective?.id ?? "");
	retrospectiveQuery = createQuery(() => ({
		...getRetrospectiveOptions({ path: { id: this.incidentRetrospectiveId } }),
		enabled: !!this.incidentRetrospectiveId,
	}));
	retrospective = $derived(this.retrospectiveQuery.data?.data);
	retrospectiveId = $derived(this.retrospective?.id);

	retrospectiveDocumentId = $derived(this.retrospective?.attributes.documentId);
	retrospectiveDocumentSessionQuery = createQuery(() => ({
		...getDocumentSessionOptions({ path: { id: this.retrospectiveDocumentId ?? "" } }),
		enabled: !!this.retrospectiveDocumentId,
	}));
	retrospectiveDocument = $derived(this.retrospectiveDocumentSessionQuery.data?.data);

	reviewClosed = $derived(this.retrospective?.attributes.state === "closed");
	closedReviewLine = $derived(closedReviewLine(this.retrospective?.attributes));
	documentAccess = $derived(this.closedDocumentAccess());

	refetchRetrospective = () => {
		void this.retrospectiveQuery.refetch();
	};

	private closedDocumentAccess() {
		const access = this.retrospectiveDocument?.access;
		if (!access || !this.reviewClosed) {
			return access;
		}
		return { ...access, canEdit: false, canManage: false };
	}

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
export const initIncidentViewController = (refFn: Getter<string>) =>
	ctx.set(new IncidentViewController(refFn));
export const useIncidentView = () => ctx.get();
