import { DisplayMode, getMapCategoryDisplay } from "./category";
import type { GraphEntity, GraphRelationship, GraphSubset } from "./graph";
import type { InspectionTarget, MapAnnotation, MapProjection } from "./presentation";

export type InspectedRelationship = {
	relationship: GraphRelationship;
	source?: GraphEntity;
	target?: GraphEntity;
};

export type InspectionVisibility = "visible" | "represented" | "not-rendered" | "unavailable";

export type SystemMapInspection = {
	target: InspectionTarget;
	visibility: InspectionVisibility;
	representativeId?: string;
	entity?: GraphEntity;
	relationship?: InspectedRelationship;
	summary?: {
		relationships: readonly InspectedRelationship[];
		missingRelationshipIds: readonly string[];
	};
	annotation?: {
		annotation?: MapAnnotation;
		entity?: GraphEntity;
		relationship?: InspectedRelationship;
		originalTarget?: GraphEntity;
	};
	/** All supplied facts touching an entity, including membership and self-relations. */
	relationships: readonly InspectedRelationship[];
	/** The supplied containment facts involving an entity. */
	memberships: readonly InspectedRelationship[];
};

export type SystemMapInspectionItem = {
	id: string;
	label: string;
	detail: string;
	target: InspectionTarget;
};

/** Checks that a selection still belongs to the currently supplied source snapshot. */
export const inspectionTargetAvailable = (graph: GraphSubset, target: InspectionTarget): boolean => {
	const entityIds = new Set(graph.entities.map((entity) => entity.id));
	const relationshipsById = new Map(
		graph.relationships.map((relationship) => [relationship.id, relationship])
	);

	switch (target.kind) {
		case "entity":
			return entityIds.has(target.entityId);
		case "relationship":
			return relationshipsById.has(target.relationshipId);
		case "summary":
			return target.sourceRelationshipIds.some((relationshipId) =>
				relationshipsById.has(relationshipId)
			);
		case "annotation": {
			const relationship = relationshipsById.get(target.relationshipId);
			return entityIds.has(target.entityId) && relationship?.source === target.entityId;
		}
	}
};

const inspectedRelationship = (
	relationship: GraphRelationship,
	entitiesById: ReadonlyMap<string, GraphEntity>
): InspectedRelationship => ({
	relationship,
	source: entitiesById.get(relationship.source),
	target: entitiesById.get(relationship.target),
});

const visibilityFor = (
	entityId: string,
	projection: MapProjection,
	visibleIds: ReadonlySet<string>
): Pick<SystemMapInspection, "visibility" | "representativeId"> => {
	const representativeId = projection.representativeByEntityId.get(entityId);
	if (visibleIds.has(entityId)) return { visibility: "visible", representativeId };
	if (representativeId) return { visibility: "represented", representativeId };
	return { visibility: "not-rendered" };
};

const entityInspection = (
	target: Extract<InspectionTarget, { kind: "entity" }>,
	graph: GraphSubset,
	projection: MapProjection,
	entitiesById: ReadonlyMap<string, GraphEntity>,
	visibleIds: ReadonlySet<string>
): SystemMapInspection => {
	const relationships = graph.relationships
		.filter(
			(relationship) =>
				relationship.source === target.entityId || relationship.target === target.entityId
		)
		.map((relationship) => inspectedRelationship(relationship, entitiesById));
	const memberships = relationships.filter(({ relationship }) => relationship.predicate === "contains");
	const entity = entitiesById.get(target.entityId);

	return {
		target,
		...visibilityFor(target.entityId, projection, visibleIds),
		entity,
		relationships,
		memberships,
	};
};

const relationshipInspection = (
	target: Extract<InspectionTarget, { kind: "relationship" }>,
	projection: MapProjection,
	entitiesById: ReadonlyMap<string, GraphEntity>,
	relationshipsById: ReadonlyMap<string, GraphRelationship>,
	visibleIds: ReadonlySet<string>
): SystemMapInspection => {
	const relationship = relationshipsById.get(target.relationshipId);
	const inspected = relationship ? inspectedRelationship(relationship, entitiesById) : undefined;
	const visibility = relationship
		? visibilityFor(relationship.source, projection, visibleIds)
		: { visibility: "unavailable" as const };

	return {
		target,
		...visibility,
		relationship: inspected,
		relationships: inspected ? [inspected] : [],
		memberships: inspected?.relationship.predicate === "contains" && inspected ? [inspected] : [],
	};
};

const summaryInspection = (
	target: Extract<InspectionTarget, { kind: "summary" }>,
	projection: MapProjection,
	entitiesById: ReadonlyMap<string, GraphEntity>,
	relationshipsById: ReadonlyMap<string, GraphRelationship>,
	visibleIds: ReadonlySet<string>
): SystemMapInspection => {
	const relationships: InspectedRelationship[] = [];
	const missingRelationshipIds: string[] = [];

	for (const relationshipId of target.sourceRelationshipIds) {
		const relationship = relationshipsById.get(relationshipId);
		if (!relationship) {
			missingRelationshipIds.push(relationshipId);
			continue;
		}
		relationships.push(inspectedRelationship(relationship, entitiesById));
	}

	const first = relationships[0]?.relationship;
	const representativeId = first
		? (projection.representativeByEntityId.get(first.source) ??
			projection.representativeByEntityId.get(first.target))
		: undefined;

	return {
		target,
		visibility: representativeId
			? visibleIds.has(representativeId)
				? "represented"
				: "not-rendered"
			: "unavailable",
		representativeId,
		summary: { relationships, missingRelationshipIds },
		relationships,
		memberships: relationships.filter(({ relationship }) => relationship.predicate === "contains"),
	};
};

