import { describe, expect, test } from "bun:test";
import { regionOutline, type Box, type Point } from "../geometry";
import {
	CARD_HEIGHT,
	CARD_WIDTH,
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
import { largeFixture } from "$features/dev/views/system-map-fixture/large-fixture";
import { fullFixture, mainFixture } from "$features/dev/views/system-map-fixture/terrain-fixture";

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
	test("places entities under their primary parents and orphans in the nearest drawn ancestor or the root", () => {
		const tree = buildStructureTree(fullFixture);
		const fulfilment = Array.from({ length: 12 }, (_, index) => ({
			id: `rt-fulfilment-${String(index + 1).padStart(2, "0")}`,
		}));

		expect(tree.roots.map(outline)).toEqual([
			{
				id: "fn-commerce",
				children: [
					{
						id: "sys-checkout",
						children: [
							{ id: "rt-checkout-api" },
							{ id: "rt-checkout-db" },
							{ id: "rt-ledger-db" },
						],
					},
					{ id: "sys-fulfilment", children: fulfilment },
					{
						id: "sys-payments",
						children: [{ id: "rt-payment-service" }, { id: "rt-payment-worker" }],
					},
				],
			},
			{
				id: "fn-discovery",
				// A runtime whose parent is a landscape sits directly in it, after its systems.
				children: [
					{ id: "sys-search", children: [{ id: "rt-search-api" }, { id: "rt-search-index" }] },
					{ id: "rt-crawler" },
				],
			},
			{
				id: "fn-platform",
				children: [
					{
						id: "sys-compute",
						children: [{ id: "rt-production-cluster" }, { id: "rt-standby-cluster" }],
					},
				],
			},
			{ id: "sys-notifications", children: [] },
			{ id: "rt-legacy-cron" },
		]);
	});

	test("places a runtime with several parents once, in the parent with the lowest id", () => {
		const tree = buildStructureTree(fullFixture);
		const ledger = cards(tree.roots).filter((card) => card.id === "rt-ledger-db");

		expect(ledger).toHaveLength(1);
		expect(tree.primaryParents.get("rt-ledger-db")).toBe("sys-checkout");
		expect(ledger[0].alsoIn).toEqual(["sys-payments"]);
	});

	test("counts each implementation entity once, under its runtime ancestor", () => {
		const counts = Object.fromEntries(
			cards(buildStructureTree(fullFixture).roots).map((card) => [card.id, card.implementationCount])
		);

		expect(counts["rt-checkout-api"]).toBe(8);
		expect(Object.values(counts).reduce((total, count) => total + count, 0)).toBe(8);
	});
});

describe("system map layout", () => {
	test("packs the example tree into the worked-example boxes", () => {
		const layout = layoutStructure(buildStructureTree(mainFixture));
		const box = (id: string) => layout.boxes.get(id);

		expect(box("fn-commerce")).toEqual({ x: 0, y: 0, width: 1705, height: 517.5 });
		expect(box("fn-discovery")).toEqual({ x: 1801, y: 0, width: 867.5, height: 517.5 });
		expect(box("sys-checkout")).toEqual({ x: 202.5, y: 123.75, width: 630, height: 310 });
		expect(box("sys-payments")).toEqual({ x: 872.5, y: 123.75, width: 630, height: 310 });
		expect(box("sys-search")).toEqual({ x: 1919.75, y: 123.75, width: 630, height: 310 });
		expect(box("rt-checkout-api")).toEqual({
			x: 297.5,
			y: 226.75,
			width: CARD_WIDTH,
			height: CARD_HEIGHT,
		});
		expect({ width: layout.width, height: layout.height }).toEqual({ width: 2668.5, height: 517.5 });
	});

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
		const ids = ["fn-commerce", "sys-checkout", "sys-payments", "sys-search"];

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
