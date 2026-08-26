import type { ComponentProps } from "svelte";
import { Context } from "runed";

import IncidentTimelineContextMenu from "./incident-timeline/IncidentTimelineContextMenu.svelte";

type ContextMenuProps = {
	timeline?: ComponentProps<typeof IncidentTimelineContextMenu>;
};

export class IncidentAnalysisController {
	contextMenu = $state.raw<ContextMenuProps>({});

	setContextMenu(props: ContextMenuProps) {
		this.contextMenu = props;
	}

	clearContextMenu() {
		this.contextMenu = {};
	}
}

const ctx = new Context<IncidentAnalysisController>("IncidentAnalysisController");
export const initIncidentAnalysisController = () => ctx.set(new IncidentAnalysisController());
export const useIncidentAnalysis = () => ctx.get();
