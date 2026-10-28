import { expect, test } from "bun:test";
import { fullFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import type { GraphSubset } from "../graph";
import {
	isPlacementUnknown,
	placementChips,
	placementCountLine,
	placementFootprint,
	placementTargets,
} from "../placement";
import { buildStructureTree } from "../structure";

const graph: GraphSubset = {
	...fullFixture,
	relationships: [
		...fullFixture.relationships,
		{ id: "on-index", source: "rt-checkout-api", target: "rt-search-index", predicate: "runs_on" },
	],
	unresolvedRelationships: [
		{ id: "on-missing", source: "rt-checkout-api", target: "rt-missing", predicate: "runs_on" },
	],
};

test("a card shows its first two targets by display name as chips and collapses the rest", () => {
	const targets = placementTargets(graph);
	const chips = placementChips(targets.get("rt-checkout-api")!);

	expect(chips.shown.map(({ id }) => id)).toEqual(["rt-production-cluster", "rt-search-index"]);
	expect(chips.collapsed.map(({ id }) => id)).toEqual(["rt-standby-cluster"]);
	expect(placementChips(targets.get("rt-payment-worker")!).collapsed).toEqual([]);
	expect(targets.has("rt-checkout-db")).toBe(false);
});

test("the popover counts the loaded entities that run on a target", () => {
	expect(placementCountLine(graph, "rt-production-cluster")).toBe("3 entities run here");
	expect(placementCountLine(graph, "rt-standby-cluster")).toBe("1 entity runs here");
});

test("on a partial graph the count is a lower bound", () => {
	const partial: GraphSubset = { ...graph, enumeration: { scope: "query", stopReason: "entity-limit" } };

	expect(placementCountLine(partial, "rt-production-cluster")).toBe("At least 3 entities run here");
});

test("placement is unknown only for containers without a loaded target", () => {
	const targets = placementTargets(graph);
	const unknown = (id: string) =>
		isPlacementUnknown(
			graph.entities.find((entity) => entity.id === id)!,
			targets.get(id)
		);

	expect(unknown("rt-search-api")).toBe(true);
	expect(unknown("rt-checkout-api")).toBe(false);
	// Infrastructure is not expected to run anywhere.
	expect(unknown("rt-checkout-db")).toBe(false);
	expect(unknown("sys-checkout")).toBe(false);
});

test("the footprint marks the cards that run on a target, or their systems at the landscape band", () => {
	const tree = buildStructureTree(fullFixture);
	const footprint = (band: "landscape" | "runtime") =>
		[...placementFootprint(fullFixture, tree, "rt-production-cluster", band)].sort();

	expect(footprint("runtime")).toEqual(["rt-checkout-api", "rt-payment-service", "rt-payment-worker"]);
	expect(footprint("landscape")).toEqual(["sys-checkout", "sys-payments"]);
	expect(placementFootprint(fullFixture, tree, "rt-checkout-api", "runtime").size).toBe(0);
});
