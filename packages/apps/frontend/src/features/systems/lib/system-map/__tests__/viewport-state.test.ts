import { describe, expect, test } from "bun:test";

import { MapCategory, NodeDetailLevel } from "../category";
import { viewportCenteredOn, worldBoundsByNodeId, worldCenter, type Point } from "../geometry";
import {
	alignLayoutToPrevious,
	nearestSurvivingArchitectureAnchor,
} from "$features/systems/components/system-map/layout";
import {
	preferredLayoutAnchorId,
	settleLayoutChain,
	MapViewportState,
} from "$features/systems/components/system-map/viewport-state";
import type { FlowNode, LayoutResult } from "$features/systems/components/system-map/flow-model";

const node = (id: string, x: number, y: number, parentId?: string): FlowNode => ({
	id,
	position: { x, y } satisfies Point,
	...(parentId ? { parentId } : {}),
	width: id.startsWith("group") ? 280 : 160,
	height: id.startsWith("group") ? 180 : 68,
	data: {} as FlowNode["data"],
	type: "system-map-node",
});

describe("system map viewport state", () => {
	test("retains pointer B through a completion-triggered follow-up before selected A can anchor", () => {
		const viewportState = new MapViewportState(NodeDetailLevel.Landscape);
		viewportState.updateViewport({ ...viewportState.lastProcessedViewport, zoom: 0.9 }, true, "pointerB");
		const selectedEntityIds = ["selectedA"];
		let currentRequestId = 1;
		const clearAnchor = () => viewportState.clearLayoutAnchor();
		const anchorForNextLayout = () =>
			preferredLayoutAnchorId(viewportState.pendingLayoutAnchorId, selectedEntityIds);

		expect(anchorForNextLayout()).toBe("pointerB");
		expect(
			settleLayoutChain({
				isCurrent: () => currentRequestId === 1,
				updateNearby: () => {
					currentRequestId = 2;
					return true;
				},
				clearAnchor,
			})
		).toBe(true);
		expect(anchorForNextLayout()).toBe("pointerB");

		expect(
			settleLayoutChain({
				isCurrent: () => currentRequestId === 2,
				updateNearby: () => false,
				clearAnchor,
			})
		).toBe(false);
		expect(anchorForNextLayout()).toBe("selectedA");
	});

	test("uses the last processed viewport across bound onmove events and pending layouts", () => {
		const viewportState = new MapViewportState(NodeDetailLevel.Landscape);
		const boundViewport = { ...viewportState.lastProcessedViewport, zoom: 0.9 };

		// Svelte Flow has already written boundViewport to the component binding; viewport state still
		// owns the last processed viewport and must detect this threshold crossing.
		expect(viewportState.hasZoomChanged(boundViewport)).toBe(true);
		expect(viewportState.updateViewport(boundViewport, true, "group-a")).toBe(true);
		expect(viewportState.pendingLayoutAnchorId).toBe("group-a");

		// A following wheel event can be same-level and have no resolvable node under its pointer.
		// It must not erase the anchor needed by the delayed layout response.
		expect(viewportState.updateViewport({ ...boundViewport, zoom: 1 }, true)).toBe(false);
		expect(viewportState.pendingLayoutAnchorId).toBe("group-a");
		expect(
			viewportState.updateViewport({ ...viewportState.lastProcessedViewport, zoom: 1.4 }, true)
		).toBe(true);
		expect(viewportState.pendingLayoutAnchorId).toBe("group-a");
		expect(viewportState.pendingLayoutAnchorId).toBe("group-a");

		viewportState.clearLayoutAnchor();
		expect(viewportState.pendingLayoutAnchorId).toBeUndefined();
	});

	test("explicit Reveal raises detail zoom and keeps it through recenter and pan", () => {
		const viewportState = new MapViewportState(NodeDetailLevel.Landscape);
		const transition = viewportState.prepareReveal("code", NodeDetailLevel.Implementation, {
			x: 220,
			y: 180,
		});

		expect(transition.zoomChanged).toBe(true);
		expect(transition.viewport.zoom).toBe(1.74);
		expect(viewportState.updateViewport(transition.viewport)).toBe(true);
		expect(viewportState.structuralDetail).toBe(NodeDetailLevel.Implementation);

		const recentered = viewportCenteredOn(
			viewportState.lastProcessedViewport,
			{ x: 120, y: 80 },
			{ width: 800, height: 600 }
		);
		viewportState.updateViewport(recentered);
		viewportState.updateViewport({ ...recentered, x: recentered.x + 80 });

		expect(viewportState.structuralDetail).toBe(NodeDetailLevel.Implementation);
		expect(viewportState.continuousDetail).toBeGreaterThanOrEqual(2.99);
	});

	test("pointer and keyboard anchors survive nested layout realignment", () => {
		const previous: LayoutResult = {
			nodes: [node("group-a", 0, 0), node("group-b", 300, 0), node("member-b", 20, 60, "group-b")],
			edges: [],
		};
		const next: LayoutResult = {
			nodes: [node("group-a", 700, 0), node("group-b", 900, 0), node("member-b", 20, 60, "group-b")],
			edges: [],
		};
		const viewportState = new MapViewportState(NodeDetailLevel.Runtime);
		const before = worldCenter(worldBoundsByNodeId(previous.nodes).get("member-b")!);

		expect(
			viewportState.updateViewport(
				{ ...viewportState.lastProcessedViewport, zoom: 1.8 },
				true,
				"member-b"
			)
		).toBe(true);
		const anchorId = viewportState.pendingLayoutAnchorId;
		const aligned = alignLayoutToPrevious(next, previous, anchorId, anchorId);
		const after = worldCenter(worldBoundsByNodeId(aligned.nodes).get("member-b")!);
		expect(after).toEqual(before);

		viewportState.clearLayoutAnchor();
		const keyboardViewport = viewportState.prepareKeyboardZoom(
			1.1,
			"member-b",
			worldBoundsByNodeId(previous.nodes).get("member-b"),
			{ width: 800, height: 600 }
		);
		const focusedScreenPoint = {
			x: before.x * viewportState.lastProcessedViewport.zoom + viewportState.lastProcessedViewport.x,
			y: before.y * viewportState.lastProcessedViewport.zoom + viewportState.lastProcessedViewport.y,
		};
		expect(focusedScreenPoint.x - keyboardViewport.x).toBeCloseTo(before.x * keyboardViewport.zoom);
		expect(focusedScreenPoint.y - keyboardViewport.y).toBeCloseTo(before.y * keyboardViewport.zoom);
	});

	test("falls back to the nearest surviving architectural anchor when context disappears", () => {
		const architectureNode = (id: string, x: number, y: number): FlowNode => ({
			id,
			position: { x, y },
			width: 200,
			height: 100,
			data: {
				entity: { id, category: MapCategory.System, label: id, kind: "system" },
				appearance: "group",
				annotationCount: 0,
			},
			type: "system-map-node",
		});
		const contextNode = (x: number): FlowNode => ({
			...architectureNode("team", x, 100),
			data: {
				...architectureNode("team", x, 100).data,
				entity: { id: "team", category: MapCategory.Event, label: "team", kind: "context" },
				appearance: "compact",
			},
		});
		const previous: LayoutResult = {
			nodes: [architectureNode("group-a", 100, 100), architectureNode("group-b", 400, 100), contextNode(800)],
			edges: [],
		};
		const next: LayoutResult = {
			nodes: [architectureNode("group-a", 20, 40), architectureNode("group-b", 320, 40)],
			edges: [],
		};
		const anchor = nearestSurvivingArchitectureAnchor(
			next,
			previous,
			{ x: 0, y: 0, zoom: 1 },
			{ x: 850, y: 150 }
		);

		expect(anchor).toEqual({ nextId: "group-b", previousId: "group-b" });
		const aligned = alignLayoutToPrevious(next, previous, anchor?.nextId, anchor?.previousId);
		expect(aligned.nodes.map((node) => [node.id, node.position])).toEqual([
			["group-a", { x: 100, y: 100 }],
			["group-b", { x: 400, y: 100 }],
		]);
	});
});
