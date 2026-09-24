import { describe, expect, test } from "bun:test";
import { graphToFlow } from "$features/system/components/system-map/flow-graph-model";
import type { GraphSubset } from "../graph";

const graph: GraphSubset = {
	entities: [
		{ id: "checkout", category: "system", kind: "Checkout" },
		{ id: "billing", category: "system", kind: "Billing" },
		{ id: "database", category: "infrastructure", kind: "Orders" },
		{ id: "incident", category: "event", kind: "Incident" },
	],
	relationships: [
		{ id: "owns-a", source: "checkout", target: "database", predicate: "contains" },
		{ id: "owns-b", source: "billing", target: "database", predicate: "contains" },
		{ id: "reads", source: "checkout", target: "database", predicate: "reads" },
		{ id: "context", source: "incident", target: "checkout", predicate: "affects" },
	],
	unresolvedRelationships: [{ id: "missing", source: "database", target: "absent", predicate: "runs_on" }],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
};

describe("flat map model", () => {
	test("keeps shared entities once and preserves each resolved architectural relationship", () => {
		const model = graphToFlow(graph, []);
		expect(model.nodes.map((node) => node.id)).toEqual(["checkout", "billing", "database"]);
		expect(model.nodes.every((node) => !node.parentId)).toBe(true);
		expect(model.edges.flatMap((edge) => edge.data!.connection.relationshipIds)).toEqual([
			"owns-a",
			"owns-b",
			"reads",
		]);
	});

	test("graph refresh preserves live drag positions, even with older supplied coordinates", () => {
		const supplied = { checkout: { x: 15, y: 25 } };
		const before = graphToFlow(graph, [], supplied);
		before.nodes[0].position = { x: 420, y: 240 };
		const refreshed = graphToFlow(graph, before.nodes, supplied);
		expect(refreshed.nodes[0].position).toEqual({ x: 420, y: 240 });
	});

	test("places added entities outside existing nodes without moving survivors; prunes removed entities", () => {
		const before = graphToFlow(graph, []).nodes;
		const changed = {
			...graph,
			entities: [graph.entities[0], { id: "new", category: "container", kind: "Worker" }],
		};
		const model = graphToFlow(changed, before);
		expect(model.nodes.map((node) => node.id)).toEqual(["checkout", "new"]);
		expect(model.nodes[0].position).toEqual(before[0].position);
		expect(model.nodes[1].position.x).toBeGreaterThan(Math.max(...before.map((node) => node.position.x)));
		expect(model.edges).toEqual([]);
	});

	test("uses finite supplied positions exactly and ignores invalid coordinates", () => {
		const model = graphToFlow(graph, [], {
			checkout: { x: -123, y: 456 },
			billing: { x: Number.NaN, y: 0 },
		});
		expect(model.nodes[0].position).toEqual({ x: -123, y: 456 });
		expect(
			model.nodes.every((node) => Number.isFinite(node.position.x) && Number.isFinite(node.position.y))
		).toBe(true);
	});
});
