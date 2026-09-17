import type { ELK as ElkInstance, ElkEdgeSection, ElkExtendedEdge, ElkNode } from "elkjs/lib/elk-api.js";
import { MarkerType } from "@xyflow/svelte";

import { isArchitectureCategory } from "$features/systems/lib/system-map/category";
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
import type { MapNode, MapProjection } from "$features/systems/lib/system-map/presentation";
import type {
	ConnectionRoute,
	ConnectionRouteSection,
	FlowEdge,
	FlowNode,
	FlowNodeData,
	LayoutResult,
} from "./flow-model";

export const COMPACT_NODE_SIZE = { width: 240, height: 96 } as const satisfies Readonly<Size>;
export const GROUP_NODE_MIN_SIZE = { width: 280, height: 168 } as const satisfies Readonly<Size>;

const GROUP_PADDING = "[top=48,left=28,bottom=28,right=28]";
const ROOT_PADDING = "[top=32,left=32,bottom=32,right=32]";
const NODE_SPACING = "28";
const LAYER_SPACING = "72";

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

	return { rootNodes };
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

const pointIsFinite = (point: Point | undefined): point is Point =>
	point !== undefined && Number.isFinite(point.x) && Number.isFinite(point.y);

const pointOnBoundsBoundary = (point: Point, bounds: Bounds): boolean => {
	const epsilon = 0.001;
	const onHorizontalBoundary =
		(Math.abs(point.y - bounds.y) <= epsilon || Math.abs(point.y - (bounds.y + bounds.height)) <= epsilon) &&
		point.x >= bounds.x - epsilon &&
		point.x <= bounds.x + bounds.width + epsilon;
	const onVerticalBoundary =
		(Math.abs(point.x - bounds.x) <= epsilon || Math.abs(point.x - (bounds.x + bounds.width)) <= epsilon) &&
		point.y >= bounds.y - epsilon &&
		point.y <= bounds.y + bounds.height + epsilon;
	return onHorizontalBoundary || onVerticalBoundary;
};

const routeSectionPoints = (section: ConnectionRouteSection): Point[] => [
	section.start,
	...section.bends,
	section.end,
];

const labelPositionForSections = (sections: readonly ConnectionRouteSection[]): Point => {
	type Segment = { start: Point; end: Point; length: number; horizontal: boolean };

	const segments: Segment[] = [];
	for (const section of sections) {
		const points = routeSectionPoints(section);
		for (let index = 1; index < points.length; index += 1) {
			const start = points[index - 1];
			const end = points[index];
			const length = Math.hypot(end.x - start.x, end.y - start.y);
			if (length === 0) continue;
			segments.push({
				start,
				end,
				length,
				horizontal: start.y === end.y,
			});
		}
	}

	const horizontalSegments = segments.filter((segment) => segment.horizontal);
	const candidates = horizontalSegments.length ? horizontalSegments : segments;
	if (!candidates.length) throw new Error("ELK returned a route without a usable segment");
	const segment = candidates.reduce((longest, candidate) =>
		candidate.length > longest.length ? candidate : longest
	);
	return {
		x: (segment.start.x + segment.end.x) / 2,
		y: (segment.start.y + segment.end.y) / 2,
	};
};

const collectElkEdges = (node: ElkNode): ElkExtendedEdge[] => [
	...(node.edges ?? []),
	...(node.children ?? []).flatMap(collectElkEdges),
];

const routeFromElkEdge = (
	edge: ElkExtendedEdge,
	connection: MapProjection["connections"][number],
	worldOriginByNodeId: ReadonlyMap<string, Point>,
	boundsByNodeId: ReadonlyMap<string, Bounds>
): ConnectionRoute => {
	const elkSections = edge.sections ?? [];
	if (!elkSections.length) {
		throw new Error(`ELK returned no route sections for connection ${connection.id}`);
	}

	const containerOrigin = edge.container
		? worldOriginByNodeId.get(edge.container)
		: worldOriginByNodeId.get("system-map") ?? { x: 0, y: 0 };
	if (!containerOrigin) {
		throw new Error(`ELK returned an unknown edge container for connection ${connection.id}`);
	}

	const toWorldPoint = (point: { x: number; y: number }): Point => ({
		x: point.x + containerOrigin.x,
		y: point.y + containerOrigin.y,
	});

	const sections = elkSections.map((section: ElkEdgeSection): ConnectionRouteSection => {
		const start = toWorldPoint(section.startPoint);
		const bends = (section.bendPoints ?? []).map(toWorldPoint);
		const end = toWorldPoint(section.endPoint);
		const points = [start, ...bends, end];
		if (!points.every(pointIsFinite)) {
			throw new Error(`ELK returned a malformed route for connection ${connection.id}`);
		}
		for (let index = 1; index < points.length; index += 1) {
			const previous = points[index - 1];
			const current = points[index];
			if (previous.x !== current.x && previous.y !== current.y) {
				throw new Error(`ELK returned a non-orthogonal route for connection ${connection.id}`);
			}
		}
		return { start, bends, end };
	});

	const sourceSectionIndexes = elkSections
		.map((section, index) => (section.incomingShape === connection.endpoints[0] ? index : -1))
		.filter((index) => index >= 0);
	const targetSectionIndexes = elkSections
		.map((section, index) => (section.outgoingShape === connection.endpoints[1] ? index : -1))
		.filter((index) => index >= 0);
	if (sourceSectionIndexes.length !== 1 || targetSectionIndexes.length !== 1) {
		throw new Error(`ELK returned an ambiguous terminal route for connection ${connection.id}`);
	}

	const sourceBounds = boundsByNodeId.get(connection.endpoints[0]);
	const targetBounds = boundsByNodeId.get(connection.endpoints[1]);
	if (!sourceBounds || !targetBounds) {
		throw new Error(`ELK route endpoints are not displayed for connection ${connection.id}`);
	}
	if (
		!pointOnBoundsBoundary(sections[sourceSectionIndexes[0]].start, sourceBounds) ||
		!pointOnBoundsBoundary(sections[targetSectionIndexes[0]].end, targetBounds)
	) {
		throw new Error(`ELK route does not terminate at connection ${connection.id}'s node boundaries`);
	}

	return {
		sections,
		labelPosition: labelPositionForSections(sections),
		targetSectionIndex: targetSectionIndexes[0],
	};
};

