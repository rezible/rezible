import { Position, type Edge, type Node } from "@xyflow/svelte";

import type { GraphEntity, GraphSubset } from "$features/system/lib/system-map/graph";
import { isArchitectureCategory } from "$features/system/lib/system-map/category";
import { getMapConnections, type MapConnection } from "$features/system/lib/system-map/connections";
import type { Point } from "$features/system/lib/system-map/geometry";

import { connectionLabel } from "./map-connection/presentation";

export const NODE_WIDTH = 220;
export const NODE_HEIGHT = 72;

export type FlowNode = Node<
	{
		entity: GraphEntity;
		highlighted?: boolean;
	},
	"system-map-node"
>;

export type FlowEdge = Edge<
	{
		connection: MapConnection;
		emphasized?: boolean;
		hovered?: boolean;
		dimmed?: boolean;
		showLabel?: boolean;
	},
	"system-map-connection"
>;

export const validPosition = (point: Point | undefined): point is Point =>
	point !== undefined && Number.isFinite(point.x) && Number.isFinite(point.y);

/** Preserve existing positions; put new records in a staging column without moving existing nodes. */
export function graphToFlow(
	graph: GraphSubset,
	previous: readonly FlowNode[],
	positions: Readonly<Partial<Record<string, Point>>> = {}
) {
	const existingNodes = new Map(previous.map((node) => [node.id, node]));
	let nextX = 0;
	if (previous.length > 0) {
		const rightmostX = Math.max(...previous.map((node) => node.position.x));
		nextX = rightmostX + NODE_WIDTH + 100;
	}

	let added = 0;
	const nodes: FlowNode[] = graph.entities
		.filter((entity) => isArchitectureCategory(entity.category))
		.map((entity) => {
			const existingNode = existingNodes.get(entity.id);
			const suppliedPosition = positions[entity.id];
			let position: Point;

			if (existingNode) {
				position = existingNode.position;
			} else if (validPosition(suppliedPosition)) {
				position = suppliedPosition;
			} else {
				position = { x: nextX, y: added * (NODE_HEIGHT + 32) };
				added += 1;
			}

			const handles: FlowNode["handles"] = [
				{
					id: "input",
					type: "target",
					position: Position.Left,
					x: -3,
					y: NODE_HEIGHT / 2 - 3,
					width: 6,
					height: 6,
				},
				{
					id: "output",
					type: "source",
					position: Position.Right,
					x: NODE_WIDTH - 3,
					y: NODE_HEIGHT / 2 - 3,
					width: 6,
					height: 6,
				},
			];

			return {
				...existingNode,
				id: entity.id,
				type: "system-map-node",
				hidden: false,
				position,
				width: NODE_WIDTH,
				height: NODE_HEIGHT,
				// Cards and ports have fixed dimensions; edges do not need a DOM measurement pass.
				measured: { width: NODE_WIDTH, height: NODE_HEIGHT },
				handles,
				focusable: true,
				ariaRole: "button",
				ariaLabel: entity.kind,
				data: { entity },
			};
		});

	const connections = getMapConnections(graph);
	
	return { nodes, edges: connectionsToFlow(connections) };
}

export function connectionsToFlow(connections: readonly MapConnection[]): FlowEdge[] {
	return connections.map((connection) => ({
		id: connection.id,
		source: connection.source,
		target: connection.target,
		type: "system-map-connection",
		sourceHandle: "output",
		targetHandle: "input",
		interactionWidth: 20,
		focusable: true,
		ariaRole: "button",
		data: { connection },
		ariaLabel: connectionLabel(connection),
	}));
}
