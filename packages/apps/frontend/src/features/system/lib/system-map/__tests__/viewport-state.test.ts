import { describe, expect, test } from "bun:test";

import { MapCategory, NodeDetailLevel } from "../category";
import { worldBoundsByNodeId, worldCenter, type Point, type Viewport } from "../geometry";
import { structuralDetailForZoom } from "../interaction";
import { MapViewportState } from "$features/system/components/system-map/viewport-state";
import type { FlowNode } from "$features/system/components/system-map/flow-model";

const node = (id: string, x: number, y: number): FlowNode => ({
	id,
	position: { x, y } satisfies Point,
	width: 160,
	height: 68,
	data: {
		entity: { id, category: MapCategory.System, kind: "system" },
		appearance: "compact",
	},
	type: "system-map-node",
});

describe("system map viewport state", () => {
	test("keeps camera zoom through same-detail pan updates", () => {
		const state = new MapViewportState(NodeDetailLevel.Landscape);
		const programmaticViewport: Viewport = { x: -150, y: 72, zoom: 1.25 };
		state.updateViewport(programmaticViewport);

		const afterFirstPan: Viewport = { x: -94, y: 118, zoom: programmaticViewport.zoom };
		expect(state.updateViewport(afterFirstPan)).toBe(false);
		expect(state.lastProcessedViewport).toEqual(afterFirstPan);
		expect(state.continuousDetail).toBeGreaterThan(1.9);
	});

	test("zooms keyboard camera around the selected representation", () => {
		const state = new MapViewportState(NodeDetailLevel.Landscape);
		const bounds = worldBoundsByNodeId([node("selected", 100, 80)]).get("selected")!;
		const before = {
			x: worldCenter(bounds).x * state.lastProcessedViewport.zoom + state.lastProcessedViewport.x,
			y: worldCenter(bounds).y * state.lastProcessedViewport.zoom + state.lastProcessedViewport.y,
		};
		const next = state.prepareKeyboardZoom(1.2, bounds, { width: 800, height: 600 });

		expect(next.zoom).toBeCloseTo(state.lastProcessedViewport.zoom * 1.2);
		expect(before.x - next.x).toBeCloseTo(worldCenter(bounds).x * next.zoom);
		expect(before.y - next.y).toBeCloseTo(worldCenter(bounds).y * next.zoom);
	});

	test("Reveal raises visible detail while leaving its target in the same screen area", () => {
		const state = new MapViewportState(NodeDetailLevel.Landscape);
		const transition = state.prepareReveal(NodeDetailLevel.Implementation, { x: 220, y: 180 });

		expect(transition.zoomChanged).toBe(true);
		expect(transition.viewport.zoom).toBe(1.74);
		expect(state.updateViewport(transition.viewport)).toBe(true);
		expect(state.structuralDetail).toBe(NodeDetailLevel.Implementation);
		expect(state.updateViewport({ ...transition.viewport, x: transition.viewport.x + 80 })).toBe(false);
		expect(state.structuralDetail).toBe(NodeDetailLevel.Implementation);
	});

	test("keeps enter and exit thresholds stable around each detail boundary", () => {
		expect(structuralDetailForZoom(1.29, NodeDetailLevel.Systems)).toBe(NodeDetailLevel.Systems);
		expect(structuralDetailForZoom(1.3, NodeDetailLevel.Systems)).toBe(NodeDetailLevel.Runtime);
		expect(structuralDetailForZoom(1.11, NodeDetailLevel.Runtime)).toBe(NodeDetailLevel.Runtime);
		expect(structuralDetailForZoom(1.1, NodeDetailLevel.Runtime)).toBe(NodeDetailLevel.Systems);
	});
});
