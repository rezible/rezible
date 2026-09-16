import { NodeDetailLevel, MapCategory } from "../category";

export type TestEntity = {
	id: string;
	category: MapCategory;
	label: string;
	kind: string;
};

export type TestFact = {
	id: string;
	source: string;
	target: string;
	predicate: string;
	supportReferences?: readonly string[];
};

export type ExpectedConnection = {
	endpoints: readonly [string, string];
	predicate: string;
	classification: "direct" | "summary";
	sourceRelationshipIds: readonly string[];
	count: number;
};

export type GraphExample = {
	name: string;
	detailLevel: NodeDetailLevel;
	showActors: boolean;
	showAnnotations: boolean;
	source: {
		entities: readonly TestEntity[];
		membership: readonly TestFact[];
		relationships: readonly TestFact[];
		coverage?: "complete" | "partial";
	};
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

const entity = (id: string, category: MapCategory, label = id, kind = "subject"): TestEntity => ({
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
): TestFact => ({
	id,
	source,
	target,
	predicate,
	supportReferences,
});

export const sharedGroupsExample: GraphExample = {
	name: "two groups share one member",
	detailLevel: NodeDetailLevel.Systems,
	showActors: false,
	showAnnotations: false,
	source: {
		coverage: "complete",
		entities: [
			entity("root", MapCategory.Function),
			entity("group-a", MapCategory.System),
			entity("group-b", MapCategory.System),
			entity("member-a", MapCategory.Container),
			entity("member-b", MapCategory.Container),
			entity("shared-resource", MapCategory.Container),
			entity("code", MapCategory.Code),
		],
		membership: [
			fact("m-root-a", "root", "group-a", "contains"),
			fact("m-root-b", "root", "group-b", "contains"),
			fact("m-a", "group-a", "member-a", "contains"),
			fact("m-b", "group-b", "member-b", "contains"),
			fact("m-shared-a", "group-a", "shared-resource", "contains"),
			fact("m-shared-b", "group-b", "shared-resource", "contains"),
			fact("m-code", "member-a", "code", "contains"),
		],
		relationships: [],
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
	detailLevel: NodeDetailLevel.Systems,
	showActors: false,
	showAnnotations: false,
	source: {
		entities: [
			entity("group-a", MapCategory.System),
			entity("group-b", MapCategory.System),
			entity("member-a", MapCategory.Container),
			entity("member-a2", MapCategory.Container),
			entity("member-b", MapCategory.Container),
		],
		membership: [
			fact("m-a", "group-a", "member-a", "contains"),
			fact("m-a2", "group-a", "member-a2", "contains"),
			fact("m-b", "group-b", "member-b", "contains"),
		],
		relationships: [
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
	detailLevel: NodeDetailLevel.Systems,
	showActors: false,
	showAnnotations: true,
	source: {
		// Parent membership coverage is incomplete for every entity in this example.
		coverage: "partial",
		entities: [
			entity("parent", MapCategory.System),
			entity("child", MapCategory.Container),
			entity("nested", MapCategory.System),
			entity("isolated", MapCategory.Unknown),
			entity("actor", MapCategory.Actor),
			entity("event", MapCategory.Event),
			entity("code", MapCategory.Code),
		],
		membership: [
			fact("m-child", "parent", "child", "contains"),
			fact("m-nested", "parent", "nested", "contains"),
			fact("m-code", "child", "code", "contains"),
			fact("m-cycle", "nested", "parent", "contains"),
		],
		relationships: [
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
		detailLevel: NodeDetailLevel.Landscape,
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
		detailLevel: NodeDetailLevel.Runtime,
		expected: {
			...sharedGroupsExample.expected,
			visibleRepresentatives: ["root", "group-a", "group-b", "member-a", "member-b", "shared-resource"],
			enclosureMembershipIds: ["m-root-a", "m-root-b", "m-a", "m-b"],
		},
	},
	{
		...sharedGroupsExample,
		name: "hierarchy at implementation detail",
		detailLevel: NodeDetailLevel.Implementation,
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
