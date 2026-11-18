import { expect, test } from "bun:test";
import { fullFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";
import { withTestModel } from "$features/system/lib/system-map/__tests__/test-model";
import { resolveInspection } from "../inspection";
import { buildInspectionViewData } from "../presentation";

const graph = withTestModel(fullFixture);

const inspect = (entityId: string) =>
	buildInspectionViewData(resolveInspection(graph, { kind: "entity", entityId }))!;

test("a landscape counts its systems and every container beneath it, including cards held directly", () => {
	expect(inspect("fn-discovery").structureFacts).toEqual(["1 system", "2 runtime entities"]);
});

test("a container shows its components, placements with ancestry, code and undrawn relationships", () => {
	const checkout = inspect("rt-checkout");

	expect(checkout.structureFacts).toEqual(["2 implementation entities grouped"]);
	expect(checkout.runsOn.map(({ text }) => text)).toEqual([
		"commerce-prod › prod-use1 › commerce · ns-commerce",
		"commerce-prod › prod-euw1 › commerce · ns-commerce-eu",
	]);
	expect(checkout.builtFrom.map(({ id }) => id)).toEqual(["repo-checkout"]);
	expect(checkout.incidentRelationships.map(({ id }) => id)).toEqual([
		"calls:rt-checkout:rt-payment-service",
		"calls:rt-checkout:rt-search-api",
		"calls:rt-payment-worker:rt-checkout",
	]);
	expect(checkout.undrawnRelationships.map(({ id }) => id)).toContain("writes_to:rt-checkout:db-orders");
	expect(checkout.undrawnRelationships.map(({ id }) => id)).toContain("owns:team-commerce:rt-checkout");
});

test("other parent claims and unknown placements are shown", () => {
	expect(inspect("rt-search-api").alsoClaimedBy.map(({ id }) => id)).toEqual(["sys-commerce-core"]);
	expect(inspect("rt-crawler").placementUnknown).toBe(true);
	expect(inspect("db-orders").placementUnknown).toBe(false);
});
