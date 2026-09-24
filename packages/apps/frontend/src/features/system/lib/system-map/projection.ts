import { getMapCategoryDisplay, isArchitectureCategory, NodeDetailLevel } from "./category";
import { groupOpacityAtDetail, layerOpacityAtDetail } from "./interaction";
import type { GraphEntity, GraphRelationship, GraphSubset } from "./graph";
import type { MapConnection, MapNode, MapProjection, MapRevealState } from "./presentation";

type GraphIndex = {
	entitiesById: ReadonlyMap<string, GraphEntity>;
	relationships: readonly GraphRelationship[];
	resolvedRelationships: readonly GraphRelationship[];
	memberships: readonly GraphRelationship[];
	parentsByChild: ReadonlyMap<string, readonly GraphRelationship[]>;
	childrenByParent: ReadonlyMap<string, readonly GraphRelationship[]>;
	parentIdsByChild: ReadonlyMap<string, ReadonlySet<string>>;
};

type SummaryConnection = {
	endpoints: readonly [string, string];
	predicate: string;
	sourceRelationshipIds: Set<string>;
};

const architectureLevel = (entity: GraphEntity): number | undefined =>
	getMapCategoryDisplay(entity.category).level;

const uniqueRelationships = (graph: GraphSubset) => {
	const seen = new Set<string>();
	const all: GraphRelationship[] = [];
	const resolved: GraphRelationship[] = [];

	for (const relationship of graph.relationships) {
		if (seen.has(relationship.id)) continue;
		seen.add(relationship.id);
		all.push(relationship);
		resolved.push(relationship);
	}

	for (const relationship of graph.unresolvedRelationships) {
		if (seen.has(relationship.id)) continue;
		seen.add(relationship.id);
		all.push(relationship);
	}

	return { all, resolved };
};

const indexGraph = (graph: GraphSubset): GraphIndex => {
	const entitiesById = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const { all: relationships, resolved: resolvedRelationships } = uniqueRelationships(graph);
	const memberships: GraphRelationship[] = [];
	const parentsByChild = new Map<string, GraphRelationship[]>();
	const childrenByParent = new Map<string, GraphRelationship[]>();
	const parentIdsByChild = new Map<string, Set<string>>();

	for (const relationship of relationships) {
		if (relationship.predicate !== "contains") continue;
		memberships.push(relationship);

		const parentIds = parentIdsByChild.get(relationship.target) ?? new Set<string>();
		parentIds.add(relationship.source);
		parentIdsByChild.set(relationship.target, parentIds);

		const parents = parentsByChild.get(relationship.target) ?? [];
		parents.push(relationship);
		parentsByChild.set(relationship.target, parents);

		if (!entitiesById.has(relationship.source) || !entitiesById.has(relationship.target)) continue;
		const children = childrenByParent.get(relationship.source) ?? [];
		children.push(relationship);
		childrenByParent.set(relationship.source, children);
	}

	return {
		entitiesById,
		relationships,
		resolvedRelationships,
		memberships,
		parentsByChild,
		childrenByParent,
		parentIdsByChild,
	};
};

const clampDetail = (detail: number): number => Math.max(0, Math.min(3, detail));

const validArchitectureMembershipLevels = (
	membership: GraphRelationship,
	index: GraphIndex
): { parentLevel: number; childLevel: number } | undefined => {
	const parent = index.entitiesById.get(membership.source);
	const child = index.entitiesById.get(membership.target);
	if (!parent || !child) return undefined;
	const parentLevel = architectureLevel(parent);
	const childLevel = architectureLevel(child);
	if (parentLevel === undefined || childLevel === undefined || parentLevel > childLevel) return undefined;
	return { parentLevel, childLevel };
};

