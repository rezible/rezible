import { DisplayMode, getMapCategoryDisplay, isActorCategory, isArchitectureCategory } from "./category";
import type { GraphEntity, GraphRelationship, GraphSubset } from "./graph";
import type {
	MapAnnotation,
	MapConnection,
	MapDisplayOptions,
	MapNode,
	MapProjection,
	MapRevealState,
} from "./presentation";

type GraphIndex = {
	entitiesById: ReadonlyMap<string, GraphEntity>;
	memberships: readonly GraphRelationship[];
	parentsByChild: ReadonlyMap<string, readonly GraphRelationship[]>;
	childrenByParent: ReadonlyMap<string, readonly GraphRelationship[]>;
};

type SummaryConnection = {
	endpoints: [string, string];
	predicate: string;
	sourceRelationshipIds: string[];
};

const architectureLevel = (entity: GraphEntity): number | undefined =>
	getMapCategoryDisplay(entity.category).level;

const uniqueRelationships = (relationships: readonly GraphRelationship[]): GraphRelationship[] => {
	const seen = new Set<string>();
	const unique: GraphRelationship[] = [];

	for (const relationship of relationships) {
		if (seen.has(relationship.id)) continue;

		seen.add(relationship.id);
		unique.push(relationship);
	}

	return unique;
};

const indexGraph = (graph: GraphSubset): GraphIndex => {
	const entitiesById = new Map<string, GraphEntity>();
	for (const entity of graph.entities) {
		entitiesById.set(entity.id, entity);
	}

	const memberships: GraphRelationship[] = [];
	const parentsByChild = new Map<string, GraphRelationship[]>();
	const childrenByParent = new Map<string, GraphRelationship[]>();

	for (const relationship of uniqueRelationships(graph.relationships)) {
		if (relationship.predicate !== "contains") continue;
		if (!entitiesById.has(relationship.source) || !entitiesById.has(relationship.target)) continue;

		memberships.push(relationship);
		const parents = parentsByChild.get(relationship.target) ?? [];
		parents.push(relationship);
		parentsByChild.set(relationship.target, parents);
		const children = childrenByParent.get(relationship.source) ?? [];
		children.push(relationship);
		childrenByParent.set(relationship.source, children);
	}

	return { entitiesById, memberships, parentsByChild, childrenByParent };
};

const clampDetail = (detail: number): number => Math.max(0, Math.min(3, detail));

const hasMembershipPath = (
	startId: string,
	targetId: string,
	index: GraphIndex,
	excludedRelationshipId?: string
): boolean => {
	const visited = new Set<string>();
	const pending = [startId];

	while (pending.length > 0) {
		const currentId = pending.pop()!;
		if (currentId === targetId) return true;
		if (visited.has(currentId)) continue;

		visited.add(currentId);
		for (const membership of index.childrenByParent.get(currentId) ?? []) {
			if (membership.id !== excludedRelationshipId) pending.push(membership.target);
		}
	}

	return false;
};

const findRelevantVisibleArchitectureAncestors = (
	entityId: string,
	visibleIds: ReadonlySet<string>,
	index: GraphIndex
): Set<string> => {
	const visited = new Set<string>([entityId]);
	const pending = [entityId];
	const candidates = new Set<string>();

	while (pending.length > 0) {
		const currentId = pending.pop()!;

		for (const membership of index.parentsByChild.get(currentId) ?? []) {
			const parentId = membership.source;
			const parent = index.entitiesById.get(parentId);
			if (parent && isArchitectureCategory(parent.category) && visibleIds.has(parentId)) {
				// This path has found its first visible architecture ancestor. Other
				// paths may still need to continue through hidden parents.
				candidates.add(parentId);
				continue;
			}

			if (visited.has(parentId)) continue;
			visited.add(parentId);
			pending.push(parentId);
		}
	}

	// A visible candidate that contains another candidate is context, not a
	// competing sibling group, even when the candidates occur at different
	// membership-path depths.
	const relevant = new Set(candidates);
	for (const candidateId of candidates) {
		for (const otherId of candidates) {
			if (candidateId === otherId) continue;
			if (hasMembershipPath(otherId, candidateId, index)) {
				relevant.delete(otherId);
			}
		}
	}

	return relevant;
};

const addContextualAncestors = (visibleIds: Set<string>, index: GraphIndex, detail: number): void => {
	let changed = true;

	while (changed) {
		changed = false;

		for (const [childId, parents] of index.parentsByChild) {
			if (!visibleIds.has(childId)) continue;

			for (const membership of parents) {
				const parent = index.entitiesById.get(membership.source);
				if (!parent || !isArchitectureCategory(parent.category) || parent.id === childId) continue;
				if (architectureLevel(parent)! > detail || visibleIds.has(parent.id)) continue;

				visibleIds.add(parent.id);
				changed = true;
			}
		}
	}
};

