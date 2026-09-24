import { describe, expect, test } from "bun:test";
import { getViewportForBounds } from "@xyflow/svelte";

import { MapCategory, NodeDetailLevel } from "../category";
import type { GraphSubset } from "../graph";
import {
	deriveNearbyEntityIds,
	detailFromZoom,
	groupOpacityAtDetail,
	initialFrameMaxZoom,
	layerOpacityAtDetail,
	LAYER_FADE_WIDTH,
	MAP_MIN_ZOOM,
	minimumDetailForGraph,
	revealZoomForDetail,
	structuralDetailForZoom,
	systemMapOverviewEmptyState,
} from "../interaction";
import { viewportCenteredOn, viewportForZoomAtPoint } from "../geometry";
import { projectMap } from "../projection";
import { nodePresentationForEntity } from "$features/system/components/system-map/map-node/presentation";
import { MapViewportState } from "$features/system/components/system-map/viewport-state";
import type { LayoutResult } from "$features/system/components/system-map/flow-model";
import { alignLayoutToPrevious } from "$features/system/components/system-map/layout";
import { edgeCasesExample, relationshipExample, sharedGroupsExample } from "./test-fixtures";

describe("system map reveal interaction", () => {
	test("maps zoom continuously while structural reveal stays discrete and hysteretic", () => {
		expect(detailFromZoom(0.45)).toBe(0);
		expect(detailFromZoom(1.035)).toBeGreaterThan(1);
		expect(detailFromZoom(1.035)).toBeLessThan(2);
		expect(structuralDetailForZoom(0.9, 0)).toBe(1);
		expect(structuralDetailForZoom(0.8, 1)).toBe(1);
		expect(structuralDetailForZoom(0.67, 1)).toBe(0);
		expect(structuralDetailForZoom(2, 0)).toBe(NodeDetailLevel.Implementation);
		expect(revealZoomForDetail(NodeDetailLevel.Implementation)).toBe(1.74);
	});

	test("fades category layers within a narrow bounded detail interval", () => {
		const layer = NodeDetailLevel.Systems;
		const fadeStart = layer - LAYER_FADE_WIDTH;
		const midway = fadeStart + LAYER_FADE_WIDTH / 2;
		const startOpacity = layerOpacityAtDetail(layer, fadeStart);
		const middleOpacity = layerOpacityAtDetail(layer, midway);
		const fullOpacity = layerOpacityAtDetail(layer, layer);

		expect(startOpacity).toBe(0);
		expect(middleOpacity).toBeGreaterThan(startOpacity);
		expect(middleOpacity).toBeLessThan(fullOpacity);
		expect(fullOpacity).toBe(1);
		for (const opacity of [startOpacity, middleOpacity, fullOpacity]) {
			expect(opacity).toBeGreaterThanOrEqual(0);
			expect(opacity).toBeLessThanOrEqual(1);
		}
		expect(groupOpacityAtDetail(NodeDetailLevel.Landscape, 2)).toBe(0.5);
	});

	test("caps the first automatic frame below the next reveal level", () => {
		expect(initialFrameMaxZoom(NodeDetailLevel.Landscape)).toBeLessThan(0.86);
		expect(initialFrameMaxZoom(NodeDetailLevel.Systems)).toBeLessThan(1.3);
		expect(initialFrameMaxZoom(NodeDetailLevel.Runtime)).toBeLessThan(1.74);

		for (const detail of [0, 1, 2, 3]) {
			for (const size of [
				{ width: 16, height: 16 },
				{ width: 1600, height: 900 },
			]) {
				const framed = getViewportForBounds(
					{ x: 0, y: 0, width: 1600, height: 900 },
					size.width,
					size.height,
					MAP_MIN_ZOOM,
					initialFrameMaxZoom(detail),
					0.3
				);
				const state = new MapViewportState(detail);
				state.updateViewport(framed);
				expect(state.structuralDetail).toBe(detail);
				if (size.width === 16) expect(framed.zoom).toBe(MAP_MIN_ZOOM);
			}
		}
	});

	test("renders source kind and category directly on a projected node", () => {
		const entity = sharedGroupsExample.source.entities.find(({ id }) => id === "shared-resource");
		expect(nodePresentationForEntity(entity!)).toEqual({
			label: "orders database",
			categoryLabel: "Container",
		});
	});

	test("keeps a pointer world position fixed while zooming", () => {
		const viewport = { x: -80, y: 35, zoom: 1 };
		const screenPoint = { x: 220, y: 145 };
		const next = viewportForZoomAtPoint(viewport, 1.8, screenPoint);
		const worldBefore = {
			x: (screenPoint.x - viewport.x) / viewport.zoom,
			y: (screenPoint.y - viewport.y) / viewport.zoom,
		};
		const worldAfter = {
			x: (screenPoint.x - next.x) / next.zoom,
			y: (screenPoint.y - next.y) / next.zoom,
		};
		expect(worldAfter.x).toBeCloseTo(worldBefore.x);
		expect(worldAfter.y).toBeCloseTo(worldBefore.y);
	});

	test("centers a world point without changing zoom", () => {
		const centered = viewportCenteredOn(
			{ x: 10, y: 20, zoom: 1.4 },
			{ x: 100, y: 80 },
			{ width: 500, height: 300 }
		);
		expect(centered.zoom).toBe(1.4);
		expect(centered.x + 100 * centered.zoom).toBe(250);
		expect(centered.y + 80 * centered.zoom).toBe(150);
	});

	test("derives nearby descendants from their current representative", () => {
		const graph = sharedGroupsExample.source;
		const overview = projectMap(graph, { detail: 0, nearbyEntityIds: [] });
		const nearby = deriveNearbyEntityIds(
			graph,
			overview,
			[{ id: "root", position: { x: 0, y: 0 }, width: 600, height: 400 }],
			{ x: 0, y: 0, zoom: 1 },
			{ width: 400, height: 300 },
			0
		);
		expect(nearby).toContain("group-a");
		expect(nearby).toContain("shared-resource");
		expect(nearby).toContain("code");
	});

	test("retains revealed detail while its group remains inside the viewport margin", () => {
		const graph: GraphSubset = {
			entities: [
				{ id: "root", category: MapCategory.SystemFunction, kind: "root" },
				{ id: "group", category: MapCategory.System, kind: "group" },
				{ id: "child", category: MapCategory.Container, kind: "service" },
			],
			relationships: [
				{ id: "m-root", source: "root", target: "group", predicate: "contains" },
				{ id: "m-child", source: "group", target: "child", predicate: "contains" },
			],
			unresolvedRelationships: [],
			enumeration: { scope: "fixture", stopReason: "exhausted" },
		};
		const reveal = { detail: NodeDetailLevel.Runtime, nearbyEntityIds: ["group", "child"] };
		const projection = projectMap(graph, reveal);
		const nodes = [
			{ id: "root", position: { x: 0, y: 0 }, width: 1000, height: 500 },
			{ id: "group", parentId: "root", position: { x: 100, y: 100 }, width: 220, height: 180 },
			{ id: "child", parentId: "group", position: { x: 20, y: 230 }, width: 160, height: 68 },
		];

		const retained = deriveNearbyEntityIds(
			graph,
			projection,
			nodes,
			{ x: 0, y: 0, zoom: 1 },
			{ width: 400, height: 300 },
			0
		);
		expect(retained).toContain("child");
		const away = deriveNearbyEntityIds(
			graph,
			projection,
			nodes,
			{ x: -1400, y: -1000, zoom: 1 },
			{ width: 400, height: 300 },
			0
		);
		expect(away).not.toContain("child");
	});

	test("keeps the selected/pointer layout anchor while asynchronous layouts settle", () => {
		const node = (id: string, x: number, y: number): LayoutResult["nodes"][number] => ({
			id,
			position: { x, y },
			width: 200,
			height: 100,
			data: { entity: { id, category: "system", kind: id }, appearance: "compact" },
			type: "system-map-node",
		});
		const previous: LayoutResult = { nodes: [node("group", 0, 0)], edges: [] };
		const next: LayoutResult = { nodes: [node("group", 700, 300)], edges: [] };
		const aligned = alignLayoutToPrevious(next, previous, "group");
		expect(aligned.nodes[0].position).toEqual(previous.nodes[0].position);
	});

	test("derives minimum detail and empty canvas states from supplied architecture", () => {
		expect(minimumDetailForGraph(relationshipExample.source)).toBe(NodeDetailLevel.Systems);
		expect(minimumDetailForGraph(sharedGroupsExample.source)).toBe(NodeDetailLevel.Landscape);
		expect(minimumDetailForGraph(edgeCasesExample.source)).toBe(NodeDetailLevel.Systems);

		const unsupported: GraphSubset = {
			entities: [{ id: "actor", category: "actor", kind: "person" }],
			relationships: [],
			unresolvedRelationships: [],
			enumeration: { scope: "fixture", stopReason: "exhausted" },
		};
		const emptyProjection = projectMap(unsupported, { detail: 0, nearbyEntityIds: [] });
		expect(systemMapOverviewEmptyState(unsupported, emptyProjection)).toBe("no-architecture");
		expect(
			systemMapOverviewEmptyState(
				{ ...unsupported, entities: [] },
				projectMap({ ...unsupported, entities: [] }, { detail: 0, nearbyEntityIds: [] })
			)
		).toBe("truly-empty");
	});
});
