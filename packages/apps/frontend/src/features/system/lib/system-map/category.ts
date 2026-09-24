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

export type MapCategoryDisplay = {
	// Human-readable category name, distinct from an entity's name or detail-level title.
	categoryLabel: string;
};

const architectureCategories = new Set<string>([
	MapCategory.SystemFunction,
	MapCategory.System,
	MapCategory.Container,
	MapCategory.Infrastructure,
	MapCategory.Component,
	MapCategory.Code,
]);

const categoryLabels: Readonly<Record<MapCategory, string>> = {
	[MapCategory.SystemFunction]: "Function",
	[MapCategory.System]: "System",
	[MapCategory.Container]: "Container",
	[MapCategory.Infrastructure]: "Infrastructure",
	[MapCategory.Component]: "Component",
	[MapCategory.Code]: "Code",
	[MapCategory.Actor]: "Actor",
	[MapCategory.Concern]: "Concern",
	[MapCategory.Decision]: "Decision",
	[MapCategory.Event]: "Event",
	[MapCategory.Signal]: "Signal",
	[MapCategory.Process]: "Workflow",
	[MapCategory.Unknown]: "Unsupported category",
};

export const parseMapCategory = (value: string) =>
	categories.has(value) ? (value as MapCategory) : MapCategory.Unknown;

export function getMapCategoryDisplay(category: string): MapCategoryDisplay {
	const parsed = parseMapCategory(category);
	return { categoryLabel: categoryLabels[parsed] };
}

export const isArchitectureCategory = (category: string): boolean => architectureCategories.has(category);