const selectArchitectureNodes = (
	graph: GraphSubset,
	index: GraphIndex,
	reveal: MapRevealState
): Set<string> => {
	const detail = clampDetail(reveal.detail);
	const nearbyIds = new Set(reveal.nearbyEntityIds);
	const visibleIds = new Set<string>();

	for (const entity of graph.entities) {
		if (!isArchitectureCategory(entity.category) || architectureLevel(entity)! > detail) continue;

		// Landscape functions form the stable overview. Finer levels are local to
		// the viewport margin or an explicitly revealed subject.
		if (architectureLevel(entity) === 0 || nearbyIds.has(entity.id)) {
			visibleIds.add(entity.id);
		}
	}

	addContextualAncestors(visibleIds, index, detail);

	// Shared dependencies can be promoted when multiple relevant visible ancestors
	// establish shared context, even before the entity's own category would normally
	// appear. Membership coverage is not used to invent an exclusive parent.
	let addedSharedEntity = true;
	while (addedSharedEntity) {
		addedSharedEntity = false;

		for (const [childId] of index.parentsByChild) {
			if (visibleIds.has(childId)) continue;

			const child = index.entitiesById.get(childId);
			if (!child || !isArchitectureCategory(child.category)) continue;

			const visibleParentIds = findRelevantVisibleArchitectureAncestors(childId, visibleIds, index);
			if (visibleParentIds.size < 2) continue;

			visibleIds.add(childId);
			addedSharedEntity = true;
		}

		if (addedSharedEntity) addContextualAncestors(visibleIds, index, detail);
	}

	return visibleIds;
};

const cyclicMembershipIds = (index: GraphIndex): ReadonlySet<string> => {
	const cyclicIds = new Set<string>();

	for (const membership of index.memberships) {
		if (
			membership.source === membership.target ||
			hasMembershipPath(membership.target, membership.source, index, membership.id)
		) {
			cyclicIds.add(membership.id);
		}
	}

	return cyclicIds;
};

const uniqueParentIds = (memberships: readonly GraphRelationship[]): Set<string> => {
	const parentIds = new Set<string>();
	for (const membership of memberships) parentIds.add(membership.source);
	return parentIds;
};

const findEnclosures = (
	visibleIds: ReadonlySet<string>,
	index: GraphIndex,
	cyclicIds: ReadonlySet<string>,
	parentMembershipCoverage: GraphSubset["coverage"]["parentMembership"]
): ReadonlyMap<string, GraphRelationship> => {
	const enclosureByChild = new Map<string, GraphRelationship>();

	if (parentMembershipCoverage !== "complete") return enclosureByChild;

	for (const membership of index.memberships) {
		if (cyclicIds.has(membership.id)) continue;
		if (!visibleIds.has(membership.source) || !visibleIds.has(membership.target)) continue;

		const parent = index.entitiesById.get(membership.source);
		const child = index.entitiesById.get(membership.target);
		if (!parent || !child || !isArchitectureCategory(parent.category) || !isArchitectureCategory(child.category)) continue;
		if (architectureLevel(parent)! > architectureLevel(child)!) continue;

		const parentIds = uniqueParentIds(index.parentsByChild.get(child.id) ?? []);
		if (parentIds.size !== 1 || !parentIds.has(parent.id)) continue;

		const current = enclosureByChild.get(child.id);
		if (!current || membership.id.localeCompare(current.id) < 0) {
			enclosureByChild.set(child.id, membership);
		}
	}

	return enclosureByChild;
};

const findNearestVisibleAncestor = (
	entityId: string,
	visibleIds: ReadonlySet<string>,
	index: GraphIndex
): string | undefined => {
	const candidates = findRelevantVisibleArchitectureAncestors(entityId, visibleIds, index);
	return candidates.size === 1 ? [...candidates][0] : undefined;
};

const buildRepresentatives = (visibleIds: ReadonlySet<string>, index: GraphIndex): Map<string, string> => {
	const representatives = new Map<string, string>();

	for (const entity of index.entitiesById.values()) {
		const canBeRepresentative =
			isArchitectureCategory(entity.category) ||
			(isActorCategory(entity.category) && visibleIds.has(entity.id));
		if (!canBeRepresentative) continue;

		if (visibleIds.has(entity.id)) {
			representatives.set(entity.id, entity.id);
			continue;
		}
		if (!isArchitectureCategory(entity.category)) continue;

		const representativeId = findNearestVisibleAncestor(entity.id, visibleIds, index);
		if (representativeId) representatives.set(entity.id, representativeId);
	}

	return representatives;
};

const addVisibleActors = (
	visibleIds: Set<string>,
	index: GraphIndex,
	relationships: readonly GraphRelationship[],
	representatives: ReadonlyMap<string, string>
): void => {
	for (const relationship of relationships) {
		if (relationship.predicate === "contains") continue;

		const source = index.entitiesById.get(relationship.source);
		const target = index.entitiesById.get(relationship.target);
		if (!source || !target) continue;
		if (
			isActorCategory(source.category) &&
			isArchitectureCategory(target.category) &&
			representatives.has(target.id)
		) {
			visibleIds.add(source.id);
		}
		if (
			isActorCategory(target.category) &&
			isArchitectureCategory(source.category) &&
			representatives.has(source.id)
		) {
			visibleIds.add(target.id);
		}
	}
};

