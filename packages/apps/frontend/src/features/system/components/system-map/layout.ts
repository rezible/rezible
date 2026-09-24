import type { ELK as ElkInstance, ElkNode } from "elkjs/lib/elk-api.js";
import { MarkerType } from "@xyflow/svelte";

import {
	getMapCategoryDisplay,
	isArchitectureCategory,
	NodeDetailLevel,
} from "$features/system/lib/system-map/category";
import type { GraphEntity, GraphSubset } from "$features/system/lib/system-map/graph";
import {
	worldBoundsByNodeId,
	worldCenter,
	screenPointFromWorld,
	type Bounds,
	type Point,
	type Size,
	type Viewport,
} from "$features/system/lib/system-map/geometry";
import type { MapNode, MapProjection } from "$features/system/lib/system-map/presentation";
import type {
	ConnectionRoute,
	ConnectionRouteVariant,
	FlowEdge,
	FlowNode,
	FlowNodeData,
	LayoutResult,
} from "./flow-model";
import { ProjectedRouteCache } from "./routing";

const NODE_MIN_SIZES = {
	[NodeDetailLevel.Landscape]: {
		compact: { width: 320, height: 136 },
		group: { width: 360, height: 256 },
	},
	[NodeDetailLevel.Systems]: {
		compact: { width: 280, height: 120 },
		group: { width: 320, height: 240 },
	},
	[NodeDetailLevel.Runtime]: {
		compact: { width: 240, height: 104 },
		group: { width: 280, height: 224 },
	},
	[NodeDetailLevel.Implementation]: {
		compact: { width: 200, height: 88 },
		group: { width: 240, height: 208 },
	},
} as const satisfies Readonly<Record<NodeDetailLevel, Readonly<Record<"compact" | "group", Readonly<Size>>>>>;

export const GROUP_HEADER_LAYOUT = {
	contentPadding: 12,
	categoryHeight: 16,
	titleGap: 8,
	titleHeight: 36,
	titleLineHeight: 18,
	childGap: 16,
} as const;

export const nodeMinSizeForCategory = (category: string, appearance: "compact" | "group"): Readonly<Size> => {
	const level = getMapCategoryDisplay(category).level ?? NodeDetailLevel.Systems;
	return NODE_MIN_SIZES[level][appearance];
};

const GROUP_PADDING_TOP =
	GROUP_HEADER_LAYOUT.contentPadding +
	GROUP_HEADER_LAYOUT.categoryHeight +
	GROUP_HEADER_LAYOUT.titleGap +
	GROUP_HEADER_LAYOUT.titleHeight +
	GROUP_HEADER_LAYOUT.childGap;
const GROUP_PADDING_SIDE = 16;
const GROUP_PADDING_BOTTOM = 24;
const GROUP_PADDING = `[top=${GROUP_PADDING_TOP},left=${GROUP_PADDING_SIDE},bottom=${GROUP_PADDING_BOTTOM},right=${GROUP_PADDING_SIDE}]`;
const ROOT_PADDING = "[top=32,left=32,bottom=32,right=32]";
const NODE_SPACING = "28";
const LAYER_SPACING = "72";
const INTERACTIVE_LAYOUT_OPTIONS = {
	"org.eclipse.elk.interactive": "true",
	"org.eclipse.elk.layered.layering.strategy": "INTERACTIVE",
	"org.eclipse.elk.layered.crossingMinimization.strategy": "INTERACTIVE",
	"org.eclipse.elk.layered.nodePlacement.strategy": "INTERACTIVE",
};

type LayoutModelNode = {
	mapNode: MapNode;
	entity: GraphEntity;
	children: LayoutModelNode[];
};

type PositionedNode = { node: FlowNode } & Bounds;

export type LayoutPositionInputs = {
	/** Host-owned sparse world-space starting hints. */
	positions?: Readonly<Partial<Record<string, Point>>>;
	/** Previous full-layout coordinates used only to translate nested world-space hints for ELK. */
	previousWorldPositions?: Readonly<Partial<Record<string, Point>>>;
};

