/** Canonical knowledge graph records supplied by a host; entity IDs are unique within a subset. */
export type GraphEntity = {
	readonly id: string;
	readonly category: string;
	readonly kind: string;
};

/** One canonical directed relationship. IDs are unique across both relationship arrays in a subset. */
export type GraphRelationship = {
	readonly id: string;
	readonly source: string;
	readonly target: string;
	readonly predicate: string;
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
	readonly entities: readonly GraphEntity[];
	readonly relationships: readonly GraphRelationship[];
	readonly unresolvedRelationships: readonly GraphRelationship[];
	readonly enumeration: GraphEnumeration;
};