const annotationInspection = (
	target: Extract<InspectionTarget, { kind: "annotation" }>,
	projection: MapProjection,
	entitiesById: ReadonlyMap<string, GraphEntity>,
	relationshipsById: ReadonlyMap<string, GraphRelationship>,
	visibleIds: ReadonlySet<string>
): SystemMapInspection => {
	const relationship = relationshipsById.get(target.relationshipId);
	const annotation = projection.annotations.find(
		(candidate) =>
			candidate.entityId === target.entityId && candidate.relationshipId === target.relationshipId
	);
	const inspected = relationship ? inspectedRelationship(relationship, entitiesById) : undefined;
	const originalTarget = relationship ? entitiesById.get(relationship.target) : undefined;
	const visibility = originalTarget
		? visibilityFor(originalTarget.id, projection, visibleIds)
		: { visibility: "unavailable" as const };

	return {
		target,
		...visibility,
		annotation: {
			annotation,
			entity: entitiesById.get(target.entityId),
			relationship: inspected,
			originalTarget,
		},
		relationships: inspected ? [inspected] : [],
		memberships: [],
	};
};

/** Resolves source records at inspection time; it never captures projected records as identity. */
export const resolveInspectionTarget = (
	graph: GraphSubset,
	projection: MapProjection,
	target: InspectionTarget
): SystemMapInspection => {
	const entitiesById = new Map(graph.entities.map(ent => [ent.id, ent]));
	const relationshipsById = new Map(graph.relationships.map(rel => [rel.id, rel]));
	const visibleIds = new Set(projection.nodes.map((node) => node.id));

	switch (target.kind) {
		case "entity":
			return entityInspection(target, graph, projection, entitiesById, visibleIds);
		case "relationship":
			return relationshipInspection(target, projection, entitiesById, relationshipsById, visibleIds);
		case "summary":
			return summaryInspection(target, projection, entitiesById, relationshipsById, visibleIds);
		case "annotation":
			return annotationInspection(target, projection, entitiesById, relationshipsById, visibleIds);
	}
};

export const inspectionItemId = (target: InspectionTarget | undefined): string | undefined => {
	if (!target) return undefined;
	if (target.kind === "entity") return `entity:${target.entityId}`;
	if (target.kind === "relationship") return `relationship:${target.relationshipId}`;
	if (target.kind === "annotation") return `annotation:${target.entityId}:${target.relationshipId}`;
	return undefined;
};

const mapGraphEntityToSystemMapInspectionItem = ({id, label, category, kind}: GraphEntity): SystemMapInspectionItem => ({
	id: `entity:${id}`,
	label: label,
	detail: `${getMapCategoryDisplay(category).categoryLabel} · ${kind}`,
	target: { kind: "entity", entityId: id },
});

const makeAnnotationId = (entityId: string, relationshipId: string) => `${entityId}:${relationshipId}`;

/** Builds the keyboard list from the supplied subset, including non-rendered subjects and facts. */
export const buildInspectionItems = (graph: GraphSubset, {annotations}: MapProjection): SystemMapInspectionItem[] => {
	const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const items = graph.entities.map(mapGraphEntityToSystemMapInspectionItem);
	const projectedAnnotationIds = new Set(annotations.map(a => makeAnnotationId(a.entityId, a.relationshipId)));

	for (const relationship of graph.relationships) {
		const source = entitiesById.get(relationship.source);
		const target = entitiesById.get(relationship.target);
		items.push({
			id: `relationship:${relationship.id}`,
			label: relationship.predicate.replaceAll("_", " "),
			detail: `${source?.label || relationship.source} → ${target?.label || relationship.target}`,
			target: { kind: "relationship", relationshipId: relationship.id },
		});
	}

	for (const {id: relId, source: sourceId, target: targetId} of graph.relationships) {
		const source = entitiesById.get(sourceId);
		if (getMapCategoryDisplay(source?.category ?? "").mode !== DisplayMode.Annotation) {
			continue;
		}
		const target = entitiesById.get(targetId);
		const annotationId = makeAnnotationId(sourceId, relId);
		const isProjected = projectedAnnotationIds.has(annotationId);
		items.push({
			id: `annotation:${annotationId}`,
			label: source?.label || sourceId,
			detail: `${isProjected ? "annotation" : "hidden annotation"} → ${target?.label || targetId}`,
			target: {
				kind: "annotation",
				entityId: sourceId,
				relationshipId: relId,
			},
		});
	}

	return items;
};
