import type { GraphSubset } from "$features/systems/lib/system-map/graph";
import {
	edgeCasesExample,
	relationshipExample,
	sharedGroupsExample,
} from "$features/systems/lib/system-map/__tests__/test-fixtures";

export type FixtureScenarioId = "hierarchy" | "connections" | "context" | "stress";

export type FixtureScenario = {
	source: GraphSubset;
};

const denseRelationships: GraphSubset["relationships"] = [
	{ id: "r-dense-calls-1", source: "member-a", target: "member-b", predicate: "calls" },
	{ id: "r-dense-calls-2", source: "member-a2", target: "member-b", predicate: "calls" },
	{ id: "r-dense-calls-3", source: "member-b", target: "member-a", predicate: "calls" },
	{ id: "r-dense-depends-1", source: "member-a", target: "member-b", predicate: "depends_on" },
	{ id: "r-dense-depends-2", source: "member-b", target: "member-a", predicate: "depends_on" },
	{ id: "r-dense-emits-1", source: "member-a", target: "member-b", predicate: "emits" },
	{ id: "r-dense-emits-2", source: "member-a2", target: "member-b", predicate: "emits" },
	{ id: "r-dense-reads-1", source: "member-b", target: "member-a2", predicate: "reads" },
	{ id: "r-dense-reads-2", source: "member-b", target: "member-a", predicate: "reads" },
	{ id: "r-dense-owns", source: "member-a2", target: "member-b", predicate: "owns" },
];

const longAndDense = (source: GraphSubset): GraphSubset => ({
	...source,
	entities: source.entities.map((entity) => ({
		...entity,
		label:
			entity.id === "group-a"
				? "Order Fulfillment and Customer Notification — Regional Control Plane"
				: entity.id === "group-b"
					? "Billing and Entitlements — Shared Platform Operations"
					: entity.id === "member-a"
						? "Checkout request orchestration and idempotency worker"
						: entity.id === "member-a2"
							? "Promotion eligibility and pricing rule evaluation service"
							: entity.id === "member-b"
								? "Invoice generation, tax calculation, and ledger synchronization"
								: entity.label,
	})),
	relationships: [...source.relationships, ...denseRelationships],
});

const connectionsSource = relationshipExample.source;
const stressSource = longAndDense(relationshipExample.source);

export const systemMapFixtureScenarios: Readonly<Record<FixtureScenarioId, FixtureScenario>> = {
	hierarchy: {
		source: sharedGroupsExample.source,
	},
	connections: {
		source: connectionsSource,
	},
	context: {
		source: edgeCasesExample.source,
	},
	stress: {
		source: stressSource,
	},
};
