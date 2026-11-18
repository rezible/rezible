import { describe, expect, test } from "bun:test";
import { fullFixture as fixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import { connectionAnchors } from "$features/system/lib/system-map/geometry";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import { withTestModel } from "$features/system/lib/system-map/__tests__/test-model";
import { applyMapState, connectionLabel, projectMap, type CardFlowNode, type FlowNode } from "../projection";

const fullFixture = withTestModel(fixture);

const node = (nodes: FlowNode[], id: string) => nodes.find((candidate) => candidate.id === id)!;

const box = (flowNode: FlowNode) => ({
	x: flowNode.position.x,
	y: flowNode.position.y,
	width: flowNode.width!,
	height: flowNode.height!,
});

describe("system map projection", () => {
	test("layers landscapes, systems, connections and cards in that order", () => {
		const projection = projectMap(fullFixture);
		const { nodes, edges } = applyMapState(projection, {
			graph: fullFixture,
			band: "runtime",
			relationshipMode: "all",
		});

		expect(node(nodes, "fn-commerce").zIndex).toBe(0);
		expect(node(nodes, "sys-commerce-core").zIndex).toBe(1);
		expect(node(nodes, "rt-checkout").zIndex).toBe(3);
		expect(edges.length).toBe(3);
		expect(edges.every((edge) => edge.zIndex === 2)).toBe(true);
	});

	test("cards read their parent container, other claims and placement ancestry", () => {
		const { nodes } = projectMap(fullFixture);
		const card = (id: string) => (node(nodes, id) as CardFlowNode).data;

		expect(card("rt-checkout-proxy").partOfLabel).toBe("Part of checkout");
		expect(card("rt-search-api").alsoClaimedLabel).toBe("Also claimed by Commerce Core");
		expect(card("rt-checkout").placement.shown.map(({ breadcrumb }) => breadcrumb)).toEqual([
			"commerce-prod › prod-use1",
			"commerce-prod › prod-euw1",
		]);
	});

	test("the landscape band hides cards except a selected one", () => {
		const { nodes } = applyMapState(projectMap(fullFixture), {
			graph: fullFixture,
			band: "landscape",
			relationshipMode: "focused",
			selection: { kind: "entity", entityId: "rt-checkout" },
		});
		const cards = nodes.filter((flowNode) => flowNode.type === "system-map-card");

		expect(cards.filter((card) => !card.hidden).map((card) => card.id)).toEqual(["rt-checkout"]);
		expect(nodes.some((flowNode) => flowNode.type === "system-map-region" && flowNode.hidden)).toBe(
			false
		);
	});

	test("separates only connections that have a counterpart in the opposite direction", () => {
		const graph: GraphSubset = {
			...fullFixture,
			relationships: [
				...fullFixture.relationships,
				{
					id: "reply",
					source: "rt-payment-service",
					target: "rt-checkout",
					predicate: "calls",
				},
			],
		};
		const { nodes, edges } = applyMapState(projectMap(graph), {
			graph,
			band: "runtime",
			relationshipMode: "all",
		});
		const anchorsOf = (source: string, target: string) =>
			edges.find((edge) => edge.source === source && edge.target === target)!.data!.anchors;
		const api = box(node(nodes, "rt-checkout"));
		const payments = box(node(nodes, "rt-payment-service"));
		const search = box(node(nodes, "rt-search-api"));

		expect(anchorsOf("rt-checkout", "rt-payment-service")).toEqual(
			connectionAnchors(api, payments, true)
		);
		expect(anchorsOf("rt-payment-service", "rt-checkout")).toEqual(
			connectionAnchors(payments, api, true)
		);
		expect(anchorsOf("rt-checkout", "rt-search-api")).toEqual(connectionAnchors(api, search, false));
	});

	test("system-level connections count relationships in the singular and plural", () => {
		const connection = {
			id: "c",
			source: "a",
			target: "b",
			predicates: ["calls"],
			relationshipIds: ["r1"],
		};

		expect(connectionLabel(connection, "system")).toBe("1 relationship");
		expect(connectionLabel({ ...connection, relationshipIds: ["r1", "r2"] }, "system")).toBe(
			"2 relationships"
		);
		expect(connectionLabel({ ...connection, relationshipIds: ["r1", "r2"] }, "runtime")).toBe(
			"calls × 2"
		);
	});

	test("a selected canonical relationship stays selected when the band changes its group", () => {
		const projection = projectMap(fullFixture);
		const selection = {
			kind: "relationship" as const,
			relationshipId: "calls:rt-checkout:rt-payment-service",
		};
		for (const band of ["landscape", "runtime"] as const) {
			for (const relationshipMode of ["focused", "all"] as const) {
				const { edges } = applyMapState(projection, {
					graph: fullFixture,
					band,
					relationshipMode,
					selection,
				});
				const selected = edges.filter((edge) => edge.selected);

				expect(selected).toHaveLength(1);
				expect(selected[0].data!.connection.relationshipIds).toContain(selection.relationshipId);
			}
		}
	});

	test("in All, selecting a node emphasizes its connections and dims the rest", () => {
		const { nodes, edges } = applyMapState(projectMap(fullFixture), {
			graph: fullFixture,
			band: "runtime",
			relationshipMode: "all",
			selection: { kind: "entity", entityId: "rt-payment-worker" },
		});

		expect(nodes.filter((flowNode) => flowNode.selected).map((flowNode) => flowNode.id)).toEqual([
			"rt-payment-worker",
		]);
		const emphasized = edges.filter((edge) => edge.data!.emphasized);
		expect(emphasized.map((edge) => [edge.source, edge.target])).toEqual([
			["rt-payment-worker", "rt-checkout"],
		]);
		expect(edges.filter((edge) => edge.data!.dimmed).length).toBe(2);
	});

	test("Focused at rest labels a connection only while it is hovered", () => {
		const projection = projectMap(fullFixture);
		const state = { graph: fullFixture, band: "runtime" as const, relationshipMode: "focused" as const };
		const resting = applyMapState(projection, state);
		const hoveredEdgeId = resting.edges[0].id;
		const hovered = applyMapState(projection, { ...state, hoveredEdgeId });

		expect(resting.edges.some((edge) => edge.data!.showLabel)).toBe(false);
		expect(hovered.edges.filter((edge) => edge.data!.showLabel).map((edge) => edge.id)).toEqual([
			hoveredEdgeId,
		]);
	});
});
