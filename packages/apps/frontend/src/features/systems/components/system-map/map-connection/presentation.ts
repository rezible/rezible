import type { FlowEdge } from "../flow-model";

export type ConnectionSelection = {
	edgeId?: string;
	endpointIds: ReadonlySet<string>;
};

export type ConnectionPresentationState = {
	selected: boolean;
	highlighted: boolean;
	dimmed: boolean;
};

/** Keeps edge emphasis derived from stable edge IDs and displayed endpoint IDs. */
export const connectionPresentationState = (
	edge: Pick<FlowEdge, "id" | "source" | "target">,
	selection: ConnectionSelection | undefined
): ConnectionPresentationState => {
	const selected = selection?.edgeId === edge.id;
	const connectedToSelection =
		selection?.endpointIds.has(edge.source) || selection?.endpointIds.has(edge.target) || false;
	const highlighted = selected || (!selection?.edgeId && connectedToSelection);

	return {
		selected,
		highlighted,
		dimmed: Boolean(selection && !highlighted),
	};
};
