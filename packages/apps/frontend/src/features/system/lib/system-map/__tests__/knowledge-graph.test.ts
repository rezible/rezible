import { describe, expect, test } from "bun:test";
import type { GraphEntity, GraphRelationship, GraphSubset } from "../graph";
import type { MapSelection } from "../presentation";
import { reconcileSelection, selectedEntityIds, selectionIsAvailable } from "../selection";

const graph = (
	entities: readonly GraphEntity[],
	relationships: readonly GraphRelationship[],
	unresolvedRelationships: readonly GraphRelationship[] = [],
	stopReason: GraphSubset["enumeration"]["stopReason"] = "exhausted"
): GraphSubset => ({
	entities,
	relationships,
	unresolvedRelationships,
	enumeration: { scope: "fixture", stopReason },
});

describe("neutral host selection helpers", () => {
	test("availability includes unresolved source relationships", () => {
		const source = graph(
			[{ id: "present", category: "system", kind: "system" }],
			[],
			[{ id: "unresolved", source: "present", target: "missing", predicate: "contains" }]
		);
		const selection: MapSelection = { kind: "relationship", relationshipId: "unresolved" };

		expect(selectionIsAvailable(source, selection)).toBe(true);
		expect(selectedEntityIds(source, selection)).toEqual(["present"]);
		expect(reconcileSelection(source, selection)).toBe(selection);
	});

	test("summary reconciliation retains every original ID while any contributor remains", () => {
		const source = graph(
			[
				{ id: "a", category: "system", kind: "A" },
				{ id: "b", category: "system", kind: "B" },
			],
			[{ id: "still-present", source: "a", target: "b", predicate: "calls" }]
		);
		const selection: MapSelection = {
			kind: "summary",
			relationshipIds: ["still-present", "no-longer-supplied"],
		};

		expect(reconcileSelection(source, selection)).toEqual(selection);
		expect(reconcileSelection(graph(source.entities, []), selection)).toBeUndefined();
		expect(reconcileSelection(source, { kind: "entity", entityId: "missing" })).toBeUndefined();
	});
});
