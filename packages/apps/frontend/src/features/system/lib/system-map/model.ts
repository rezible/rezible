/**
 * The knowledge graph's category and predicate rules, from GET /knowledge/model. The map keeps no rules of
 * its own; it maps levels and families to drawing.
 */
export type MapModel = {
	readonly categories: ReadonlyMap<string, { readonly hierarchy?: string; readonly level?: number }>;
	readonly predicates: ReadonlyMap<string, { readonly family: string; readonly rollsUp: boolean }>;
	/** Categories by hierarchy. */
	readonly hierarchies: ReadonlyMap<string, readonly string[]>;
};

/** The served model, structurally, so hosts and tests can pass it without the API client. */
export type MapModelSource = {
	readonly categories: readonly { value: string; hierarchy?: string; level?: number }[];
	readonly predicates: readonly { value: string; family: string; rollsUp: boolean }[];
	readonly hierarchies: readonly { value: string; categories: readonly string[] }[];
};

export const mapModel = (source: MapModelSource): MapModel => ({
	categories: new Map(
		source.categories.map(({ value, hierarchy, level }) => [value, { hierarchy, level }])
	),
	predicates: new Map(source.predicates.map(({ value, family, rollsUp }) => [value, { family, rollsUp }])),
	hierarchies: new Map(source.hierarchies.map(({ value, categories }) => [value, categories])),
});

export const emptyMapModel: MapModel = mapModel({ categories: [], predicates: [], hierarchies: [] });

/** Logical levels the map draws: regions at 0 and 1, cards at 2, components folded into cards at 3. */
export const LANDSCAPE_LEVEL = 0;
export const SYSTEM_LEVEL = 1;
export const CONTAINER_LEVEL = 2;
export const COMPONENT_LEVEL = 3;

export const LOGICAL_HIERARCHY = "logical";
export const DEPLOYMENT_HIERARCHY = "deployment";

export const categoryLevel = (model: MapModel, category: string): number | undefined =>
	model.categories.get(category)?.level;

export const inHierarchy = (model: MapModel, category: string, hierarchy: string): boolean =>
	model.categories.get(category)?.hierarchy === hierarchy;

export const rollsUp = (model: MapModel, predicate: string): boolean =>
	model.predicates.get(predicate)?.rollsUp ?? false;

export const predicateFamily = (model: MapModel, predicate: string): string | undefined =>
	model.predicates.get(predicate)?.family;

export const isPlacement = (model: MapModel, predicate: string): boolean =>
	predicateFamily(model, predicate) === "placement";

/** The categories the map draws: the logical and deployment hierarchies'. */
export const mapCategories = (model: MapModel): string[] => [
	...(model.hierarchies.get(LOGICAL_HIERARCHY) ?? []),
	...(model.hierarchies.get(DEPLOYMENT_HIERARCHY) ?? []),
];
