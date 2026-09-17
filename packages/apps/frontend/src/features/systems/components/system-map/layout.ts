import type { ELK as ElkInstance, ElkNode } from "elkjs/lib/elk-api.js";
import { MarkerType } from "@xyflow/svelte";

import {
	isActorCategory,
	isArchitectureCategory,
} from "$features/systems/lib/system-map/category";
import type { GraphEntity, GraphSubset } from "$features/systems/lib/system-map/graph";
import {
	worldBoundsByNodeId,
	worldCenter,
	screenPointFromWorld,
	type Bounds,
	type Point,
	type Size,
	type Viewport,
} from "$features/systems/lib/system-map/geometry";
import {
	type MapConnection,
	type MapNode,
	type MapProjection,
} from "$features/systems/lib/system-map/presentation";
import { buildConnectionLanes, connectionHandleAssignment } from "./map-connection/geometry";
import type { FlowEdge, FlowNode, FlowNodeData, LayoutResult } from "./flow-model";

export const COMPACT_NODE_SIZE = { width: 240, height: 96 } as const satisfies Readonly<Size>;
export const GROUP_NODE_MIN_SIZE = { width: 280, height: 168 } as const satisfies Readonly<Size>;

const GROUP_PADDING = "[top=48,left=28,bottom=28,right=28]";
const ROOT_PADDING = "[top=32,left=32,bottom=32,right=32]";
const NODE_SPACING = "28";
const LAYER_SPACING = "72";
const ACTOR_GAP = 96;

type LayoutModelNode = {
	mapNode: MapNode;
	entity: GraphEntity;
	children: LayoutModelNode[];
};

type PositionedNode = { node: FlowNode } & Bounds;

const nodeData = (modelNode: LayoutModelNode, annotationCount: number): FlowNodeData => ({
	entity: modelNode.entity,
	appearance: modelNode.mapNode.appearance,
	annotationCount,
});

const createModel = (graph: GraphSubset, projection: MapProjection) => {
	const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const mapNodesById = new Map(projection.nodes.map((mapNode) => [mapNode.id, mapNode]));
	const childrenByParent = new Map<string, MapNode[]>();

	for (const mapNode of projection.nodes) {
		const parentId = mapNode.enclosure?.parentId;
		if (!parentId || !mapNodesById.has(parentId)) continue;

		const children = childrenByParent.get(parentId) ?? [];
		children.push(mapNode);
		childrenByParent.set(parentId, children);
	}

	const toModelNode = (mapNode: MapNode): LayoutModelNode => ({
		mapNode,
		entity: entitiesById.get(mapNode.id)!,
		children: (childrenByParent.get(mapNode.id) ?? [])
			.slice()
			.sort((left, right) => left.id.localeCompare(right.id))
			.map(toModelNode),
	});

	const rootNodes = projection.nodes
		.filter((mapNode) => !mapNode.enclosure)
		.map(toModelNode)
		.filter((modelNode) => isArchitectureCategory(modelNode.entity.category))
		.sort((left, right) => left.entity.id.localeCompare(right.entity.id));

	const actors = projection.nodes
		.map(toModelNode)
		.filter((modelNode) => isActorCategory(modelNode.entity.category))
		.sort((left, right) => left.entity.id.localeCompare(right.entity.id));

	return { rootNodes, actors };
};

const createElkNode = (modelNode: LayoutModelNode): ElkNode => {
	const isGroup = modelNode.mapNode.appearance === "group";
	const dimensions: Partial<Size> = isGroup ? {} : COMPACT_NODE_SIZE;
	const children = modelNode.children.map((child) => createElkNode(child));

	return {
		id: modelNode.entity.id,
		...(dimensions.width ? { width: dimensions.width } : {}),
		...(dimensions.height ? { height: dimensions.height } : {}),
		...(children.length ? { children } : {}),
		...(isGroup
			? {
					layoutOptions: {
						"org.eclipse.elk.padding": GROUP_PADDING,
						"org.eclipse.elk.layered.nodePlacement.strategy": "BRANDES_KOEPF",
						"org.eclipse.elk.hierarchyHandling": "INCLUDE_CHILDREN",
					},
				}
			: {}),
	};
};

const isGroupNode = (node: LayoutModelNode) => node.mapNode.appearance === "group";

const fallbackDimensions = (modelNode: LayoutModelNode): Size =>
	isGroupNode(modelNode)
		? GROUP_NODE_MIN_SIZE
		: COMPACT_NODE_SIZE;

