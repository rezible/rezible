import { getMapCategoryDisplay, MapCategory } from "./category";
import type { GraphEntity } from "./graph";

const knownCategories = new Set<string>(Object.values(MapCategory));

/** Canvas entity text follows the canonical kind without adding identity or inferred names. */
export const entityLabel = (entity: GraphEntity): string => entity.kind;

/** Uses the existing readable category name while preserving unrecognized source values. */
export const entityCategoryLabel = (category: string): string =>
	knownCategories.has(category) ? getMapCategoryDisplay(category).categoryLabel : category;

/** Relationship labels are presentation text; callers retain the original predicate value. */
export const relationshipPredicateLabel = (predicate: string): string => predicate.replaceAll("_", " ");
