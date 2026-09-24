import type { GraphEntity } from "$features/system/lib/system-map/graph";
import { entityLabel, relationshipPredicateLabel } from "$features/system/lib/system-map/labels";
import type { MapSelection } from "$features/system/lib/system-map/presentation";
import { selectedRelationshipIds } from "$features/system/lib/system-map/selection";
import type { InspectedRelationship, MapInspection, MapInspectorItem } from "./inspection";

export type InspectorEndpointView = {
	id: string;
	entity?: GraphEntity;
	text: string;
};

export type InspectorRelationshipView = {
	id: string;
	predicate: string;
	source: InspectorEndpointView;
	target: InspectorEndpointView;
	text: string;
};

export type InspectorViewData = {
	title: string;
	entity?: GraphEntity;
	relationship?: InspectorRelationshipView;
	summary?: {
		relationships: readonly InspectorRelationshipView[];
		missingRelationshipIds: readonly string[];
		distinctContributorCount: number;
	};
	incidentRelationships: readonly InspectorRelationshipView[];
	selectionAvailable: boolean;
};

const endpointView = (id: string, entity?: GraphEntity): InspectorEndpointView => ({
	id,
	entity,
	text: entity ? `${entityLabel(entity)} · ${entity.id}` : `${id} (unavailable in this subset)`,
});

const relationshipView = ({
	relationship,
	source,
	target,
}: InspectedRelationship): InspectorRelationshipView => {
	const predicate = relationshipPredicateLabel(relationship.predicate);
	const sourceView = endpointView(relationship.source, source);
	const targetView = endpointView(relationship.target, target);
	return {
		id: relationship.id,
		predicate,
		source: sourceView,
		target: targetView,
		text: `${sourceView.text} → ${predicate} → ${targetView.text}`,
	};
};

export const buildInspectionViewData = (inspection: MapInspection): InspectorViewData | undefined => {
	if (!inspection.selection) return undefined;

	const relationship = inspection.relationship ? relationshipView(inspection.relationship) : undefined;
	const summary = inspection.summary
		? {
				relationships: inspection.summary.relationships.map(relationshipView),
				missingRelationshipIds: inspection.summary.missingRelationshipIds,
				distinctContributorCount: inspection.summary.distinctContributorCount,
			}
		: undefined;
	const entity = inspection.entity;

	return {
		title: entity
			? entityLabel(entity)
			: (relationship?.predicate ?? (summary ? "Connection summary" : "Unavailable selection")),
		entity,
		relationship,
		summary,
		incidentRelationships: inspection.incidentRelationships.map(relationshipView),
		selectionAvailable: inspection.available,
	};
};

export const isInspectorItemSelected = (
	item: MapInspectorItem,
	selection: MapSelection | undefined
): boolean => {
	if (!selection) return false;
	if (item.target.kind === "entity") {
		return selection.kind === "entity" && item.target.entityId === selection.entityId;
	}
	if (item.target.kind !== "relationship") return false;
	if (selection.kind === "relationship") {
		return item.target.relationshipId === selection.relationshipId;
	}
	if (selection.kind === "summary") {
		return selectedRelationshipIds(selection).includes(item.target.relationshipId);
	}
	return false;
};

export const loadedCountsText = (entityCount: number, relationshipCount: number): string =>
	`${entityCount} loaded ${entityCount === 1 ? "entity" : "entities"} · ${relationshipCount} loaded ${relationshipCount === 1 ? "relationship" : "relationships"}`;
