import { expect, test } from "bun:test";
import { fullFixture as fixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import type { GraphSubset } from "../graph";
import { visibleConnections, type MapConnection } from "../rollup";
import { buildStructureTree } from "../structure";
import { withTestModel } from "./test-model";

const fullFixture = withTestModel(fixture);

const withRelationships = (...relationships: GraphSubset["relationships"]): GraphSubset => ({
	...fullFixture,
	relationships: [...fullFixture.relationships, ...relationships],
});

const pairs = (connections: readonly MapConnection[]) =>
	connections.map(({ source, target, relationshipIds }) => [source, target, relationshipIds.length]);

test("connections roll up to the band's level: nearest regions, or container cards", () => {
	const graph = withRelationships({
		id: "payments-search",
		source: "sys-payments",
		target: "sys-search",
		predicate: "depends_on",
	});
	const tree = buildStructureTree(graph);
	const atRest = (band: "landscape" | "runtime") =>
		visibleConnections(graph, tree, { band, mode: "focused", focus: { kind: "none" } });

	// The component's call rolls up with checkout's; placements, ownership, code and the infrastructure write
	// are not drawn.
	expect(atRest("landscape").level).toBe("system");
	expect(pairs(atRest("landscape").connections)).toEqual([
		["sys-commerce-core", "sys-payments", 2],
		["sys-commerce-core", "sys-search", 1],
		["sys-payments", "sys-commerce-core", 1],
		["sys-payments", "sys-search", 1],
	]);
	// A system's own relationship has no card to attach to.
	expect(atRest("runtime").level).toBe("runtime");
	expect(pairs(atRest("runtime").connections)).toEqual([
		["rt-checkout", "rt-payment-service", 2],
		["rt-checkout", "rt-search-api", 1],
		["rt-payment-worker", "rt-checkout", 1],
	]);
});

test("a selected container draws its runtime-level connections, its components' included", () => {
	const tree = buildStructureTree(fullFixture);
	const view = visibleConnections(fullFixture, tree, {
		band: "runtime",
		mode: "focused",
		focus: { kind: "node", nodeId: "rt-checkout" },
	});

	expect(view.focused).toBe(true);
	expect(pairs(view.connections)).toEqual([
		["rt-checkout", "rt-payment-service", 2],
		["rt-checkout", "rt-search-api", 1],
		["rt-payment-worker", "rt-checkout", 1],
	]);
});

test("a container with no region has no node at the system level", () => {
	const graph = withRelationships({
		id: "cron-search",
		source: "rt-legacy-cron",
		target: "rt-search-api",
		predicate: "calls",
	});
	const tree = buildStructureTree(graph);
	const drawn = (band: "landscape" | "runtime") =>
		visibleConnections(graph, tree, { band, mode: "all", focus: { kind: "none" } }).connections.some(
			({ relationshipIds }) => relationshipIds.includes("cron-search")
		);

	expect(drawn("landscape")).toBe(false);
	expect(drawn("runtime")).toBe(true);
});
