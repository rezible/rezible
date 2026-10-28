import type { Edge, Node, NodeProps } from "@xyflow/svelte";

import type { DetailBand } from "$features/system/lib/system-map/detail";
import {
	visibleConnections,
	type ConnectionFocus,
	type ConnectionLevel,
	type MapConnection,
	type RelationshipMode,
} from "$features/system/lib/system-map/rollup";
import {
	connectionAnchors,
	regionPath,
	type Box,
	type ConnectionAnchors,
} from "$features/system/lib/system-map/geometry";
import type { GraphEntity, GraphSubset } from "$features/system/lib/system-map/graph";
import { entityLabel, relationshipPredicateLabel } from "$features/system/lib/system-map/labels";
import { layoutStructure } from "$features/system/lib/system-map/layout";
import {
	placementChips,
	placementCountLine,
	placementFootprint,
	placementTargets,
} from "$features/system/lib/system-map/placement";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { selectedRelationshipIds } from "$features/system/lib/system-map/selection";
import {
	buildStructureTree,
	type StructureCard,
	type StructureNode,
	type StructureRegion,
	type StructureTree,
} from "$features/system/lib/system-map/structure";

/**
 * Back to front: landscapes, systems, connections, cards, then the landscape band's region labels. Connections
 * cross region fills under cards.
 */
export const LANDSCAPE_Z = 0;
export const SYSTEM_Z = 1;
export const CONNECTION_Z = 2;
export const CARD_Z = 3;
export const REGION_LABEL_Z = 4;

type NodeState = {
	band: DetailBand;
	/** Endpoint of a focused or hovered connection, or named by the host. */
	highlighted?: boolean;
	/** Runs on the selected entity. */
	footprint?: boolean;
};

export type RegionNodeData = NodeState & {
	region: StructureRegion;
	label: string;
	/** The region's outline relative to its own box, computed once with the layout. */
	path: string;
};

/** A loaded target the card's entity runs on. */
export type CardPlacement = {
	entity: GraphEntity;
	label: string;
	/** "3 entities run here", or "At least 3" on a partial graph. */
	countLine: string;
};

export type CardNodeData = NodeState & {
	card: StructureCard;
	label: string;
	/** Every target by display name, then id; the first few as chips and the rest behind "+N". */
	placement: {
		targets: readonly CardPlacement[];
		shown: readonly CardPlacement[];
		collapsed: readonly CardPlacement[];
	};
	/** "Also in Payments" when the entity has structural parents besides the one it is drawn in. */
	alsoInLabel?: string;
};

export type RegionFlowNode = Node<RegionNodeData, "system-map-region">;
export type CardFlowNode = Node<CardNodeData, "system-map-card">;
export type FlowNode = RegionFlowNode | CardFlowNode;

export type RegionNodeProps = NodeProps<RegionFlowNode>;
export type CardNodeProps = NodeProps<CardFlowNode>;

export type FlowEdge = Edge<
	{
		connection: MapConnection;
		/** Computed from the endpoint boxes; the edge component draws between these and ignores handles. */
		anchors: ConnectionAnchors;
		label: string;
		showLabel?: boolean;
		emphasized?: boolean;
		dimmed?: boolean;
	},
	"system-map-connection"
>;

/** The fixed geometry of a graph; interaction state is applied over it. */
export type MapProjection = {
	nodes: FlowNode[];
	tree: StructureTree;
	boxes: ReadonlyMap<string, Box>;
	/** The extent of the whole map. */
	bounds: Box;
};

const placement = (box: Box) => ({
	position: { x: box.x, y: box.y },
	width: box.width,
	height: box.height,
	// Boxes are fixed by the layout; nothing waits for a DOM measurement pass.
	measured: { width: box.width, height: box.height },
});

function regionNode(region: StructureRegion, box: Box): RegionFlowNode {
	let zIndex = LANDSCAPE_Z;
	if (region.level === "system") {
		zIndex = SYSTEM_Z;
	}

	const label = entityLabel(region.entity);
	const path = regionPath({ x: 0, y: 0, width: box.width, height: box.height }, region.id);
	return {
		id: region.id,
		type: "system-map-region",
		...placement(box),
		zIndex,
		focusable: true,
		ariaRole: "button",
		ariaLabel: `${label}, ${region.level}`,
		data: { band: "landscape", region, label, path },
	};
}

function cardNode(
	card: StructureCard,
	box: Box,
	entities: ReadonlyMap<string, GraphEntity>,
	targets: readonly CardPlacement[]
): CardFlowNode {
	const label = entityLabel(card.entity);

	let alsoInLabel: string | undefined;
	if (card.alsoIn.length > 0) {
		const parents = card.alsoIn.map((id) => {
			const parent = entities.get(id);
			return parent ? entityLabel(parent) : id;
		});
		alsoInLabel = `Also in ${parents.join(", ")}`;
	}

	return {
		id: card.id,
		type: "system-map-card",
		...placement(box),
		zIndex: CARD_Z,
		focusable: true,
		ariaRole: "button",
		ariaLabel: `${label}, ${card.entity.kind}`,
		data: {
			band: "landscape",
			card,
			label,
			placement: { targets, ...placementChips(targets) },
			alsoInLabel,
		},
	};
}

/** "N relationships" between systems; the verb, with "× N" when aggregated, between runtimes. */
export function connectionLabel(connection: MapConnection, level: ConnectionLevel): string {
	const count = connection.relationshipIds.length;
	if (level === "system") {
		return `${count} ${count === 1 ? "relationship" : "relationships"}`;
	}
	const predicate = relationshipPredicateLabel(connection.predicates[0]);
	if (count > 1) {
		return `${predicate} × ${count}`;
	}
	return predicate;
}

