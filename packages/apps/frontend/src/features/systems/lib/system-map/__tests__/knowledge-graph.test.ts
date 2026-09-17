import { describe, expect, test } from "bun:test";
import {
	categoryDisplay,
	NodeDetailLevel,
	DisplayMode,
	getMapCategoryDisplay,
	MapCategory,
	parseMapCategory,
} from "../category";
import { Coverage } from "../graph";
import { projectMap } from "../projection";
import {
	edgeCasesExample,
	graphExamples,
	relationshipExample,
	sharedGroupsExample,
	type ExpectedConnection,
	type GraphExample,
} from "./test-fixtures";

describe("system map category policy", () => {
	test("assigns architectural nodes to the intended detail levels", () => {
		const levels = {
			[MapCategory.Function]: NodeDetailLevel.Landscape,
			[MapCategory.System]: NodeDetailLevel.Systems,
			[MapCategory.Container]: NodeDetailLevel.Runtime,
			[MapCategory.Infrastructure]: NodeDetailLevel.Runtime,
			[MapCategory.Component]: NodeDetailLevel.Implementation,
			[MapCategory.Code]: NodeDetailLevel.Implementation,
		};
		for (const [category, level] of Object.entries(levels)) {
			expect(getMapCategoryDisplay(category)).toMatchObject({ mode: DisplayMode.Node, level });
		}
	});

	test("keeps actors, annotations, and inspection-only categories unlevelled", () => {
		const modes = {
			[MapCategory.Actor]: DisplayMode.Node,
			[MapCategory.Concern]: DisplayMode.Annotation,
			[MapCategory.Decision]: DisplayMode.Annotation,
			[MapCategory.Event]: DisplayMode.Annotation,
			[MapCategory.Signal]: DisplayMode.Annotation,
			[MapCategory.Process]: DisplayMode.DetailsOnly,
			[MapCategory.Unknown]: DisplayMode.DetailsOnly,
		};
		for (const [category, mode] of Object.entries(modes)) {
			expect(getMapCategoryDisplay(category)).toMatchObject({ mode, level: undefined });
		}
	});

	test("parses every category identifier and resolves its display", () => {
		for (const category of Object.values(MapCategory)) {
			expect(parseMapCategory(category)).toBe(category);
			expect(getMapCategoryDisplay(category)).toBe(categoryDisplay[category]);
		}
	});

	test("uses the inspection-only fallback for unrecognised strings", () => {
		for (const category of ["future_category", "toString", "__proto__", "", "System", " system "]) {
			expect(parseMapCategory(category)).toBe(MapCategory.Unknown);
			expect(getMapCategoryDisplay(category)).toBe(categoryDisplay[MapCategory.Unknown]);
		}
	});
});

