import type { DetailBand } from "./detail";
import type { GraphRelationship, GraphSubset } from "./graph";
import { rollsUp } from "./model";
import type { StructureNode, StructureTree } from "./structure";

/** The level connections attach to: endpoints below it roll up to their ancestor at it. */
export type ConnectionLevel = "system" | "runtime";

export type MapConnection = {
	id: string;
	source: string;
	target: string;
	predicates: string[];
	relationshipIds: string[];
};

export type RelationshipMode = "focused" | "all";

/** What the map is focused on: a drawn node, a set of relationships, or nothing. */
export type ConnectionFocus =
	| { readonly kind: "none" }
	| { readonly kind: "node"; readonly nodeId: string }
	| { readonly kind: "relationships"; readonly relationshipIds: readonly string[] };

export type VisibleConnections = {
	connections: MapConnection[];
	level: ConnectionLevel;
	/** Every connection was drawn because it answers the focus. */
	focused: boolean;
	/** Drawable relationships with an endpoint in the focused node's subtree. */
	incidentRelationshipIds: ReadonlySet<string>;
};

/** The band's visible level: systems at the landscape band, runtimes at the runtime band. */
const bandLevel = (band: DetailBand): ConnectionLevel => (band === "landscape" ? "system" : "runtime");

type Resolve = (entityId: string, level: ConnectionLevel) => string | undefined;

/**
 * Resolves an entity along its parent chain to its drawn node at a level: the nearest container card at the
 * runtime level, and the nearest region at the system level. Entities outside the logical tree, such as
 * infrastructure, and entities with no node at the level resolve to nothing.
 */
function endpointResolver(tree: StructureTree): Resolve {
	return (entityId, level) => {
		const nodeId = tree.drawnNodes.get(entityId);
		if (!nodeId) {
			return undefined;
		}
		const isRegion = tree.nodes.get(nodeId)?.type === "region";
		if (level === "runtime") {
			return isRegion ? undefined : nodeId;
		}
		return isRegion ? nodeId : tree.enclosingRegions.get(nodeId);
	};
}

/**
 * Groups relationships between their drawn nodes at a level, dropping those with an endpoint that has no node
 * there or whose endpoints meet. The system level groups by node pair; the runtime level also by predicate.
 */
function groupConnections(
	relationships: readonly GraphRelationship[],
	resolve: Resolve,
	level: ConnectionLevel
): MapConnection[] {
	const connections = new Map<string, MapConnection>();
	for (const relationship of relationships) {
		const source = resolve(relationship.source, level);
		const target = resolve(relationship.target, level);
		if (!source || !target || source === target) {
			continue;
		}

		let key = [source, target];
		if (level === "runtime") {
			key = [...key, relationship.predicate];
		}
		const id = JSON.stringify(key);
		let connection = connections.get(id);
		if (!connection) {
			connection = {
				id,
				source,
				target,
				predicates: [],
				relationshipIds: [],
			};
			connections.set(id, connection);
		}
		connection.relationshipIds.push(relationship.id);
		if (!connection.predicates.includes(relationship.predicate)) {
			connection.predicates.push(relationship.predicate);
		}
	}
	return [...connections.values()];
}

function subtreeIds(tree: StructureTree, nodeId: string): ReadonlySet<string> {
	const ids = new Set<string>();
	const visit = (node: StructureNode) => {
		ids.add(node.id);
		if (node.type === "region") {
			node.children.forEach(visit);
		}
	};
	const node = tree.nodes.get(nodeId);
	if (node) {
		visit(node);
	}
	return ids;
}

type VisibleConnectionOptions = {
	band: DetailBand;
	mode: RelationshipMode;
	focus: ConnectionFocus;
};

/** The relationships the structure view can draw: those that roll up, with both endpoints in the tree. */
export function drawableRelationships(graph: GraphSubset, tree: StructureTree): GraphRelationship[] {
	return graph.relationships.filter(
		({ predicate, source, target }) =>
			rollsUp(graph.model, predicate) && tree.drawnNodes.has(source) && tree.drawnNodes.has(target)
	);
}

/**
 * The connections drawn for a band, relationship mode and focus, always at the band's level. Only predicates
 * that roll up are drawn.
 */
export function visibleConnections(
	graph: GraphSubset,
	tree: StructureTree,
	{ band, mode, focus }: VisibleConnectionOptions
): VisibleConnections {
	const resolve = endpointResolver(tree);
	const level = bandLevel(band);
	const drawable = drawableRelationships(graph, tree);

	let incident: GraphRelationship[] = [];
	if (focus.kind === "node") {
		const subtree = subtreeIds(tree, focus.nodeId);
		const inSubtree = (entityId: string) => {
			const nodeId = tree.drawnNodes.get(entityId);
			return nodeId !== undefined && subtree.has(nodeId);
		};
		incident = drawable.filter(({ source, target }) => inSubtree(source) || inSubtree(target));
	}
	const incidentRelationshipIds = new Set(incident.map((relationship) => relationship.id));

	if (mode === "all" || focus.kind === "none") {
		return {
			connections: groupConnections(drawable, resolve, level),
			level,
			focused: false,
			incidentRelationshipIds,
		};
	}

	if (focus.kind === "node") {
		return {
			connections: groupConnections(incident, resolve, level),
			level,
			focused: true,
			incidentRelationshipIds,
		};
	}

	const ids = new Set(focus.relationshipIds);
	return {
		connections: groupConnections(
			drawable.filter((relationship) => ids.has(relationship.id)),
			resolve,
			level
		),
		level,
		focused: true,
		incidentRelationshipIds,
	};
}
