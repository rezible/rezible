import type { IncidentMilestone } from "$lib/api";
import { Context } from "runed";

export class MilestonesDialogState {
	open = $state(false);
	editingMilestone = $state<IncidentMilestone>();
	editorOpen = $state(false);

	close() {
		this.open = false;
		this.editingMilestone = undefined;
	}
}

const milestonesDialogCtx = new Context<MilestonesDialogState>("incidentMilestonesDialog");
export const setMilestonesDialog = (s: MilestonesDialogState) => milestonesDialogCtx.set(s);
export const useMilestonesDialog = () => milestonesDialogCtx.get();