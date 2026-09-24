import type { Edge, Node } from "@xyflow/svelte";

import type { GraphEntity } from "$features/system/lib/system-map/graph";
import type { Point } from "$features/system/lib/system-map/geometry";
import type { MapConnection, MapNodeAppearance } from "$features/system/lib/system-map/presentation";

export type FlowNodeData = {
	entity: GraphEntity;
	appearance: MapNodeAppearance;
	opacity?: number;
	isDraggable?: boolean;
	isInteractive?: boolean;
	isConnectionEndpoint?: boolean;
	isHighlighted?: boolean;
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
	sourceSectionIndex: number;
	/** The section whose end point is the ELK target boundary. */
	targetSectionIndex: number;
};

export type ConnectionRouteVariant = {
	route: ConnectionRoute;
	/** Layer blend opacity, applied independently from endpoint fade and edge emphasis. */
	opacity: number;
};

export type FlowEdgeData = {
	connection: MapConnection;
	route: ConnectionRoute;
	routeVariants?: readonly ConnectionRouteVariant[];
	opacity?: number;
	isDimmed?: boolean;
	isHighlighted?: boolean;
	isHovered?: boolean;
	isLabelVisible?: boolean;
};

export type FlowEdge = Edge<FlowEdgeData, "system-map-connection">;

export type LayoutResult = {
	nodes: FlowNode[];
	edges: FlowEdge[];
};
