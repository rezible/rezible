/** Optional context to display without changing the supplied architecture. */
export type MapDisplayOptions = {
	showAnnotations: boolean;
};

/** Continuous reveal input; projection uses discrete thresholds, rendering interpolates appearance. */
export type MapRevealState = {
	/** Zoom-derived position from Landscape (0) to Implementation (3), including fractional values. */
	detail: number;
	/** Entities eligible for local reveal, including the viewport margin and explicit Reveal targets. */
	nearbyEntityIds: readonly string[];
};

export type MapNodeAppearance = "compact" | "group";

/** A display assignment backed by exclusive membership, not a replacement for source memberships. */
type MapNodeDisplayEnclosure = {
	parentId: string;
	membershipId: string;
};

/** One displayed architectural entity; actor records remain source-inspectable and are not projected. */
export type MapNode = {
	/** The canonical entity ID, unchanged when the entity becomes a group boundary. */
	id: string;
	/** Both appearances are nodes representing the same entity. */
	appearance: MapNodeAppearance;
	enclosure?: MapNodeDisplayEnclosure;
};

/** A displayed connection; summaries are presentations of facts, not new source relationships. */
export type MapConnection = {
	/** Stable display identity; summary identity excludes counts and contributing relationship IDs. */
	id: string;
	/** Ordered source and target MapNode IDs. */
	endpoints: readonly [string, string];
	predicate: string;
	/** Direct preserves original endpoints; summary substitutes representatives, even for one fact. */
	classification: "direct" | "summary";
	/** Unique contributing canonical IDs; their length is the count of supplied relationships. */
	sourceRelationshipIds: readonly string[];
};

/** A graph entity annotating an original target through one source relationship. */
export type MapAnnotation = {
	entityId: string;
	relationshipId: string;
	originalTargetId: string;
	/** Visible node carrying the marker; a hidden descendant does not imply whole-group impact. */
	representativeId: string;
};

/** Presentation output before layout; the source subset remains available for full inspection. */
export type MapProjection = {
	nodes: readonly MapNode[];
	/** Includes shared membership connections as well as ordinary and summary connections. */
	connections: readonly MapConnection[];
	annotations: readonly MapAnnotation[];
	/** Source entity to its single visible representative; missing entries have no representative. */
	representativeByEntityId: ReadonlyMap<string, string>;
};

/** Stable source references being inspected, independent of current canvas visibility. */
export type InspectionTarget =
	| { kind: "entity"; entityId: string }
	| { kind: "relationship"; relationshipId: string }
	/** Retain contributing source IDs so inspection survives regrouping or disappearance of the summary. */
	| { kind: "summary"; sourceRelationshipIds: readonly string[] }
	| { kind: "annotation"; entityId: string; relationshipId: string };
