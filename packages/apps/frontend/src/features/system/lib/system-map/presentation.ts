/**
 * Canonical subject selected by the host. Summary selections retain their original relationship IDs as
 * projection regroups them.
 */
export type MapSelection =
	| { readonly kind: "entity"; readonly entityId: string }
	| { readonly kind: "relationship"; readonly relationshipId: string }
	| { readonly kind: "summary"; readonly relationshipIds: readonly string[] };

/** Canonical subjects a host asks the map to emphasize; unavailable IDs are ignored. */
export type MapHighlights = {
	readonly entityIds: readonly string[];
	readonly relationshipIds: readonly string[];
};

/** Continuous reveal input; projection uses discrete thresholds while rendering stays continuous. */
export type MapRevealState = {
	/** Discrete layer currently used for the structural projection. */
	detail: number;
	/** Continuous zoom-derived detail used for opacity at layer boundaries. */
	visualDetail?: number;
	/** Supplied entities eligible for local reveal from viewport proximity or explicit Reveal. */
	nearbyEntityIds: readonly string[];
};

export type MapNodeAppearance = "compact" | "group";

type MapNodeDisplayEnclosure = {
	parentId: string;
	membershipId: string;
};

/** One displayed architectural entity; a group remains the same canonical entity. */
export type MapNode = {
	id: string;
	appearance: MapNodeAppearance;
	/** Direct opacity for this node; sibling descendants are not affected. */
	opacity?: number;
	enclosure?: MapNodeDisplayEnclosure;
};

/** A direct source relationship or summary of source relationships between visible representatives. */
export type MapConnection = {
	id: string;
	endpoints: readonly [string, string];
	predicate: string;
	classification: "direct" | "summary";
	sourceRelationshipIds: readonly string[];
	/** Minimum opacity of the currently rendered endpoint nodes. */
	opacity?: number;
};

/** Projection output before layout; all IDs remain canonical source IDs except stable summary IDs. */
export type MapProjection = {
	nodes: readonly MapNode[];
	connections: readonly MapConnection[];
	/** Source entity ID to its single visible representative. */
	representativeByEntityId: ReadonlyMap<string, string>;
};
