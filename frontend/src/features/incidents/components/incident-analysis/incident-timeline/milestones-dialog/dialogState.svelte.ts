import type { IncidentMilestone } from "$lib/api";

const createMilestonesDialogState = () => {
	let open = $state(false);
	let editingMilestone = $state<IncidentMilestone>();
	let editorOpen = $state(false);

	const close = () => {
		open = false;
		editingMilestone = undefined;
	}

	return {
		get open() { return open },
		set open(o: boolean) { open = o },
		close,
		get editingMilestone() { return editingMilestone },
		set editingMilestone(m: IncidentMilestone | undefined) { editingMilestone = m },
		get editorOpen() { return editorOpen },
		set editorOpen(e: boolean) { editorOpen = e },
	};
};

export const milestonesDialog = createMilestonesDialogState();