const absolutePositionedNodes = (
	node: ElkNode,
	parentX: number,
	parentY: number,
	positioned: PositionedNode[],
	modelById: ReadonlyMap<string, LayoutModelNode>,
	annotationCounts: ReadonlyMap<string, number>,
	parentId?: string
): void => {
	for (const child of node.children ?? []) {
		const modelNode = modelById.get(child.id);
		if (!modelNode) continue;

		const localX = child.x ?? 0;
		const localY = child.y ?? 0;
		const x = parentX + localX;
		const y = parentY + localY;
		const fallback = fallbackDimensions(modelNode);
		const width = child.width ?? fallback.width;
		const height = child.height ?? fallback.height;
		const bounds: Bounds = { x, y, width, height };

		positioned.push({
			node: {
				id: modelNode.entity.id,
				position: { x: localX, y: localY },
				data: nodeData(modelNode, annotationCounts.get(modelNode.entity.id) ?? 0),
				type: "system-map-node",
				...(parentId ? { parentId } : {}),
				width,
				height,
				ariaLabel: modelNode.entity.label,
			},
			...bounds,
		});

		absolutePositionedNodes(child, x, y, positioned, modelById, annotationCounts, modelNode.entity.id);
	}
};

const flowNodesFromLayout = (
	root: ElkNode,
	rootModels: readonly LayoutModelNode[],
	annotationCounts: ReadonlyMap<string, number>
) => {
	const modelById = new Map<string, LayoutModelNode>();
	const register = (modelNode: LayoutModelNode) => {
		modelById.set(modelNode.entity.id, modelNode);
		modelNode.children.forEach(register);
	};
	rootModels.forEach(register);

	const positioned: PositionedNode[] = [];
	absolutePositionedNodes(root, 0, 0, positioned, modelById, annotationCounts);

	const positionedById = new Map(positioned.map((entry) => [entry.node.id, entry]));
	const nodes: FlowNode[] = [];

	const appendInFlowOrder = (modelNode: LayoutModelNode) => {
		const entry = positionedById.get(modelNode.entity.id);
		if (!entry) return;

		const parentId = modelNode.mapNode.enclosure?.parentId;
		const node: FlowNode = {
			...entry.node,
			...(parentId ? { extent: "parent" as const } : {}),
			zIndex: modelNode.mapNode.appearance === "group" ? 0 : 2,
		};
		nodes.push(node);
		modelNode.children.forEach(appendInFlowOrder);
	};
	rootModels.forEach(appendInFlowOrder);

	return { nodes, architecturePositions: positioned };
};

const actorNodes = (
	actors: readonly LayoutModelNode[],
	positions: readonly PositionedNode[],
	annotationCounts: ReadonlyMap<string, number>
): FlowNode[] => {
	// Actors are positioned downstream of architecture so toggling context cannot repack it.
	const maxX = positions.reduce((curr, {x, width}) => Math.max(curr, x + width), 0);
	const minY = positions.reduce((curr, {y}) => Math.min(curr, y), 0);

	return actors.map((actor, index) => {
		const { width, height } = COMPACT_NODE_SIZE;
		const position = {
			x: maxX + ACTOR_GAP,
			y: minY + index * (height + 24),
		};
		return {
			id: actor.entity.id,
			data: nodeData(actor, annotationCounts.get(actor.entity.id) ?? 0),
			type: "system-map-node",
			position,
			width,
			height,
			zIndex: 3,
			ariaLabel: actor.entity.label,
		};
	});
};

const flowEdges = (connections: readonly MapConnection[], nodes: readonly FlowNode[]): FlowEdge[] => {
	const lanes = buildConnectionLanes(connections);
	const boundsByNodeId = worldBoundsByNodeId(nodes);

	return connections.map((connection) => {
		const lane = lanes.get(connection.id)!;
		const sourceBounds = boundsByNodeId.get(connection.endpoints[0]);
		const targetBounds = boundsByNodeId.get(connection.endpoints[1]);
		const handles = sourceBounds && targetBounds
			? connectionHandleAssignment(sourceBounds, targetBounds)
			: undefined;
		return {
			id: connection.id,
			source: connection.endpoints[0],
			target: connection.endpoints[1],
			...(handles
				? { sourceHandle: handles.sourceHandle, targetHandle: handles.targetHandle }
				: {}),
			type: "system-map-connection",
			markerEnd: { type: MarkerType.ArrowClosed },
			interactionWidth: 24,
			ariaLabel: `${connection.classification} ${connection.predicate.replaceAll("_", " ")}, ${connection.sourceRelationshipIds.length} relationship${connection.sourceRelationshipIds.length === 1 ? "" : "s"}`,
			data: {
				connection,
				laneOffset: lane.offset,
			},
		};
	});
};

