import type { GraphEntity, GraphRelationship, GraphSubset } from "$features/system/lib/system-map/graph";

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

/** Runtimes per system, in the order above; 300 in total. */
const RUNTIME_COUNTS = [36, 12, 24, 8, 20, 30, 16, 10, 28, 14, 26, 18, 22, 6, 30];

const PREDICATES = ["calls", "reads_from", "writes_to", "publishes_to"];

const BEHAVIOUR_RELATIONSHIP_COUNT = 100;

/** Shared clusters in the Scheduling system that every service runs on. */
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

function buildLargeFixture(): GraphSubset {
	const random = seededRandom(20261029);
	const pick = <T>(items: readonly T[]) => items[Math.floor(random() * items.length)];
	const entities: GraphEntity[] = [];
	const relationships: GraphRelationship[] = [];
	const runtimeIds: string[] = [];
	const codeIds: string[] = [];

	for (const { id, name } of CLUSTERS) {
		entities.push({ id, category: "infrastructure", kind: "cluster", name });
		relationships.push(relationship("sys-scheduling", "contains", id));
	}

	LANDSCAPES.forEach((name) => {
		entities.push({ id: `fn-${slug(name)}`, category: "system_function", kind: "function", name });
	});

	SYSTEMS.forEach((name, systemIndex) => {
		const systemId = `sys-${slug(name)}`;
		const landscapeId = `fn-${slug(LANDSCAPES[Math.floor(systemIndex / 5)])}`;
		entities.push({ id: systemId, category: "system", kind: "system", name });
		relationships.push(relationship(landscapeId, "contains", systemId));

		for (let index = 1; index <= RUNTIME_COUNTS[systemIndex]; index++) {
			const number = String(index).padStart(2, "0");
			const runtimeId = `rt-${slug(name)}-${number}`;
			if (index % 5 === 0) {
				entities.push({
					id: runtimeId,
					category: "infrastructure",
					kind: "database",
					name: `${name} store ${number}`,
				});
			} else {
				entities.push({
					id: runtimeId,
					category: "container",
					kind: "service",
					name: `${name} service ${number}`,
				});
				relationships.push(relationship(runtimeId, "runs_on", CLUSTERS[systemIndex % 2].id));
			}
			relationships.push(relationship(systemId, "contains", runtimeId));
			runtimeIds.push(runtimeId);
		}

		// The first runtime of each system carries two components and a source file.
		const firstRuntimeId = `rt-${slug(name)}-01`;
		const codeId = `code-${slug(name)}`;
		entities.push({ id: codeId, category: "code", kind: "file", name: `${slug(name)}.go` });
		codeIds.push(codeId);
		for (const component of ["handlers", "client"]) {
			const componentId = `cmp-${slug(name)}-${component}`;
			entities.push({ id: componentId, category: "component", kind: "module", name: component });
			relationships.push(relationship(firstRuntimeId, "contains", componentId));
		}
		relationships.push(relationship(`cmp-${slug(name)}-handlers`, "contains", codeId));
	});

	// A tenth of the behaviour starts in source files, which roll up to their runtimes.
	const behaviour = new Map<string, GraphRelationship>();
	while (behaviour.size < BEHAVIOUR_RELATIONSHIP_COUNT) {
		let source = pick(runtimeIds);
		if (random() < 0.1) {
			source = pick(codeIds);
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
 * Three landscapes, fifteen systems and 302 runtimes, for checking the map zoomed fully out. Services run on
 * one of two shared clusters.
 */
export const largeFixture = buildLargeFixture();
