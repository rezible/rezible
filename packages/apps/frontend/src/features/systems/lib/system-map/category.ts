export enum MapCategory {
	/** A business/organisational ability */
	Function = "function",
	/** A system or subsystem grouping units that provide a function */
	System = "system",
	/** An independently runnable unit within a system */
	Container = "container",
	/** An internal "module" within a runtime unit */
	Component = "component",
	/** A source artifact, such as a repository, package, or file */
	Code = "code",
	/** A runtime platform resource that hosts units */
	Infrastructure = "infrastructure",
	/** A participant shown as optional related context; examples include people and organizations. */
	Actor = "actor",
	/** A risk, requirement, or constraint that can annotate architectural subjects. */
	Concern = "concern",
	/** A recorded design or operational choice that can annotate related subjects. */
	Decision = "decision",
	/** An occurrence, such as an incident or deployment, that can provide annotation context. */
	Event = "event",
	/** An observation or notification, such as an alert, that can provide annotation context. */
	Signal = "signal",
	/** A workflow kept as related context while its architectural classification is unresolved. */
	Process = "process",
	/** An unsupported category that remains inspectable without an inferred level. */
	Unknown = "unknown",
}

const categories = new Set<string>(Object.values(MapCategory));

export enum DisplayMode {
	// Drawn as a node, or as a group boundary
	Node = 0,
	// Attached to a related subject without occupying a node position
	Annotation = 1,
	// Available in inspection; not drawn on the map
	DetailsOnly = 2,
}

export enum NodeDetailLevel {
	// Broad business and organizational functions.
	Landscape = 0,
	// Systems grouping runtime units.
	Systems = 1,
	// Runtime units and supporting infrastructure.
	Runtime = 2,
	// Internal components and source artifacts.
	Implementation = 3,
}

export type MapCategoryDisplay = {
	// How entities in this category appear on the map.
	mode: DisplayMode;
	// Human-readable category name, distinct from an entity's name or detail-level title.
	categoryLabel: string;
	// Usual reveal level for architecture; omitted for actors and other context.
	level?: NodeDetailLevel;
};

const defineCategoryDisplay = (
	mode: DisplayMode,
	categoryLabel: string,
	level?: NodeDetailLevel
): MapCategoryDisplay => ({ mode, categoryLabel, level });
const node = (name: string, level?: NodeDetailLevel) => defineCategoryDisplay(DisplayMode.Node, name, level);
const annotation = (name: string) => defineCategoryDisplay(DisplayMode.Annotation, name);
const detailsOnly = (name: string) => defineCategoryDisplay(DisplayMode.DetailsOnly, name);

export const categoryDisplay: Readonly<Record<MapCategory, MapCategoryDisplay>> = {
	[MapCategory.Function]: node("Function", NodeDetailLevel.Landscape),
	[MapCategory.System]: node("System", NodeDetailLevel.Systems),
	[MapCategory.Container]: node("Container", NodeDetailLevel.Runtime),
	[MapCategory.Infrastructure]: node("Infrastructure", NodeDetailLevel.Runtime),
	[MapCategory.Component]: node("Component", NodeDetailLevel.Implementation),
	[MapCategory.Code]: node("Source Artifact", NodeDetailLevel.Implementation),
	[MapCategory.Actor]: node("Actor"),
	[MapCategory.Concern]: annotation("Concern"),
	[MapCategory.Decision]: annotation("Decision"),
	[MapCategory.Event]: annotation("Event"),
	[MapCategory.Signal]: annotation("Signal"),
	[MapCategory.Process]: detailsOnly("Workflow"),
	[MapCategory.Unknown]: detailsOnly("Unsupported category"),
};

export const parseMapCategory = (value: string) =>
	categories.has(value) ? (value as MapCategory) : MapCategory.Unknown;

export const getMapCategoryDisplay = (category: string): MapCategoryDisplay =>
	categoryDisplay[parseMapCategory(category)];

export const isArchitectureCategory = (category: string): boolean => {
	const display = getMapCategoryDisplay(category);
	return display.mode === DisplayMode.Node && display.level !== undefined;
};

export const isActorCategory = (category: string): boolean => {
	const display = getMapCategoryDisplay(category);
	return display.mode === DisplayMode.Node && display.level === undefined;
};
