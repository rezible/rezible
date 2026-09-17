/** Completeness reported by retrieval or fixture metadata; never inferred from graph shape. */
export enum Coverage {
	/** The requested enumeration finished without remaining pages, truncation, or unresolved limits. */
	Complete = "complete",
	/** More matching facts are known to exist than were supplied. */
	Partial = "partial",
	/** The producer cannot establish whether the requested enumeration is complete. */
	Unknown = "unknown",
}

/** Canonical source entity supplied to the map, independent of API response wrappers. */
export type GraphEntity = {
	/** Opaque canonical identity; the map does not require UUID syntax. */
	id: string;
	/** Original identifier; parseMapCategory selects presentation without replacing this value. */
	category: string;
	label: string;
	kind: string;
	properties?: Readonly<Record<string, unknown>>;
};

/** One canonical directed fact; contains means membership from parent to member. */
export type GraphRelationship = {
	/** Opaque canonical identity, independent of any displayed connection ID. */
	id: string;
	source: string;
	target: string;
	predicate: string;
	/** Supporting evidence references do not increase the canonical relationship count. */
	supportReferences?: readonly string[];
};

/** Supplied graph content, independent of viewport, display options, and inspection selection. */
export type GraphSubset = {
	entities: readonly GraphEntity[];
	/** Includes membership; each relationship ID occurs once and both endpoints must be supplied. */
	relationships: readonly GraphRelationship[];
	coverage: {
		/** All parents of supplied entities, including outside the chosen scope; not children of one parent. */
		parentMembership: Coverage;
		/** Ordinary relationships within the requested scope, not completeness of knowledge about reality. */
		relationships: Coverage;
	};
};
