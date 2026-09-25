import { Context } from "runed";
import { useSituationController } from "../controller.svelte";

export class SituationInvestigationsController {
	private situationController = useSituationController();

	investigation = $derived(this.situationController.investigation);
	report = $derived(this.situationController.investigationReport?.attributes);
}

const ctx = new Context<SituationInvestigationsController>("SituationInvestigationsController");
export const initSituationInvestigationController = () => ctx.set(new SituationInvestigationsController());
export const useSituationInvestigationController = () => ctx.get();
