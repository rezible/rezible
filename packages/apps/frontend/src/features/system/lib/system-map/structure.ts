import { MapCategory } from "./category";
import type { GraphEntity, GraphSubset } from "./graph";
import { entityLabel } from "./labels";

export type StructureLevel = "landscape" | "system" | "runtime" | "implementation";

export type RegionLevel = "landscape" | "system";

/** A landscape or system region. Its children may mix systems and cards when a level is missing. */
export type StructureRegion = {
	readonly type: "region";
	readonly id: string;
	readonly level: RegionLevel;
	readonly entity: GraphEntity;
	/** Structural parents other than the primary one, by id. */
	readonly alsoIn: readonly string[];
	readonly children: readonly StructureNode[];
};

/** A runtime entity, with its implementation subtree folded into a count. */
export type StructureCard = {
	readonly type: "card";
	readonly id: string;
	readonly entity: GraphEntity;
	readonly alsoIn: readonly string[];
	readonly implementationCount: number;
};

export type StructureNode = StructureRegion | StructureCard;

export type StructureTree = {
	/** Top-level nodes in packing order: landscapes, then systems and cards with no drawn parent. */
	readonly roots: readonly StructureNode[];
	/** Every drawn node by id. */
	readonly nodes: ReadonlyMap<string, StructureNode>;
	/** The primary structural parent of every map entity that has one. */
	readonly primaryParents: ReadonlyMap<string, string>;
};

const categoryLevels: Readonly<Record<string, StructureLevel>> = {
	[MapCategory.SystemFunction]: "landscape",
	[MapCategory.System]: "system",
	[MapCategory.Container]: "runtime",
	[MapCategory.Infrastructure]: "runtime",
	[MapCategory.Component]: "implementation",
	[MapCategory.Code]: "implementation",
};

/**
 * Containment depth by category. Components rank above code so component → code chains carry implementation
 * entities up to their runtime; the structure levels themselves stay as categoryLevels defines them.
 */
const categoryDepths: Readonly<Record<string, number>> = {
	[MapCategory.SystemFunction]: 0,
	[MapCategory.System]: 1,
	[MapCategory.Container]: 2,
	[MapCategory.Infrastructure]: 2,
	[MapCategory.Component]: 3,
	[MapCategory.Code]: 4,
};

/** The structure level of a category, or undefined when the category is not a map entity. */
export const structureLevel = (category: string): StructureLevel | undefined => categoryLevels[category];

/** Plain code-point comparison, the repository tie-breaker. */
export const compareIds = (a: string, b: string): number => {
	if (a < b) {
		return -1;
	}
	if (a > b) {
		return 1;
	}
	return 0;
};

/** Landscapes, then systems, then cards. */
function siblingRank(node: StructureNode): number {
	if (node.type === "card") {
		return 2;
	}
	if (node.level === "system") {
		return 1;
	}
	return 0;
}

/** Regions before cards, each by display name, then id. */
function compareSiblings(a: StructureNode, b: StructureNode): number {
	return (
		siblingRank(a) - siblingRank(b) ||
		compareIds(entityLabel(a.entity), entityLabel(b.entity)) ||
		compareIds(a.id, b.id)
	);
}

/** Candidate parents per map entity: loaded contains sources at a shallower depth. */
function collectParentCandidates(graph: GraphSubset, depths: ReadonlyMap<string, number>) {
	const candidates = new Map<string, Set<string>>();
	for (const relationship of graph.relationships) {
		if (relationship.predicate !== "contains") {
			continue;
		}
		const sourceDepth = depths.get(relationship.source);
		const targetDepth = depths.get(relationship.target);
		if (sourceDepth === undefined || targetDepth === undefined || sourceDepth >= targetDepth) {
			continue;
		}

		let parents = candidates.get(relationship.target);
		if (!parents) {
			parents = new Set();
			candidates.set(relationship.target, parents);
		}
		parents.add(relationship.source);
	}
	return candidates;
}

/** Nearest depth first, then lowest id. */
function rankParents(parents: Iterable<string>, depths: ReadonlyMap<string, number>): string[] {
	return [...parents].sort((a, b) => depths.get(b)! - depths.get(a)! || compareIds(a, b));
}

type MutableRegion = Omit<StructureRegion, "children"> & { children: StructureNode[] };

function sortRegion(region: MutableRegion): StructureRegion {
	const children = region.children.map((child) => {
		if (child.type === "region") {
			return sortRegion(child as MutableRegion);
		}
		return child;
	});
	children.sort(compareSiblings);
	return { ...region, children };
}

/** Builds the containment tree of the map entities in a graph. */
export function buildStructureTree(graph: GraphSubset): StructureTree {
	const entities = new Map<string, GraphEntity>();
	const depths = new Map<string, number>();
	for (const entity of graph.entities) {
		const depth = categoryDepths[entity.category];
		if (depth !== undefined) {
			entities.set(entity.id, entity);
			depths.set(entity.id, depth);
		}
	}

	const primaryParents = new Map<string, string>();
	const alsoIn = new Map<string, string[]>();
	for (const [id, parents] of collectParentCandidates(graph, depths)) {
		const [primary, ...others] = rankParents(parents, depths);
		primaryParents.set(id, primary);
		alsoIn.set(id, others.sort(compareIds));
	}

	const implementationCounts = new Map<string, number>();
	for (const [id, entity] of entities) {
		if (structureLevel(entity.category) !== "implementation") {
			continue;
		}
		let ancestor = primaryParents.get(id);
		while (ancestor && structureLevel(entities.get(ancestor)!.category) === "implementation") {
			ancestor = primaryParents.get(ancestor);
		}
		if (ancestor && structureLevel(entities.get(ancestor)!.category) === "runtime") {
			implementationCounts.set(ancestor, (implementationCounts.get(ancestor) ?? 0) + 1);
		}
	}

	const regions = new Map<string, MutableRegion>();
	const roots: StructureNode[] = [];

	/** Places a node directly in its primary parent when that parent is a drawn region, else at the root. */
	const place = (node: StructureNode) => {
		const parentId = primaryParents.get(node.id);
		const parent = parentId ? regions.get(parentId) : undefined;
		if (parent) {
			parent.children.push(node);
		} else {
			roots.push(node);
		}
	};

	const addRegion = (id: string, level: RegionLevel, entity: GraphEntity) => {
		const region: MutableRegion = {
			type: "region",
			id,
			level,
			entity,
			alsoIn: alsoIn.get(id) ?? [],
			children: [],
		};
		regions.set(id, region);
		place(region);
	};

	// Each level is placed after the level above it exists.
	for (const level of ["landscape", "system"] as const) {
		for (const [id, entity] of entities) {
			if (structureLevel(entity.category) === level) {
				addRegion(id, level, entity);
			}
		}
	}

	for (const [id, entity] of entities) {
		if (structureLevel(entity.category) !== "runtime") {
			continue;
		}
		place({
			type: "card",
			id,
			entity,
			alsoIn: alsoIn.get(id) ?? [],
			implementationCount: implementationCounts.get(id) ?? 0,
		});
	}

	const sorted = roots.map((node) => (node.type === "region" ? sortRegion(node as MutableRegion) : node));
	sorted.sort(compareSiblings);

	const nodes = new Map<string, StructureNode>();
	const index = (node: StructureNode) => {
		nodes.set(node.id, node);
		if (node.type === "region") {
			node.children.forEach(index);
		}
	};
	sorted.forEach(index);

	return { roots: sorted, nodes, primaryParents };
}
