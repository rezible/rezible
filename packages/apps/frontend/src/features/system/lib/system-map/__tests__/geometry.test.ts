import { describe, expect, test } from "bun:test";
import { connectionAnchors, regionPath, type Box, type ConnectionAnchor } from "../geometry";

const onSide = (box: Box, anchor: ConnectionAnchor) => {
	if (anchor.side === "left" || anchor.side === "right") {
		const x = anchor.side === "left" ? box.x : box.x + box.width;
		return anchor.x === x && anchor.y >= box.y && anchor.y <= box.y + box.height;
	}
	const y = anchor.side === "top" ? box.y : box.y + box.height;
	return anchor.y === y && anchor.x >= box.x && anchor.x <= box.x + box.width;
};

describe("connection anchors", () => {
	test("separates an opposing diagonal pair while keeping every end on its box edge", () => {
		const a: Box = { x: 0, y: 0, width: 200, height: 72 };
		const b: Box = { x: 300, y: 200, width: 200, height: 72 };
		const forward = connectionAnchors(a, b, true);
		const backward = connectionAnchors(b, a, true);

		expect(onSide(a, forward.source) && onSide(b, forward.target)).toBe(true);
		expect(onSide(b, backward.source) && onSide(a, backward.target)).toBe(true);
		expect(forward.source).not.toEqual(backward.target);
		expect(forward.target).not.toEqual(backward.source);
	});

	test("shifts the labels of an opposing pair to opposite sides and leaves a lone connection's in place", () => {
		const a: Box = { x: 0, y: 0, width: 200, height: 72 };
		const b: Box = { x: 300, y: 0, width: 200, height: 72 };
		const forward = connectionAnchors(a, b, true).separation;
		const backward = connectionAnchors(b, a, true).separation;

		expect(forward).not.toEqual({ x: 0, y: 0 });
		expect([forward.x + backward.x, forward.y + backward.y]).toEqual([0, 0]);
		expect(connectionAnchors(a, b, false).separation).toEqual({ x: 0, y: 0 });
	});
});

describe("region path", () => {
	test("draws the same outline for the same id and a different one for another id", () => {
		const box: Box = { x: 0, y: 0, width: 600, height: 300 };

		expect(regionPath(box, "sys-checkout")).toBe(regionPath(box, "sys-checkout"));
		expect(regionPath(box, "sys-checkout")).not.toBe(regionPath(box, "sys-payments"));
	});
});
