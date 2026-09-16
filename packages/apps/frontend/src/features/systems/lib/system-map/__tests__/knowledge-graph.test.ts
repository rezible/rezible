import { describe, expect, test } from "bun:test";
import {
	categoryDisplay,
	NodeDetailLevel,
	DisplayMode,
	getMapCategoryDisplay,
	MapCategory,
	parseMapCategory,
} from "../category";
import { graphExamples } from "./test-fixtures";

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

// These checks validate example consistency. Projection behavior belongs in future projection tests.
describe("test-only graph examples", () => {
	for (const example of graphExamples) {
		test(`${example.name} has consistent source and expected references`, () => {
			const entities = new Map(example.source.entities.map((entity) => [entity.id, entity]));
			const facts = [...example.source.membership, ...example.source.relationships];
			const factsById = new Map(facts.map((fact) => [fact.id, fact]));
			const visibleIds = new Set(example.expected.visibleRepresentatives);
			expect(entities.size).toBe(example.source.entities.length);
			expect(factsById.size).toBe(facts.length);
			expect(visibleIds.size).toBe(example.expected.visibleRepresentatives.length);
			for (const fact of facts) {
				expect(entities.has(fact.source)).toBe(true);
				expect(entities.has(fact.target)).toBe(true);
			}
			for (const id of visibleIds) {
				expect(entities.has(id)).toBe(true);
				const category = entities.get(id)?.category ?? MapCategory.Unknown;
				expect(getMapCategoryDisplay(category).mode).toBe(DisplayMode.Node);
				if (category === MapCategory.Actor) expect(example.showActors).toBe(true);
			}
			for (const id of example.expected.enclosureMembershipIds ?? []) {
				const membership = factsById.get(id);
				expect(membership?.predicate).toBe("contains");
				expect(visibleIds.has(membership!.source)).toBe(true);
				expect(visibleIds.has(membership!.target)).toBe(true);
			}
			for (const exclusion of example.expected.excludedEnclosures ?? []) {
				expect(factsById.get(exclusion.membershipId)?.predicate).toBe("contains");
				expect(example.expected.enclosureMembershipIds ?? []).not.toContain(exclusion.membershipId);
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
			}
			for (const connection of example.expected.connections) {
				const [source, target] = connection.endpoints;
				expect(visibleIds.has(source)).toBe(true);
				expect(visibleIds.has(target)).toBe(true);
				expect(connection.count).toBeGreaterThan(0);
				expect(connection.count).toBe(connection.sourceRelationshipIds.length);
				expect(new Set(connection.sourceRelationshipIds).size).toBe(connection.count);
				for (const id of connection.sourceRelationshipIds) {
					const fact = factsById.get(id);
					expect(fact).toMatchObject({ predicate: connection.predicate });
					if (connection.classification === "direct") {
						expect(fact).toMatchObject({ source, target });
					}
				}
			}
			for (const annotation of example.expected.annotations) {
				expect(example.showAnnotations).toBe(true);
				const category = entities.get(annotation.entityId)?.category ?? MapCategory.Unknown;
				expect(getMapCategoryDisplay(category).mode).toBe(DisplayMode.Annotation);
				expect(factsById.get(annotation.relationshipId)).toMatchObject({
					source: annotation.entityId,
					target: annotation.originalTargetId,
				});
			}
		});
	}
});
