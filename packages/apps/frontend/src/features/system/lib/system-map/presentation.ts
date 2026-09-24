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