describe("test-only graph examples", () => {
	const revealFor = (example: GraphExample) => ({
		detail: example.detail,
		nearbyEntityIds: example.source.entities.map((entity) => entity.id),
	});

	const expectedConnectionsFor = (example: GraphExample): ExpectedConnection[] => [
		...(example.expected.sharedMembership ?? []).map((membership) => ({
			endpoints: [membership.parentId, membership.sharedId] as const,
			predicate: "contains",
			classification: "direct" as const,
			sourceRelationshipIds: [membership.membershipId],
			count: 1,
		})),
		...example.expected.connections,
	];

	const connectionExpectation = (connection: {
		endpoints: readonly [string, string];
		predicate: string;
		classification: "direct" | "summary";
		sourceRelationshipIds: readonly string[];
	}): ExpectedConnection => ({
		endpoints: connection.endpoints,
		predicate: connection.predicate,
		classification: connection.classification,
		sourceRelationshipIds: connection.sourceRelationshipIds,
		count: connection.sourceRelationshipIds.length,
	});

	for (const example of graphExamples) {
		test(`${example.name} projects the expected nodes, memberships, and relationships`, () => {
			const sourceBeforeProjection = structuredClone(example.source);
			const projection = projectMap(example.source, revealFor(example), example.displayOptions);
			const entities = new Map(example.source.entities.map((entity) => [entity.id, entity]));
			const facts = example.source.relationships;
			const factsById = new Map(facts.map((fact) => [fact.id, fact]));
			const visibleIds = new Set(example.expected.visibleRepresentatives);
			const projectedIds = projection.nodes.map((node) => node.id);
			expect(projectedIds).toEqual([...example.expected.visibleRepresentatives]);
			expect(entities.size).toBe(example.source.entities.length);
			expect(factsById.size).toBe(facts.length);
			expect(visibleIds.size).toBe(example.expected.visibleRepresentatives.length);
			expect(example.source).toEqual(sourceBeforeProjection);
			for (const fact of facts) {
				expect(entities.has(fact.source)).toBe(true);
				expect(entities.has(fact.target)).toBe(true);
			}
			for (const id of visibleIds) {
				expect(entities.has(id)).toBe(true);
				const category = entities.get(id)?.category ?? MapCategory.Unknown;
				expect(getMapCategoryDisplay(category).mode).toBe(DisplayMode.Node);
				if (category === MapCategory.Actor) expect(example.displayOptions.showActors).toBe(true);
			}
			for (const id of example.expected.enclosureMembershipIds ?? []) {
				const membership = factsById.get(id);
				expect(membership?.predicate).toBe("contains");
				expect(visibleIds.has(membership!.source)).toBe(true);
				expect(visibleIds.has(membership!.target)).toBe(true);
			}
			expect(
				projection.nodes.flatMap((node) => (node.enclosure ? [node.enclosure.membershipId] : []))
			).toEqual([...(example.expected.enclosureMembershipIds ?? [])]);
			for (const exclusion of example.expected.excludedEnclosures ?? []) {
				expect(factsById.get(exclusion.membershipId)?.predicate).toBe("contains");
				expect(example.expected.enclosureMembershipIds ?? []).not.toContain(exclusion.membershipId);
				expect(
					projection.nodes.some((node) => node.enclosure?.membershipId === exclusion.membershipId)
				).toBe(false);
			}
			for (const id of example.expected.unsupportedIds ?? []) {
				expect(entities.has(id)).toBe(true);
				expect(visibleIds.has(id)).toBe(false);
				const category = entities.get(id)?.category ?? MapCategory.Unknown;
				expect(getMapCategoryDisplay(category)).toBe(categoryDisplay[MapCategory.Unknown]);
			}
			for (const membership of example.expected.sharedMembership) {
				expect(visibleIds.has(membership.parentId)).toBe(true);
				expect(visibleIds.has(membership.sharedId)).toBe(true);
				expect(factsById.get(membership.membershipId)).toMatchObject({
					source: membership.parentId,
					target: membership.sharedId,
					predicate: "contains",
				});
				const connection = projection.connections.find(
					(candidate) =>
						candidate.sourceRelationshipIds.length === 1 &&
						candidate.sourceRelationshipIds[0] === membership.membershipId
				);
				expect(connection).toMatchObject({
					endpoints: [membership.parentId, membership.sharedId],
					predicate: "contains",
					classification: "direct",
				});
			}
			const expectedConnections = expectedConnectionsFor(example);
			expect(projection.connections.map(connectionExpectation)).toEqual(expectedConnections);
			for (const connection of projection.connections) {
				const [source, target] = connection.endpoints;
				expect(visibleIds.has(source)).toBe(true);
				expect(visibleIds.has(target)).toBe(true);
				expect(connection.sourceRelationshipIds.length).toBeGreaterThan(0);
				expect(new Set(connection.sourceRelationshipIds).size).toBe(
					connection.sourceRelationshipIds.length
				);
				expect("count" in connection).toBe(false);
				for (const id of connection.sourceRelationshipIds) {
					const fact = factsById.get(id);
					expect(fact).toMatchObject({ predicate: connection.predicate });
					const sourceRepresentative = projection.representativeByEntityId.get(fact!.source);
					const targetRepresentative = projection.representativeByEntityId.get(fact!.target);
					expect([sourceRepresentative, targetRepresentative]).toEqual([source, target]);
					if (connection.classification === "direct") {
						expect(fact).toMatchObject({ source, target });
					}
				}
			}
			expect(
				projection.annotations.map(({ entityId, relationshipId, originalTargetId }) => ({
					entityId,
					relationshipId,
					originalTargetId,
				}))
			).toEqual([...example.expected.annotations]);
			for (const annotation of projection.annotations) {
				expect(example.displayOptions.showAnnotations).toBe(true);
				const category = entities.get(annotation.entityId)?.category ?? MapCategory.Unknown;
				expect(getMapCategoryDisplay(category).mode).toBe(DisplayMode.Annotation);
				expect(factsById.get(annotation.relationshipId)).toMatchObject({
					source: annotation.entityId,
					target: annotation.originalTargetId,
				});
				expect(projection.nodes.some((node) => node.id === annotation.representativeId)).toBe(true);
			}
		});
	}
});

