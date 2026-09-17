import { describe, expect, test } from "bun:test";

import { NodeDetailLevel } from "../category";
import {
	buildInspectionItems,
	inspectionItemId,
	inspectionTargetAvailable,
	resolveInspectionTarget,
} from "../inspection";
import { buildInspectionViewData } from "$features/systems/components/system-map/map-inspector/presentation";
import { projectMap } from "../projection";
import { edgeCasesExample, relationshipExample } from "./test-fixtures";

describe("system map source inspection", () => {
	test("keeps source inspection on the latest supplied graph when a layout replacement fails", () => {
		const oldSource = edgeCasesExample.source;
		const newSource = {
			...relationshipExample.source,
			entities: [
				...relationshipExample.source.entities,
				{ id: "new-subject", category: "unknown", label: "New subject", kind: "unknown" },
			],
		};
		const oldProjection = projectMap(
			oldSource,
			{ detail: NodeDetailLevel.Landscape, nearbyEntityIds: [] },
			edgeCasesExample.displayOptions
		);
		const newProjection = projectMap(
			newSource,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: [] },
			relationshipExample.displayOptions
		);
		const oldTarget = { kind: "entity" as const, entityId: "isolated" };
		const newTarget = { kind: "entity" as const, entityId: "new-subject" };
		const oldRelationshipTarget = { kind: "relationship" as const, relationshipId: "r-event" };
		const newRelationshipTarget = { kind: "relationship" as const, relationshipId: "r-direct" };

		expect(inspectionTargetAvailable(oldSource, oldTarget)).toBe(true);
		expect(inspectionTargetAvailable(newSource, oldTarget)).toBe(false);
		expect(inspectionTargetAvailable(newSource, newTarget)).toBe(true);
		expect(inspectionTargetAvailable(newSource, oldRelationshipTarget)).toBe(false);
		expect(inspectionTargetAvailable(newSource, newRelationshipTarget)).toBe(true);
		expect(
			buildInspectionItems(newSource, newProjection).some((item) => item.id === "entity:new-subject")
		).toBe(true);
		expect(resolveInspectionTarget(newSource, newProjection, newTarget).entity?.id).toBe("new-subject");
		expect(resolveInspectionTarget(oldSource, oldProjection, oldTarget).entity?.id).toBe("isolated");
	});

	test("keeps original endpoint pairs for a summary", () => {
		const source = relationshipExample.source;
		const projection = projectMap(
			source,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: source.entities.map((entity) => entity.id) },
			relationshipExample.displayOptions
		);
		const summary = projection.connections.find(
			(connection) => connection.classification === "summary" && connection.predicate === "calls"
		);
		expect(summary).toBeDefined();

		const inspection = resolveInspectionTarget(source, projection, {
			kind: "summary",
			sourceRelationshipIds: summary!.sourceRelationshipIds,
		});

		expect(
			inspection.summary?.relationships.map(({ relationship }) => [
				relationship.id,
				relationship.source,
				relationship.target,
			])
		).toEqual([
			["r-one", "member-a", "member-b"],
			["r-two", "member-a2", "member-b"],
		]);
	});

	test("inspects hidden and unsupported subjects from the supplied graph", () => {
		const source = edgeCasesExample.source;
		const projection = projectMap(
			source,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: source.entities.map((entity) => entity.id) },
			edgeCasesExample.displayOptions
		);
		const inspection = resolveInspectionTarget(source, projection, {
			kind: "entity",
			entityId: "isolated",
		});

		expect(inspection.entity?.id).toBe("isolated");
		expect(inspection.visibility).toBe("not-rendered");
		expect(inspection.representativeId).toBeUndefined();
	});

	test("preserves annotation source and original target identity", () => {
		const source = edgeCasesExample.source;
		const projection = projectMap(
			source,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: source.entities.map((entity) => entity.id) },
			edgeCasesExample.displayOptions
		);
		const annotation = projection.annotations[0];
		const inspection = resolveInspectionTarget(source, projection, {
			kind: "annotation",
			entityId: annotation.entityId,
			relationshipId: annotation.relationshipId,
		});

		expect(inspection.annotation?.entity?.id).toBe("event");
		expect(inspection.annotation?.originalTarget?.id).toBe("child");
		expect(inspection.annotation?.relationship?.relationship.id).toBe("r-event");
	});

	test("prepares readable relationship and annotation inspection values", () => {
		const relationshipSource = relationshipExample.source;
		const relationshipProjection = projectMap(
			relationshipSource,
			{
				detail: NodeDetailLevel.Systems,
				nearbyEntityIds: relationshipSource.entities.map((entity) => entity.id),
			},
			relationshipExample.displayOptions
		);
		const relationshipInspection = resolveInspectionTarget(relationshipSource, relationshipProjection, {
			kind: "relationship",
			relationshipId: "r-direct",
		});
		const relationshipView = buildInspectionViewData(relationshipInspection);

		expect(relationshipView?.title).toBe("calls");
		expect(relationshipView?.visibilityText).toBe("Visible");
		expect(relationshipView?.relationship?.text).toBe("group-a → calls → group-b");
		expect(inspectionItemId(relationshipInspection.target)).toBe("relationship:r-direct");

		const annotationSource = edgeCasesExample.source;
		const annotationProjection = projectMap(
			annotationSource,
			{
				detail: NodeDetailLevel.Systems,
				nearbyEntityIds: annotationSource.entities.map((entity) => entity.id),
			},
			edgeCasesExample.displayOptions
		);
		const annotation = resolveInspectionTarget(annotationSource, annotationProjection, {
			kind: "annotation",
			entityId: "event",
			relationshipId: "r-event",
		});

		expect(buildInspectionViewData(annotation)?.annotationTargetLabel).toBe("child");
	});

	test("lists every supplied entity and relationship, including non-rendered facts", () => {
		const source = edgeCasesExample.source;
		const projection = projectMap(
			source,
			{ detail: NodeDetailLevel.Landscape, nearbyEntityIds: [] },
			edgeCasesExample.displayOptions
		);
		const items = buildInspectionItems(source, projection);

		expect(items.some((item) => item.id === "entity:isolated")).toBe(true);
		expect(items.some((item) => item.id === "relationship:m-cycle")).toBe(true);
		expect(items.find((item) => item.id === "annotation:event:r-event")?.detail).toStartWith(
			"hidden annotation"
		);
	});
});
