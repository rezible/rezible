import { createSystemMapLayoutEngine, type LayoutResult } from "$src/features/systems/components/system-map";
import { NodeDetailLevel, MapCategory } from "../category";
import { Coverage, type GraphEntity, type GraphRelationship, type GraphSubset } from "../graph";
import type { MapConnection, MapDisplayOptions, MapProjection } from "../presentation";

/** Convenience boundary for pure layout tests; production controllers retain one engine. */
export const layoutProjection = async (graph: GraphSubset, projection: MapProjection): Promise<LayoutResult> => {
	const engine = createSystemMapLayoutEngine();
	try {
		return await engine.layout(graph, projection);
	} finally {
		engine.dispose();
	}
};

export type ExpectedConnection = Omit<MapConnection, "id"> & { count: number };

export type GraphExample = {
	name: string;
	detail: number;
	displayOptions: MapDisplayOptions;
	source: GraphSubset;
	expected: {
		visibleRepresentatives: readonly string[];
		sharedMembership: readonly { parentId: string; sharedId: string; membershipId: string }[];
		connections: readonly ExpectedConnection[];
		annotations: readonly { entityId: string; relationshipId: string; originalTargetId: string }[];
		unsupportedIds?: readonly string[];
		// Memberships used for enclosure, not additional source relationships.
		enclosureMembershipIds?: readonly string[];
		excludedEnclosures?: readonly { membershipId: string; reason: "cycle" | "partial coverage" }[];
	};
};

const entity = (id: string, category: MapCategory, label = id, kind = "subject"): GraphEntity => ({
	id,
	category,
	label,
	kind,
});

const fact = (
	id: string,
	source: string,
	target: string,
	predicate: string,
	supportReferences?: readonly string[]
): GraphRelationship => ({
	id,
	source,
	target,
	predicate,
	supportReferences,
});

export const sharedGroupsExample: GraphExample = {
	name: "two groups share one member",
	detail: NodeDetailLevel.Systems,
	displayOptions: { showActors: false, showAnnotations: false },
	source: {
		coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
		entities: [
			entity("root", MapCategory.Function),
			entity("group-a", MapCategory.System),
			entity("group-b", MapCategory.System),
			entity("member-a", MapCategory.Container),
			entity("member-b", MapCategory.Container),
			entity("shared-resource", MapCategory.Container),
			entity("code", MapCategory.Code),
		],
		relationships: [
			fact("m-root-a", "root", "group-a", "contains"),
			fact("m-root-b", "root", "group-b", "contains"),
			fact("m-a", "group-a", "member-a", "contains"),
			fact("m-b", "group-b", "member-b", "contains"),
			fact("m-shared-a", "group-a", "shared-resource", "contains"),
			fact("m-shared-b", "group-b", "shared-resource", "contains"),
			fact("m-code", "member-a", "code", "contains"),
		],
	},
	expected: {
		visibleRepresentatives: ["root", "group-a", "group-b", "shared-resource"],
		enclosureMembershipIds: ["m-root-a", "m-root-b"],
		sharedMembership: [
			{ parentId: "group-a", sharedId: "shared-resource", membershipId: "m-shared-a" },
			{ parentId: "group-b", sharedId: "shared-resource", membershipId: "m-shared-b" },
		],
		connections: [],
		annotations: [],
	},
};