function connectionEdges(
	connections: readonly MapConnection[],
	boxes: ReadonlyMap<string, Box>,
	level: ConnectionLevel
): FlowEdge[] {
	const directions = new Set(connections.map(({ source, target }) => JSON.stringify([source, target])));

	return connections.map((connection) => {
		const { source, target } = connection;
		const opposed = directions.has(JSON.stringify([target, source]));
		const anchors = connectionAnchors(boxes.get(source)!, boxes.get(target)!, opposed);
		const label = connectionLabel(connection, level);
		return {
			id: connection.id,
			type: "system-map-connection",
			source,
			target,
			zIndex: CONNECTION_Z,
			focusable: true,
			ariaRole: "button",
			ariaLabel: label,
			data: { connection, anchors, label },
		};
	});
}

/** Projects a graph to Svelte Flow nodes at fixed, computed positions, regions before their children. */
export function projectMap(graph: GraphSubset): MapProjection {
	const tree = buildStructureTree(graph);
	const layout = layoutStructure(tree);
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const targets = placementTargets(graph);
	const placements = new Map<string, CardPlacement>();
	const cardPlacements = (id: string) =>
		(targets.get(id) ?? []).map((entity) => {
			let placement = placements.get(entity.id);
			if (!placement) {
				placement = {
					entity,
					label: entityLabel(entity),
					countLine: placementCountLine(graph, entity.id),
				};
				placements.set(entity.id, placement);
			}
			return placement;
		});
	const nodes: FlowNode[] = [];

	const visit = (node: StructureNode) => {
		const box = layout.boxes.get(node.id)!;
		if (node.type === "card") {
			nodes.push(cardNode(node, box, entities, cardPlacements(node.id)));
			return;
		}

		nodes.push(regionNode(node, box));
		node.children.forEach(visit);
	};
	tree.roots.forEach(visit);

	return {
		nodes,
		tree,
		boxes: layout.boxes,
		bounds: { x: 0, y: 0, width: layout.width, height: layout.height },
	};
}

type MapState = {
	graph: GraphSubset;
	band: DetailBand;
	relationshipMode: RelationshipMode;
	selection?: MapSelection;
	highlights?: MapHighlights;
	hoveredEdgeId?: string;
};

function connectionFocus(tree: StructureTree, selection: MapSelection | undefined): ConnectionFocus {
	if (selection?.kind === "entity" && tree.nodes.has(selection.entityId)) {
		return { kind: "node", nodeId: selection.entityId };
	}
	if (selection?.kind === "relationship" || selection?.kind === "summary") {
		return { kind: "relationships", relationshipIds: selectedRelationshipIds(selection) };
	}
	return { kind: "none" };
}

/** Applies the detail band, relationship mode, selection, host highlights and hover over the fixed geometry. */
export function applyMapState(
	projection: MapProjection,
	state: MapState
): { nodes: FlowNode[]; edges: FlowEdge[] } {
	const { graph, band, relationshipMode, selection, highlights, hoveredEdgeId } = state;
	const focus = connectionFocus(projection.tree, selection);
	const view = visibleConnections(graph, projection.tree, { band, mode: relationshipMode, focus });
	const allEdges = connectionEdges(view.connections, projection.boxes, view.level);

	const selectedRelationships = new Set(selectedRelationshipIds(selection));
	const hostRelationships = new Set(highlights?.relationshipIds);
	const hostEntities = new Set(highlights?.entityIds);
	const hoveredEdge = allEdges.find((edge) => edge.id === hoveredEdgeId);

	const edges = allEdges.map((edge) => {
		const data = edge.data!;
		const relationshipIds = data.connection.relationshipIds;
		const selected = relationshipIds.some((id) => selectedRelationships.has(id));
		const highlighted = relationshipIds.some((id) => hostRelationships.has(id));
		const incidentToSelection = relationshipIds.some((id) => view.incidentRelationshipIds.has(id));
		const incidentToHighlight = hostEntities.has(edge.source) || hostEntities.has(edge.target);
		const hovered = edge.id === hoveredEdgeId;
		const emphasized =
			view.focused || selected || highlighted || incidentToSelection || incidentToHighlight || hovered;
		return {
			...edge,
			selected,
			data: {
				...data,
				showLabel: emphasized,
				emphasized,
				dimmed: relationshipMode === "all" && !emphasized && Boolean(selection || hoveredEdge),
			},
		};
	});

	const highlightedIds = new Set(hostEntities);
	const endpointEdges = edges.filter((edge) => edge.selected || edge.id === hoveredEdge?.id);
	for (const edge of endpointEdges) {
		highlightedIds.add(edge.source);
		highlightedIds.add(edge.target);
	}

	let selectedNodeId: string | undefined;
	if (focus.kind === "node") {
		selectedNodeId = focus.nodeId;
		// The persistent selection treatment supersedes the endpoint highlight.
		highlightedIds.delete(selectedNodeId);
	}

	let footprint: ReadonlySet<string> = new Set();
	if (selection?.kind === "entity") {
		footprint = placementFootprint(graph, projection.tree, selection.entityId, band);
	}

	const nodes = projection.nodes.map((node) => {
		const selected = node.id === selectedNodeId;
		// Cards appear at the runtime band; a selected card stays visible in compact form.
		const hidden = node.type === "system-map-card" && band === "landscape" && !selected;
		return {
			...node,
			selected,
			hidden,
			data: {
				...node.data,
				band,
				highlighted: highlightedIds.has(node.id),
				footprint: footprint.has(node.id),
			},
		};
	}) as FlowNode[];

	return { nodes, edges };
}
