import type { GraphEntity, GraphRelationship } from "$features/system/lib/system-map/graph";
import type { FixtureGraph } from "./fixtures";

const LANDSCAPES = ["Commerce", "Platform", "Insights"];

const SYSTEMS = [
	"Storefront",
	"Checkout",
	"Payments",
	"Fulfilment",
	"Returns",
	"Identity",
	"Messaging",
	"Storage",
	"Scheduling",
	"Edge",
	"Analytics",
	"Reporting",
	"Search",
	"Recommendations",
	"Experiments",
];

/** Runtimes per system, in the order above; 300 in total, every fifth a store. */
const RUNTIME_COUNTS = [36, 12, 24, 8, 20, 30, 16, 10, 28, 14, 26, 18, 22, 6, 30];

const PREDICATES = ["calls", "reads_from", "writes_to", "publishes_to"];

const BEHAVIOUR_RELATIONSHIP_COUNT = 100;

/** Shared clusters, in one environment, that every service runs on; stores sit in the clusters. */
const CLUSTERS = [
	{ id: "rt-primary-cluster", name: "Primary cluster" },
	{ id: "rt-secondary-cluster", name: "Secondary cluster" },
];

/** Mulberry32: a fixed seed gives the same graph on every load. */
function seededRandom(seed: number) {
	let state = seed;
	return () => {
		state = (state + 0x6d2b79f5) | 0;
		let value = Math.imul(state ^ (state >>> 15), 1 | state);
		value = (value + Math.imul(value ^ (value >>> 7), 61 | value)) ^ value;
		return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
	};
}

const slug = (name: string) => name.toLowerCase();

const relationship = (source: string, predicate: string, target: string): GraphRelationship => ({
	id: `${predicate}:${source}:${target}`,
	source,
	target,
	predicate,
});

function buildLargeFixture(): FixtureGraph {
	const random = seededRandom(20261029);
	const pick = <T>(items: readonly T[]) => items[Math.floor(random() * items.length)];
	const entities: GraphEntity[] = [];
	const relationships: GraphRelationship[] = [];
	const runtimeIds: string[] = [];
	const componentIds: string[] = [];
	const add = (entity: GraphEntity) => {
		entities.push(entity);
		if (entity.parentId) {
			relationships.push(relationship(entity.parentId, "contains", entity.id));
		}
	};

	add({ id: "env-production", category: "infrastructure", kind: "environment", name: "production" });
	for (const { id, name } of CLUSTERS) {
		add({ id, category: "infrastructure", kind: "kubernetes_cluster", name, parentId: "env-production" });
	}

	LANDSCAPES.forEach((name) => {
		add({ id: `fn-${slug(name)}`, category: "system_function", kind: "function", name });
	});

	SYSTEMS.forEach((name, systemIndex) => {
		const systemId = `sys-${slug(name)}`;
		const landscapeId = `fn-${slug(LANDSCAPES[Math.floor(systemIndex / 5)])}`;
		add({ id: systemId, category: "system", kind: "system", name, parentId: landscapeId });

		for (let index = 1; index <= RUNTIME_COUNTS[systemIndex]; index++) {
			const number = String(index).padStart(2, "0");
			const runtimeId = `rt-${slug(name)}-${number}`;
			const cluster = CLUSTERS[systemIndex % 2].id;
			if (index % 5 === 0) {
				add({
					id: runtimeId,
					category: "infrastructure",
					kind: "database",
					name: `${name} store ${number}`,
					parentId: cluster,
				});
				continue;
			}
			add({
				id: runtimeId,
				category: "container",
				kind: "service",
				name: `${name} service ${number}`,
				parentId: systemId,
			});
			relationships.push(relationship(runtimeId, "runs_on", cluster));
			runtimeIds.push(runtimeId);
		}

		// The first container of each system carries two components.
		for (const component of ["handlers", "client"]) {
			const componentId = `cmp-${slug(name)}-${component}`;
			add({
				id: componentId,
				category: "component",
				kind: "module",
				name: component,
				parentId: `rt-${slug(name)}-01`,
			});
			componentIds.push(componentId);
		}
	});

	// A tenth of the flow starts in components, which roll up to their containers.
	const behaviour = new Map<string, GraphRelationship>();
	while (behaviour.size < BEHAVIOUR_RELATIONSHIP_COUNT) {
		let source = pick(runtimeIds);
		if (random() < 0.1) {
			source = pick(componentIds);
		}
		const target = pick(runtimeIds);
		if (source === target) {
			continue;
		}
		const next = relationship(source, pick(PREDICATES), target);
		behaviour.set(next.id, next);
	}

	return {
		entities,
		relationships: [...relationships, ...behaviour.values()],
		unresolvedRelationships: [],
		enumeration: { scope: "fixture", stopReason: "exhausted" },
	};
}

/**
 * Three landscapes, fifteen systems and 240 containers, for checking the map zoomed fully out. Services run on
 * one of two shared clusters, which also hold 60 stores.
 */
export const largeFixture = buildLargeFixture();
