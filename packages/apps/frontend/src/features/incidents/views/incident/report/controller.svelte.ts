import { Context } from "runed";
import { useIncidentView } from "../controller.svelte";

export class IncidentReportController {
	viewController = useIncidentView();
	retrospectiveId = $derived(this.viewController.retrospectiveId);
	retrospective = $derived(this.viewController.retrospective);
	sections = $derived(this.retrospective?.attributes.reportSections ?? []);

	documentAccess = $derived(this.viewController.documentAccess);
	canView = $derived(this.documentAccess?.canView ?? false);
	canEdit = $derived(this.documentAccess?.canEdit ?? false);

	editMode = $state(false);
	toggleEdit = () => {
		if (this.canEdit) this.editMode = !this.editMode;
	};
}

const ctx = new Context<IncidentReportController>("IncidentReportController");
export const initIncidentReportController = () => ctx.set(new IncidentReportController());
export const useIncidentReportController = () => ctx.get();
