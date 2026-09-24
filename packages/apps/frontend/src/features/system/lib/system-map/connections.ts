import { isArchitectureCategory } from "./category";
import type { GraphSubset } from "./graph";

export type MapConnection = {
	id: string;
	source: string;
	target: string;
	predicate: string;
	relationshipIds: string[];
};

/** Aggregate parallel records without changing their canonical endpoints. */
export function getMapConnections(graph: GraphSubset): MapConnection[] {
	const architectureIds = new Set(
		graph.entities.filter((entity) => isArchitectureCategory(entity.category)).map((entity) => entity.id)
	);
	const connections = new Map<string, MapConnection>();

	for (const relationship of graph.relationships) {
		const { source, target } = relationship;
		if (!architectureIds.has(source) || !architectureIds.has(target)) {
			continue;
		}

		// Direction and predicate are part of identity; each canonical record contributes once.
		const id = JSON.stringify([source, target, relationship.predicate]);
		let connection = connections.get(id);
		if (!connection) {
			connection = { id, source, target, predicate: relationship.predicate, relationshipIds: [] };
			connections.set(id, connection);
		}
		connection.relationshipIds.push(relationship.id);
	}

	return [...connections.values()];
}