const connectionIdentity = (
	classification: MapConnection["classification"],
	endpoints: readonly [string, string],
	predicate: string
): string => JSON.stringify([classification, endpoints[0], endpoints[1], predicate]);

const projectConnections = (
	relationships: readonly GraphRelationship[],
	index: GraphIndex,
	representatives: ReadonlyMap<string, string>,
	enclosureByChild: ReadonlyMap<string, GraphRelationship>,
	cyclicMembershipIds: ReadonlySet<string>
): MapConnection[] => {
	const connections: MapConnection[] = [];
	const summaries = new Map<string, SummaryConnection>();

	for (const relationship of relationships) {
		if (
			enclosureByChild.has(relationship.target) &&
			enclosureByChild.get(relationship.target)?.id === relationship.id
		) {
			continue;
		}
		if (relationship.predicate === "contains" && cyclicMembershipIds.has(relationship.id)) continue;

		const source = index.entitiesById.get(relationship.source);
		const target = index.entitiesById.get(relationship.target);
		if (!source || !target) continue;
		if (getMapCategoryDisplay(source.category).mode === DisplayMode.Annotation) continue;

		const sourceRepresentative = representatives.get(source.id);
		const targetRepresentative = representatives.get(target.id);
		if (!sourceRepresentative || !targetRepresentative) continue;
		if (sourceRepresentative === targetRepresentative) continue;

		const endpoints: [string, string] = [sourceRepresentative, targetRepresentative];
		const classification =
			sourceRepresentative === source.id && targetRepresentative === target.id ? "direct" : "summary";

		if (classification === "direct") {
			connections.push({
				id: `direct:${relationship.id}`,
				endpoints,
				predicate: relationship.predicate,
				classification,
				sourceRelationshipIds: [relationship.id],
			});
			continue;
		}

		const key = connectionIdentity(classification, endpoints, relationship.predicate);
		const summary = summaries.get(key);
		if (summary) {
			if (!summary.sourceRelationshipIds.includes(relationship.id)) {
				summary.sourceRelationshipIds.push(relationship.id);
			}
			continue;
		}

		summaries.set(key, {
			endpoints,
			predicate: relationship.predicate,
			sourceRelationshipIds: [relationship.id],
		});
	}

	for (const [key, summary] of summaries) {
		connections.push({
			id: `summary:${key}`,
			endpoints: summary.endpoints,
			predicate: summary.predicate,
			classification: "summary",
			sourceRelationshipIds: summary.sourceRelationshipIds,
		});
	}

	return connections;
};

const projectAnnotations = (
	relationships: readonly GraphRelationship[],
	index: GraphIndex,
	representatives: ReadonlyMap<string, string>,
	showAnnotations: boolean
): MapAnnotation[] => {
	if (!showAnnotations) return [];

	const annotations: MapAnnotation[] = [];

	for (const relationship of relationships) {
		const entity = index.entitiesById.get(relationship.source);
		if (getMapCategoryDisplay(entity?.category ?? "").mode !== DisplayMode.Annotation) continue;

		const representativeId = representatives.get(relationship.target);
		if (!representativeId) continue;

		annotations.push({
			entityId: relationship.source,
			relationshipId: relationship.id,
			originalTargetId: relationship.target,
			representativeId,
		});
	}

	return annotations;
};

/** Projects supplied graph facts into source-backed map representations before layout. */
export const projectMap = (
	graph: GraphSubset,
	reveal: MapRevealState,
	displayOptions: MapDisplayOptions
): MapProjection => {
	const index = indexGraph(graph);
	const relationships = uniqueRelationships(graph.relationships);
	const visibleIds = selectArchitectureNodes(graph, index, reveal);

	let representatives = buildRepresentatives(visibleIds, index);
	if (displayOptions.showActors) {
		addVisibleActors(visibleIds, index, relationships, representatives);
		representatives = buildRepresentatives(visibleIds, index);
	}

	const cyclicIds = cyclicMembershipIds(index);
	const enclosureByChild = findEnclosures(visibleIds, index, cyclicIds, graph.coverage.parentMembership);
	const enclosedParentIds = new Set<string>();
	for (const membership of enclosureByChild.values()) enclosedParentIds.add(membership.source);

	const nodes: MapNode[] = [];
	for (const entity of graph.entities) {
		if (!visibleIds.has(entity.id)) continue;

		const enclosure = enclosureByChild.get(entity.id);
		nodes.push({
			id: entity.id,
			appearance: enclosedParentIds.has(entity.id) ? "group" : "compact",
			...(enclosure
				? {
						enclosure: {
							parentId: enclosure.source,
							membershipId: enclosure.id,
						},
					}
				: {}),
		});
	}

	return {
		nodes,
		connections: projectConnections(relationships, index, representatives, enclosureByChild, cyclicIds),
		annotations: projectAnnotations(
			relationships,
			index,
			representatives,
			displayOptions.showAnnotations
		),
		representativeByEntityId: representatives,
	};
};
