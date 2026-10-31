import { loadedAncestors, type GraphEntity, type GraphSubset } from "./graph";
import { entityLabel } from "./labels";
import {
	categoryLevel,
	COMPONENT_LEVEL,
	CONTAINER_LEVEL,
	LANDSCAPE_LEVEL,
	LOGICAL_HIERARCHY,
	inHierarchy,
	SYSTEM_LEVEL,
} from "./model";

export type RegionLevel = "landscape" | "system";

/** A system function or system region. Regions nest to any depth, and may hold regions and cards together. */
export type StructureRegion = {
	readonly type: "region";
	readonly id: string;
	readonly level: RegionLevel;
	readonly entity: GraphEntity;
	readonly otherParentIds: readonly string[];
	readonly children: readonly StructureNode[];
};

/** A container, with the components below it folded into a count. */
export type StructureCard = {
	readonly type: "card";
	readonly id: string;
	readonly entity: GraphEntity;
	readonly otherParentIds: readonly string[];
	/** The loaded container this one is part of; the card still sits in its nearest region. */
	readonly partOf?: GraphEntity;
	readonly implementationCount: number;
};

export type StructureNode = StructureRegion | StructureCard;

export type StructureTree = {
	/** Top-level nodes in packing order: landscapes, then systems and cards with no drawn region. */
	readonly roots: readonly StructureNode[];
	/** Every drawn node by id. */
	readonly nodes: ReadonlyMap<string, StructureNode>;
	/** The drawn node each logical entity resolves to: itself when drawn, else its nearest drawn ancestor. */
	readonly drawnNodes: ReadonlyMap<string, string>;
	/** The region a drawn node sits in directly. */
	readonly enclosingRegions: ReadonlyMap<string, string>;
};

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

/** Builds the logical tree from each entity's parentId. A parent outside the subset makes a root. */
export function buildStructureTree(graph: GraphSubset): StructureTree {
	const { model } = graph;
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const level = (entity: GraphEntity) =>
		inHierarchy(model, entity.category, LOGICAL_HIERARCHY)
			? categoryLevel(model, entity.category)
			: undefined;
	const isRegion = (entity: GraphEntity) => {
		const entityLevel = level(entity);
		return entityLevel === LANDSCAPE_LEVEL || entityLevel === SYSTEM_LEVEL;
	};
	const isDrawn = (entity: GraphEntity) => isRegion(entity) || level(entity) === CONTAINER_LEVEL;

	const ancestors = new Map<string, GraphEntity[]>();
	const drawnNodes = new Map<string, string>();
	const implementationCounts = new Map<string, number>();
	for (const entity of graph.entities) {
		if (level(entity) === undefined) {
			continue;
		}
		const chain = loadedAncestors(entities, entity.id);
		ancestors.set(entity.id, chain);
		const drawn = isDrawn(entity) ? entity : chain.find(isDrawn);
		if (drawn) {
			drawnNodes.set(entity.id, drawn.id);
		}
		if (level(entity) === COMPONENT_LEVEL && drawn && level(drawn) === CONTAINER_LEVEL) {
			implementationCounts.set(drawn.id, (implementationCounts.get(drawn.id) ?? 0) + 1);
		}
	}

	const regions = new Map<string, MutableRegion>();
	const pending: { node: StructureNode; regionId?: string }[] = [];
	for (const entity of graph.entities) {
		if (!isDrawn(entity)) {
			continue;
		}
		const chain = ancestors.get(entity.id)!;
		const regionId = chain.find(isRegion)?.id;
		const otherParentIds = entity.otherParentIds ?? [];
		if (isRegion(entity)) {
			const region: MutableRegion = {
				type: "region",
				id: entity.id,
				level: level(entity) === LANDSCAPE_LEVEL ? "landscape" : "system",
				entity,
				otherParentIds,
				children: [],
			};
			regions.set(entity.id, region);
			pending.push({ node: region, regionId });
			continue;
		}
		const parent = chain[0];
		pending.push({
			node: {
				type: "card",
				id: entity.id,
				entity,
				otherParentIds,
				partOf: parent && level(parent) === CONTAINER_LEVEL ? parent : undefined,
				implementationCount: implementationCounts.get(entity.id) ?? 0,
			},
			regionId,
		});
	}

	// The backend keeps parents acyclic; regions caught in a cycle written concurrently are not drawn.
	const roots: StructureNode[] = [];
	const enclosingRegions = new Map<string, string>();
	for (const { node, regionId } of pending) {
		if (regionId) {
			regions.get(regionId)!.children.push(node);
			enclosingRegions.set(node.id, regionId);
		} else {
			roots.push(node);
		}
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

	return { roots: sorted, nodes, drawnNodes, enclosingRegions };
}