const hasMembershipPath = (
	startId: string,
	targetId: string,
	index: GraphIndex,
	excludedRelationshipId?: string,
	excludedRelationshipIds: ReadonlySet<string> = new Set()
): boolean => {
	const visited = new Set<string>();
	const pending = [startId];

	while (pending.length > 0) {
		const currentId = pending.pop()!;
		if (currentId === targetId) return true;
		if (visited.has(currentId)) continue;
		visited.add(currentId);

		for (const membership of index.childrenByParent.get(currentId) ?? []) {
			if (membership.id !== excludedRelationshipId && !excludedRelationshipIds.has(membership.id)) {
				pending.push(membership.target);
			}
		}
	}

	return false;
};

const hasValidArchitectureMembershipPath = (
	startId: string,
	targetId: string,
	index: GraphIndex,
	excludedMembershipIds: ReadonlySet<string>
): boolean => {
	const pending = [startId];
	const visited = new Set<string>();

	while (pending.length > 0) {
		const currentId = pending.pop()!;
		if (visited.has(currentId)) continue;
		visited.add(currentId);

		for (const membership of index.childrenByParent.get(currentId) ?? []) {
			if (excludedMembershipIds.has(membership.id)) continue;
			if (!validArchitectureMembershipLevels(membership, index)) continue;
			if (membership.target === targetId) return true;
			pending.push(membership.target);
		}
	}

	return false;
};

const findRelevantVisibleArchitectureAncestors = (
	entityId: string,
	visibleIds: ReadonlySet<string>,
	index: GraphIndex,
	maximumParentLevel?: number,
	excludedMembershipIds: ReadonlySet<string> = new Set()
): Set<string> => {
	const entity = index.entitiesById.get(entityId);
	const initialLevel = entity ? architectureLevel(entity) : undefined;
	const parentLevelLimit = maximumParentLevel ?? initialLevel;
	const pending = [entityId];
	const visited = new Set<string>();
	const candidates = new Set<string>();

	while (pending.length > 0) {
		const currentId = pending.pop()!;
		if (visited.has(currentId)) continue;
		visited.add(currentId);

		for (const membership of index.parentsByChild.get(currentId) ?? []) {
			if (excludedMembershipIds.has(membership.id)) continue;
			const levels = validArchitectureMembershipLevels(membership, index);
			if (!levels) continue;
			const parentId = membership.source;
			const parentLevel = levels.parentLevel;
			if (
				(parentLevelLimit === undefined || parentLevel <= parentLevelLimit) &&
				visibleIds.has(parentId)
			) {
				candidates.add(parentId);
			}
			pending.push(parentId);
		}
	}

	const nearest = new Set(candidates);
	for (const candidateId of candidates) {
		for (const otherId of candidates) {
			if (
				candidateId !== otherId &&
				hasValidArchitectureMembershipPath(candidateId, otherId, index, excludedMembershipIds)
			) {
				nearest.delete(candidateId);
				break;
			}
		}
	}
	return nearest;
};

const findCommonVisibleArchitectureAncestors = (
	parentIds: ReadonlySet<string>,
	visibleIds: ReadonlySet<string>,
	index: GraphIndex,
	maximumParentLevel: number,
	excludedMembershipIds: ReadonlySet<string>
): Set<string> => {
	const ancestorsByParent = [...parentIds].map((parentId) => {
		const pending = [parentId];
		const visited = new Set<string>();
		const visibleAncestors = new Set<string>();

		while (pending.length > 0) {
			const currentId = pending.pop()!;
			if (visited.has(currentId)) continue;
			visited.add(currentId);
			const currentEntity = index.entitiesById.get(currentId);
			if (!currentEntity) continue;
			const currentLevel = architectureLevel(currentEntity);
			if (currentLevel === undefined || currentLevel > maximumParentLevel) continue;
			if (visibleIds.has(currentId)) {
				visibleAncestors.add(currentId);
			}

			for (const membership of index.parentsByChild.get(currentId) ?? []) {
				if (excludedMembershipIds.has(membership.id)) continue;
				const levels = validArchitectureMembershipLevels(membership, index);
				if (!levels || levels.childLevel !== currentLevel) continue;
				const ancestorId = membership.source;
				pending.push(ancestorId);
			}
		}

		return visibleAncestors;
	});
	if (ancestorsByParent.length === 0) return new Set();

	const common = new Set(ancestorsByParent[0]);
	for (const ancestors of ancestorsByParent.slice(1)) {
		for (const ancestorId of common) {
			if (!ancestors.has(ancestorId)) common.delete(ancestorId);
		}
	}

	for (const candidateId of common) {
		for (const otherId of common) {
			if (
				candidateId !== otherId &&
				hasValidArchitectureMembershipPath(candidateId, otherId, index, excludedMembershipIds)
			) {
				common.delete(candidateId);
				break;
			}
		}
	}
	return common;
};