describe("system map projection boundaries", () => {
	const allEntityIds = relationshipExample.source.entities.map((entity) => entity.id);
	const transitiveSharedSource = {
		...sharedGroupsExample.source,
		entities: [
			...sharedGroupsExample.source.entities,
			{ id: "member-a-alt", category: MapCategory.Container, label: "member-a-alt", kind: "subject" },
		],
		relationships: [
			...sharedGroupsExample.source.relationships.filter(
				(relationship) => relationship.id !== "m-shared-a" && relationship.id !== "m-shared-b"
			),
			{ id: "m-a-alt", source: "group-a", target: "member-a-alt", predicate: "contains" },
			{ id: "m-shared-a", source: "member-a", target: "shared-resource", predicate: "contains" },
			{
				id: "m-shared-a-alt",
				source: "member-a-alt",
				target: "shared-resource",
				predicate: "contains",
			},
			{ id: "m-shared-b", source: "member-b", target: "shared-resource", predicate: "contains" },
			{ id: "r-shared-calls", source: "shared-resource", target: "member-a", predicate: "calls" },
		],
	};

	test("promotes a transitive shared entity once without treating ancestor paths as sibling parents", () => {
		const projection = projectMap(
			transitiveSharedSource,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: ["group-a", "group-b"] },
			sharedGroupsExample.displayOptions
		);
		const sharedCalls = projection.connections.filter((connection) =>
			connection.sourceRelationshipIds.includes("r-shared-calls")
		);

		expect(projection.nodes.map((node) => node.id)).toEqual([
			"root",
			"group-a",
			"group-b",
			"shared-resource",
		]);
		expect(projection.representativeByEntityId.get("shared-resource")).toBe("shared-resource");
		expect(sharedCalls).toHaveLength(1);
		expect(sharedCalls[0]).toMatchObject({
			endpoints: ["shared-resource", "group-a"],
			predicate: "calls",
			classification: "summary",
		});
	});

	test("considers visible ancestors at different path depths for a shared entity", () => {
		const source = {
			coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
			entities: [
				{ id: "group-a", category: MapCategory.System, label: "group-a", kind: "subject" },
				{ id: "group-b", category: MapCategory.System, label: "group-b", kind: "subject" },
				{
					id: "intermediate-b",
					category: MapCategory.Container,
					label: "intermediate-b",
					kind: "subject",
				},
				{ id: "shared", category: MapCategory.Container, label: "shared", kind: "subject" },
			],
			relationships: [
				{ id: "m-a-shared", source: "group-a", target: "shared", predicate: "contains" },
				{
					id: "m-b-intermediate",
					source: "group-b",
					target: "intermediate-b",
					predicate: "contains",
				},
				{
					id: "m-intermediate-shared",
					source: "intermediate-b",
					target: "shared",
					predicate: "contains",
				},
				{ id: "r-shared-calls-b", source: "shared", target: "intermediate-b", predicate: "calls" },
			],
		};
		const projection = projectMap(
			source,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: ["group-a", "group-b"] },
			sharedGroupsExample.displayOptions
		);

		expect(projection.nodes.filter((node) => node.id === "shared")).toHaveLength(1);
		expect(projection.connections).toContainEqual({
			id: 'summary:["summary","shared","group-b","calls"]',
			endpoints: ["shared", "group-b"],
			predicate: "calls",
			classification: "summary",
			sourceRelationshipIds: ["r-shared-calls-b"],
		});
	});

	test("keeps a shared resource represented when both groups are nearby at Runtime detail", () => {
		const projection = projectMap(
			sharedGroupsExample.source,
			{ detail: NodeDetailLevel.Runtime, nearbyEntityIds: ["group-a", "group-b"] },
			sharedGroupsExample.displayOptions
		);

		expect(projection.nodes.map((node) => node.id)).toEqual([
			"root",
			"group-a",
			"group-b",
			"shared-resource",
		]);
		expect(projection.nodes.some((node) => node.id === "member-a")).toBe(false);
		expect(projection.nodes.some((node) => node.id === "member-b")).toBe(false);
		expect(
			projection.connections.filter(
				(connection) => connection.predicate === "contains" && connection.classification === "direct"
			)
		).toHaveLength(2);
	});

	test("does not let support evidence inflate a summary count", () => {
		const projection = projectMap(
			relationshipExample.source,
			{ detail: relationshipExample.detail, nearbyEntityIds: allEntityIds },
			relationshipExample.displayOptions
		);
		const summary = projection.connections.find(
			(connection) =>
				connection.classification === "summary" &&
				connection.predicate === "calls" &&
				connection.endpoints[0] === "group-a"
		);

		expect(summary?.sourceRelationshipIds).toEqual(["r-one", "r-two"]);
		expect(summary?.sourceRelationshipIds.length).toBe(2);
	});

	test("keeps a summary ID stable when another source relationship joins its group", () => {
		const reveal = { detail: relationshipExample.detail, nearbyEntityIds: allEntityIds };
		const baseline = projectMap(relationshipExample.source, reveal, relationshipExample.displayOptions);
		const expandedSource = {
			...relationshipExample.source,
			entities: [
				...relationshipExample.source.entities,
				{ id: "member-b2", category: MapCategory.Container, label: "member-b2", kind: "subject" },
			],
			relationships: [
				...relationshipExample.source.relationships,
				{ id: "m-b2", source: "group-b", target: "member-b2", predicate: "contains" },
				{ id: "r-three", source: "member-a", target: "member-b2", predicate: "calls" },
			],
		};
		const expanded = projectMap(
			expandedSource,
			{ ...reveal, nearbyEntityIds: [...reveal.nearbyEntityIds, "member-b2"] },
			relationshipExample.displayOptions
		);
		const baselineSummary = baseline.connections.find(
			(connection) =>
				connection.classification === "summary" &&
				connection.predicate === "calls" &&
				connection.endpoints[0] === "group-a"
		);
		const expandedSummary = expanded.connections.find(
			(connection) =>
				connection.classification === "summary" &&
				connection.predicate === "calls" &&
				connection.endpoints[0] === "group-a"
		);

		expect(expandedSummary?.id).toBe(baselineSummary?.id);
		expect(expandedSummary?.sourceRelationshipIds).toEqual(["r-one", "r-two", "r-three"]);
	});

	test("shows an enabled actor once and keeps an annotation on its original hidden target", () => {
		const example = graphExamples.find(
			(candidate) => candidate.name === "partial graph with cyclic membership and annotation context"
		)!;
		const projection = projectMap(
			example.source,
			{ detail: example.detail, nearbyEntityIds: example.source.entities.map((entity) => entity.id) },
			{ showActors: true, showAnnotations: true }
		);

		expect(projection.nodes.filter((node) => node.id === "actor")).toHaveLength(1);
		const actorConnection = projection.connections.find(
			(connection) => connection.sourceRelationshipIds[0] === "r-actor"
		);
		expect(actorConnection).toEqual({
			id: 'summary:["summary","actor","parent","owns"]',
			endpoints: ["actor", "parent"],
			predicate: "owns",
			classification: "summary",
			sourceRelationshipIds: ["r-actor"],
		});
		expect(projection.annotations).toContainEqual({
			entityId: "event",
			relationshipId: "r-event",
			originalTargetId: "child",
			representativeId: "parent",
		});
	});

	test("does not give unsupported entities or disabled actors architectural representatives", () => {
		const source = {
			...edgeCasesExample.source,
			relationships: [
				...edgeCasesExample.source.relationships,
				{ id: "m-unsupported", source: "parent", target: "isolated", predicate: "contains" },
				{ id: "m-disabled-actor", source: "parent", target: "actor", predicate: "contains" },
				{
					id: "r-unsupported-summary",
					source: "isolated",
					target: "nested",
					predicate: "relates_to",
				},
				{ id: "r-disabled-actor-summary", source: "actor", target: "nested", predicate: "owns" },
			],
		};
		const projection = projectMap(
			source,
			{ detail: edgeCasesExample.detail, nearbyEntityIds: source.entities.map((entity) => entity.id) },
			edgeCasesExample.displayOptions
		);

		expect(projection.representativeByEntityId.has("isolated")).toBe(false);
		expect(projection.representativeByEntityId.has("actor")).toBe(false);
		expect(projection.nodes.some((node) => node.id === "actor")).toBe(false);
		expect(
			projection.connections.flatMap((connection) => connection.sourceRelationshipIds)
		).not.toContain("r-unsupported-summary");
		expect(
			projection.connections.flatMap((connection) => connection.sourceRelationshipIds)
		).not.toContain("r-disabled-actor-summary");
		expect(source.relationships.some((relationship) => relationship.id === "m-unsupported")).toBe(true);
		expect(source.relationships.some((relationship) => relationship.id === "m-disabled-actor")).toBe(
			true
		);
	});

	test("rejects cyclic enclosure assignments even with complete membership coverage", () => {
		const source = {
			...edgeCasesExample.source,
			coverage: { ...edgeCasesExample.source.coverage, parentMembership: Coverage.Complete },
		};
		const sourceBeforeProjection = structuredClone(source);
		const projection = projectMap(
			source,
			{ detail: edgeCasesExample.detail, nearbyEntityIds: source.entities.map((entity) => entity.id) },
			edgeCasesExample.displayOptions
		);

		expect(
			projection.nodes.some(
				(node) =>
					node.enclosure?.membershipId === "m-nested" || node.enclosure?.membershipId === "m-cycle"
			)
		).toBe(false);
		expect(
			source.relationships.filter(
				(relationship) => relationship.id === "m-nested" || relationship.id === "m-cycle"
			)
		).toHaveLength(2);
		expect(source).toEqual(sourceBeforeProjection);
	});

	test("uses category, not label or kind, for projection", () => {
		const relabeledSource = {
			...relationshipExample.source,
			entities: relationshipExample.source.entities.map((entity) => ({
				...entity,
				label: `renamed ${entity.id}`,
				kind: "unrelated-description",
			})),
		};
		const reveal = { detail: relationshipExample.detail, nearbyEntityIds: allEntityIds };
		const baseline = projectMap(relationshipExample.source, reveal, relationshipExample.displayOptions);
		const relabeled = projectMap(relabeledSource, reveal, relationshipExample.displayOptions);

		expect(relabeled.nodes).toEqual(baseline.nodes);
		expect(relabeled.connections).toEqual(baseline.connections);
	});

	test("accepts fractional reveal detail without treating it as a new category level", () => {
		const example = graphExamples.find((candidate) => candidate.name === "two groups share one member")!;
		const projection = projectMap(
			example.source,
			{ detail: 1.5, nearbyEntityIds: example.source.entities.map((entity) => entity.id) },
			example.displayOptions
		);

		expect(projection.nodes.map((node) => node.id)).toEqual([
			"root",
			"group-a",
			"group-b",
			"shared-resource",
		]);
	});

	test("uses nearby entities for local reveal while retaining their architectural ancestors", () => {
		const example = graphExamples.find((candidate) => candidate.name === "two groups share one member")!;
		const projection = projectMap(
			example.source,
			{ detail: NodeDetailLevel.Runtime, nearbyEntityIds: ["member-a"] },
			example.displayOptions
		);

		expect(projection.nodes.map((node) => node.id)).toEqual(["root", "group-a", "member-a"]);
		expect(projection.nodes.find((node) => node.id === "member-a")?.enclosure).toEqual({
			parentId: "group-a",
			membershipId: "m-a",
		});
	});
});
