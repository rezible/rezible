import type { Edge, Node } from "@xyflow/svelte";

import type { GraphEntity } from "$features/systems/lib/system-map/graph";
import type { MapConnection, MapNodeAppearance } from "$features/systems/lib/system-map/presentation";

export type FlowNodeData = {
	entity: GraphEntity;
	appearance: MapNodeAppearance;
	annotationCount: number;
	isConnectionEndpoint?: boolean;
};

export type FlowNode = Node<FlowNodeData, "system-map-node">;

export type FlowEdgeData = {
	connection: MapConnection;
	laneOffset?: number;
	isDimmed?: boolean;
	isHighlighted?: boolean;
	isLabelVisible?: boolean;
};

export type FlowEdge = Edge<FlowEdgeData, "system-map-connection">;

export type LayoutResult = {
	nodes: FlowNode[];
	edges: FlowEdge[];
};
