import type { MapConnection } from "$features/systems/lib/system-map/presentation";
import type { FlowEdge } from "../flow-model";

export type ConnectionSelection = {
	edgeId?: string;
	endpointIds: ReadonlySet<string>;
};

export type ConnectionPresentationState = {
	selected: boolean;
	hovered: boolean;
	highlighted: boolean;
	dimmed: boolean;
	endpointEmphasized: boolean;
	labelVisible: boolean;
};

export type ConnectionPresentationOptions = {
	selection?: ConnectionSelection;
	hoveredConnectionId?: string;
	showAllConnectionLabels?: boolean;
};

type ConnectionEdge = Pick<FlowEdge, "id" | "source" | "target">;

/** Formats only the readable canvas label; counts remain derived from the source IDs. */
export const connectionLabel = (connection: MapConnection): string => {
	const predicate = connection.predicate.replaceAll("_", " ");
	return connection.classification === "summary"
		? `${predicate} · ${connection.sourceRelationshipIds.length}`
		: predicate;
};

/** Keeps edge emphasis and label visibility derived from stable IDs and displayed endpoints. */
export const connectionPresentationState = (
	edge: ConnectionEdge,
	{
		selection,
		hoveredConnectionId,
		showAllConnectionLabels = false,
	}: ConnectionPresentationOptions = {}
): ConnectionPresentationState => {
	const selected = selection?.edgeId === edge.id;
	const hovered = hoveredConnectionId === edge.id;
	const connectedToSelection =
		selection?.endpointIds.has(edge.source) || selection?.endpointIds.has(edge.target) || false;
	const nodeSelectionActive = Boolean(selection && !selection.edgeId && selection.endpointIds.size > 0);
	const highlighted =
		selected || hovered || (!hoveredConnectionId && nodeSelectionActive && connectedToSelection);
	const hasPresentationContext = Boolean(hovered || selection?.edgeId || nodeSelectionActive);

	return {
		selected,
		hovered,
		highlighted,
		dimmed: hasPresentationContext && !highlighted,
		endpointEmphasized: selected || hovered,
		labelVisible: showAllConnectionLabels || selected || hovered,
	};
};

/** Returns only the displayed endpoints of hovered or selected edges. */
export const connectionEndpointIds = (
	edges: readonly ConnectionEdge[],
	options: ConnectionPresentationOptions
): ReadonlySet<string> => {
	const endpointIds = new Set<string>();

	for (const edge of edges) {
		if (!connectionPresentationState(edge, options).endpointEmphasized) continue;
		endpointIds.add(edge.source);
		endpointIds.add(edge.target);
	}

	return endpointIds;
};
