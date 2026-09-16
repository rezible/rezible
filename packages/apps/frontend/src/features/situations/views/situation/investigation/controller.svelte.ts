import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import { getSituationQueryKey, startSituationInvestigationMutation, type InvestigationReportAttributes } from "$lib/api";
import { useSituationController } from "../controller.svelte";
import { evidenceChanged } from "../model";
import { Context } from "runed";

export type InvestigationForm = {
	query: string;
	pending: boolean;
};

const idPath = (id: string = "") => ({id});

const makeReportSections = (ra?: InvestigationReportAttributes) => {
	return [
		{ title: "Likely cause", items: [ra?.likelyCause ?? ""] },
		{ title: "Best next step", items: [ra?.bestNextStep ?? ""] },
		{ title: "Recommended actions", items: ra?.recommendedActions ?? [] },
		{ title: "Suggested checks", items: ra?.suggestedChecks ?? [] },
		{ title: "Limitations", items: ra?.limitations ?? [] },
	]
		.map((section) => ({ ...section, items: section.items.filter(Boolean) }))
		.filter((section) => section.items.length)
}

export class SituationInvestigationsController {
	queryClient = useQueryClient();

	situationController = useSituationController();
	
	private situationId = $derived(this.situationController.situationId ?? "");
	sitAttrs = $derived(this.situationController.situation?.attributes);

	form = $state<InvestigationForm>({ query: "", pending: false });

	startInvestigationMutation = createMutation(() => ({
		...startSituationInvestigationMutation(),
		onSuccess: async () => {
			await this.queryClient.invalidateQueries({
				queryKey: getSituationQueryKey({ path: idPath(this.situationId) }),
			});
		},
	}));

	investigation = $derived(this.situationController.investigation);
	report = $derived(this.investigation?.attributes?.report?.attributes);
	evidenceChanged = $derived(evidenceChanged(
		this.sitAttrs?.evidenceRevision ?? 0,
		this.sitAttrs?.investigation?.completedRevision
	));
	reportChanged = $derived(!!this.report && this.evidenceChanged);

	reportSections = $derived(makeReportSections(this.report));

	runInvestigation = async () => {
		if (this.form.pending) return;
		this.form.pending = true;
		try {
			await this.startInvestigationMutation.mutateAsync({
				path: { id: this.situationId },
				body: { 
					attributes: { 
						query: this.form.query.trim() || undefined,
					},
				},
			});
		} catch {
			// The mutation exposes its API error. Keep the input available for another submission.
		} finally {
			this.form.pending = false;
		}
	};
}

const ctx = new Context<SituationInvestigationsController>("SituationInvestigationsController")
export const initSituationInvestigationController = () => ctx.set(new SituationInvestigationsController());
export const useSituationInvestigationController = () => ctx.get();