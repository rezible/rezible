import type { Edge, Node } from "@xyflow/svelte";

import type { GraphEntity } from "$features/systems/lib/system-map/graph";
import type { Point } from "$features/systems/lib/system-map/geometry";
import type { MapConnection, MapNodeAppearance } from "$features/systems/lib/system-map/presentation";

export type FlowNodeData = {
	entity: GraphEntity;
	appearance: MapNodeAppearance;
	annotationCount: number;
	isConnectionEndpoint?: boolean;
};

export type FlowNode = Node<FlowNodeData, "system-map-node">;

export type ConnectionRouteSection = {
	start: Point;
	bends: readonly Point[];
	end: Point;
};

export type ConnectionRoute = {
	sections: readonly ConnectionRouteSection[];
	labelPosition: Point;
	/** The section whose end point is the ELK target boundary. */
	targetSectionIndex: number;
};

export type FlowEdgeData = {
	connection: MapConnection;
	route: ConnectionRoute;
	isDimmed?: boolean;
	isHighlighted?: boolean;
	isLabelVisible?: boolean;
};

export type FlowEdge = Edge<FlowEdgeData, "system-map-connection">;

export type LayoutResult = {
	nodes: FlowNode[];
	edges: FlowEdge[];
};
