import type { GraphEntity, GraphRelationship, GraphSubset } from "$features/system/lib/system-map/graph";

type EntityRow = [id: string, category: string, kind: string, name: string];

const entity = ([id, category, kind, name]: EntityRow): GraphEntity => ({ id, category, kind, name });

const contains = (source: string, target: string): GraphRelationship => ({
	id: `contains:${source}:${target}`,
	source,
	target,
	predicate: "contains",
});

const relates = (source: string, predicate: string, target: string): GraphRelationship => ({
	id: `${predicate}:${source}:${target}`,
	source,
	target,
	predicate,
});

const mainEntities: EntityRow[] = [
	["fn-commerce", "system_function", "function", "Commerce"],
	["fn-discovery", "system_function", "function", "Discovery"],
	["sys-checkout", "system", "system", "Checkout"],
	["sys-payments", "system", "system", "Payments"],
	["sys-search", "system", "system", "Search"],
	["rt-checkout-api", "container", "service", "Checkout API"],
	["rt-checkout-db", "infrastructure", "database", "Checkout database"],
	["rt-payment-service", "container", "service", "Payment service"],
	["rt-payment-worker", "container", "service", "Payment worker"],
	["rt-search-api", "container", "service", "Search API"],
	["rt-search-index", "infrastructure", "database", "Search index"],
	["cmp-handlers", "component", "module", "handlers"],
	["cmp-pricing", "component", "module", "pricing"],
	["cmp-client", "component", "module", "client"],
	["code-routes", "code", "file", "routes.go"],
	["code-handlers", "code", "file", "handlers.go"],
	["code-pricing", "code", "file", "pricing.go"],
	["code-tax", "code", "file", "tax.go"],
	["code-main", "code", "file", "main.go"],
];

const mainRelationships: GraphRelationship[] = [
	contains("fn-commerce", "sys-checkout"),
	contains("fn-commerce", "sys-payments"),
	contains("fn-discovery", "sys-search"),
	contains("sys-checkout", "rt-checkout-api"),
	contains("sys-checkout", "rt-checkout-db"),
	contains("sys-payments", "rt-payment-service"),
	contains("sys-payments", "rt-payment-worker"),
	contains("sys-search", "rt-search-api"),
	contains("sys-search", "rt-search-index"),
	contains("rt-checkout-api", "cmp-handlers"),
	contains("rt-checkout-api", "cmp-pricing"),
	contains("rt-checkout-api", "cmp-client"),
	contains("cmp-handlers", "code-routes"),
	contains("cmp-handlers", "code-handlers"),
	contains("cmp-pricing", "code-pricing"),
	contains("cmp-pricing", "code-tax"),
	// main.go is also contained by the runtimes directly; its primary parent is the nearer component.
	contains("cmp-handlers", "code-main"),
	contains("rt-checkout-api", "code-main"),
	contains("rt-payment-service", "code-main"),
	relates("rt-checkout-api", "calls", "rt-payment-service"),
	relates("rt-checkout-api", "writes_to", "rt-checkout-db"),
	relates("rt-checkout-api", "calls", "rt-search-api"),
	relates("rt-payment-worker", "calls", "rt-checkout-api"),
];

const fulfilmentRuntimes: EntityRow[] = Array.from({ length: 12 }, (_, index) => {
	const number = String(index + 1).padStart(2, "0");
	return [`rt-fulfilment-${number}`, "container", "service", `Fulfilment worker ${number}`];
});

const extraEntities: EntityRow[] = [
	["rt-ledger-db", "infrastructure", "database", "Ledger database"],
	["rt-legacy-cron", "container", "job", "Legacy cron"],
	["rt-crawler", "container", "service", "Crawler"],
	["sys-notifications", "system", "system", "Notifications"],
	["sys-fulfilment", "system", "system", "Fulfilment"],
	["cmp-checkout-scripts", "component", "module", "scripts"],
	["fn-platform", "system_function", "function", "Platform"],
	["sys-compute", "system", "system", "Compute"],
	["rt-production-cluster", "infrastructure", "cluster", "Production cluster"],
	["rt-standby-cluster", "infrastructure", "cluster", "Standby cluster"],
	...fulfilmentRuntimes,
];

const extraRelationships: GraphRelationship[] = [
	contains("sys-payments", "rt-ledger-db"),
	contains("sys-checkout", "rt-ledger-db"),
	contains("fn-discovery", "rt-crawler"),
	contains("fn-commerce", "sys-fulfilment"),
	// An implementation entity with no runtime ancestor is not counted.
	contains("sys-checkout", "cmp-checkout-scripts"),
	...fulfilmentRuntimes.map(([id]) => contains("sys-fulfilment", id)),
	contains("fn-platform", "sys-compute"),
	contains("sys-compute", "rt-production-cluster"),
	contains("sys-compute", "rt-standby-cluster"),
	// Checkout API runs on both clusters; Search API, Crawler and the fulfilment workers have no placement.
	relates("rt-checkout-api", "runs_on", "rt-production-cluster"),
	relates("rt-checkout-api", "runs_on", "rt-standby-cluster"),
	relates("rt-payment-service", "runs_on", "rt-production-cluster"),
	relates("rt-payment-worker", "runs_on", "rt-production-cluster"),
];

const subset = (entities: EntityRow[], relationships: GraphRelationship[]): GraphSubset => ({
	entities: entities.map(entity),
	relationships,
	unresolvedRelationships: [contains("fn-missing", "sys-search")],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
});

/** The spec's example tree, before the extra cases are added. */
export const mainFixture = subset(mainEntities, mainRelationships);

/**
 * The example tree with a shared database, orphans, a twelve-runtime system and a platform the commerce
 * workloads run on.
 */
export const fullFixture = subset(
	[...mainEntities, ...extraEntities],
	[...mainRelationships, ...extraRelationships]
);
