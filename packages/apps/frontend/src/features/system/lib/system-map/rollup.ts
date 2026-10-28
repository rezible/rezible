import type { DetailBand } from "./detail";
import type { GraphRelationship, GraphSubset } from "./graph";
import { PLACEMENT_PREDICATE } from "./placement";
import { structureLevel, type StructureNode, type StructureTree } from "./structure";

/** The level connections attach to: endpoints below it roll up to their ancestor at it. */
export type ConnectionLevel = "system" | "runtime";

export type MapConnection = {
	id: string;
	source: string;
	target: string;
	predicates: string[];
	relationshipIds: string[];
	/** A placement drawn for its selected source; placements never roll up. */
	placement: boolean;
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
	/** Behaviour relationships with an endpoint in the focused node's subtree. */
	incidentRelationshipIds: ReadonlySet<string>;
};

/** The band's visible level: systems at the landscape band, runtimes at the runtime band. */
const bandLevel = (band: DetailBand): ConnectionLevel => (band === "landscape" ? "system" : "runtime");

type Resolve = (entityId: string, level: ConnectionLevel) => string | undefined;

/** Resolves an entity to its drawn node at a level, walking primary ancestry. */
function endpointResolver(graph: GraphSubset, tree: StructureTree): Resolve {
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const systemLevelIds = new Map<string, string>();
	const visit = (node: StructureNode, systemId?: string) => {
		if (node.type === "card") {
			// A runtime with no system ancestor has no node at the system level.
			if (systemId) {
				systemLevelIds.set(node.id, systemId);
			}
			return;
		}

		systemLevelIds.set(node.id, node.id);
		const container = node.level === "system" ? node.id : systemId;
		node.children.forEach((child) => visit(child, container));
	};
	tree.roots.forEach((node) => visit(node));

	return (entityId, level) => {
		let entity = entities.get(entityId);
		if (!entity || !structureLevel(entity.category)) {
			return undefined;
		}
		// Implementation entities must reach a runtime, even when contained directly by a region.
		while (structureLevel(entity.category) === "implementation") {
			const parentId = tree.primaryParents.get(entity.id);
			entity = parentId ? entities.get(parentId) : undefined;
			if (!entity) {
				return undefined;
			}
			const parentLevel = structureLevel(entity.category);
			if (parentLevel !== "implementation" && parentLevel !== "runtime") {
				return undefined;
			}
		}

		if (!tree.nodes.has(entity.id)) {
			return undefined;
		}
		if (level === "runtime") {
			return entity.id;
		}
		return systemLevelIds.get(entity.id);
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
				placement: relationship.predicate === PLACEMENT_PREDICATE,
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

/**
 * The connections drawn for a band, relationship mode and focus. Placements are never rolled up or drawn in
 * All; Focused draws them only for a selected source at the runtime band.
 */
export function visibleConnections(
	graph: GraphSubset,
	tree: StructureTree,
	{ band, mode, focus }: VisibleConnectionOptions
): VisibleConnections {
	const resolve = endpointResolver(graph, tree);
	const level = bandLevel(band);
	const behaviour = graph.relationships.filter(
		({ predicate }) => predicate !== "contains" && predicate !== PLACEMENT_PREDICATE
	);

	let incident: GraphRelationship[] = [];
	if (focus.kind === "node") {
		const subtree = subtreeIds(tree, focus.nodeId);
		const inSubtree = (entityId: string) => {
			const nodeId = resolve(entityId, "runtime");
			return nodeId !== undefined && subtree.has(nodeId);
		};
		incident = behaviour.filter(({ source, target }) => inSubtree(source) || inSubtree(target));
	}
	const incidentRelationshipIds = new Set(incident.map((relationship) => relationship.id));

	if (mode === "all") {
		return {
			connections: groupConnections(behaviour, resolve, level),
			level,
			focused: false,
			incidentRelationshipIds,
		};
	}

	if (focus.kind === "node") {
		let relationships = incident;
		if (level === "runtime") {
			const placements = graph.relationships.filter(
				(relationship) =>
					relationship.predicate === PLACEMENT_PREDICATE &&
					resolve(relationship.source, "runtime") === focus.nodeId
			);
			relationships = [...incident, ...placements];
		}
		return {
			connections: groupConnections(relationships, resolve, level),
			level,
			focused: true,
			incidentRelationshipIds,
		};
	}

	if (focus.kind === "relationships") {
		const ids = new Set(focus.relationshipIds);
		const relationships = graph.relationships.filter((relationship) => {
			if (!ids.has(relationship.id) || relationship.predicate === "contains") {
				return false;
			}
			return level === "runtime" || relationship.predicate !== PLACEMENT_PREDICATE;
		});
		return {
			connections: groupConnections(relationships, resolve, level),
			level,
			focused: true,
			incidentRelationshipIds,
		};
	}

	// At rest the map shows how systems relate, at either band.
	return {
		connections: groupConnections(behaviour, resolve, "system"),
		level: "system",
		focused: false,
		incidentRelationshipIds,
	};
}
