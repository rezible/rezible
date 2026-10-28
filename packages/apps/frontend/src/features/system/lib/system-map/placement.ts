import { MapCategory } from "./category";
import type { DetailBand } from "./detail";
import type { GraphEntity, GraphSubset } from "./graph";
import { entityLabel } from "./labels";
import { compareIds, type StructureTree } from "./structure";

/** Says where a workload runs, not what it talks to; drawn as a placement rather than a connection. */
export const PLACEMENT_PREDICATE = "runs_on";

/** Loaded targets of each source's placements, by display name, then id. */
export function placementTargets(graph: GraphSubset): ReadonlyMap<string, readonly GraphEntity[]> {
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const targets = new Map<string, GraphEntity[]>();
	for (const relationship of graph.relationships) {
		const target = entities.get(relationship.target);
		if (relationship.predicate !== PLACEMENT_PREDICATE || !target) {
			continue;
		}

		let sourceTargets = targets.get(relationship.source);
		if (!sourceTargets) {
			sourceTargets = [];
			targets.set(relationship.source, sourceTargets);
		}
		if (!sourceTargets.includes(target)) {
			sourceTargets.push(target);
		}
	}

	for (const sourceTargets of targets.values()) {
		sourceTargets.sort((a, b) => compareIds(entityLabel(a), entityLabel(b)) || compareIds(a.id, b.id));
	}
	return targets;
}

/** Cards show this many placement chips; further targets collapse into a "+N" chip. */
export const MAX_PLACEMENT_CHIPS = 2;

/** The targets a card shows as chips, and those collapsed into its "+N" chip. */
export function placementChips<T>(targets: readonly T[]): { shown: readonly T[]; collapsed: readonly T[] } {
	return { shown: targets.slice(0, MAX_PLACEMENT_CHIPS), collapsed: targets.slice(MAX_PLACEMENT_CHIPS) };
}

/** Loaded entities that run on the target. */
export function placementSourceCount(graph: GraphSubset, targetId: string): number {
	const sources = new Set<string>();
	for (const relationship of graph.relationships) {
		if (relationship.predicate === PLACEMENT_PREDICATE && relationship.target === targetId) {
			sources.add(relationship.source);
		}
	}
	return sources.size;
}

/** "N entities run here"; a partial enumeration may have missed some, so it reads "At least N". */
export function placementCountLine(graph: GraphSubset, targetId: string): string {
	const count = placementSourceCount(graph, targetId);
	let line = `${count} ${count === 1 ? "entity runs" : "entities run"} here`;
	if (graph.enumeration.stopReason !== "exhausted") {
		line = `At least ${line}`;
	}
	return line;
}

/** A container whose placement no loaded runs_on states. */
export function isPlacementUnknown(
	entity: GraphEntity,
	targets: readonly GraphEntity[] | undefined
): boolean {
	return entity.category === MapCategory.Container && !targets?.length;
}

/**
 * Drawn nodes that run on the target: its source cards at the runtime band, and the systems holding those
 * cards at the landscape band, where cards are hidden.
 */
export function placementFootprint(
	graph: GraphSubset,
	tree: StructureTree,
	targetId: string,
	band: DetailBand
): ReadonlySet<string> {
	const footprint = new Set<string>();
	for (const relationship of graph.relationships) {
		if (relationship.predicate !== PLACEMENT_PREDICATE || relationship.target !== targetId) {
			continue;
		}
		if (tree.nodes.get(relationship.source)?.type !== "card") {
			continue;
		}

		if (band === "runtime") {
			footprint.add(relationship.source);
			continue;
		}

		const parentId = tree.primaryParents.get(relationship.source);
		const parent = parentId ? tree.nodes.get(parentId) : undefined;
		if (parent?.type === "region" && parent.level === "system") {
			footprint.add(parent.id);
		}
	}
	return footprint;
}
