import { expect, test } from "bun:test";
import { fullFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import { deploymentGroups } from "../deployment";
import { withTestModel } from "./test-model";

// A component placed on its own is not listed; deployment rows are containers.
const graph = withTestModel({
	...fullFixture,
	relationships: [
		...fullFixture.relationships,
		{ id: "orchestrator-on", source: "cmp-orchestrator", target: "ns-commerce", predicate: "runs_on" },
	],
});

const ids = (entries: readonly { entity: { id: string } }[]) => entries.map(({ entity }) => entity.id);

test("the infrastructure tree lists what runs on each node by its most specific placement", () => {
	const { sections } = deploymentGroups(graph);
	const section = (id: string) => sections.find(({ entity }) => entity.id === id)!;

	expect(sections.map(({ entity, depth }) => [entity.id, depth])).toEqual([
		["acct-commerce-prod", 0],
		["cl-prod-euw1", 1],
		["ns-commerce-eu", 2],
		["cl-prod-use1", 1],
		["ns-commerce", 2],
		["ns-search", 2],
		["db-orders", 0],
	]);
	expect(section("ns-commerce").breadcrumb).toBe("commerce-prod › prod-use1");
	expect(ids(section("ns-commerce").rows)).toEqual([
		"rt-payment-service",
		"rt-payment-worker",
		"rt-checkout",
	]);
	expect(section("acct-commerce-prod").rows).toEqual([]);

	const checkout = section("ns-commerce").rows.find(({ entity }) => entity.id === "rt-checkout")!;
	expect(checkout.breadcrumb).toBe("Online Commerce › Commerce Core");
	expect(ids(checkout.alsoOn)).toEqual(["ns-commerce-eu"]);
});

test("placement unknown lists the containers without a placement", () => {
	const unknown = ids(deploymentGroups(graph).placementUnknown);

	expect(unknown).toContain("rt-checkout-proxy");
	expect(unknown).toContain("rt-legacy-cron");
	expect(unknown).not.toContain("rt-checkout");
	expect(unknown).not.toContain("db-orders");
	expect(unknown).toHaveLength(16);
});
