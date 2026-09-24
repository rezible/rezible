import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import { selectedEntityIds, selectedRelationshipIds } from "$features/system/lib/system-map/selection";
import type { FlowNode, FlowEdge } from "./flow-graph-model";

type InteractionOptions = {
	graph: GraphSubset;
	selection?: MapSelection;
	highlights?: MapHighlights;
	hoveredEdgeId?: string;
	showLabels: boolean;
};

/** Interaction flags preserve live geometry and never change canonical selection. */
export function applyInteractionState(nodes: FlowNode[], edges: FlowEdge[], options: InteractionOptions) {
	const { graph, selection, highlights, hoveredEdgeId, showLabels } = options;
	const selectedRelationships = selectedRelationshipIds(selection);
	const selectedEntities = selectedEntityIds(graph, selection);
	const highlightedEntities = [...selectedEntities, ...(highlights?.entityIds ?? [])];
	const highlightedIds = new Set(highlightedEntities);
	const hostHighlightedIds = new Set(highlights?.entityIds);
	const hoveredEdge = edges.find((edge) => edge.id === hoveredEdgeId);

	if (hoveredEdge) {
		highlightedIds.add(hoveredEdge.source);
		highlightedIds.add(hoveredEdge.target);
	}

	let selectedNodeId: string | undefined;
	if (selection?.kind === "entity") {
		selectedNodeId = selection.entityId;
	}

	const updatedNodes = nodes.map((node) => ({
		...node,
		selected: node.id === selectedNodeId,
		data: { ...node.data, highlighted: highlightedIds.has(node.id) },
	}));
	const updatedEdges = edges.map((edge) => {
		const connection = edge.data!.connection;
		const selected = connection.relationshipIds.some((id) => selectedRelationships.includes(id));
		const highlighted = connection.relationshipIds.some((id) => highlights?.relationshipIds.includes(id));
		const connectedToSelection = edge.source === selectedNodeId || edge.target === selectedNodeId;
		const connectedToHighlight =
			hostHighlightedIds.has(edge.source) || hostHighlightedIds.has(edge.target);
		const hovered = edge.id === hoveredEdgeId;
		const emphasized = selected || highlighted || connectedToSelection || connectedToHighlight || hovered;
		return {
			...edge,
			selected,
			data: {
				connection,
				emphasized,
				hovered,
				dimmed: !emphasized && Boolean(selection || hoveredEdge),
				showLabel: showLabels || emphasized,
			},
		};
	});

	return { nodes: updatedNodes, edges: updatedEdges };
}
