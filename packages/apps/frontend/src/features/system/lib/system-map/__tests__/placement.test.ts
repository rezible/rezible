import { expect, test } from "bun:test";
import { fullFixture as fixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import {
	breadcrumb,
	isPlacementUnknown,
	placementFootprint,
	placementSources,
	placementTargets,
} from "../placement";
import { buildStructureTree } from "../structure";
import { withTestModel } from "./test-model";

const graph = withTestModel(fixture);
const entities = new Map(graph.entities.map((entity) => [entity.id, entity]));

test("a source shows only its most specific placements, each with its infrastructure ancestry", () => {
	const targets = placementTargets(graph).get("rt-checkout")!;

	// checkout also runs_on production, which its namespaces imply.
	expect(targets.map(({ id }) => id)).toEqual(["ns-commerce", "ns-commerce-eu"]);
	expect(targets.map(({ id }) => breadcrumb(entities, id))).toEqual([
		"production › prod-use1",
		"production › prod-euw1",
	]);
	expect(placementSources(graph, placementTargets(graph)).has("env-production")).toBe(false);
});

test("placement is unknown only for containers without a loaded target", () => {
	const targets = placementTargets(graph);
	const unknown = (id: string) => isPlacementUnknown(graph, entities.get(id)!, targets.get(id));

	expect(unknown("rt-crawler")).toBe(true);
	expect(unknown("rt-checkout")).toBe(false);
	expect(unknown("db-orders")).toBe(false);
	expect(unknown("sys-search")).toBe(false);
});

test("the footprint marks the cards that run on a target, or their regions at the landscape band", () => {
	const tree = buildStructureTree(graph);
	const footprint = (band: "landscape" | "runtime") =>
		[...placementFootprint(graph, tree, "ns-commerce", band)].sort();

	expect(footprint("runtime")).toEqual(["rt-checkout", "rt-payment-service", "rt-payment-worker"]);
	expect(footprint("landscape")).toEqual(["sys-commerce-core", "sys-payments"]);
});