const addContextualAncestors = (
	visibleIds: Set<string>,
	index: GraphIndex,
	detail: number,
	excludedMembershipIds: ReadonlySet<string>
): void => {
	let changed = true;
	while (changed) {
		changed = false;
		for (const [childId, parents] of index.parentsByChild) {
			if (!visibleIds.has(childId)) continue;
			for (const membership of parents) {
				if (excludedMembershipIds.has(membership.id)) continue;
				const parent = index.entitiesById.get(membership.source);
				if (!parent) continue;
				const level = architectureLevel(parent);
				if (level === undefined || level > detail || visibleIds.has(parent.id)) continue;
				visibleIds.add(parent.id);
				changed = true;
			}
		}
	}
};

const shallowestArchitectureLevel = (entities: readonly GraphEntity[]): number | undefined => {
	let minimum: number | undefined;
	for (const entity of entities) {
		const level = architectureLevel(entity);
		if (level !== undefined && (minimum === undefined || level < minimum)) minimum = level;
	}
	return minimum;
};

const selectArchitectureNodes = (
	graph: GraphSubset,
	index: GraphIndex,
	reveal: MapRevealState,
	excludedMembershipIds: ReadonlySet<string>
): { visibleIds: Set<string>; baselineLevel: number | undefined } => {
	const detail = clampDetail(reveal.detail);
	const visualDetail = clampDetail(reveal.visualDetail ?? reveal.detail);
	const nearbyIds = new Set(reveal.nearbyEntityIds);
	const baselineLevel = shallowestArchitectureLevel(graph.entities);
	const visibleIds = new Set<string>();

	for (const entity of graph.entities) {
		const level = architectureLevel(entity);
		if (level === undefined) continue;
		const isNearby = nearbyIds.has(entity.id);
		const inTransition = layerOpacityAtDetail(level, visualDetail) > 0;
		if (level === baselineLevel || (isNearby && (level <= detail || inTransition))) {
			visibleIds.add(entity.id);
		}
	}

	addContextualAncestors(visibleIds, index, detail, excludedMembershipIds);

	let addedSharedEntity = true;
	while (addedSharedEntity) {
		addedSharedEntity = false;
		for (const [childId] of index.parentsByChild) {
			if (visibleIds.has(childId)) continue;
			const child = index.entitiesById.get(childId);
			if (!child || !isArchitectureCategory(child.category)) continue;
			if (
				findRelevantVisibleArchitectureAncestors(
					childId,
					visibleIds,
					index,
					undefined,
					excludedMembershipIds
				).size < 2
			)
				continue;
			visibleIds.add(childId);
			addedSharedEntity = true;
		}
		if (addedSharedEntity) addContextualAncestors(visibleIds, index, detail, excludedMembershipIds);
	}

	// A shared relationship endpoint must have one canonical displayed instance even when
	// its visible enclosing groups would otherwise compete to represent it.
	let promotedEndpoint = true;
	while (promotedEndpoint) {
		promotedEndpoint = false;
		for (const relationship of index.resolvedRelationships) {
			for (const entityId of [relationship.source, relationship.target]) {
				if (visibleIds.has(entityId)) continue;
				const entity = index.entitiesById.get(entityId);
				if (!entity || !isArchitectureCategory(entity.category)) continue;
				if (
					findRelevantVisibleArchitectureAncestors(
						entityId,
						visibleIds,
						index,
						undefined,
						excludedMembershipIds
					).size < 2
				)
					continue;
				visibleIds.add(entityId);
				promotedEndpoint = true;
			}
		}
	}

	return { visibleIds, baselineLevel };
};

