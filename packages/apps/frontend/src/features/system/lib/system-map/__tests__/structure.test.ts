import { describe, expect, test } from "bun:test";
import { regionOutline, type Box, type Point } from "../geometry";
import {
	layoutStructure,
	regionContentBox,
	REGION_LABEL_HEIGHT,
	REGION_PADDING,
	REGION_SHAPE_SCALE,
} from "../layout";
import {
	buildStructureTree,
	type StructureCard,
	type StructureNode,
	type StructureRegion,
} from "../structure";
import { largeFixture as largeFixtureGraph } from "$features/dev/views/system-map-fixture/large-fixture";
import { fullFixture as fullFixtureGraph } from "$features/dev/views/system-map-fixture/terrain-fixture";
import { withTestModel } from "./test-model";

const fullFixture = withTestModel(fullFixtureGraph);
const largeFixture = withTestModel(largeFixtureGraph);

type Outline = { id: string; children?: Outline[] };

const outline = (node: StructureNode): Outline => {
	if (node.type === "card") {
		return { id: node.id };
	}
	return { id: node.id, children: node.children.map(outline) };
};

const regions = (nodes: readonly StructureNode[]): StructureRegion[] =>
	nodes.flatMap((node) => (node.type === "region" ? [node, ...regions(node.children)] : []));

const cards = (nodes: readonly StructureNode[]): StructureCard[] =>
	nodes.flatMap((node) => (node.type === "card" ? [node] : cards(node.children)));

const overlaps = (a: Box, b: Box) =>
	a.x < b.x + b.width && b.x < a.x + a.width && a.y < b.y + b.height && b.y < a.y + a.height;

const corners = (box: Box): Point[] => [
	{ x: box.x, y: box.y },
	{ x: box.x + box.width, y: box.y },
	{ x: box.x, y: box.y + box.height },
	{ x: box.x + box.width, y: box.y + box.height },
];

/** Even-odd ray casting against the outline's polygon. */
const insidePolygon = (polygon: readonly Point[], point: Point) => {
	let inside = false;
	for (let index = 0, previous = polygon.length - 1; index < polygon.length; previous = index++) {
		const a = polygon[index];
		const b = polygon[previous];
		const crosses = a.y > point.y !== b.y > point.y;
		if (crosses && point.x < ((b.x - a.x) * (point.y - a.y)) / (b.y - a.y) + a.x) {
			inside = !inside;
		}
	}
	return inside;
};

describe("system map structure", () => {
	test("nests regions to any depth, from parentId, with parents outside the subset as roots", () => {
		const fulfilment = Array.from({ length: 12 }, (_, index) => ({
			id: `rt-fulfilment-${String(index + 1).padStart(2, "0")}`,
		}));

		expect(buildStructureTree(fullFixture).roots.map(outline)).toEqual([
			{
				id: "fn-discovery",
				children: [{ id: "sys-search", children: [{ id: "rt-search-api" }] }, { id: "rt-crawler" }],
			},
			{
				id: "fn-commerce",
				children: [
					{
						id: "sys-commerce-core",
						children: [
							{
								id: "sys-payments",
								children: [{ id: "rt-payment-service" }, { id: "rt-payment-worker" }],
							},
							// A container in a container is its own card, in the nearest region.
							{ id: "rt-checkout" },
							{ id: "rt-checkout-proxy" },
						],
					},
					{ id: "sys-fulfilment", children: fulfilment },
				],
			},
			{ id: "rt-legacy-cron" },
			{ id: "rt-reports" },
		]);
	});

	test("a card names its parent container, its other claims and its component count", () => {
		const byId = new Map(cards(buildStructureTree(fullFixture).roots).map((card) => [card.id, card]));

		expect(byId.get("rt-checkout-proxy")!.partOf?.id).toBe("rt-checkout");
		expect(byId.get("rt-checkout")!.partOf).toBeUndefined();
		expect(byId.get("rt-search-api")!.otherParentIds).toEqual(["sys-commerce-core"]);
		expect(byId.get("rt-checkout")!.implementationCount).toBe(2);
	});
});

describe("system map layout", () => {
	test("keeps children inside their region's content area without overlapping siblings", () => {
		const tree = buildStructureTree(fullFixture);
		const layout = layoutStructure(tree);
		const fulfilment = regions(tree.roots).find((region) => region.id === "sys-fulfilment")!;
		const parent = regionContentBox(layout.boxes.get(fulfilment.id)!);
		const children = fulfilment.children.map((child) => layout.boxes.get(child.id)!);

		expect(children).toHaveLength(12);
		for (const child of children) {
			expect(child.x).toBeGreaterThanOrEqual(parent.x + REGION_PADDING);
			expect(child.y).toBeGreaterThanOrEqual(parent.y + REGION_PADDING + REGION_LABEL_HEIGHT);
			expect(child.x + child.width).toBeLessThanOrEqual(parent.x + parent.width - REGION_PADDING);
			expect(child.y + child.height).toBeLessThanOrEqual(parent.y + parent.height - REGION_PADDING);
		}
		for (const [index, child] of children.entries()) {
			for (const sibling of children.slice(index + 1)) {
				expect(overlaps(child, sibling)).toBe(false);
			}
		}
	});

	test("keeps a region's padded content inside its drawn outline at every aspect ratio", () => {
		const paddedSizes = [
			[400, 400],
			[1200, 400],
			[400, 1200],
			[3200, 400],
			[12800, 800],
		];
		const ids = ["fn-commerce", "sys-commerce-core", "sys-payments", "sys-search"];

		for (const [width, height] of paddedSizes) {
			const box = {
				x: 0,
				y: 0,
				width: width * REGION_SHAPE_SCALE,
				height: height * REGION_SHAPE_SCALE,
			};
			for (const id of ids) {
				const outline = regionOutline(box, id);
				for (const corner of corners(regionContentBox(box))) {
					expect(insidePolygon(outline, corner)).toBe(true);
				}
			}
		}
	});

	test("keeps every child's corners inside its parent's drawn outline", () => {
		const tree = buildStructureTree(largeFixture);
		const layout = layoutStructure(tree);
		const parents = regions(tree.roots);

		expect(parents.length).toBeGreaterThan(10);
		for (const parent of parents) {
			const outline = regionOutline(layout.boxes.get(parent.id)!, parent.id);
			for (const child of parent.children) {
				for (const corner of corners(layout.boxes.get(child.id)!)) {
					expect(insidePolygon(outline, corner)).toBe(true);
				}
			}
		}
	});
});
