import type { MapModel } from "./model";

/** Canonical knowledge graph records supplied by a host; entity IDs are unique within a subset. */
export type GraphEntity = {
	readonly id: string;
	readonly category: string;
	readonly kind: string;
	/** Display name from the entity's state; empty when no source has named it. */
	readonly name: string;
	/** The entity's one parent, chosen by the backend from its contains claims. */
	readonly parentId?: string;
	/** Sources of the entity's other current contains claims. */
	readonly otherParentIds?: readonly string[];
	/** Every source that still lists it has stopped observing it, so it may no longer exist. */
	readonly stale?: boolean;
};

/** One canonical directed relationship. IDs are unique across both relationship arrays in a subset. */
export type GraphRelationship = {
	readonly id: string;
	readonly source: string;
	readonly target: string;
	readonly predicate: string;
	readonly stale?: boolean;
};

/** Describes enumeration progress only within the named query, analysis, or fixture scope. */
export type GraphEnumeration = {
	readonly scope: "query" | "analysis" | "fixture";
	readonly stopReason: "exhausted" | "entity-limit" | "relationship-limit" | "request-limit";
};

/**
 * Supplied canonical records partitioned by endpoint availability. Every resolved relationship has both
 * endpoints in entities; every unresolved relationship has at least one endpoint missing from entities.
 * Partial state is derived from enumeration.stopReason and unresolvedRelationships.
 */
export type GraphSubset = {
	readonly model: MapModel;
	readonly entities: readonly GraphEntity[];
	readonly relationships: readonly GraphRelationship[];
	readonly unresolvedRelationships: readonly GraphRelationship[];
	readonly enumeration: GraphEnumeration;
};

/** Ancestry walks stop here, as the backend's do, against cycles from concurrent writes. */
export const MAX_ANCESTRY_DEPTH = 16;

/** Loaded ancestors along parentId, nearest first. Stops at a parent outside the subset or a repeat. */
export function loadedAncestors(entities: ReadonlyMap<string, GraphEntity>, id: string): GraphEntity[] {
	const ancestors: GraphEntity[] = [];
	const seen = new Set([id]);
	let parentId = entities.get(id)?.parentId;
	while (parentId && !seen.has(parentId) && ancestors.length < MAX_ANCESTRY_DEPTH) {
		const parent = entities.get(parentId);
		if (!parent) {
			break;
		}
		ancestors.push(parent);
		seen.add(parentId);
		parentId = parent.parentId;
	}
	return ancestors;
}
