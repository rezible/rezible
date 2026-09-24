import type { GraphRelationship, GraphSubset } from "./graph";
import type { MapSelection } from "./presentation";

const relationshipsById = (graph: GraphSubset): ReadonlyMap<string, GraphRelationship> =>
	new Map(
		[...graph.relationships, ...graph.unresolvedRelationships].map((relationship) => [
			relationship.id,
			relationship,
		])
	);

/** Whether a canonical host selection still refers to at least one supplied record. */
export const selectionIsAvailable = (graph: GraphSubset, selection: MapSelection | undefined): boolean => {
	if (!selection) return false;
	if (selection.kind === "entity") return graph.entities.some((entity) => entity.id === selection.entityId);

	const relationships = relationshipsById(graph);
	if (selection.kind === "relationship") return relationships.has(selection.relationshipId);
	return selection.relationshipIds.some((relationshipId) => relationships.has(relationshipId));
};

/** Host reconciliation keeps a summary's original contributor IDs while any contributor remains supplied. */
export const reconcileSelection = (
	graph: GraphSubset,
	selection: MapSelection | undefined
): MapSelection | undefined => (selectionIsAvailable(graph, selection) ? selection : undefined);

/** Returns relationship IDs represented by a relationship or summary selection. */
export const selectedRelationshipIds = (selection: MapSelection | undefined): readonly string[] => {
	if (!selection) return [];
	if (selection.kind === "relationship") return [selection.relationshipId];
	if (selection.kind === "summary") return selection.relationshipIds;
	return [];
};

/** Returns only supplied endpoints for a selected relationship or summary. */
export const selectedEntityIds = (
	graph: GraphSubset,
	selection: MapSelection | undefined
): readonly string[] => {
	if (!selection) return [];
	if (selection.kind === "entity") {
		return graph.entities.some((entity) => entity.id === selection.entityId) ? [selection.entityId] : [];
	}

	const entityIds = new Set(graph.entities.map((entity) => entity.id));
	const relationshipIds = new Set(selectedRelationshipIds(selection));
	const selected = new Set<string>();
	for (const relationship of [...graph.relationships, ...graph.unresolvedRelationships]) {
		if (!relationshipIds.has(relationship.id)) continue;
		if (entityIds.has(relationship.source)) selected.add(relationship.source);
		if (entityIds.has(relationship.target)) selected.add(relationship.target);
	}
	return [...selected];
};
