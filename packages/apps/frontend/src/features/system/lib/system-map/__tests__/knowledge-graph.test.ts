import { describe, expect, test } from "bun:test";

import {
	categoryDisplay,
	getMapCategoryDisplay,
	getArchitectureLayerLabel,
	getDetailLayerLabel,
	isArchitectureCategory,
	MapCategory,
	NodeDetailLevel,
	parseMapCategory,
} from "../category";
import type { GraphEntity, GraphRelationship, GraphSubset } from "../graph";
import { projectMap } from "../projection";
import type { MapConnection, MapSelection } from "../presentation";
import { reconcileSelection, selectedEntityIds, selectionIsAvailable } from "../selection";

const graph = (
	entities: readonly GraphEntity[],
	relationships: readonly GraphRelationship[],
	unresolvedRelationships: readonly GraphRelationship[] = [],
	stopReason: GraphSubset["enumeration"]["stopReason"] = "exhausted"
): GraphSubset => ({
	entities,
	relationships,
	unresolvedRelationships,
	enumeration: { scope: "fixture", stopReason },
});

const revealAll = (source: GraphSubset) => ({
	detail: NodeDetailLevel.Implementation,
	nearbyEntityIds: source.entities.map((entity) => entity.id),
});

const connectionFor = (
	projection: ReturnType<typeof projectMap>,
	relationshipId: string
): MapConnection | undefined =>
	projection.connections.find((connection) => connection.sourceRelationshipIds.includes(relationshipId));

describe("system map category policy", () => {
	test("uses only the named architectural categories and levels", () => {
		const levels = {
			[MapCategory.SystemFunction]: NodeDetailLevel.Landscape,
			[MapCategory.System]: NodeDetailLevel.Systems,
			[MapCategory.Container]: NodeDetailLevel.Runtime,
			[MapCategory.Infrastructure]: NodeDetailLevel.Runtime,
			[MapCategory.Component]: NodeDetailLevel.Implementation,
			[MapCategory.Code]: NodeDetailLevel.Implementation,
		};

		for (const [category, level] of Object.entries(levels)) {
			expect(getMapCategoryDisplay(category).level).toBe(level);
			expect(isArchitectureCategory(category)).toBe(true);
		}
		expect(getArchitectureLayerLabel(NodeDetailLevel.Landscape)).toBe("Landscape · Function");
		expect(getDetailLayerLabel(2.08)).toBe("Runtime");
	});

	test("retains contextual and unknown categories without assigning architecture levels", () => {
		for (const category of [
			MapCategory.Actor,
			MapCategory.Process,
			MapCategory.Concern,
			"future_category",
		]) {
			expect(isArchitectureCategory(category)).toBe(false);
			expect(getMapCategoryDisplay(category).level).toBeUndefined();
		}
		for (const category of Object.values(MapCategory)) {
			expect(parseMapCategory(category)).toBe(category);
			expect(getMapCategoryDisplay(category)).toBe(categoryDisplay[category]);
		}
		expect(parseMapCategory("future_category")).toBe(MapCategory.Unknown);
	});

	test("keeps incoming category nodes and incident edges through the fade interval", () => {
		const source = graph(
			[
				{ id: "landscape", category: "system_function", kind: "payments" },
				{ id: "system", category: "system", kind: "billing" },
			],
			[{ id: "relationship", source: "landscape", target: "system", predicate: "supports" }]
		);
		const atStart = projectMap(source, {
			detail: NodeDetailLevel.Landscape,
			visualDetail: 0.81,
			nearbyEntityIds: ["system"],
		});
		const fading = projectMap(source, {
			detail: NodeDetailLevel.Landscape,
			visualDetail: 0.9,
			nearbyEntityIds: ["system"],
		});
		const fullyProminent = projectMap(source, {
			detail: NodeDetailLevel.Systems,
			visualDetail: 1,
			nearbyEntityIds: ["system"],
		});
		const fadingSystem = fading.nodes.find(({ id }) => id === "system");
		const fadingEdge = fading.connections.find(({ sourceRelationshipIds }) =>
			sourceRelationshipIds.includes("relationship")
		);

		expect(atStart.nodes.map(({ id }) => id)).toContain("system");
		expect(fadingSystem?.opacity).toBeGreaterThan(0);
		expect(fadingSystem?.opacity).toBeLessThan(1);
		expect(fullyProminent.nodes.find(({ id }) => id === "system")?.opacity).toBe(1);
		expect(fadingEdge?.opacity).toBe(fadingSystem?.opacity);
	});

	test("subdues enclosing regions without reducing descendant opacity", () => {
		const source = graph(
			[
				{ id: "landscape", category: "system_function", kind: "payments" },
				{ id: "runtime", category: "container", kind: "billing service" },
			],
			[{ id: "membership", source: "landscape", target: "runtime", predicate: "contains" }]
		);
		const projection = projectMap(source, {
			detail: NodeDetailLevel.Implementation,
			visualDetail: NodeDetailLevel.Implementation,
			nearbyEntityIds: ["landscape", "runtime"],
		});

		expect(projection.nodes.find(({ id }) => id === "landscape")?.opacity).toBe(0.5);
		expect(projection.nodes.find(({ id }) => id === "runtime")?.opacity).toBe(1);
	});
});

