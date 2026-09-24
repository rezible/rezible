import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { Point } from "$features/system/lib/system-map/geometry";

export type FixtureScenarioId = "shared" | "empty" | "sparse-positions";

export type FixtureScenario = {
	graph: GraphSubset;
	positions?: Readonly<Partial<Record<string, Point>>>;
};

const sharedGraph: GraphSubset = {
	entities: [
		{ id: "function-orders", category: "system_function", kind: "order processing" },
		{ id: "system-checkout", category: "system", kind: "checkout" },
		{ id: "system-billing", category: "system", kind: "billing" },
		{ id: "service-checkout", category: "container", kind: "checkout service" },
		{ id: "database-orders", category: "infrastructure", kind: "orders database" },
	],
	relationships: [
		{
			id: "contains-orders",
			source: "function-orders",
			target: "system-checkout",
			predicate: "contains",
		},
		{
			id: "contains-billing",
			source: "function-orders",
			target: "system-billing",
			predicate: "contains",
		},
		{
			id: "contains-checkout",
			source: "system-checkout",
			target: "service-checkout",
			predicate: "contains",
		},
		{
			id: "contains-orders-db",
			source: "system-checkout",
			target: "database-orders",
			predicate: "contains",
		},
		{
			id: "contains-billing-db",
			source: "system-billing",
			target: "database-orders",
			predicate: "contains",
		},
		{ id: "reads-orders-db", source: "service-checkout", target: "database-orders", predicate: "reads" },
	],
	unresolvedRelationships: [
		{
			id: "contains-platform-checkout",
			source: "system-platform",
			target: "service-checkout",
			predicate: "contains",
		},
	],
	enumeration: { scope: "fixture", stopReason: "relationship-limit" },
};

const emptyGraph: GraphSubset = {
	entities: [],
	relationships: [],
	unresolvedRelationships: [],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
};

export const systemMapFixtureScenarios: Readonly<Record<FixtureScenarioId, FixtureScenario>> = {
	shared: {
		graph: sharedGraph,
	},
	empty: {
		graph: emptyGraph,
	},
	"sparse-positions": {
		graph: sharedGraph,
		positions: {
			"function-orders": { x: 0, y: 0 },
			"system-checkout": { x: -260, y: 150 },
		},
	},
};
