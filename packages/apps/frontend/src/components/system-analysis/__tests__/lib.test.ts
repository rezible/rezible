import { describe, expect, test } from "bun:test";
import type { SystemAnalysisEdge, SystemAnalysisEntry, SystemAnalysisNode } from "$lib/api";
import { buildAnalysisGraph, mapEntryAttachments } from "../lib";

const node = (
	id: string,
	entityId: string,
	position: { x: number; y: number },
	category = "system",
	kind = "service"
): SystemAnalysisNode =>
	({
		id,
		attributes: {
			hidden: false,
			position,
			knowledgeEntity: {
				id: entityId,
				attributes: { category, kind },
			},
		},
	}) as unknown as SystemAnalysisNode;

const edge = (
	id: string,
	relationshipId: string,
	sourceEntityId: string,
	targetEntityId: string,
	predicate = "calls"
): SystemAnalysisEdge =>
	({
		id,
		attributes: {
			hidden: false,
			knowledgeRelationship: {
				id: relationshipId,
				attributes: { sourceEntityId, targetEntityId, predicate },
			},
		},
	}) as unknown as SystemAnalysisEdge;

const entry = (
	id: string,
	subjects: Array<{ knowledgeEntityId?: string; knowledgeRelationshipId?: string }>
): SystemAnalysisEntry =>
	({
		id,
		attributes: {
			subjects: subjects.map((attributes, index) => ({
				id: `${id}-subject-${index}`,
				attributes: { ...attributes, role: "subject" },
			})),
		},
	}) as unknown as SystemAnalysisEntry;

describe("analysis graph adaptation", () => {
	test("uses canonical IDs and sorts duplicate memberships by analysis record ID", () => {
		const model = buildAnalysisGraph(
			[
				node("node-z", "entity-1", { x: 90, y: 80 }, "component", "worker"),
				node("node-a", "entity-1", { x: 10, y: 20 }, "system", "service"),
				node("node-b", "entity-2", { x: 40, y: 50 }),
			],
			[
				edge("edge-b", "relationship-1", "entity-1", "missing", "reads_from"),
				edge("edge-a", "relationship-1", "entity-1", "entity-2", "calls"),
			]
		);

		expect(model.graph.entities).toEqual([
			{ id: "entity-1", category: "system", kind: "service" },
			{ id: "entity-2", category: "system", kind: "service" },
		]);
		expect(model.graph.relationships).toEqual([
			{ id: "relationship-1", source: "entity-1", target: "entity-2", predicate: "calls" },
		]);
		expect(model.graph.unresolvedRelationships).toEqual([]);
		expect(model.positions).toEqual({
			"entity-1": { x: 10, y: 20 },
			"entity-2": { x: 40, y: 50 },
		});
		expect(model.nodeByEntityId.get("entity-1")?.id).toBe("node-a");
		expect(model.edgeByRelationshipId.get("relationship-1")?.id).toBe("edge-a");
		expect(model.duplicateEntityIds).toEqual(["entity-1"]);
		expect(model.duplicateRelationshipIds).toEqual(["relationship-1"]);
		expect(model.graph.enumeration).toEqual({ scope: "analysis", stopReason: "exhausted" });
	});

	test("partitions relationships with unavailable endpoints and accepts an empty completed analysis", () => {
		const model = buildAnalysisGraph(
			[node("node-1", "entity-1", { x: 10, y: 20 })],
			[
				edge("edge-present", "resolved", "entity-1", "entity-1"),
				edge("edge-missing", "unresolved", "missing", "entity-1"),
			]
		);

		expect(model.graph.relationships.map(({ id }) => id)).toEqual(["resolved"]);
		expect(model.graph.unresolvedRelationships.map(({ id }) => id)).toEqual(["unresolved"]);
		expect(buildAnalysisGraph([], []).graph).toEqual({
			entities: [],
			relationships: [],
			unresolvedRelationships: [],
			enumeration: { scope: "analysis", stopReason: "exhausted" },
		});
	});

	test("maps entry attachments to canonical IDs and deduplicates subjects per entry", () => {
		const model = buildAnalysisGraph(
			[node("node-1", "entity-1", { x: 10, y: 20 }), node("node-2", "entity-2", { x: 30, y: 40 })],
			[edge("edge-1", "relationship-1", "entity-1", "entity-2")]
		);
		const attached = entry("entry-1", [
			{ knowledgeEntityId: "entity-1" },
			{ knowledgeEntityId: "entity-1" },
			{ knowledgeRelationshipId: "relationship-1" },
		]);
		const unavailable = entry("entry-2", [{ knowledgeEntityId: "not-in-analysis" }]);
		const attachments = mapEntryAttachments(model.nodeByEntityId, model.edgeByRelationshipId, [
			attached,
			unavailable,
		]);

		expect(attachments.byEntityId.get("entity-1")).toEqual([attached]);
		expect(attachments.byRelationshipId.get("relationship-1")).toEqual([attached]);
		expect(attachments.unattached).toEqual([unavailable]);
	});
});
