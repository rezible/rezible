import { page } from "$app/state";
import { goto, replaceState } from "$app/navigation";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import {
	getSituationInvestigationQueryKey,
	getSituationQueryKey,
	getSituationInvestigationOptions, 
	listSituationInvestigationsOptions,
	listSituationInvestigationsQueryKey,
	startSituationInvestigationMutation,
	type StartSituationInvestigationResponse,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { watch } from "runed";
import { useSituationController } from "../controller.svelte";
import { investigationHref, situationHref } from "../../../lib/routes";
import { evidenceChanged, selectInvestigation } from "../model";

export type InvestigationForm = { query: string; pending: boolean };

export class SituationInvestigationsController {
	queryClient = useQueryClient();
	situationController = useSituationController();
	private situation = $derived(this.situationController.situation);
	private situationId = $derived(this.situation?.id ?? "");
	attrs = $derived(this.situation?.attributes);
	
	form = $state<InvestigationForm>({ query: "", pending: false });

	selectedId = $derived(
		selectInvestigation(
			page.url.searchParams.get("investigation"),
			undefined,
			this.attrs?.investigations ?? []
		)
	);

	investigationsQuery = createPaginatedQuery({
		queryOptions: (pagination) =>
			listSituationInvestigationsOptions({ path: { id: this.situationId }, query: pagination }),
	});
	selectedQuery = createQuery(() => ({
		...getSituationInvestigationOptions({ path: { id: this.selectedId ?? "" } }),
		enabled: !!this.selectedId,
	}));
	startMutation = createMutation(() => ({
		...startSituationInvestigationMutation(),
		onSuccess: async (result: StartSituationInvestigationResponse) => {
			const selectedKey = getSituationInvestigationQueryKey({ path: { id: result.data.id } });
			this.queryClient.setQueryData(selectedKey, result);
			await Promise.all([
				this.queryClient.invalidateQueries({ queryKey: getSituationQueryKey({ path: { id: this.situationId } }) }),
				this.queryClient.invalidateQueries({ queryKey: listSituationInvestigationsQueryKey({ path: { id: this.situationId } }) }),
				this.queryClient.invalidateQueries({ queryKey: selectedKey }),
			]);
			if (page.params.id !== this.situationId) return;
			await goto(this.investigationHref(this.situationId), { noScroll: true, keepFocus: true });
		},
	}));
	investigations = $derived(this.investigationsQuery.query.data?.data ?? []);

	constructor() {
		watch(
			() => [this.attrs?.investigations, page.url.search] as const,
			([items]) => {
				if (page.url.searchParams.get("investigation")) return;
				const id = selectInvestigation(null, undefined, items ?? []);
				if (!id) return;
				const params = new URLSearchParams(page.url.searchParams);
				params.set("investigation", id);
				void replaceState(`${page.url.pathname}?${params}${page.url.hash}`, {});
			}
		);
	}
	
	selected = $derived(this.selectedQuery.data?.data);
	selectedBelongs = $derived(this.selected?.attributes.situationId === this.situationId);
	selectedChanged = $derived(evidenceChanged(this.attrs?.evidenceRevision ?? 0, this.selected));
	selectedReport = $derived(this.selectedBelongs ? this.selected?.attributes.report : undefined);
	
	investigationsHref = $derived(situationHref(this.situationId, "investigations", page.url.search));
	investigationHref = (id: string) => investigationHref(this.situationId, id, page.url.search);
	reportSections = $derived(
		[
			{ title: "Likely cause", items: [this.selectedReport?.likelyCause ?? ""] },
			{ title: "Best next step", items: [this.selectedReport?.bestNextStep ?? ""] },
			{ title: "Recommended actions", items: this.selectedReport?.recommendedActions ?? [] },
			{ title: "Suggested checks", items: this.selectedReport?.suggestedChecks ?? [] },
			{ title: "Limitations", items: this.selectedReport?.limitations ?? [] },
		]
			.map((section) => ({ ...section, items: section.items.filter(Boolean) }))
			.filter((section) => section.items.length)
	);
	runInvestigation = async () => {
		if (this.form.pending) return;
		this.form.pending = true;
		try {
			const query = this.form.query.trim() || undefined;
			await this.startMutation.mutateAsync({
				path: { id: this.situationId },
				body: { attributes: { query } },
			})
		} catch {
			// The mutation exposes its API error. Keep the input available for another submission.
		} finally {
			this.form.pending = false;
		}
	}
}

export const initSituationInvestigationsController = () => new SituationInvestigationsController();
