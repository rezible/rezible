import type { GraphEntity, GraphSubset } from "./graph";
import { entityLabel } from "./labels";
import { isPlacementUnknown, PLACEMENT_PREDICATE, placementCountLine, placementTargets } from "./placement";
import { buildStructureTree, compareIds, type StructureTree } from "./structure";

/** One entity as the deployment view lists it. */
export type DeploymentEntry = {
	readonly entity: GraphEntity;
	readonly label: string;
	/** Display names of its structural ancestors, outermost first: ["Commerce", "Checkout"]. */
	readonly breadcrumb: readonly string[];
};

export type DeploymentRow = DeploymentEntry & {
	/** The source's other loaded targets, by display name, then id. */
	readonly alsoOn: readonly DeploymentEntry[];
};

/** A placement target and the loaded entities that run on it. */
export type DeploymentSection = DeploymentEntry & {
	/** "3 entities run here", or "At least 3" on a partial graph. */
	readonly countLine: string;
	readonly rows: readonly DeploymentRow[];
};

export type DeploymentGroups = {
	readonly sections: readonly DeploymentSection[];
	/** Containers with no loaded placement that are not targets themselves. */
	readonly placementUnknown: readonly DeploymentEntry[];
};

const compareEntities = (a: GraphEntity, b: GraphEntity) =>
	compareIds(entityLabel(a), entityLabel(b)) || compareIds(a.id, b.id);

function breadcrumb(tree: StructureTree, entities: ReadonlyMap<string, GraphEntity>, id: string): string[] {
	const names: string[] = [];
	let ancestorId = tree.primaryParents.get(id);
	while (ancestorId) {
		const ancestor = entities.get(ancestorId);
		if (!ancestor) {
			break;
		}
		names.unshift(entityLabel(ancestor));
		ancestorId = tree.primaryParents.get(ancestorId);
	}
	return names;
}

/** Groups the loaded entities by the direct target of their runs_on relationships. */
export function deploymentGroups(graph: GraphSubset): DeploymentGroups {
	const tree = buildStructureTree(graph);
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const targetsBySource = placementTargets(graph);
	const entry = (entity: GraphEntity): DeploymentEntry => ({
		entity,
		label: entityLabel(entity),
		breadcrumb: breadcrumb(tree, entities, entity.id),
	});

	const sourcesByTarget = new Map<string, Set<GraphEntity>>();
	for (const relationship of graph.relationships) {
		const source = entities.get(relationship.source);
		if (relationship.predicate !== PLACEMENT_PREDICATE || !source || !entities.has(relationship.target)) {
			continue;
		}

		let sources = sourcesByTarget.get(relationship.target);
		if (!sources) {
			sources = new Set();
			sourcesByTarget.set(relationship.target, sources);
		}
		sources.add(source);
	}

	const targets = [...sourcesByTarget.keys()].map((id) => entities.get(id)!).sort(compareEntities);
	const sections = targets.map((target): DeploymentSection => {
		const sources = [...sourcesByTarget.get(target.id)!].sort(compareEntities);
		const rows = sources.map((source): DeploymentRow => {
			const others = (targetsBySource.get(source.id) ?? []).filter((other) => other.id !== target.id);
			return { ...entry(source), alsoOn: others.map(entry) };
		});
		return { ...entry(target), countLine: placementCountLine(graph, target.id), rows };
	});

	const placementUnknown = graph.entities
		.filter((entity) => !sourcesByTarget.has(entity.id))
		.filter((entity) => isPlacementUnknown(entity, targetsBySource.get(entity.id)))
		.sort(compareEntities)
		.map(entry);

	return { sections, placementUnknown };
}
