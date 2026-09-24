import type { SystemAnalysisEdge, SystemAnalysisEntry, SystemAnalysisNode } from "$lib/api";
import type { Point } from "$features/system/lib/system-map/geometry";
import type { GraphEntity, GraphRelationship, GraphSubset } from "$features/system/lib/system-map/graph";
import type { ResponsePage } from "$lib/api/utils";

export function flattenPages<T extends { id: string }>(pages: ResponsePage<T>[] | undefined) {
	const items = new Map<string, T>();
	for (const page of pages ?? []) {
		for (const item of page.data) {
			items.set(item.id, item);
		}
	}
	return [...items.values()];
}

export type AnalysisGraph = {
	graph: GraphSubset;
	positions: Readonly<Partial<Record<string, Point>>>;
	nodeByEntityId: ReadonlyMap<string, SystemAnalysisNode>;
	edgeByRelationshipId: ReadonlyMap<string, SystemAnalysisEdge>;
	duplicateEntityIds: readonly string[];
	duplicateRelationshipIds: readonly string[];
};

function preferLowestAnalysisId<T extends { id: string }>(current: T, candidate: T): T {
	return candidate.id < current.id ? candidate : current;
}

function canonicalNodeEntity(node: SystemAnalysisNode): GraphEntity {
	const entity = node.attributes.knowledgeEntity;
	return {
		id: entity.id,
		category: entity.attributes.category,
		kind: entity.attributes.kind,
	};
}

function canonicalEdgeRelationship(edge: SystemAnalysisEdge): GraphRelationship {
	const relationship = edge.attributes.knowledgeRelationship;
	return {
		id: relationship.id,
		source: relationship.attributes.sourceEntityId,
		target: relationship.attributes.targetEntityId,
		predicate: relationship.attributes.predicate,
	};
}

/** Adapts included analysis records to canonical map records without enriching or filtering them. */
export function buildAnalysisGraph(
	analysisNodes: readonly SystemAnalysisNode[],
	analysisEdges: readonly SystemAnalysisEdge[]
): AnalysisGraph {
	const nodeByEntityId = new Map<string, SystemAnalysisNode>();
	const duplicateEntityIds = new Set<string>();
	for (const node of analysisNodes) {
		const entityId = node.attributes.knowledgeEntity.id;
		const existing = nodeByEntityId.get(entityId);
		if (existing) {
			duplicateEntityIds.add(entityId);
			nodeByEntityId.set(entityId, preferLowestAnalysisId(existing, node));
		} else {
			nodeByEntityId.set(entityId, node);
		}
	}

	const edgeByRelationshipId = new Map<string, SystemAnalysisEdge>();
	const duplicateRelationshipIds = new Set<string>();
	for (const edge of analysisEdges) {
		const relationshipId = edge.attributes.knowledgeRelationship.id;
		const existing = edgeByRelationshipId.get(relationshipId);
		if (existing) {
			duplicateRelationshipIds.add(relationshipId);
			edgeByRelationshipId.set(relationshipId, preferLowestAnalysisId(existing, edge));
		} else {
			edgeByRelationshipId.set(relationshipId, edge);
		}
	}

	const entities = [...nodeByEntityId.entries()]
		.sort(([left], [right]) => (left < right ? -1 : left > right ? 1 : 0))
		.map(([, node]) => canonicalNodeEntity(node));
	const entityIds = new Set(nodeByEntityId.keys());
	const positions = Object.fromEntries(
		[...nodeByEntityId].map(([entityId, node]) => [entityId, { ...node.attributes.position }])
	) as Partial<Record<string, Point>>;

	const relationships = [...edgeByRelationshipId.entries()]
		.sort(([left], [right]) => (left < right ? -1 : left > right ? 1 : 0))
		.map(([, edge]) => canonicalEdgeRelationship(edge));
	const resolvedRelationships: GraphRelationship[] = [];
	const unresolvedRelationships: GraphRelationship[] = [];
	for (const relationship of relationships) {
		if (entityIds.has(relationship.source) && entityIds.has(relationship.target)) {
			resolvedRelationships.push(relationship);
		} else {
			unresolvedRelationships.push(relationship);
		}
	}

	return {
		graph: {
			entities,
			relationships: resolvedRelationships,
			unresolvedRelationships,
			enumeration: { scope: "analysis", stopReason: "exhausted" },
		},
		positions,
		nodeByEntityId,
		edgeByRelationshipId,
		duplicateEntityIds: [...duplicateEntityIds].sort(),
		duplicateRelationshipIds: [...duplicateRelationshipIds].sort(),
	};
}

export type EntryAttachments = {
	byEntityId: Map<string, SystemAnalysisEntry[]>;
	byRelationshipId: Map<string, SystemAnalysisEntry[]>;
	unattached: SystemAnalysisEntry[];
};

export function mapEntryAttachments(
	nodeByEntityId: ReadonlyMap<string, SystemAnalysisNode>,
	edgeByRelationshipId: ReadonlyMap<string, SystemAnalysisEdge>,
	entries: readonly SystemAnalysisEntry[]
): EntryAttachments {
	const byEntityId = new Map<string, SystemAnalysisEntry[]>();
	const byRelationshipId = new Map<string, SystemAnalysisEntry[]>();
	const unattached: SystemAnalysisEntry[] = [];

	for (const entry of entries) {
		const entityIds = new Set<string>();
		const relationshipIds = new Set<string>();
		for (const subject of entry.attributes.subjects) {
			const { knowledgeEntityId, knowledgeRelationshipId } = subject.attributes;
			if (knowledgeEntityId && nodeByEntityId.has(knowledgeEntityId)) {
				entityIds.add(knowledgeEntityId);
			}
			if (knowledgeRelationshipId && edgeByRelationshipId.has(knowledgeRelationshipId)) {
				relationshipIds.add(knowledgeRelationshipId);
			}
		}

		for (const entityId of entityIds) {
			byEntityId.set(entityId, [...(byEntityId.get(entityId) ?? []), entry]);
		}
		for (const relationshipId of relationshipIds) {
			byRelationshipId.set(relationshipId, [...(byRelationshipId.get(relationshipId) ?? []), entry]);
		}
		if (!entityIds.size && !relationshipIds.size) {
			unattached.push(entry);
		}
	}

	return { byEntityId, byRelationshipId, unattached };
}
