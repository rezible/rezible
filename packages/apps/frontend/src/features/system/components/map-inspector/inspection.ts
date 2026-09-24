import type { GraphEntity, GraphRelationship, GraphSubset } from "$features/system/lib/system-map/graph";
import type { MapSelection } from "$features/system/lib/system-map/presentation";
import {
	entityCategoryLabel,
	entityLabel,
	relationshipPredicateLabel,
} from "$features/system/lib/system-map/labels";
import { selectionIsAvailable } from "$features/system/lib/system-map/selection";

export type InspectedRelationship = {
	relationship: GraphRelationship;
	source?: GraphEntity;
	target?: GraphEntity;
};

export type MapInspection = {
	selection?: MapSelection;
	available: boolean;
	entity?: GraphEntity;
	relationship?: InspectedRelationship;
	summary?: {
		relationships: readonly InspectedRelationship[];
		missingRelationshipIds: readonly string[];
		distinctContributorCount: number;
	};
	/** Supplied relationships incident to the selected entity, including unresolved records. */
	incidentRelationships: readonly InspectedRelationship[];
};

export type MapInspectorItem = {
	id: string;
	label: string;
	detail: string;
	target: MapSelection;
};

const allRelationships = (graph: GraphSubset): readonly GraphRelationship[] => [
	...graph.relationships,
	...graph.unresolvedRelationships,
];

const inspectedRelationship = (
	relationship: GraphRelationship,
	entitiesById: ReadonlyMap<string, GraphEntity>
): InspectedRelationship => ({
	relationship,
	source: entitiesById.get(relationship.source),
	target: entitiesById.get(relationship.target),
});

const endpointInventoryLabel = (entityId: string, entity?: GraphEntity): string =>
	entity ? `${entityLabel(entity)} · ${entity.id}` : `${entityId} (unavailable in this subset)`;

/** Resolves canonical supplied records only. Missing summary contributors remain explicit. */
export const resolveInspection = (graph: GraphSubset, selection: MapSelection | undefined): MapInspection => {
	const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const relationshipsById = new Map(
		allRelationships(graph).map((relationship) => [relationship.id, relationship])
	);
	const available = selectionIsAvailable(graph, selection);

	if (!selection) return { available, incidentRelationships: [] };

	switch (selection.kind) {
		case "entity": {
			const entity = entitiesById.get(selection.entityId);
			const incidentRelationships = entity
				? allRelationships(graph)
						.filter(
							(relationship) =>
								relationship.source === entity.id || relationship.target === entity.id
						)
						.map((relationship) => inspectedRelationship(relationship, entitiesById))
				: [];

			return { selection, available, entity, incidentRelationships };
		}
		case "relationship": {
			const relationship = relationshipsById.get(selection.relationshipId);
			return {
				selection,
				available,
				relationship: relationship ? inspectedRelationship(relationship, entitiesById) : undefined,
				incidentRelationships: [],
			};
		}
		case "summary": {
			const relationshipIds = [...new Set(selection.relationshipIds)];
			const relationships: InspectedRelationship[] = [];
			const missingRelationshipIds: string[] = [];

			for (const relationshipId of relationshipIds) {
				const relationship = relationshipsById.get(relationshipId);
				if (relationship) {
					relationships.push(inspectedRelationship(relationship, entitiesById));
				} else {
					missingRelationshipIds.push(relationshipId);
				}
			}

			return {
				selection,
				available,
				summary: {
					relationships,
					missingRelationshipIds,
					distinctContributorCount: relationshipIds.length,
				},
				incidentRelationships: [],
			};
		}
	}
};

/** Lists every supplied canonical record without consulting map projection or layout. */
export const buildInspectionItems = (graph: GraphSubset): MapInspectorItem[] => {
	const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const items = graph.entities.map<MapInspectorItem>((entity) => ({
		id: `entity:${entity.id}`,
		label: entityLabel(entity),
		detail: `${entityCategoryLabel(entity.category)} · ${entity.id}`,
		target: { kind: "entity", entityId: entity.id },
	}));

	for (const relationship of allRelationships(graph)) {
		items.push({
			id: `relationship:${relationship.id}`,
			label: relationshipPredicateLabel(relationship.predicate),
			detail: `${endpointInventoryLabel(relationship.source, entitiesById.get(relationship.source))} → ${endpointInventoryLabel(relationship.target, entitiesById.get(relationship.target))} · ${relationship.id}`,
			target: { kind: "relationship", relationshipId: relationship.id },
		});
	}

	return items;
};