const cyclicMembershipIds = (index: GraphIndex): ReadonlySet<string> => {
	const cyclicIds = new Set<string>();
	for (const { id, source, target } of index.memberships) {
		if (!index.entitiesById.has(source) || !index.entitiesById.has(target)) continue;
		if (source === target || hasMembershipPath(target, source, index, id)) cyclicIds.add(id);
	}
	return cyclicIds;
};

const findEnclosures = (
	visibleIds: ReadonlySet<string>,
	index: GraphIndex,
	cyclicIds: ReadonlySet<string>
): ReadonlyMap<string, GraphRelationship> => {
	const enclosureByChild = new Map<string, GraphRelationship>();
	for (const membership of index.memberships) {
		if (cyclicIds.has(membership.id)) continue;
		if (!visibleIds.has(membership.source) || !visibleIds.has(membership.target)) continue;

		const parent = index.entitiesById.get(membership.source);
		const child = index.entitiesById.get(membership.target);
		if (
			!parent ||
			!child ||
			!isArchitectureCategory(parent.category) ||
			!isArchitectureCategory(child.category)
		) {
			continue;
		}
		if (architectureLevel(parent)! > architectureLevel(child)!) continue;

		const parentIds = index.parentIdsByChild.get(child.id) ?? new Set<string>();
		if (parentIds.size !== 1 || !parentIds.has(parent.id)) continue;
		const current = enclosureByChild.get(child.id);
		if (!current || membership.id.localeCompare(current.id) < 0) {
			enclosureByChild.set(child.id, membership);
		}
	}
	return enclosureByChild;
};

const buildRepresentatives = (
	visibleIds: ReadonlySet<string>,
	index: GraphIndex,
	excludedMembershipIds: ReadonlySet<string>
): Map<string, string> => {
	const representatives = new Map<string, string>();
	for (const entity of index.entitiesById.values()) {
		if (!isArchitectureCategory(entity.category)) continue;
		if (visibleIds.has(entity.id)) {
			representatives.set(entity.id, entity.id);
			continue;
		}
		const level = architectureLevel(entity);
		if (level === undefined) continue;
		const parentIds = index.parentIdsByChild.get(entity.id) ?? new Set<string>();
		let ancestors: ReadonlySet<string>;
		if (parentIds.size > 1) {
			if ([...parentIds].some((parentId) => !index.entitiesById.has(parentId))) continue;
			ancestors = findCommonVisibleArchitectureAncestors(
				parentIds,
				visibleIds,
				index,
				level,
				excludedMembershipIds
			);
		} else {
			ancestors = findRelevantVisibleArchitectureAncestors(
				entity.id,
				visibleIds,
				index,
				level,
				excludedMembershipIds
			);
		}
		if (ancestors.size === 1) representatives.set(entity.id, [...ancestors][0]);
	}
	return representatives;
};

const summaryIdentity = (endpoints: readonly [string, string], predicate: string): string =>
	JSON.stringify(["summary", endpoints[0], endpoints[1], predicate]);

