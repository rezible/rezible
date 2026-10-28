import { expect, test } from "bun:test";
import { fullFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import type { GraphSubset } from "../graph";
import { visibleConnections, type MapConnection } from "../rollup";
import { buildStructureTree } from "../structure";

const withRelationships = (...relationships: GraphSubset["relationships"]): GraphSubset => ({
	...fullFixture,
	relationships: [...fullFixture.relationships, ...relationships],
});

const pairs = (connections: readonly MapConnection[]) =>
	connections.map(({ source, target, relationshipIds }) => [source, target, relationshipIds.length]);

test("Focused with nothing selected draws the system-level roll-up at both bands", () => {
	const graph = withRelationships(
		{
			id: "payment-data",
			source: "rt-checkout-db",
			target: "rt-payment-worker",
			predicate: "reads_from",
		},
		{ id: "internal", source: "rt-payment-service", target: "rt-payment-worker", predicate: "calls" }
	);
	const tree = buildStructureTree(graph);
	for (const band of ["landscape", "runtime"] as const) {
		const view = visibleConnections(graph, tree, { band, mode: "focused", focus: { kind: "none" } });

		expect(view.level).toBe("system");
		expect(view.focused).toBe(false);
		expect(pairs(view.connections)).toEqual([
			["sys-checkout", "sys-payments", 2],
			["sys-checkout", "sys-search", 1],
			["sys-payments", "sys-checkout", 1],
		]);
		expect(view.connections[0].predicates).toEqual(["calls", "reads_from"]);
	}
});

test("Focused with a runtime selected draws only its runtime-level connections, placements dashed", () => {
	const tree = buildStructureTree(fullFixture);
	const view = visibleConnections(fullFixture, tree, {
		band: "runtime",
		mode: "focused",
		focus: { kind: "node", nodeId: "rt-checkout-api" },
	});

	expect(view.focused).toBe(true);
	expect(
		view.connections.map(({ source, target, predicates, placement }) => [
			source,
			target,
			predicates,
			placement,
		])
	).toEqual([
		["rt-checkout-api", "rt-payment-service", ["calls"], false],
		["rt-checkout-api", "rt-checkout-db", ["writes_to"], false],
		["rt-checkout-api", "rt-search-api", ["calls"], false],
		["rt-payment-worker", "rt-checkout-api", ["calls"], false],
		["rt-checkout-api", "rt-production-cluster", ["runs_on"], true],
		["rt-checkout-api", "rt-standby-cluster", ["runs_on"], true],
	]);
});

test("placements never roll up, never appear in All and are not drawn for their target", () => {
	const tree = buildStructureTree(fullFixture);
	const views = [
		...(["landscape", "runtime"] as const).flatMap((band) => [
			visibleConnections(fullFixture, tree, { band, mode: "all", focus: { kind: "none" } }),
			visibleConnections(fullFixture, tree, { band, mode: "focused", focus: { kind: "none" } }),
		]),
		visibleConnections(fullFixture, tree, {
			band: "runtime",
			mode: "focused",
			focus: { kind: "node", nodeId: "rt-production-cluster" },
		}),
	];

	for (const view of views) {
		expect(view.connections.some(({ predicates }) => predicates.includes("runs_on"))).toBe(false);
	}
});

test("an implementation endpoint follows its primary chain up to its runtime", () => {
	const graph = withRelationships({
		id: "code-call",
		source: "code-main",
		target: "rt-payment-worker",
		predicate: "calls",
	});
	const view = visibleConnections(graph, buildStructureTree(graph), {
		band: "runtime",
		mode: "all",
		focus: { kind: "none" },
	});

	const connection = view.connections.find(({ relationshipIds }) => relationshipIds.includes("code-call"));
	expect(connection).toMatchObject({ source: "rt-checkout-api", target: "rt-payment-worker" });
});

test("a selected region's connections include its descendants' at the band's level", () => {
	const graph = withRelationships(
		{ id: "code-payment", source: "code-main", target: "rt-payment-worker", predicate: "calls" },
		{ id: "system-search", source: "sys-payments", target: "sys-search", predicate: "depends_on" }
	);
	const tree = buildStructureTree(graph);
	for (const [band, level] of [
		["landscape", "system"],
		["runtime", "runtime"],
	] as const) {
		const view = visibleConnections(graph, tree, {
			band,
			mode: "focused",
			focus: { kind: "node", nodeId: "sys-payments" },
		});

		expect(view.level).toBe(level);
		expect(view.connections.flatMap(({ relationshipIds }) => relationshipIds).sort()).toEqual([
			"calls:rt-checkout-api:rt-payment-service",
			"calls:rt-payment-worker:rt-checkout-api",
			"code-payment",
			"system-search",
		]);
	}
});

test("a runtime with no system has no node at the system level", () => {
	const graph = withRelationships({
		id: "crawl",
		source: "rt-crawler",
		target: "rt-search-api",
		predicate: "calls",
	});
	const tree = buildStructureTree(graph);
	const resting = visibleConnections(graph, tree, {
		band: "landscape",
		mode: "all",
		focus: { kind: "none" },
	});
	const runtime = visibleConnections(graph, tree, {
		band: "runtime",
		mode: "all",
		focus: { kind: "none" },
	});

	expect(resting.connections.some(({ relationshipIds }) => relationshipIds.includes("crawl"))).toBe(false);
	expect(runtime.connections.some(({ relationshipIds }) => relationshipIds.includes("crawl"))).toBe(true);
});
