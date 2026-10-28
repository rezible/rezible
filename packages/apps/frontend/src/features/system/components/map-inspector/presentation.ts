import type { GraphEntity } from "$features/system/lib/system-map/graph";
import { entityLabel, relationshipPredicateLabel } from "$features/system/lib/system-map/labels";
import type { MapSelection } from "$features/system/lib/system-map/presentation";
import { selectedRelationshipIds } from "$features/system/lib/system-map/selection";
import type {
	InspectedRelationship,
	InspectedStructure,
	MapInspection,
	MapInspectorItem,
} from "./inspection";

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
	/** "2 systems", "4 runtime entities", "8 implementation entities grouped". */
	structureFacts: readonly string[];
	/** Where a runtime entity runs. */
	runsOn: readonly InspectorEndpointView[];
	/** A container with no loaded placement reads "Placement unknown" under Runs on. */
	placementUnknown: boolean;
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

const plural = (count: number, singular: string, pluralForm: string) =>
	`${count} ${count === 1 ? singular : pluralForm}`;

function structureFacts(structure: InspectedStructure | undefined): string[] {
	if (!structure) {
		return [];
	}
	if (structure.level === "landscape") {
		return [
			plural(structure.systemCount, "system", "systems"),
			plural(structure.runtimeCount, "runtime entity", "runtime entities"),
		];
	}
	if (structure.level === "system") {
		return [plural(structure.runtimeCount, "runtime entity", "runtime entities")];
	}
	if (structure.implementationCount > 0) {
		const count = plural(
			structure.implementationCount,
			"implementation entity",
			"implementation entities"
		);
		return [`${count} grouped`];
	}
	return [];
}

function runsOnViews(structure: InspectedStructure | undefined): InspectorEndpointView[] {
	if (structure?.level !== "runtime") {
		return [];
	}
	return structure.runsOn.map((target) => endpointView(target.id, target));
}

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
		structureFacts: structureFacts(inspection.structure),
		runsOn: runsOnViews(inspection.structure),
		placementUnknown: inspection.structure?.level === "runtime" && inspection.structure.placementUnknown,
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