const projectConnections = (
	index: GraphIndex,
	representatives: ReadonlyMap<string, string>,
	enclosureByChild: ReadonlyMap<string, GraphRelationship>
): MapConnection[] => {
	const direct: MapConnection[] = [];
	const summaries = new Map<string, SummaryConnection>();
	const relationships = [...index.resolvedRelationships].sort((left, right) =>
		left.id.localeCompare(right.id)
	);

	for (const relationship of relationships) {
		if (enclosureByChild.get(relationship.target)?.id === relationship.id) continue;
		const sourceRepresentative = representatives.get(relationship.source);
		const targetRepresentative = representatives.get(relationship.target);
		if (!sourceRepresentative || !targetRepresentative || sourceRepresentative === targetRepresentative)
			continue;

		const endpoints = [sourceRepresentative, targetRepresentative] as const;
		const isDirect =
			sourceRepresentative === relationship.source && targetRepresentative === relationship.target;
		if (isDirect) {
			direct.push({
				id: `direct:${relationship.id}`,
				endpoints,
				predicate: relationship.predicate,
				classification: "direct",
				sourceRelationshipIds: [relationship.id],
			});
			continue;
		}

		const key = summaryIdentity(endpoints, relationship.predicate);
		const summary = summaries.get(key);
		if (summary) summary.sourceRelationshipIds.add(relationship.id);
		else
			summaries.set(key, {
				endpoints,
				predicate: relationship.predicate,
				sourceRelationshipIds: new Set([relationship.id]),
			});
	}

	const projectedSummaries = [...summaries.entries()].map(([key, summary]): MapConnection => ({
		id: `summary:${key}`,
		endpoints: summary.endpoints,
		predicate: summary.predicate,
		classification: "summary",
		sourceRelationshipIds: [...summary.sourceRelationshipIds].sort(),
	}));
	return [...direct, ...projectedSummaries];
};

/** Projects only supplied architectural entities and resolved relationships into map representations. */
export const projectMap = (graph: GraphSubset, reveal: MapRevealState): MapProjection => {
	const index = indexGraph(graph);
	const cyclicIds = cyclicMembershipIds(index);
	const { visibleIds, baselineLevel } = selectArchitectureNodes(graph, index, reveal, cyclicIds);
	const enclosureByChild = findEnclosures(visibleIds, index, cyclicIds);
	const enclosedParentIds = new Set([...enclosureByChild.values()].map((membership) => membership.source));
	const fullArchitectureIds = new Set(
		[...index.entitiesById.values()]
			.filter((entity) => isArchitectureCategory(entity.category))
			.map((entity) => entity.id)
	);
	const fullEnclosures = findEnclosures(fullArchitectureIds, index, cyclicIds);
	const fullGroupIds = new Set([...fullEnclosures.values()].map((membership) => membership.source));
	const detail = clampDetail(reveal.visualDetail ?? reveal.detail);
	const nearbyIds = new Set(reveal.nearbyEntityIds);
	const nodes: MapNode[] = [...visibleIds]
		.sort((left, right) => left.localeCompare(right))
		.map((id) => {
			const membership = enclosureByChild.get(id);
			const entity = index.entitiesById.get(id)!;
			const level = architectureLevel(entity)!;
			const localReveal = level > reveal.detail && nearbyIds.has(id);
			const layerOpacity =
				level === baselineLevel || (!localReveal && level > reveal.detail)
					? 1
					: layerOpacityAtDetail(level, detail);
			const opacity = fullGroupIds.has(id)
				? layerOpacity * groupOpacityAtDetail(level, detail)
				: layerOpacity;
			return {
				id,
				appearance: enclosedParentIds.has(id) ? "group" : "compact",
				opacity,
				...(membership
					? { enclosure: { parentId: membership.source, membershipId: membership.id } }
					: {}),
			};
		});
	const representativeByEntityId = buildRepresentatives(visibleIds, index, cyclicIds);
	const opacityById = new Map(nodes.map((node) => [node.id, node.opacity ?? 1]));
	const connections = projectConnections(index, representativeByEntityId, enclosureByChild).map(
		(connection) => ({
			...connection,
			opacity: Math.min(
				opacityById.get(connection.endpoints[0]) ?? 1,
				opacityById.get(connection.endpoints[1]) ?? 1
			),
		})
	);

	return { nodes, connections, representativeByEntityId };
};

/** Projects every supplied architectural entity for the map's zoom-independent full layout. */
export const projectFullArchitecture = (graph: GraphSubset): MapProjection =>
	projectMap(graph, {
		detail: NodeDetailLevel.Implementation,
		nearbyEntityIds: graph.entities
			.filter((entity) => isArchitectureCategory(entity.category))
			.map((entity) => entity.id),
	});
