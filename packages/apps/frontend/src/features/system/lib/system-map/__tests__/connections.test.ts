import { describe, expect, test } from "bun:test";
import type { GraphSubset } from "../graph";
import { getMapConnections } from "../connections";
import { graphToFlow, connectionsToFlow } from "$features/system/components/system-map/flow-graph-model";
import { applyInteractionState } from "$features/system/components/system-map/interaction-state";

const graph: GraphSubset = {
	entities: [
		{ id: "a", category: "system", kind: "A" },
		{ id: "b", category: "system", kind: "B" },
		{ id: "worker", category: "container", kind: "Worker" },
		{ id: "api", category: "container", kind: "API" },
	],
	relationships: [
		{ id: "owns-worker", source: "a", target: "worker", predicate: "contains" },
		{ id: "owns-api", source: "b", target: "api", predicate: "contains" },
		{ id: "request-1", source: "worker", target: "api", predicate: "calls" },
		{ id: "request-2", source: "worker", target: "api", predicate: "calls" },
		{ id: "response", source: "api", target: "worker", predicate: "calls" },
		{ id: "reads", source: "worker", target: "api", predicate: "reads" },
	],
	unresolvedRelationships: [],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
};

describe("map connections", () => {
	test("aggregates parallel records without mixing direction or predicates", () => {
		const connections = getMapConnections(graph);
		const calls = connections.filter((connection) => connection.predicate === "calls");
		expect(
			calls.map(({ source, target, relationshipIds }) => ({ source, target, relationshipIds }))
		).toEqual([
			{ source: "worker", target: "api", relationshipIds: ["request-1", "request-2"] },
			{ source: "api", target: "worker", relationshipIds: ["response"] },
		]);
		expect(connections.flatMap((connection) => connection.relationshipIds).sort()).toEqual(
			graph.relationships.map((relationship) => relationship.id).sort()
		);
	});

	test("edge hover reveals predicate/count and endpoints without changing selection or positions", () => {
		const connections = getMapConnections(graph);
		const nodes = graphToFlow(graph, []).nodes;
		const edges = connectionsToFlow(connections).filter(
			(edge) => edge.data!.connection.predicate === "calls"
		);
		const options = { graph, showLabels: false };
		const hovered = applyInteractionState(nodes, edges, { ...options, hoveredEdgeId: edges[0].id });
		expect(hovered.edges[0].data?.showLabel).toBe(true);
		expect(hovered.edges[0].ariaLabel).toBe("calls · 2 relationships");
		expect(hovered.nodes.filter((node) => node.data.highlighted).map((node) => node.id)).toEqual([
			"worker",
			"api",
		]);
		expect(hovered.edges.some((edge) => edge.selected)).toBe(false);
		expect(hovered.nodes.map((node) => node.position)).toEqual(nodes.map((node) => node.position));
		const cleared = applyInteractionState(hovered.nodes, hovered.edges, options);
		expect(cleared.edges[0].data?.showLabel).toBe(false);
		expect(cleared.nodes.some((node) => node.data.highlighted)).toBe(false);
	});
});
