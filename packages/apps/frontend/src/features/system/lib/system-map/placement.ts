import type { DetailBand } from "./detail";
import { loadedAncestors, type GraphEntity, type GraphSubset } from "./graph";
import { entityLabel } from "./labels";
import { categoryLevel, CONTAINER_LEVEL, isPlacement } from "./model";
import { compareIds, type StructureTree } from "./structure";

const compareEntities = (a: GraphEntity, b: GraphEntity) =>
	compareIds(entityLabel(a), entityLabel(b)) || compareIds(a.id, b.id);

/**
 * Each source's most specific loaded placement targets, by display name, then id. A target is dropped when
 * the source is also placed on one of its descendants, which implies it.
 */
export function placementTargets(graph: GraphSubset): ReadonlyMap<string, readonly GraphEntity[]> {
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const targets = new Map<string, Set<GraphEntity>>();
	for (const relationship of graph.relationships) {
		const target = entities.get(relationship.target);
		if (!target || !isPlacement(graph.model, relationship.predicate)) {
			continue;
		}
		let sourceTargets = targets.get(relationship.source);
		if (!sourceTargets) {
			sourceTargets = new Set();
			targets.set(relationship.source, sourceTargets);
		}
		sourceTargets.add(target);
	}

	const specific = new Map<string, GraphEntity[]>();
	for (const [source, sourceTargets] of targets) {
		const implied = new Set<string>();
		for (const target of sourceTargets) {
			loadedAncestors(entities, target.id).forEach((ancestor) => implied.add(ancestor.id));
		}
		specific.set(
			source,
			[...sourceTargets].filter((target) => !implied.has(target.id)).sort(compareEntities)
		);
	}
	return specific;
}

/** The sources placed on each target, by its most specific placements, by display name, then id. */
export function placementSources(
	graph: GraphSubset,
	targets: ReadonlyMap<string, readonly GraphEntity[]>
): ReadonlyMap<string, readonly GraphEntity[]> {
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const sources = new Map<string, GraphEntity[]>();
	for (const [sourceId, sourceTargets] of targets) {
		const source = entities.get(sourceId);
		if (!source) {
			continue;
		}
		for (const target of sourceTargets) {
			sources.set(target.id, [...(sources.get(target.id) ?? []), source]);
		}
	}
	for (const targetSources of sources.values()) {
		targetSources.sort(compareEntities);
	}
	return sources;
}

/** Ancestor names, outermost first, joined for display: "production › prod-use1". */
export function breadcrumb(entities: ReadonlyMap<string, GraphEntity>, id: string): string {
	return loadedAncestors(entities, id).reverse().map(entityLabel).join(" › ");
}

/** Cards show this many placement chips; further targets collapse into a "+N" chip. */
export const MAX_PLACEMENT_CHIPS = 2;

/** The targets a card shows as chips, and those collapsed into its "+N" chip. */
export function placementChips<T>(targets: readonly T[]): { shown: readonly T[]; collapsed: readonly T[] } {
	return { shown: targets.slice(0, MAX_PLACEMENT_CHIPS), collapsed: targets.slice(MAX_PLACEMENT_CHIPS) };
}

/** "N entities run here"; a partial enumeration may have missed some, so it reads "At least N". */
export function placementCountLine(graph: GraphSubset, count: number): string {
	let line = `${count} ${count === 1 ? "entity runs" : "entities run"} here`;
	if (graph.enumeration.stopReason !== "exhausted") {
		line = `At least ${line}`;
	}
	return line;
}

/** A container whose placement no loaded placement relationship states. */
export function isPlacementUnknown(
	graph: GraphSubset,
	entity: GraphEntity,
	targets: readonly GraphEntity[] | undefined
): boolean {
	return categoryLevel(graph.model, entity.category) === CONTAINER_LEVEL && !targets?.length;
}

/**
 * Drawn nodes that run directly on the target: its source cards at the runtime band, and the regions holding
 * those cards at the landscape band, where cards are hidden.
 */
export function placementFootprint(
	graph: GraphSubset,
	tree: StructureTree,
	targetId: string,
	band: DetailBand
): ReadonlySet<string> {
	const footprint = new Set<string>();
	for (const relationship of graph.relationships) {
		if (!isPlacement(graph.model, relationship.predicate) || relationship.target !== targetId) {
			continue;
		}
		if (tree.nodes.get(relationship.source)?.type !== "card") {
			continue;
		}

		if (band === "runtime") {
			footprint.add(relationship.source);
			continue;
		}
		const regionId = tree.enclosingRegions.get(relationship.source);
		if (regionId) {
			footprint.add(regionId);
		}
	}
	return footprint;
}
