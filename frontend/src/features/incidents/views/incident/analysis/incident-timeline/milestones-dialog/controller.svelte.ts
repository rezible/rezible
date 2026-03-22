import type { IncidentMilestone } from "$lib/api";
import { Context } from "runed";

export class IncidentMilestonesDialogController {
	open = $state(false);
	editingMilestone = $state<IncidentMilestone>();
	editorOpen = $state(false);

	close() {
		this.open = false;
		this.editingMilestone = undefined;
	}
}

const ctx = new Context<IncidentMilestonesDialogController>("IncidentMilestonesDialogController");
export const initMilestonesDialog = () => ctx.set(new IncidentMilestonesDialogController());
export const useMilestonesDialog = () => ctx.get();