/** Stable comparison key for only the inputs that can change full canvas geometry or routes. */
export const systemMapLayoutInputKey = (
	graph: GraphSubset,
	projection: MapProjection,
	positions?: Readonly<Partial<Record<string, Point>>>
): string => {
	const architectureIds = new Set(
		graph.entities.filter((entity) => isArchitectureCategory(entity.category)).map((entity) => entity.id)
	);
	const coordinateKey = (value: number | undefined) =>
		value === undefined ? "undefined" : Number.isFinite(value) ? value : String(value);
	return JSON.stringify([
		graph.entities
			.filter((entity) => architectureIds.has(entity.id))
			.map((entity) => [entity.id, entity.category])
			.sort(([leftId], [rightId]) => String(leftId).localeCompare(String(rightId))),
		projection.nodes
			.map((node) => [node.id, node.appearance, node.enclosure?.parentId, node.enclosure?.membershipId])
			.sort(([leftId], [rightId]) => String(leftId).localeCompare(String(rightId))),
		projection.connections
			.map((connection) => [
				connection.id,
				connection.endpoints,
				connection.predicate,
				connection.classification,
				connection.sourceRelationshipIds,
			])
			.sort(([leftId], [rightId]) => String(leftId).localeCompare(String(rightId))),
		[...architectureIds]
			.filter((entityId) => Object.prototype.hasOwnProperty.call(positions ?? {}, entityId))
			.sort((left, right) => left.localeCompare(right))
			.map((entityId) => [
				entityId,
				coordinateKey(positions?.[entityId]?.x),
				coordinateKey(positions?.[entityId]?.y),
			]),
	]);
};

