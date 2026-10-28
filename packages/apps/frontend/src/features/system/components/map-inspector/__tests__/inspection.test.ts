import { expect, test } from "bun:test";
import { fullFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import { resolveInspection } from "../inspection";
import { buildInspectionViewData } from "../presentation";

const graph: GraphSubset = {
	...fullFixture,
	relationships: [
		...fullFixture.relationships,
		{ id: "on-index", source: "rt-checkout-api", target: "rt-search-index", predicate: "runs_on" },
	],
};

const inspect = (entityId: string) =>
	buildInspectionViewData(resolveInspection(graph, { kind: "entity", entityId }))!;

test("a landscape counts its systems and every runtime beneath it, including cards held directly", () => {
	expect(inspect("fn-discovery").structureFacts).toEqual(["1 system", "3 runtime entities"]);
});

test("a runtime shows its grouped implementation count and where it runs", () => {
	const api = inspect("rt-checkout-api");

	expect(api.structureFacts).toEqual(["8 implementation entities grouped"]);
	expect(api.runsOn.map(({ id }) => id)).toEqual([
		"rt-production-cluster",
		"rt-search-index",
		"rt-standby-cluster",
	]);
	expect(api.placementUnknown).toBe(false);
	expect(inspect("rt-checkout-db").structureFacts).toEqual([]);
});

test("a container with no loaded target reads placement unknown; infrastructure never does", () => {
	expect(inspect("rt-search-api").placementUnknown).toBe(true);
	expect(inspect("rt-checkout-db").placementUnknown).toBe(false);
	expect(inspect("sys-checkout").placementUnknown).toBe(false);
});
