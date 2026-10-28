import { expect, test } from "bun:test";
import { fullFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import { deploymentGroups } from "../deployment";
import type { GraphSubset } from "../graph";

const ids = (entries: readonly { entity: { id: string } }[]) => entries.map(({ entity }) => entity.id);

test("one section per placement target, by display name, with its sources in order and breadcrumbs", () => {
	const { sections } = deploymentGroups(fullFixture);

	expect(ids(sections)).toEqual(["rt-production-cluster", "rt-standby-cluster"]);
	expect(sections[0].breadcrumb).toEqual(["Platform", "Compute"]);
	expect(ids(sections[0].rows)).toEqual(["rt-checkout-api", "rt-payment-service", "rt-payment-worker"]);
	expect(sections[0].rows.map(({ breadcrumb }) => breadcrumb)).toEqual([
		["Commerce", "Checkout"],
		["Commerce", "Payments"],
		["Commerce", "Payments"],
	]);
});

test("a source on several targets links to its other targets", () => {
	const { sections } = deploymentGroups(fullFixture);
	const alsoOn = (sectionIndex: number, rowId: string) =>
		ids(sections[sectionIndex].rows.find(({ entity }) => entity.id === rowId)!.alsoOn);

	expect(alsoOn(0, "rt-checkout-api")).toEqual(["rt-standby-cluster"]);
	expect(alsoOn(1, "rt-checkout-api")).toEqual(["rt-production-cluster"]);
	expect(alsoOn(0, "rt-payment-worker")).toEqual([]);
});

test("placement unknown lists containers without a target, excluding targets and other categories", () => {
	// Search API becomes a target, so it leaves the unknown section even with no placement of its own.
	const graph: GraphSubset = {
		...fullFixture,
		relationships: [
			...fullFixture.relationships,
			{ id: "on-search", source: "rt-crawler", target: "rt-search-api", predicate: "runs_on" },
		],
	};
	const { sections, placementUnknown } = deploymentGroups(graph);
	const unknown = ids(placementUnknown);

	expect(ids(sections)).toContain("rt-search-api");
	expect(unknown).not.toContain("rt-search-api");
	expect(unknown).not.toContain("rt-crawler");
	expect(unknown).not.toContain("rt-checkout-db");
	expect(unknown.slice(0, 3)).toEqual(["rt-fulfilment-01", "rt-fulfilment-02", "rt-fulfilment-03"]);
	expect(unknown).toContain("rt-legacy-cron");
	expect(unknown).toHaveLength(13);
});

test("an empty graph has no sections and no unknown placement", () => {
	const groups = deploymentGroups({ ...fullFixture, entities: [], relationships: [] });

	expect(groups).toEqual({ sections: [], placementUnknown: [] });
});