const nodeData = (modelNode: LayoutModelNode): FlowNodeData => ({
	entity: modelNode.entity,
	appearance: modelNode.mapNode.appearance,
	opacity: modelNode.mapNode.opacity ?? 1,
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

const createElkNode = (
	modelNode: LayoutModelNode,
	startingPositions: ReadonlyMap<string, Point>
): ElkNode => {
	const isGroup = modelNode.mapNode.appearance === "group";
	const dimensions = nodeMinSizeForCategory(modelNode.entity.category, isGroup ? "group" : "compact");
	const children = modelNode.children.map((child) => createElkNode(child, startingPositions));
	const position = startingPositions.get(modelNode.entity.id);

	return {
		id: modelNode.entity.id,
		...(position ? { x: position.x, y: position.y } : {}),
		...(dimensions.width ? { width: dimensions.width } : {}),
		...(dimensions.height ? { height: dimensions.height } : {}),
		...(children.length ? { children } : {}),
		layoutOptions: {
			"org.eclipse.elk.nodeSize.constraints": "MINIMUM_SIZE",
			"org.eclipse.elk.nodeSize.minimum": `(${dimensions.width},${dimensions.height})`,
			...(isGroup
				? {
						"org.eclipse.elk.padding": GROUP_PADDING,
						"org.eclipse.elk.hierarchyHandling": "INCLUDE_CHILDREN",
						...INTERACTIVE_LAYOUT_OPTIONS,
					}
				: {}),
		},
	};
};

const isGroupNode = (node: LayoutModelNode) => node.mapNode.appearance === "group";

const fallbackDimensions = (modelNode: LayoutModelNode): Size =>
	nodeMinSizeForCategory(modelNode.entity.category, isGroupNode(modelNode) ? "group" : "compact") as Size;

const absolutePositionedNodes = (
	node: ElkNode,
	parentX: number,
	parentY: number,
	positioned: PositionedNode[],
	modelById: ReadonlyMap<string, LayoutModelNode>,
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
		const dimensions = {
			width: child.width ?? fallback.width,
			height: child.height ?? fallback.height,
		};
		const width = dimensions.width;
		const height = dimensions.height;
		const bounds: Bounds = { x, y, width, height };

		positioned.push({
			node: {
				id: modelNode.entity.id,
				position: { x: localX, y: localY },
				data: nodeData(modelNode),
				type: "system-map-node",
				...(parentId ? { parentId } : {}),
				width,
				height,
				ariaLabel: modelNode.entity.kind || modelNode.entity.id,
			},
			...bounds,
		});

		absolutePositionedNodes(child, x, y, positioned, modelById, modelNode.entity.id);
	}
};

const flowNodesFromLayout = (root: ElkNode, rootModels: readonly LayoutModelNode[]) => {
	const modelById = new Map<string, LayoutModelNode>();
	const register = (modelNode: LayoutModelNode) => {
		modelById.set(modelNode.entity.id, modelNode);
		modelNode.children.forEach(register);
	};
	rootModels.forEach(register);
	const positioned: PositionedNode[] = [];
	absolutePositionedNodes(root, 0, 0, positioned, modelById);

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

	return nodes;
};

const pointIsFinite = (point: Point | undefined): point is Point =>
	point !== undefined && Number.isFinite(point.x) && Number.isFinite(point.y);

/** Resolves explicit world-space hints to the parent-relative coordinates ELK expects. */
export const hintedStartingPositions = (
	graph: GraphSubset,
	projection: MapProjection,
	inputs: LayoutPositionInputs = {}
): ReadonlyMap<string, Point> => {
	const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
	for (const entity of graph.entities) {
		if (!isArchitectureCategory(entity.category)) continue;
		if (!Object.prototype.hasOwnProperty.call(inputs.positions ?? {}, entity.id)) continue;
		if (!pointIsFinite(inputs.positions?.[entity.id])) {
			throw new Error(
				`Position hint for architectural entity ${entity.id} must have finite coordinates.`
			);
		}
	}

	const startingPositions = new Map<string, Point>();
	for (const mapNode of projection.nodes) {
		const entity = entitiesById.get(mapNode.id);
		if (!entity || !isArchitectureCategory(entity.category)) continue;

		const explicitPosition = inputs.positions?.[entity.id];
		if (!pointIsFinite(explicitPosition)) continue;

		const parentId = mapNode.enclosure?.parentId;
		const explicitParentPosition = parentId ? inputs.positions?.[parentId] : undefined;
		const parentWorldPosition = parentId
			? pointIsFinite(explicitParentPosition)
				? explicitParentPosition
				: (inputs.previousWorldPositions?.[parentId] ?? { x: 0, y: 0 })
			: { x: 0, y: 0 };
		if (!pointIsFinite(parentWorldPosition)) continue;
		startingPositions.set(entity.id, {
			x: explicitPosition.x - parentWorldPosition.x,
			y: explicitPosition.y - parentWorldPosition.y,
		});
	}
	return startingPositions;
};

const flowEdges = (
	connections: readonly MapProjection["connections"][number][],
	routes: ReadonlyMap<string, ConnectionRoute>,
	routeVariantsById?: ReadonlyMap<string, readonly ConnectionRouteVariant[]>
): FlowEdge[] =>
	connections.map((connection) => {
		const routeVariants = routeVariantsById?.get(connection.id);
		const visibleRoute = routeVariants?.reduce(
			(best, variant) => (variant.opacity > best.opacity ? variant : best),
			routeVariants[0]
		);
		const route = visibleRoute?.route ?? routes.get(connection.id);
		if (!route) throw new Error(`Missing visible route for connection ${connection.id}`);
		const opacity = routeVariants
			? Math.max(...routeVariants.map((variant) => variant.opacity))
			: (connection.opacity ?? 1);

		return {
			id: connection.id,
			source: connection.endpoints[0],
			target: connection.endpoints[1],
			type: "system-map-connection",
			markerEnd: { type: MarkerType.ArrowClosed },
			interactionWidth: 24,
			ariaLabel: `${connection.classification} ${connection.predicate.replaceAll("_", " ")}, ${connection.sourceRelationshipIds.length} relationship${connection.sourceRelationshipIds.length === 1 ? "" : "s"}`,
			data: {
				connection,
				route,
				opacity,
				...(routeVariants ? { routeVariants } : {}),
			},
		};
	});

export type MapRouteTransition = {
	lowerProjection: MapProjection;
	upperProjection: MapProjection;
	upperLevel: number;
	progress: number;
};

const projectedNodesFromLayout = (fullLayout: LayoutResult, projection: MapProjection): FlowNode[] => {
	const fullNodesById = new Map(fullLayout.nodes.map((node) => [node.id, node]));
	const fullBoundsById = worldBoundsByNodeId(fullLayout.nodes);
	const visibleIds = new Set(projection.nodes.map((node) => node.id));
	return projection.nodes.map((mapNode) => {
		const fullNode = fullNodesById.get(mapNode.id);
		const bounds = fullBoundsById.get(mapNode.id);
		if (!fullNode || !bounds) throw new Error(`Missing full-layout geometry for entity ${mapNode.id}`);

		const parentId = mapNode.enclosure?.parentId;
		const visibleParentId = parentId && visibleIds.has(parentId) ? parentId : undefined;
		const parentBounds = visibleParentId ? fullBoundsById.get(visibleParentId) : undefined;
		return {
			...fullNode,
			data: { ...fullNode.data, appearance: mapNode.appearance, opacity: mapNode.opacity ?? 1 },
			position: {
				x: bounds.x - (parentBounds?.x ?? 0),
				y: bounds.y - (parentBounds?.y ?? 0),
			},
			...(visibleParentId
				? { parentId: visibleParentId, extent: "parent" as const }
				: { parentId: undefined, extent: undefined }),
			zIndex: mapNode.appearance === "group" ? 0 : 2,
		};
	});
};

const mergeConnections = (connections: readonly MapProjection["connections"][number][]) => {
	const merged = new Map<string, MapProjection["connections"][number]>();
	for (const connection of connections) {
		const previous = merged.get(connection.id);
		if (!previous) {
			merged.set(connection.id, connection);
			continue;
		}
		merged.set(connection.id, {
			...previous,
			sourceRelationshipIds: [
				...new Set([...previous.sourceRelationshipIds, ...connection.sourceRelationshipIds]),
			].sort(),
			opacity: Math.max(previous.opacity ?? 1, connection.opacity ?? 1),
		});
	}
	return [...merged.values()];
};

const endpointContextOpacity = (
	connection: MapProjection["connections"][number],
	projection: MapProjection,
	fullNodesById: ReadonlyMap<string, FlowNode>,
	upperLevel: number,
	isUpperProjection: boolean
): number => {
	const visibleNodes = new Map(projection.nodes.map((node) => [node.id, node]));
	const opacity = connection.endpoints.map((endpointId) => {
		const node = visibleNodes.get(endpointId);
		if (!node) return connection.opacity ?? 1;
		const entity = fullNodesById.get(endpointId)?.data.entity;
		const entityLevel = entity ? getMapCategoryDisplay(entity.category).level : undefined;
		// The upper layer's node fade is already represented by the route blend progress.
		if (isUpperProjection && entityLevel === upperLevel) return 1;
		return node.opacity ?? 1;
	});
	return Math.min(...opacity);
};

/** Applies a visible projection to fixed full-layout geometry without running ELK. */
export const presentMapProjection = (
	fullLayout: LayoutResult,
	projection: MapProjection,
	routeCache = new ProjectedRouteCache(),
	transition?: MapRouteTransition
): LayoutResult => {
	const nodes = projectedNodesFromLayout(fullLayout, projection);
	const currentRoutes = routeCache.routesFor(nodes, projection);
	if (!transition) return { nodes, edges: flowEdges(projection.connections, currentRoutes) };

	const lowerNodes = projectedNodesFromLayout(fullLayout, transition.lowerProjection);
	const upperNodes = projectedNodesFromLayout(fullLayout, transition.upperProjection);
	const lowerRoutes = routeCache.routesFor(lowerNodes, transition.lowerProjection);
	const upperRoutes = routeCache.routesFor(upperNodes, transition.upperProjection);
	const fullNodesById = new Map(fullLayout.nodes.map((node) => [node.id, node]));
	const connections = mergeConnections([
		...transition.lowerProjection.connections,
		...transition.upperProjection.connections,
		...projection.connections,
	]);
	const routes = new Map<string, ConnectionRoute>();
	const routeVariantsById = new Map<string, readonly ConnectionRouteVariant[]>();
	for (const connection of transition.lowerProjection.connections) {
		const route = lowerRoutes.get(connection.id);
		if (!route) continue;
		const currentOpacity = endpointContextOpacity(
			connection,
			projection,
			fullNodesById,
			transition.upperLevel,
			false
		);
		const variants = routeVariantsById.get(connection.id) ?? [];
		routeVariantsById.set(connection.id, [
			...variants,
			{ route, opacity: (1 - transition.progress) * currentOpacity },
		]);
	}
	for (const connection of transition.upperProjection.connections) {
		const route = upperRoutes.get(connection.id);
		if (!route) continue;
		const currentOpacity = endpointContextOpacity(
			connection,
			projection,
			fullNodesById,
			transition.upperLevel,
			true
		);
		const variants = routeVariantsById.get(connection.id) ?? [];
		routeVariantsById.set(connection.id, [
			...variants,
			{ route, opacity: transition.progress * currentOpacity },
		]);
	}
	for (const connection of projection.connections) {
		if (routeVariantsById.has(connection.id)) continue;
		const route = currentRoutes.get(connection.id);
		if (!route) continue;
		routeVariantsById.set(connection.id, [{ route, opacity: connection.opacity ?? 1 }]);
	}
	for (const connection of connections) {
		const variants = routeVariantsById.get(connection.id);
		if (variants?.length) routes.set(connection.id, variants[0].route);
		else {
			const route = currentRoutes.get(connection.id);
			if (route) {
				routes.set(connection.id, route);
				routeVariantsById.set(connection.id, [{ route, opacity: connection.opacity ?? 1 }]);
			}
		}
	}

	const visibleConnections = connections.filter((connection) => {
		const variants = routeVariantsById.get(connection.id) ?? [];
		return variants.some((variant) => variant.opacity > 0);
	});
	return {
		nodes,
		edges: flowEdges(visibleConnections, routes, routeVariantsById),
	};
};

/** Uses ELK once for full-subset node placement; visible connection routes are projected later. */
export const layoutWithElk = async (
	elk: ElkInstance,
	graph: GraphSubset,
	projection: MapProjection,
	positionInputs: LayoutPositionInputs = {}
): Promise<LayoutResult> => {
	const { rootNodes } = createModel(graph, projection);
	const startingPositions = hintedStartingPositions(graph, projection, positionInputs);

	const entityIds = new Set<string>();
	const registerArchitecture = ({ entity, children }: LayoutModelNode) => {
		entityIds.add(entity.id);
		children.forEach(registerArchitecture);
	};
	rootNodes.forEach(registerArchitecture);
	const suppliedArchitectureIds = new Set(
		graph.entities.filter((entity) => isArchitectureCategory(entity.category)).map((entity) => entity.id)
	);
	if (
		entityIds.size !== suppliedArchitectureIds.size ||
		[...suppliedArchitectureIds].some((entityId) => !entityIds.has(entityId))
	) {
		throw new Error("ELK layout requires every supplied architectural entity.");
	}

	const entityConnections = projection.connections;
	for (const connection of entityConnections) {
		if (!entityIds.has(connection.endpoints[0]) || !entityIds.has(connection.endpoints[1])) {
			throw new Error(`Connection ${connection.id} has a non-layout endpoint`);
		}
	}

	const layoutOptions = {
		...INTERACTIVE_LAYOUT_OPTIONS,
		"elk.algorithm": "layered",
		"elk.direction": "RIGHT",
		"elk.hierarchyHandling": "INCLUDE_CHILDREN",
		"org.eclipse.elk.layered.mergeEdges": "false",
		"org.eclipse.elk.layered.mergeHierarchyEdges": "false",
		"elk.padding": ROOT_PADDING,
		"elk.spacing.nodeNode": NODE_SPACING,
		"elk.layered.spacing.nodeNodeBetweenLayers": LAYER_SPACING,
		"elk.layered.considerModelOrder": "true",
	};

	const elkGraph: ElkNode = {
		id: "system-map",
		children: rootNodes.map((node) => createElkNode(node, startingPositions)),
		edges: entityConnections.map(({ id, endpoints }) => ({
			id,
			sources: [endpoints[0]],
			targets: [endpoints[1]],
		})),
		layoutOptions,
	};

	const result = await elk.layout(elkGraph);
	const nodes = flowNodesFromLayout(result, rootNodes);
	return { nodes, edges: [] };
};

const alignNodeToParent = (node: FlowNode, delta: Point) => {
	if (!!node.parentId) return node;
	const position = { 
		x: node.position.x + delta.x, 
		y: node.position.y + delta.y,
	};
	return { ...node, position };
}

/** Keeps a surviving anchor steady when a new full-subset placement changes the fixed geometry. */
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
	if (delta.x === 0 && delta.y === 0) return next;

	const alignedNodes = next.nodes.map((node) => alignNodeToParent(node, delta));
	return { ...next, nodes: alignedNodes };
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
		(node) => nextArchitectureIds.has(node.id) && isArchitectureCategory(node.data?.entity.category ?? "")
	);
	if (!previousArchitectureNodes.length) return undefined;

	const previousBounds = worldBoundsByNodeId(previous.nodes);
	let nearestId = "";
	let nearestDist = 0;
	const { x: refX, y: refY } = referenceScreenPoint;
	for (const node of previousArchitectureNodes) {
		const bounds = previousBounds.get(node.id);
		if (!bounds) continue;

		const { x: cx, y: cy } = screenPointFromWorld(viewport, worldCenter(bounds));
		const distance = (cx - refX) ** 2 + (cy - refY) ** 2;
		const isCloser = distance < nearestDist || (distance === nearestDist && node.id < nearestId);
		if (!nearestId || isCloser) {
			nearestId = node.id;
			nearestDist = distance;
		}
	}
	if (!nearestId) return;

	return { nextId: nearestId, previousId: nearestId };
};
