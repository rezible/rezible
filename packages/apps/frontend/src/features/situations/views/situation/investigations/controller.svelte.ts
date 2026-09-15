import { page } from "$app/state";
import { goto } from "$app/navigation";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import {
	getSituationInvestigationQueryKey,
	getSituationQueryKey,
	listSituationInvestigationsQueryKey,
	startSituationInvestigationMutation,
	type StartSituationInvestigationResponse,
} from "$lib/api";
import { useSituationController } from "../controller.svelte";
import { investigationHref } from "../../../lib/routes";
import { evidenceChanged } from "../model";

export type InvestigationForm = {
	query: string;
	pending: boolean;
};

export class SituationInvestigationsController {
	queryClient = useQueryClient();
	situationController = useSituationController();
	private situation = $derived(this.situationController.situation);
	private situationId = $derived(this.situation?.id ?? "");
	attrs = $derived(this.situation?.attributes);

	form = $state<InvestigationForm>({ query: "", pending: false });

	startMutation = createMutation(() => ({
		...startSituationInvestigationMutation(),
		onSuccess: async (result: StartSituationInvestigationResponse) => {
			const investigation = result.data;
			const situationId = investigation.attributes.situationId;
			const selectedKey = getSituationInvestigationQueryKey({ path: { id: investigation.id } });
			this.queryClient.setQueryData(selectedKey, result);
			await Promise.all([
				this.queryClient.invalidateQueries({
					queryKey: getSituationQueryKey({ path: { id: situationId } }),
				}),
				this.queryClient.invalidateQueries({
					queryKey: listSituationInvestigationsQueryKey({
						path: { id: situationId },
					}),
				}),
				this.queryClient.invalidateQueries({ queryKey: selectedKey }),
			]);
			if (page.params.id !== situationId) {
				return;
			}
			await goto(investigationHref(situationId, investigation.id, page.url.search), {
				noScroll: true,
				keepFocus: true,
			});
		},
	}));
	selectedChanged = $derived(
		evidenceChanged(this.attrs?.evidenceRevision ?? 0, this.situationController.selectedInvestigation)
	);
	selectedReport = $derived(
		this.situationController.selectedInvestigationBelongs
			? this.situationController.selectedInvestigation?.attributes.report
			: undefined
	);

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
		if (this.form.pending) {
			return;
		}
		this.form.pending = true;
		try {
			const query = this.form.query.trim() || undefined;
			await this.startMutation.mutateAsync({
				path: { id: this.situationId },
				body: { attributes: { query } },
			});
		} catch {
			// The mutation exposes its API error. Keep the input available for another submission.
		} finally {
			this.form.pending = false;
		}
	};
}

export const initSituationInvestigationsController = () => new SituationInvestigationsController();
