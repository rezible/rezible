import type { GraphEntity, GraphRelationship } from "$features/system/lib/system-map/graph";
import type { FixtureGraph } from "./fixtures";

type EntityRow = [id: string, category: string, kind: string, name: string, parentId?: string];

const relates = (source: string, predicate: string, target: string): GraphRelationship => ({
	id: `${predicate}:${source}:${target}`,
	source,
	target,
	predicate,
});

const entityRows: EntityRow[] = [
	// Logical: landscapes, nested systems, containers and components.
	["fn-commerce", "system_function", "function", "Online Commerce"],
	["fn-discovery", "system_function", "function", "Discovery"],
	["sys-commerce-core", "system", "system", "Commerce Core", "fn-commerce"],
	["sys-payments", "system", "system", "Payments", "sys-commerce-core"],
	["sys-search", "system", "system", "Search", "fn-discovery"],
	["sys-fulfilment", "system", "system", "Fulfilment", "fn-commerce"],
	["rt-checkout", "container", "service", "checkout", "sys-commerce-core"],
	["rt-checkout-proxy", "container", "sidecar", "checkout proxy", "rt-checkout"],
	["cmp-orchestrator", "component", "module", "Checkout Orchestrator", "rt-checkout"],
	["cmp-payment-adapter", "component", "module", "Payment Adapter", "rt-checkout"],
	["rt-payment-service", "container", "service", "Payment service", "sys-payments"],
	["rt-payment-worker", "container", "service", "Payment worker", "sys-payments"],
	["rt-search-api", "container", "service", "Search API", "sys-search"],
	["rt-crawler", "container", "service", "Crawler", "fn-discovery"],
	["rt-legacy-cron", "container", "job", "Legacy cron"],
	// Its parent is outside the subset, so it is a root.
	["rt-reports", "container", "service", "Reports", "sys-missing"],
	...Array.from({ length: 12 }, (_, index): EntityRow => {
		const number = String(index + 1).padStart(2, "0");
		return [
			`rt-fulfilment-${number}`,
			"container",
			"service",
			`Fulfilment worker ${number}`,
			"sys-fulfilment",
		];
	}),
	// Deployment: environment → cluster → namespace.
	["env-production", "infrastructure", "environment", "production"],
	["cl-prod-use1", "infrastructure", "kubernetes_cluster", "prod-use1", "env-production"],
	["cl-prod-euw1", "infrastructure", "kubernetes_cluster", "prod-euw1", "env-production"],
	["ns-commerce", "infrastructure", "kubernetes_namespace", "commerce", "cl-prod-use1"],
	["ns-search", "infrastructure", "kubernetes_namespace", "search", "cl-prod-use1"],
	["ns-commerce-eu", "infrastructure", "kubernetes_namespace", "commerce", "cl-prod-euw1"],
	["db-orders", "infrastructure", "database", "orders-postgres"],
	// Other hierarchies, shown only in the inspector.
	["repo-checkout", "code", "repository", "checkout-service"],
	["team-commerce", "actor", "team", "Commerce Team"],
];

const entities: GraphEntity[] = entityRows.map(([id, category, kind, name, parentId]) => ({
	id,
	category,
	kind,
	name,
	parentId,
	// Two catalogues disagree about Search API; the backend chose Search.
	otherParentIds: id === "rt-search-api" ? ["sys-commerce-core"] : undefined,
}));

const loaded = new Set(entities.map((entity) => entity.id));
const parentClaims = entities
	.filter((entity) => entity.parentId && loaded.has(entity.parentId))
	.map((entity) => relates(entity.parentId!, "contains", entity.id));

const relationships: GraphRelationship[] = [
	...parentClaims,
	relates("sys-commerce-core", "contains", "rt-search-api"),
	// checkout runs in two namespaces; its placement on production is implied by them, so not shown.
	relates("rt-checkout", "runs_on", "ns-commerce"),
	relates("rt-checkout", "runs_on", "ns-commerce-eu"),
	relates("rt-checkout", "runs_on", "env-production"),
	relates("rt-payment-service", "runs_on", "ns-commerce"),
	relates("rt-payment-worker", "runs_on", "ns-commerce"),
	relates("rt-search-api", "runs_on", "ns-search"),
	relates("rt-checkout", "built_from", "repo-checkout"),
	relates("team-commerce", "owns", "rt-checkout"),
	relates("rt-checkout", "calls", "rt-payment-service"),
	relates("rt-checkout", "calls", "rt-search-api"),
	relates("cmp-payment-adapter", "calls", "rt-payment-service"),
	relates("rt-payment-worker", "calls", "rt-checkout"),
	// An infrastructure endpoint: listed by the inspector, not drawn.
	relates("rt-checkout", "writes_to", "db-orders"),
];

/**
 * The checkout example with nested systems, a sidecar container, a disputed parent, roots, a twelve-container
 * system and an environment → cluster → namespace chain.
 */
export const fullFixture: FixtureGraph = {
	entities,
	relationships,
	unresolvedRelationships: [relates("fn-missing", "contains", "sys-search")],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
};