const routesFromElk = (
	result: ElkNode,
	connections: readonly MapProjection["connections"][number][],
	architecturePositions: readonly PositionedNode[]
): ReadonlyMap<string, ConnectionRoute> => {
	const edgesById = new Map(collectElkEdges(result).map((edge) => [edge.id, edge]));
	const worldOriginByNodeId = new Map<string, Point>([["system-map", { x: 0, y: 0 }]]);
	const boundsByNodeId = new Map<string, Bounds>();
	for (const positioned of architecturePositions) {
		worldOriginByNodeId.set(positioned.node.id, { x: positioned.x, y: positioned.y });
		boundsByNodeId.set(positioned.node.id, positioned);
	}

	const routes = new Map<string, ConnectionRoute>();
	for (const connection of connections) {
		const edge = edgesById.get(connection.id);
		if (!edge) throw new Error(`ELK returned no edge for connection ${connection.id}`);
		routes.set(
			connection.id,
			routeFromElkEdge(edge, connection, worldOriginByNodeId, boundsByNodeId)
		);
	}
	return routes;
};

const flowEdges = (
	connections: readonly MapProjection["connections"][number][],
	routes: ReadonlyMap<string, ConnectionRoute>
): FlowEdge[] =>
	connections.map((connection) => {
		const route = routes.get(connection.id);
		if (!route) throw new Error(`Missing ELK route for connection ${connection.id}`);

		return {
			id: connection.id,
			source: connection.endpoints[0],
			target: connection.endpoints[1],
			type: "system-map-connection",
			markerEnd: { type: MarkerType.ArrowClosed },
			interactionWidth: 24,
			ariaLabel: `${connection.classification} ${connection.predicate.replaceAll("_", " ")}, ${connection.sourceRelationshipIds.length} relationship${connection.sourceRelationshipIds.length === 1 ? "" : "s"}`,
			data: { connection, route },
		};
	});

/** Converts the projection to an ELK graph and preserves ELK's node/route output for Flow. */
export const layoutWithElk = async (
	elk: ElkInstance,
	graph: GraphSubset,
	projection: MapProjection
): Promise<LayoutResult> => {
	const { rootNodes } = createModel(graph, projection);

	const annoCounts = new Map<string, number>();
	projection.annotations.forEach(({ representativeId: repId }) => {
		annoCounts.set(repId, (annoCounts.get(repId) ?? 0) + 1);
	});

	const entityIds = new Set<string>();
	const registerArchitecture = ({ entity, children }: LayoutModelNode) => {
		entityIds.add(entity.id);
		children.forEach(registerArchitecture);
	};
	rootNodes.forEach(registerArchitecture);

	const entityConnections = projection.connections;
	for (const connection of entityConnections) {
		if (!entityIds.has(connection.endpoints[0]) || !entityIds.has(connection.endpoints[1])) {
			throw new Error(`Connection ${connection.id} has a non-layout endpoint`);
		}
	}

	const layoutOptions = {
		"elk.algorithm": "layered",
		"elk.direction": "RIGHT",
		"elk.hierarchyHandling": "INCLUDE_CHILDREN",
		"elk.edgeRouting": "ORTHOGONAL",
		"org.eclipse.elk.layered.mergeEdges": "false",
		"org.eclipse.elk.layered.mergeHierarchyEdges": "false",
		// Keep edge route points in the root coordinate system. Node coordinates remain parent-relative.
		"org.eclipse.elk.json.edgeCoords": "ROOT",
		"elk.padding": ROOT_PADDING,
		"elk.spacing.nodeNode": NODE_SPACING,
		"elk.layered.spacing.nodeNodeBetweenLayers": LAYER_SPACING,
		"elk.layered.nodePlacement.strategy": "BRANDES_KOEPF",
		"elk.layered.considerModelOrder": "true",
	};

	const elkGraph: ElkNode = {
		id: "system-map",
		children: rootNodes.map(createElkNode),
		edges: entityConnections.map(({ id, endpoints }) => ({
			id,
			sources: [endpoints[0]],
			targets: [endpoints[1]],
		})),
		layoutOptions,
	};

	const result = await elk.layout(elkGraph);
	const { nodes, architecturePositions } = flowNodesFromLayout(result, rootNodes, annoCounts);
	const routes = routesFromElk(result, entityConnections, architecturePositions);

	return {
		nodes,
		edges: flowEdges(entityConnections, routes),
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
	const edges = next.edges.map((edge) => {
		const data = edge.data;
		const route = data?.route;
		if (!data || !route) throw new Error(`Cannot align connection ${edge.id} without an ELK route`);

		return {
			...edge,
			data: {
				...data,
				route: {
					...route,
					sections: route.sections.map((section) => ({
						start: { x: section.start.x + delta.x, y: section.start.y + delta.y },
						bends: section.bends.map((point) => ({
							x: point.x + delta.x,
							y: point.y + delta.y,
						})),
						end: { x: section.end.x + delta.x, y: section.end.y + delta.y },
					})),
					labelPosition: {
						x: route.labelPosition.x + delta.x,
						y: route.labelPosition.y + delta.y,
					},
				},
			},
		};
	});
	return { ...next, nodes, edges };
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
