export enum MapCategory {
	/** A business/organisational ability */
	SystemFunction = "system_function",
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

const layerLabels: Readonly<Record<NodeDetailLevel, string>> = {
	[NodeDetailLevel.Landscape]: "Landscape · Function",
	[NodeDetailLevel.Systems]: "System",
	[NodeDetailLevel.Runtime]: "Runtime",
	[NodeDetailLevel.Implementation]: "Implementation",
};

/** Human-readable architectural layer names, independent of containment depth. */
export const getArchitectureLayerLabel = (level: NodeDetailLevel): string => layerLabels[level];

/** Names the layer represented by a continuous detail value. */
export const getDetailLayerLabel = (detail: number): string => {
	const level = Math.max(
		NodeDetailLevel.Landscape,
		Math.min(NodeDetailLevel.Implementation, Math.floor(Number.isFinite(detail) ? detail : 0))
	);
	return layerLabels[level as NodeDetailLevel];
};

export type MapCategoryDisplay = {
	// Human-readable category name, distinct from an entity's name or detail-level title.
	categoryLabel: string;
	// Architectural categories participate in map projection; all other categories remain source data.
	level?: NodeDetailLevel;
};

const category = (categoryLabel: string, level?: NodeDetailLevel): MapCategoryDisplay => ({
	categoryLabel,
	level,
});

export const categoryDisplay: Readonly<Record<MapCategory, MapCategoryDisplay>> = {
	[MapCategory.SystemFunction]: category("Function", NodeDetailLevel.Landscape),
	[MapCategory.System]: category("System", NodeDetailLevel.Systems),
	[MapCategory.Container]: category("Container", NodeDetailLevel.Runtime),
	[MapCategory.Infrastructure]: category("Infrastructure", NodeDetailLevel.Runtime),
	[MapCategory.Component]: category("Component", NodeDetailLevel.Implementation),
	[MapCategory.Code]: category("Code", NodeDetailLevel.Implementation),
	[MapCategory.Actor]: category("Actor"),
	[MapCategory.Concern]: category("Concern"),
	[MapCategory.Decision]: category("Decision"),
	[MapCategory.Event]: category("Event"),
	[MapCategory.Signal]: category("Signal"),
	[MapCategory.Process]: category("Workflow"),
	[MapCategory.Unknown]: category("Unsupported category"),
};

export const parseMapCategory = (value: string) =>
	categories.has(value) ? (value as MapCategory) : MapCategory.Unknown;

export const getMapCategoryDisplay = (category: string): MapCategoryDisplay =>
	categoryDisplay[parseMapCategory(category)];

export const isArchitectureCategory = (category: string): boolean => {
	return getMapCategoryDisplay(category).level !== undefined;
};