/** Converts the projection to ELK input and its result to renderer-owned Flow geometry. */
export const layoutWithElk = async (
	elk: ElkInstance,
	graph: GraphSubset,
	projection: MapProjection
): Promise<LayoutResult> => {
	const { rootNodes, actors } = createModel(graph, projection);
	
	const annoCounts = new Map<string, number>();
	projection.annotations.forEach(({representativeId: repId}) => {
		annoCounts.set(repId, (annoCounts.get(repId) ?? 0) + 1);
	});

	const entityIds = new Set<string>();
	const registerArchitecture = ({entity, children}: LayoutModelNode) => {
		entityIds.add(entity.id);
		children.forEach(registerArchitecture)
	};
	rootNodes.forEach(registerArchitecture);

	const entityConnections = projection.connections.filter(
		({endpoints}) => entityIds.has(endpoints[0]) && entityIds.has(endpoints[1]));

	const layoutOptions = {
		"elk.algorithm": "layered",
		"elk.direction": "RIGHT",
		"elk.hierarchyHandling": "INCLUDE_CHILDREN",
		"elk.edgeRouting": "ORTHOGONAL",
		"elk.padding": ROOT_PADDING,
		"elk.spacing.nodeNode": NODE_SPACING,
		"elk.layered.spacing.nodeNodeBetweenLayers": LAYER_SPACING,
		"elk.layered.nodePlacement.strategy": "BRANDES_KOEPF",
		"elk.layered.considerModelOrder": "true",
	};

	const elkGraph: ElkNode = {
		id: "system-map",
		children: rootNodes.map(createElkNode),
		// ELK positions nodes; the renderer computes its own connection paths from this geometry.
		edges: entityConnections.map(({id, endpoints}) => ({id, sources: [endpoints[0]], targets: [endpoints[1]]})),
		layoutOptions,
	};

	const result = await elk.layout(elkGraph);
	const { nodes, architecturePositions } = flowNodesFromLayout(result, rootNodes, annoCounts);
	const allNodes = [...nodes, ...actorNodes(actors, architecturePositions, annoCounts)];

	return {
		nodes: allNodes,
		edges: flowEdges(projection.connections, allNodes),
	};
};

/**
 * Translates the new layout around an explicit source anchor. ELK can resize a group when detail
 * changes; shifting root nodes keeps the selected or pointer representation under the same
 * viewport location without rewriting parent-relative coordinates or fitting the whole diagram.
 * Alignment preserves one meaningful anchor, not all previous positions; ELK may legitimately repack
 * the rest of the graph. Without an anchor, the fresh layout is left in its own coordinate system;
 * choosing an arbitrary common ID would move the user's context unpredictably.
 */
export const alignLayoutToPrevious = (
	next: LayoutResult,
	previous: LayoutResult | undefined,
	anchorId?: string,
	previousAnchorId = anchorId
): LayoutResult => {
	if (!previous || !next.nodes.length || !previous.nodes.length) return next;
	if (!anchorId || !previousAnchorId) return next;

	const nextBounds = worldBoundsByNodeId(next.nodes);
	const nextAnchorBounds = nextBounds.get(anchorId);
	const previousBounds = worldBoundsByNodeId(previous.nodes);
	const previousAnchorBounds = previousBounds.get(previousAnchorId);
	if (!nextAnchorBounds || !previousAnchorBounds) return next;

	const nextAnchor = worldCenter(nextAnchorBounds);
	const previousAnchor = worldCenter(previousAnchorBounds);
	const delta = {
		x: previousAnchor.x - nextAnchor.x,
		y: previousAnchor.y - nextAnchor.y,
	};
	if (delta.x === 0 && delta.y === 0) {
		return next;
	}

	const nodes = next.nodes.map((node) => {
		if (node.parentId) return node;
		const position = { x: node.position.x + delta.x, y: node.position.y + delta.y };
		return { ...node, position };
	});
	return { ...next, nodes };
};

/** Finds an architectural anchor shared by both layouts near the supplied screen point. */
export const nearestSurvivingArchitectureAnchor = (
	next: LayoutResult,
	previous: LayoutResult,
	viewport: Viewport,
	referenceScreenPoint: Point
) => {
	const nextArchitectureIds = new Set(
		next.nodes
			.filter((node) => isArchitectureCategory(node.data?.entity.category ?? ""))
			.map((node) => node.id)
	);
	const previousArchitectureNodes = previous.nodes.filter(
		(node) =>
			nextArchitectureIds.has(node.id) &&
			isArchitectureCategory(node.data?.entity.category ?? "")
	);
	if (!previousArchitectureNodes.length) return undefined;

	const previousBounds = worldBoundsByNodeId(previous.nodes);
	let nearestId = ""
	let nearestDist = 0;
	const { x: refX, y: refY } = referenceScreenPoint;
	for (const node of previousArchitectureNodes) {
		const bounds = previousBounds.get(node.id);
		if (!bounds) continue;

		const {x: cx, y: cy} = screenPointFromWorld(viewport, worldCenter(bounds));
		const distance = (cx - refX) ** 2 + (cy - refY) ** 2;
		const isCloser = distance < nearestDist || (distance === nearestDist && node.id < nearestId)
		if (!nearestId || isCloser) {
			nearestId = node.id;
			nearestDist = distance;
		}
	}
	if (!nearestId) return;

	return { nextId: nearestId, previousId: nearestId };
};
