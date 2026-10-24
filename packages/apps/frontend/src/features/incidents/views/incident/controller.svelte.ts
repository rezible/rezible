import { closedReviewLine } from "$features/incidents/lib/review";
import { getDocumentSessionOptions, getIncidentOptions, getRetrospectiveOptions } from "$lib/api";
import { getLocalTimeZone } from "@internationalized/date";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { page } from "$app/state";
import { resolve } from "$app/paths";
import { toast } from "svelte-sonner";
import type { PageRelatedLink } from "$lib/app-shell.svelte";
import { convertIncidentViewParam } from "$src/params/incidentView";
import RiArticleLine from "remixicon-svelte/icons/article-line";

export class IncidentViewController {
	slug = $state("");

	constructor(slugFn: Getter<string>) {
		this.slug = slugFn();
		watch(slugFn, (slug) => {
			this.slug = slug;
		});
	}

	incidentQuery = createQuery(() => ({
		...getIncidentOptions({ path: { id: this.slug } }),
		enabled: !!this.slug,
	}));
	incident = $derived(this.incidentQuery.data?.data);
	incidentId = $derived(this.incident?.id ?? "");

	situations = $derived(this.incident?.attributes.situations ?? []);
	relatedLinks = $derived<PageRelatedLink[]>(
		this.situations.map((situation) => ({
			key: situation.id,
			kind: "Situation",
			icon: RiArticleLine,
			label: situation.title,
			path: resolve("/situations/[id]/[[view=situationView]]", { id: situation.id }),
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
export const initIncidentViewController = (slugFn: Getter<string>) =>
	ctx.set(new IncidentViewController(slugFn));
export const useIncidentView = () => ctx.get();
