import type { GraphSubset } from "../graph";
import type { LayoutResult } from "$features/system/components/system-map/flow-model";
import type { LayoutPositionInputs } from "$features/system/components/system-map/layout";
import { createSystemMapLayoutEngine } from "$features/system/components/system-map/layout-engine";
import { projectFullArchitecture } from "../projection";

export const layoutProjection = async (
	graph: GraphSubset,
	positionInputs?: LayoutPositionInputs
): Promise<LayoutResult> => {
	const engine = createSystemMapLayoutEngine();
	try {
		return await engine.layout(graph, projectFullArchitecture(graph), positionInputs);
	} finally {
		engine.dispose();
	}
};

const fixture = (
	entities: GraphSubset["entities"],
	relationships: GraphSubset["relationships"]
): GraphSubset => ({
	entities,
	relationships,
	unresolvedRelationships: [],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
});

export const sharedGroupsExample = {
	detail: 1,
	source: fixture(
		[
			{ id: "root", category: "system_function", kind: "order processing" },
			{ id: "group-a", category: "system", kind: "checkout" },
			{ id: "group-b", category: "system", kind: "billing" },
			{ id: "member-a", category: "container", kind: "checkout service" },
			{ id: "member-b", category: "container", kind: "billing service" },
			{ id: "shared-resource", category: "container", kind: "orders database" },
			{ id: "code", category: "code", kind: "checkout source" },
		],
		[
			{ id: "m-root-a", source: "root", target: "group-a", predicate: "contains" },
			{ id: "m-root-b", source: "root", target: "group-b", predicate: "contains" },
			{ id: "m-a", source: "group-a", target: "member-a", predicate: "contains" },
			{ id: "m-b", source: "group-b", target: "member-b", predicate: "contains" },
			{ id: "m-shared-a", source: "group-a", target: "shared-resource", predicate: "contains" },
			{ id: "m-shared-b", source: "group-b", target: "shared-resource", predicate: "contains" },
			{ id: "m-code", source: "member-a", target: "code", predicate: "contains" },
			{ id: "r-reads", source: "member-a", target: "shared-resource", predicate: "reads" },
		]
	),
};

export const relationshipExample = {
	detail: 1,
	source: fixture(
		[
			{ id: "group-a", category: "system", kind: "checkout" },
			{ id: "group-b", category: "system", kind: "billing" },
			{ id: "member-a", category: "container", kind: "checkout service" },
			{ id: "member-a2", category: "container", kind: "checkout worker" },
			{ id: "member-b", category: "container", kind: "billing service" },
		],
		[
			{ id: "m-a", source: "group-a", target: "member-a", predicate: "contains" },
			{ id: "m-a2", source: "group-a", target: "member-a2", predicate: "contains" },
			{ id: "m-b", source: "group-b", target: "member-b", predicate: "contains" },
			{ id: "r-direct", source: "group-a", target: "group-b", predicate: "calls" },
			{ id: "r-one", source: "member-a", target: "member-b", predicate: "calls" },
			{ id: "r-two", source: "member-a2", target: "member-b", predicate: "calls" },
			{ id: "r-reverse", source: "member-b", target: "member-a", predicate: "calls" },
			{ id: "r-other", source: "member-a", target: "member-b", predicate: "depends_on" },
		]
	),
};

export const edgeCasesExample = {
	detail: 3,
	source: {
		entities: [
			{ id: "parent-a", category: "system", kind: "first parent" },
			{ id: "parent-b", category: "system", kind: "second parent" },
			{ id: "child", category: "container", kind: "shared child" },
			{ id: "code", category: "code", kind: "source artifact" },
			{ id: "actor", category: "actor", kind: "person" },
			{ id: "context", category: "event", kind: "event record" },
			{ id: "unsupported", category: "future_category", kind: "future kind" },
		],
		relationships: [
			{ id: "m-a", source: "parent-a", target: "child", predicate: "contains" },
			{ id: "m-b", source: "parent-b", target: "child", predicate: "contains" },
			{ id: "m-child-code", source: "child", target: "code", predicate: "contains" },
			{ id: "m-code-child", source: "code", target: "child", predicate: "contains" },
			{ id: "r-self", source: "parent-a", target: "parent-a", predicate: "calls" },
			{ id: "r-actor", source: "actor", target: "child", predicate: "relates_to" },
		],
		unresolvedRelationships: [
			{ id: "m-missing-parent", source: "missing-parent", target: "child", predicate: "contains" },
			{ id: "r-missing-endpoint", source: "child", target: "missing-target", predicate: "relates_to" },
		],
		enumeration: { scope: "fixture", stopReason: "relationship-limit" },
	} satisfies GraphSubset,
};