export const relationshipExample: GraphExample = {
	name: "direct and summary connection counts",
	detail: NodeDetailLevel.Systems,
	displayOptions: { showActors: false, showAnnotations: false },
	source: {
		coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
		entities: [
			entity("group-a", MapCategory.System),
			entity("group-b", MapCategory.System),
			entity("member-a", MapCategory.Container),
			entity("member-a2", MapCategory.Container),
			entity("member-b", MapCategory.Container),
		],
		relationships: [
			fact("m-a", "group-a", "member-a", "contains"),
			fact("m-a2", "group-a", "member-a2", "contains"),
			fact("m-b", "group-b", "member-b", "contains"),

			fact("r-direct", "group-a", "group-b", "calls"),
			fact("r-one", "member-a", "member-b", "calls", ["evidence:one", "evidence:two"]),
			fact("r-two", "member-a2", "member-b", "calls"),
			fact("r-reverse", "member-b", "member-a", "calls"),
			fact("r-other", "member-a", "member-b", "depends_on"),
		],
	},
	expected: {
		visibleRepresentatives: ["group-a", "group-b"],
		sharedMembership: [],
		connections: [
			{
				endpoints: ["group-a", "group-b"],
				predicate: "calls",
				classification: "direct",
				sourceRelationshipIds: ["r-direct"],
				count: 1,
			},
			{
				endpoints: ["group-a", "group-b"],
				predicate: "calls",
				classification: "summary",
				sourceRelationshipIds: ["r-one", "r-two"],
				count: 2,
			},
			{
				endpoints: ["group-b", "group-a"],
				predicate: "calls",
				classification: "summary",
				sourceRelationshipIds: ["r-reverse"],
				count: 1,
			},
			{
				endpoints: ["group-a", "group-b"],
				predicate: "depends_on",
				classification: "summary",
				sourceRelationshipIds: ["r-other"],
				count: 1,
			},
		],
		annotations: [],
	},
};

export const edgeCasesExample: GraphExample = {
	name: "partial graph with cyclic membership and annotation context",
	detail: NodeDetailLevel.Systems,
	displayOptions: { showActors: false, showAnnotations: true },
	source: {
		// Parent membership coverage is incomplete for every entity in this example.
		coverage: { parentMembership: Coverage.Partial, relationships: Coverage.Complete },
		entities: [
			entity("parent", MapCategory.System),
			entity("child", MapCategory.Container),
			entity("nested", MapCategory.System),
			entity("isolated", MapCategory.Unknown),
			entity("actor", MapCategory.Actor),
			entity("event", MapCategory.Event),
			entity("code", MapCategory.Code),
		],
		relationships: [
			fact("m-child", "parent", "child", "contains"),
			fact("m-nested", "parent", "nested", "contains"),
			fact("m-code", "child", "code", "contains"),
			fact("m-cycle", "nested", "parent", "contains"),

			fact("r-actor", "actor", "child", "owns"),
			fact("r-event", "event", "child", "impacts", ["evidence:one", "evidence:two"]),
		],
	},
	expected: {
		visibleRepresentatives: ["parent", "nested"],
		sharedMembership: [],
		connections: [],
		annotations: [{ entityId: "event", relationshipId: "r-event", originalTargetId: "child" }],
		unsupportedIds: ["isolated"],
		enclosureMembershipIds: [],
		// Reject both cycle directions; partial membership cannot establish exclusive enclosure.
		excludedEnclosures: [
			{ membershipId: "m-nested", reason: "cycle" },
			{ membershipId: "m-cycle", reason: "cycle" },
			{ membershipId: "m-child", reason: "partial coverage" },
			{ membershipId: "m-code", reason: "partial coverage" },
		],
	},
};

// Reuse one source hierarchy to describe reveal at all four levels.
export const graphExamples: readonly GraphExample[] = [
	{
		...sharedGroupsExample,
		name: "hierarchy at landscape detail",
		detail: NodeDetailLevel.Landscape,
		expected: {
			visibleRepresentatives: ["root"],
			enclosureMembershipIds: [],
			sharedMembership: [],
			connections: [],
			annotations: [],
		},
	},
	sharedGroupsExample,
	{
		...sharedGroupsExample,
		name: "hierarchy at runtime detail",
		detail: NodeDetailLevel.Runtime,
		expected: {
			...sharedGroupsExample.expected,
			visibleRepresentatives: ["root", "group-a", "group-b", "member-a", "member-b", "shared-resource"],
			enclosureMembershipIds: ["m-root-a", "m-root-b", "m-a", "m-b"],
		},
	},
	{
		...sharedGroupsExample,
		name: "hierarchy at implementation detail",
		detail: NodeDetailLevel.Implementation,
		expected: {
			...sharedGroupsExample.expected,
			visibleRepresentatives: [
				"root",
				"group-a",
				"group-b",
				"member-a",
				"member-b",
				"shared-resource",
				"code",
			],
			enclosureMembershipIds: ["m-root-a", "m-root-b", "m-a", "m-b", "m-code"],
		},
	},
	relationshipExample,
	edgeCasesExample,
];
