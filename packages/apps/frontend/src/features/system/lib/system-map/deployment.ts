import type { GraphEntity, GraphSubset } from "./graph";
import { entityLabel } from "./labels";
import { categoryLevel, CONTAINER_LEVEL, DEPLOYMENT_HIERARCHY, inHierarchy } from "./model";
import {
	breadcrumb,
	isPlacementUnknown,
	placementCountLine,
	placementSources,
	placementTargets,
} from "./placement";
import { compareIds } from "./structure";

/** One entity as the deployment view lists it. */
export type DeploymentEntry = {
	readonly entity: GraphEntity;
	readonly label: string;
	/** Its ancestors' names, outermost first: "Commerce › Checkout". */
	readonly breadcrumb: string;
};

export type DeploymentRow = DeploymentEntry & {
	/** The source's other most specific targets, by display name, then id. */
	readonly alsoOn: readonly DeploymentEntry[];
};

/** An infrastructure node and the loaded containers whose most specific placement it is. */
export type DeploymentSection = DeploymentEntry & {
	/** Depth in the infrastructure tree; roots are 0. */
	readonly depth: number;
	/** "3 entities run here", or "At least 3" on a partial graph. */
	readonly countLine: string;
	readonly rows: readonly DeploymentRow[];
};

export type DeploymentGroups = {
	/** The infrastructure tree, depth first, siblings by display name, then id. */
	readonly sections: readonly DeploymentSection[];
	/** Containers with no loaded placement. */
	readonly placementUnknown: readonly DeploymentEntry[];
};

const compareEntities = (a: GraphEntity, b: GraphEntity) =>
	compareIds(entityLabel(a), entityLabel(b)) || compareIds(a.id, b.id);

/** The deployment hierarchy as a tree from parentId, each node listing what runs directly on it. */
export function deploymentGroups(graph: GraphSubset): DeploymentGroups {
	const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));
	const targetsBySource = placementTargets(graph);
	const sourcesByTarget = placementSources(graph, targetsBySource);
	const entry = (entity: GraphEntity): DeploymentEntry => ({
		entity,
		label: entityLabel(entity),
		breadcrumb: breadcrumb(entities, entity.id),
	});

	const infrastructure = graph.entities.filter((entity) =>
		inHierarchy(graph.model, entity.category, DEPLOYMENT_HIERARCHY)
	);
	const children = new Map<string | undefined, GraphEntity[]>();
	for (const entity of infrastructure) {
		const parent = entity.parentId ? entities.get(entity.parentId) : undefined;
		// A parent outside the subset or outside the hierarchy makes a root.
		const key =
			parent && inHierarchy(graph.model, parent.category, DEPLOYMENT_HIERARCHY) ? parent.id : undefined;
		children.set(key, [...(children.get(key) ?? []), entity]);
	}

	const sections: DeploymentSection[] = [];
	// The backend keeps parents acyclic; nodes caught in a cycle written concurrently are not listed.
	const visit = (entity: GraphEntity, depth: number) => {
		const sources = (sourcesByTarget.get(entity.id) ?? []).filter(
			(source) => categoryLevel(graph.model, source.category) === CONTAINER_LEVEL
		);
		const rows = sources.map((source): DeploymentRow => {
			const others = (targetsBySource.get(source.id) ?? []).filter((other) => other.id !== entity.id);
			return { ...entry(source), alsoOn: others.map(entry) };
		});
		sections.push({ ...entry(entity), depth, countLine: placementCountLine(graph, rows.length), rows });
		[...(children.get(entity.id) ?? [])]
			.sort(compareEntities)
			.forEach((child) => visit(child, depth + 1));
	};
	[...(children.get(undefined) ?? [])].sort(compareEntities).forEach((root) => visit(root, 0));

	const placementUnknown = graph.entities
		.filter((entity) => isPlacementUnknown(graph, entity, targetsBySource.get(entity.id)))
		.sort(compareEntities)
		.map(entry);

	return { sections, placementUnknown };
}
