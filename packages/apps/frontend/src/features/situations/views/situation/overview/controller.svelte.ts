import { Context } from "runed";
import { useSituationController } from "../controller.svelte";
import { useSituationInvestigationController } from "../investigation/controller.svelte";
import { observationGroups, type SourceTarget } from "../model";

export class SituationOverviewController {
	private situationController = useSituationController();
	private investigationController = useSituationInvestigationController();

	private situation = $derived(this.situationController.situation);
	attrs = $derived(this.situation?.attributes);
	observations = $derived(observationGroups(this.attrs?.observationGroups ?? []));
	initialGroupId = $derived(this.attrs?.observationGroups[0]?.id);

	groupChoices = $state<Record<string, boolean>>({});

	get investigationUnavailable() {
		return this.investigationController.investigationUnavailable;
	}

	get investigationId() {
		return this.situationController.investigationId;
	}

	get investigationQuery() {
		return this.investigationController.investigationQuery;
	}

	get reportQuery() {
		return this.investigationController.reportQuery;
	}

	get sourceTarget() {
		return this.investigationController.sourceTarget;
	}

	groupOpen(id: string) {
		return this.groupChoices[id] ?? id === this.initialGroupId;
	}

	setGroupOpen = (id: string, open: boolean) => {
		this.groupChoices[id] = open;
	};

	openSource = (target: SourceTarget, trigger: HTMLElement) =>
		this.investigationController.openSource(target, trigger);
	closeSource = () => this.investigationController.closeSource();
}

const ctx = new Context<SituationOverviewController>("SituationOverviewController");
export const initSituationOverviewController = () => ctx.set(new SituationOverviewController());
export const useSituationOverviewController = () => ctx.get();