describe("supplied membership projection", () => {
	test("encloses one observed parent in partial data without a completeness gate", () => {
		const source = graph(
			[
				{ id: "parent", category: "system", kind: "system" },
				{ id: "child", category: "container", kind: "service" },
			],
			[{ id: "m-parent-child", source: "parent", target: "child", predicate: "contains" }],
			[],
			"entity-limit"
		);
		const projection = projectMap(source, { detail: 2, nearbyEntityIds: ["child"] });

		expect(projection.nodes.map((node) => node.id)).toEqual(["child", "parent"]);
		expect(projection.nodes.find((node) => node.id === "child")?.enclosure).toEqual({
			parentId: "parent",
			membershipId: "m-parent-child",
		});
		expect(projection.nodes.find((node) => node.id === "parent")?.appearance).toBe("group");
		expect(projection.connections).toEqual([]);
	});

	test("keeps a child with multiple supplied parents shared and draws each membership once", () => {
		const source = graph(
			[
				{ id: "parent-a", category: "system", kind: "A" },
				{ id: "parent-b", category: "system", kind: "B" },
				{ id: "child", category: "container", kind: "shared" },
			],
			[
				{ id: "m-a", source: "parent-a", target: "child", predicate: "contains" },
				{ id: "m-b", source: "parent-b", target: "child", predicate: "contains" },
			]
		);
		const projection = projectMap(source, { detail: 2, nearbyEntityIds: ["child"] });

		expect(projection.nodes.filter((node) => node.id === "child")).toHaveLength(1);
		expect(projection.nodes.find((node) => node.id === "child")?.enclosure).toBeUndefined();
		expect(projection.representativeByEntityId.get("child")).toBe("child");
		expect(
			projection.connections.map((connection) => [
				connection.endpoints,
				connection.sourceRelationshipIds,
			])
		).toEqual([
			[["parent-a", "child"], ["m-a"]],
			[["parent-b", "child"], ["m-b"]],
		]);
	});

	test("an unresolved second parent prevents enclosure but never becomes a connection", () => {
		const source = graph(
			[
				{ id: "parent", category: "system", kind: "system" },
				{ id: "child", category: "container", kind: "service" },
			],
			[{ id: "m-present", source: "parent", target: "child", predicate: "contains" }],
			[{ id: "m-missing", source: "missing-parent", target: "child", predicate: "contains" }]
		);
		const projection = projectMap(source, { detail: 2, nearbyEntityIds: ["child"] });

		expect(projection.nodes.find((node) => node.id === "child")?.enclosure).toBeUndefined();
		expect(projection.connections.map((connection) => connection.sourceRelationshipIds)).toEqual([
			["m-present"],
		]);
		expect(
			projection.connections.some((connection) =>
				connection.sourceRelationshipIds.includes("m-missing")
			)
		).toBe(false);
	});

	test("does not alias a hidden shared child to its one visible parent", () => {
		const source = graph(
			[
				{ id: "parent", category: "system", kind: "system" },
				{ id: "child", category: "container", kind: "service" },
			],
			[{ id: "m-present", source: "parent", target: "child", predicate: "contains" }],
			[{ id: "m-missing", source: "missing-parent", target: "child", predicate: "contains" }]
		);
		const projection = projectMap(source, { detail: 1, nearbyEntityIds: [] });

		expect(projection.nodes.map((node) => node.id)).toEqual(["parent"]);
		expect(projection.representativeByEntityId.get("child")).toBeUndefined();
		expect(projection.connections).toEqual([]);
	});

	test("does not represent a hidden child inside a more detailed parent", () => {
		const source = graph(
			[
				{ id: "landscape", category: "system_function", kind: "landscape" },
				{ id: "component", category: "component", kind: "component" },
				{ id: "system", category: "system", kind: "system" },
			],
			[{ id: "m-invalid-level", source: "component", target: "system", predicate: "contains" }]
		);
		const projection = projectMap(source, { detail: 3, nearbyEntityIds: ["component"] });

		expect(projection.nodes.map((node) => node.id)).toEqual(["component", "landscape"]);
		expect(projection.representativeByEntityId.get("system")).toBeUndefined();
		expect(projection.nodes.find((node) => node.id === "component")?.appearance).toBe("compact");
	});

	test("does not map a hidden system to a function through a non-architectural actor", () => {
		const source = graph(
			[
				{ id: "function", category: "system_function", kind: "function" },
				{ id: "other-function", category: "system_function", kind: "other function" },
				{ id: "actor", category: "actor", kind: "person" },
				{ id: "system", category: "system", kind: "system" },
			],
			[
				{
					id: "m-function-actor",
					source: "function",
					target: "actor",
					predicate: "contains",
				},
				{ id: "m-actor-system", source: "actor", target: "system", predicate: "contains" },
				{
					id: "r-system-other-function",
					source: "system",
					target: "other-function",
					predicate: "calls",
				},
			]
		);
		const projection = projectMap(source, { detail: 0, nearbyEntityIds: [] });

		expect(projection.nodes.map((node) => node.id)).toEqual(["function", "other-function"]);
		expect(projection.representativeByEntityId.get("system")).toBeUndefined();
		expect(connectionFor(projection, "r-system-other-function")).toBeUndefined();
	});

	test("does not prune a visible group through a non-architectural actor path", () => {
		const source = graph(
			[
				{ id: "function", category: "system_function", kind: "function" },
				{ id: "system", category: "system", kind: "system" },
				{ id: "actor", category: "actor", kind: "person" },
				{ id: "shared", category: "container", kind: "shared" },
			],
			[
				{
					id: "m-function-actor",
					source: "function",
					target: "actor",
					predicate: "contains",
				},
				{ id: "m-actor-system", source: "actor", target: "system", predicate: "contains" },
				{ id: "m-function-shared", source: "function", target: "shared", predicate: "contains" },
				{ id: "m-system-shared", source: "system", target: "shared", predicate: "contains" },
			]
		);
		const projection = projectMap(source, { detail: 2, nearbyEntityIds: ["system"] });

		expect(projection.nodes.map((node) => node.id)).toContain("shared");
		expect(projection.representativeByEntityId.get("shared")).toBe("shared");
		expect(connectionFor(projection, "m-function-shared")?.classification).toBe("direct");
		expect(connectionFor(projection, "m-system-shared")?.classification).toBe("direct");
	});

	test("does not find a common representative through a non-architectural actor path", () => {
		const source = graph(
			[
				{ id: "function", category: "system_function", kind: "function" },
				{ id: "parent-a", category: "system", kind: "A" },
				{ id: "parent-b", category: "system", kind: "B" },
				{ id: "actor", category: "actor", kind: "person" },
				{ id: "shared", category: "container", kind: "shared" },
			],
			[
				{ id: "m-function-a", source: "function", target: "parent-a", predicate: "contains" },
				{
					id: "m-function-actor",
					source: "function",
					target: "actor",
					predicate: "contains",
				},
				{ id: "m-actor-b", source: "actor", target: "parent-b", predicate: "contains" },
				{ id: "m-a-shared", source: "parent-a", target: "shared", predicate: "contains" },
				{ id: "m-b-shared", source: "parent-b", target: "shared", predicate: "contains" },
			]
		);
		const projection = projectMap(source, { detail: 2, nearbyEntityIds: ["parent-a"] });

		expect(projection.nodes.map((node) => node.id)).not.toContain("shared");
		expect(projection.representativeByEntityId.get("shared")).toBeUndefined();
	});

	test("rejects cyclic enclosures, keeps cyclic facts as connections, and omits source self-loops", () => {
		const source = graph(
			[
				{ id: "a", category: "system", kind: "A" },
				{ id: "b", category: "system", kind: "B" },
			],
			[
				{ id: "m-a-b", source: "a", target: "b", predicate: "contains" },
				{ id: "m-b-a", source: "b", target: "a", predicate: "contains" },
				{ id: "r-self", source: "a", target: "a", predicate: "calls" },
			]
		);
		const projection = projectMap(source, { detail: 1, nearbyEntityIds: [] });

		expect(projection.nodes.every((node) => node.enclosure === undefined)).toBe(true);
		expect(projection.connections.map((connection) => connection.sourceRelationshipIds)).toEqual([
			["m-a-b"],
			["m-b-a"],
		]);
	});

	test("does not use a cyclic membership path as a hidden endpoint representative", () => {
		const source = graph(
			[
				{ id: "landscape", category: "system_function", kind: "landscape" },
				{ id: "system", category: "system", kind: "system" },
			],
			[
				{ id: "m-down", source: "landscape", target: "system", predicate: "contains" },
				{ id: "m-up", source: "system", target: "landscape", predicate: "contains" },
			]
		);
		const projection = projectMap(source, { detail: 0, nearbyEntityIds: [] });

		expect(projection.nodes.map((node) => node.id)).toEqual(["landscape"]);
		expect(projection.representativeByEntityId.get("system")).toBeUndefined();
	});

	test("maps collapsed descendants to stable summaries with direction and predicate separation", () => {
		const source = graph(
			[
				{ id: "group-a", category: "system", kind: "A" },
				{ id: "group-b", category: "system", kind: "B" },
				{ id: "member-a", category: "container", kind: "service A" },
				{ id: "member-a2", category: "container", kind: "worker A" },
				{ id: "member-b", category: "container", kind: "service B" },
			],
			[
				{ id: "m-a", source: "group-a", target: "member-a", predicate: "contains" },
				{ id: "m-a2", source: "group-a", target: "member-a2", predicate: "contains" },
				{ id: "m-b", source: "group-b", target: "member-b", predicate: "contains" },
				{ id: "r-one", source: "member-a", target: "member-b", predicate: "calls" },
				{ id: "r-two", source: "member-a2", target: "member-b", predicate: "calls" },
				{ id: "r-reverse", source: "member-b", target: "member-a", predicate: "calls" },
				{ id: "r-other", source: "member-a", target: "member-b", predicate: "owns" },
				{ id: "r-direct", source: "group-a", target: "group-b", predicate: "calls" },
			]
		);
		const projection = projectMap(source, { detail: 1, nearbyEntityIds: [] });
		const forward = connectionFor(projection, "r-one");
		const reverse = connectionFor(projection, "r-reverse");
		const owns = connectionFor(projection, "r-other");
		const direct = connectionFor(projection, "r-direct");

		expect(forward).toMatchObject({
			endpoints: ["group-a", "group-b"],
			predicate: "calls",
			classification: "summary",
			sourceRelationshipIds: ["r-one", "r-two"],
		});
		expect(reverse).toMatchObject({
			endpoints: ["group-b", "group-a"],
			predicate: "calls",
			classification: "summary",
		});
		expect(owns).toMatchObject({
			endpoints: ["group-a", "group-b"],
			predicate: "owns",
			classification: "summary",
		});
		expect(direct).toMatchObject({ endpoints: ["group-a", "group-b"], classification: "direct" });
		expect(forward?.id).toBe(
			connectionFor(
				projectMap(
					graph(
						source.entities,
						source.relationships.filter((relationship) => relationship.id !== "r-two")
					),
					{ detail: 1, nearbyEntityIds: [] }
				),
				"r-one"
			)?.id
		);
	});

	test("does not render actors, contextual records, unknown categories, or unresolved facts", () => {
		const source = graph(
			[
				{ id: "system", category: "system", kind: "orders" },
				{ id: "actor", category: "actor", kind: "person" },
				{ id: "event", category: "event", kind: "deployment" },
				{ id: "future", category: "future_category", kind: "future" },
			],
			[],
			[{ id: "unresolved", source: "system", target: "missing", predicate: "owns" }]
		);
		const projection = projectMap(source, revealAll(source));

		expect(projection.nodes.map((node) => node.id)).toEqual(["system"]);
		expect(projection.connections).toEqual([]);
		expect(source.entities).toHaveLength(4);
		expect(source.unresolvedRelationships).toHaveLength(1);
	});
});

describe("neutral host selection helpers", () => {
	test("availability includes unresolved source relationships", () => {
		const source = graph(
			[{ id: "present", category: "system", kind: "system" }],
			[],
			[{ id: "unresolved", source: "present", target: "missing", predicate: "contains" }]
		);
		const selection: MapSelection = { kind: "relationship", relationshipId: "unresolved" };

		expect(selectionIsAvailable(source, selection)).toBe(true);
		expect(selectedEntityIds(source, selection)).toEqual(["present"]);
		expect(reconcileSelection(source, selection)).toBe(selection);
	});

	test("summary reconciliation retains every original ID while any contributor remains", () => {
		const source = graph(
			[
				{ id: "a", category: "system", kind: "A" },
				{ id: "b", category: "system", kind: "B" },
			],
			[{ id: "still-present", source: "a", target: "b", predicate: "calls" }]
		);
		const selection: MapSelection = {
			kind: "summary",
			relationshipIds: ["still-present", "no-longer-supplied"],
		};

		expect(reconcileSelection(source, selection)).toEqual(selection);
		expect(reconcileSelection(graph(source.entities, []), selection)).toBeUndefined();
		expect(reconcileSelection(source, { kind: "entity", entityId: "missing" })).toBeUndefined();
	});
});
