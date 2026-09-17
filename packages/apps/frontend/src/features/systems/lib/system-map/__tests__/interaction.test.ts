import { describe, expect, test } from "bun:test";
import { getViewportForBounds } from "@xyflow/svelte";

import { getMapCategoryDisplay, MapCategory, NodeDetailLevel } from "../category";
import { Coverage, type GraphSubset } from "../graph";
import {
	boundsForNodes,
	nearestNodeIdAtScreenPoint,
	viewportCenteredOn,
	viewportForZoomAtPoint,
	worldBoundsByNodeId,
} from "../geometry";
import {
	deriveNearbyEntityIds,
	detailFromZoom,
	MAP_MIN_ZOOM,
	minimumDetailForGraph,
	initialFrameMaxZoom,
	revealZoomForDetail,
	structuralDetailForZoom,
	systemMapOverviewEmptyState,
} from "../interaction";
import { projectMap } from "../projection";
import { nodePresentationForEntity } from "$features/systems/components/system-map/map-node/presentation";
import { alignLayoutToPrevious } from "$features/systems/components/system-map/layout";
import type { LayoutResult } from "$features/systems/components/system-map/flow-model";
import { MapViewportState } from "$features/systems/components/system-map/viewport-state";
import { relationshipExample, sharedGroupsExample, layoutProjection } from "./test-fixtures";

const displayOptions = { showActors: false, showAnnotations: false };

describe("system map reveal interaction", () => {
	test("maps zoom continuously while structural thresholds remain discrete", () => {
		expect(detailFromZoom(0.45)).toBe(0);
		expect(detailFromZoom(1.035)).toBeGreaterThan(1);
		expect(detailFromZoom(1.035)).toBeLessThan(2);
		expect(structuralDetailForZoom(0.9, 0)).toBe(1);
		expect(structuralDetailForZoom(0.8, 1)).toBe(1);
		expect(structuralDetailForZoom(0.67, 1)).toBe(0);
		expect(structuralDetailForZoom(2, 0)).toBe(NodeDetailLevel.Implementation);
		expect(structuralDetailForZoom(1.6, NodeDetailLevel.Implementation)).toBe(
			NodeDetailLevel.Implementation
		);
		expect(structuralDetailForZoom(1.5, NodeDetailLevel.Implementation)).toBe(NodeDetailLevel.Runtime);
		expect(revealZoomForDetail(NodeDetailLevel.Implementation)).toBe(1.74);
	});

	test("keeps the initial frame below the next structural reveal threshold", () => {
		expect(initialFrameMaxZoom(NodeDetailLevel.Landscape)).toBeLessThan(0.86);
		expect(initialFrameMaxZoom(NodeDetailLevel.Systems)).toBeLessThan(1.3);
		expect(initialFrameMaxZoom(NodeDetailLevel.Runtime)).toBeLessThan(1.74);
	});

	test("keeps every requested detail through tiny and roomy initial framing", () => {
		const bounds = { x: 0, y: 0, width: 1600, height: 900 };
		const sizes = [
			{ width: 16, height: 16 },
			{ width: 1600, height: 900 },
		];

		for (const detail of [
			NodeDetailLevel.Landscape,
			NodeDetailLevel.Systems,
			NodeDetailLevel.Runtime,
			NodeDetailLevel.Implementation,
		]) {
			for (const size of sizes) {
				const framed = getViewportForBounds(
					bounds,
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

	test("keeps low-detail and promoted node labels available to rendering", () => {
		const projection = projectMap(
			sharedGroupsExample.source,
			{ detail: NodeDetailLevel.Systems, nearbyEntityIds: ["group-a", "group-b"] },
			sharedGroupsExample.displayOptions
		);
		const entitiesById = new Map(sharedGroupsExample.source.entities.map((entity) => [entity.id, entity]));

		expect(projection.nodes.map((node) => node.id)).toEqual([
			"root",
			"group-a",
			"group-b",
			"shared-resource",
		]);
		for (const node of projection.nodes) {
			const entity = entitiesById.get(node.id);
			expect(entity).toBeDefined();
			expect(nodePresentationForEntity(entity!).label).toBe(entity!.label);
		}
		expect(nodePresentationForEntity(entitiesById.get("shared-resource")!).kindLabel).toBeUndefined();
	});

	test("keeps a pointer world location fixed while zooming", () => {
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
			{
				width: 500,
				height: 300,
			}
		);

		expect(centered.zoom).toBe(1.4);
		expect(centered.x + 100 * centered.zoom).toBe(250);
		expect(centered.y + 80 * centered.zoom).toBe(150);
	});

	test("derives nearby hidden descendants from visible representatives", () => {
		const graph = sharedGroupsExample.source;
		const overviewProjection = projectMap(
			graph,
			{ detail: NodeDetailLevel.Landscape, nearbyEntityIds: [] },
			sharedGroupsExample.displayOptions
		);
		const nearby = deriveNearbyEntityIds(
			graph,
			overviewProjection,
			[{ id: "root", position: { x: 0, y: 0 }, width: 600, height: 400 }],
			{ x: 0, y: 0, zoom: 1 },
			{ width: 400, height: 300 },
			0
		);

		expect(nearby).toContain("group-a");
		expect(nearby).toContain("shared-resource");
		expect(nearby).toContain("code");
	});

	test("retains a revealed child while its partially visible group remains eligible", () => {
		const graph = {
			coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
			entities: [
				{ id: "root", category: MapCategory.Function, label: "root", kind: "function" },
				{ id: "group", category: MapCategory.System, label: "group", kind: "system" },
				{ id: "child", category: MapCategory.Container, label: "child", kind: "service" },
			],
			relationships: [
				{ id: "m-root-group", source: "root", target: "group", predicate: "contains" },
				{ id: "m-group-child", source: "group", target: "child", predicate: "contains" },
			],
		};
		const reveal = { detail: NodeDetailLevel.Runtime, nearbyEntityIds: ["group", "child"] };
		const displayOptions = { showActors: false, showAnnotations: false };
		const projection = projectMap(graph, reveal, displayOptions);
		const nodes = [
			{ id: "root", position: { x: 0, y: 0 }, width: 1000, height: 500 },
			{ id: "group", parentId: "root", position: { x: 100, y: 100 }, width: 220, height: 180 },
			{ id: "child", parentId: "group", position: { x: 20, y: 230 }, width: 160, height: 68 },
		];
		const viewport = { x: 0, y: 0, zoom: 1 };
		const size = { width: 350, height: 250 };

		const nearby = deriveNearbyEntityIds(graph, projection, nodes, viewport, size, 0);
		expect(nearby).toContain("child");
		expect(deriveNearbyEntityIds(graph, projection, nodes, viewport, size, 0)).toEqual(nearby);

		const awayViewport = { x: -1200, y: 0, zoom: 1 };
		const away = deriveNearbyEntityIds(graph, projection, nodes, awayViewport, size, 0);
		expect(away).not.toContain("group");
		expect(away).not.toContain("child");

		const awayProjection = projectMap(graph, { ...reveal, nearbyEntityIds: away }, displayOptions);
		const returned = deriveNearbyEntityIds(graph, awayProjection, [nodes[0]], viewport, size, 0);
		expect(returned).toContain("group");
		expect(returned).toContain("child");
	});

	test("retains a revealed system while its visible landscape parent remains eligible", () => {
		const graph = {
			coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
			entities: [
				{ id: "root", category: MapCategory.Function, label: "root", kind: "function" },
				{ id: "group", category: MapCategory.System, label: "group", kind: "system" },
			],
			relationships: [{ id: "m-root-group", source: "root", target: "group", predicate: "contains" }],
		};
		const reveal = { detail: NodeDetailLevel.Systems, nearbyEntityIds: ["group"] };
		const displayOptions = { showActors: false, showAnnotations: false };
		const projection = projectMap(graph, reveal, displayOptions);
		const nodes = [
			{ id: "root", position: { x: 0, y: 0 }, width: 1000, height: 500 },
			{ id: "group", parentId: "root", position: { x: 600, y: 100 }, width: 220, height: 180 },
		];
		const viewport = { x: 0, y: 0, zoom: 1 };
		const size = { width: 350, height: 250 };

		const retainedIds = deriveNearbyEntityIds(graph, projection, nodes, viewport, size, 0);
		expect(retainedIds).toEqual(["group", "root"]);
		const retainedProjection = projectMap(
			graph,
			{ ...reveal, nearbyEntityIds: retainedIds },
			displayOptions
		);
		expect(deriveNearbyEntityIds(graph, retainedProjection, nodes, viewport, size, 0)).toEqual(
			retainedIds
		);

		const awayViewport = { x: -1200, y: 0, zoom: 1 };
		expect(deriveNearbyEntityIds(graph, projection, nodes, awayViewport, size, 0)).toEqual([]);

		const overviewProjection = projectMap(graph, { ...reveal, nearbyEntityIds: [] }, displayOptions);
		expect(deriveNearbyEntityIds(graph, overviewProjection, [nodes[0]], viewport, size, 0)).toEqual([
			"group",
			"root",
		]);
	});

	test("converges nested projection and layout feedback before descendants are panned away", async () => {
		const graph = {
			coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
			entities: [
				{ id: "root", category: MapCategory.Function, label: "root", kind: "function" },
				{ id: "group", category: MapCategory.System, label: "group", kind: "system" },
				{ id: "child", category: MapCategory.Container, label: "child", kind: "service" },
			],
			relationships: [
				{ id: "m-root-group", source: "root", target: "group", predicate: "contains" },
				{ id: "m-group-child", source: "group", target: "child", predicate: "contains" },
			],
		};
		const displayOptions = { showActors: false, showAnnotations: false };
		let nearbyEntityIds = ["group", "child"];
		let projection = projectMap(
			graph,
			{ detail: NodeDetailLevel.Runtime, nearbyEntityIds },
			displayOptions
		);
		let layout = await layoutProjection(graph, projection);
		const layoutBounds = worldBoundsByNodeId(layout.nodes);
		const rootBounds = layoutBounds.get("root")!;
		const groupBounds = layoutBounds.get("group")!;
		const viewport = {
			x: -rootBounds.x,
			y: -rootBounds.y,
			zoom: 1,
		};
		const size = {
			width: Math.max(1, groupBounds.x - rootBounds.x - 1),
			height: Math.max(1, groupBounds.y - rootBounds.y - 1),
		};

		for (let iteration = 0; iteration < 3; iteration += 1) {
			nearbyEntityIds = [...deriveNearbyEntityIds(graph, projection, layout.nodes, viewport, size, 0)];
			expect(nearbyEntityIds).toEqual(["child", "group", "root"]);

			projection = projectMap(
				graph,
				{ detail: NodeDetailLevel.Runtime, nearbyEntityIds },
				displayOptions
			);
			layout = await layoutProjection(graph, projection);
		}

		const awayViewport = {
			x: -(rootBounds.x + rootBounds.width + 1000),
			y: -rootBounds.y,
			zoom: 1,
		};
		expect(deriveNearbyEntityIds(graph, projection, layout.nodes, awayViewport, size, 0)).toEqual([]);
	});

	test("resolves parent-relative bounds and overall bounds", () => {
		const nodes = [
			{ id: "root", position: { x: 10, y: 20 }, width: 500, height: 300 },
			{ id: "child", parentId: "root", position: { x: 30, y: 40 }, width: 100, height: 60 },
		] as const;
		const bounds = worldBoundsByNodeId(nodes);

		expect(bounds.get("child")).toEqual({ x: 40, y: 60, width: 100, height: 60 });
		expect(boundsForNodes(nodes)).toEqual({ x: 10, y: 20, width: 500, height: 300 });
		expect(nearestNodeIdAtScreenPoint(nodes, { x: 0, y: 0, zoom: 1 }, { x: 42, y: 62 })).toBe("child");
	});

	test("aligns a new layout around a common source anchor", () => {
		const node = (id: string, x: number, y: number) => ({
			id,
			position: { x, y },
			width: 100,
			height: 60,
			data: {
				entity: { id, category: MapCategory.System, label: id, kind: "system" },
				appearance: "compact" as const,
				annotationCount: 0,
			},
			type: "system-map-node" as const,
		});
		const edge = {
			id: "edge",
			source: "anchor",
			target: "other",
			data: { connection: {} as never },
			type: "system-map-connection" as const,
		};
		const previous: LayoutResult = {
			nodes: [node("anchor", 100, 100), node("other", 300, 100)],
			edges: [edge],
		};
		const next: LayoutResult = {
			nodes: [node("anchor", 20, 40), node("other", 220, 40)],
			edges: [edge],
		};
		const aligned = alignLayoutToPrevious(next, previous, "anchor");

		expect(aligned.nodes.map((entry) => entry.position)).toEqual([
			{ x: 100, y: 100 },
			{ x: 300, y: 100 },
		]);

		const promotedNext: LayoutResult = {
			nodes: [node("promoted", 20, 40), node("other", 220, 40)],
			edges: [edge],
		};
		const promotedPrevious: LayoutResult = {
			nodes: [node("anchor", 100, 100), node("other", 300, 100)],
			edges: [edge],
		};
		const promoted = alignLayoutToPrevious(promotedNext, promotedPrevious, "promoted", "anchor");
		expect(promoted.nodes[0].position).toEqual({ x: 100, y: 100 });

		const unanchored = alignLayoutToPrevious(next, previous);
		expect(unanchored.nodes.map((entry) => entry.position)).toEqual([
			{ x: 20, y: 40 },
			{ x: 220, y: 40 },
		]);
	});
});

describe("system map minimum detail", () => {
	test("automatically shows the shallowest architecture and retains it when zoomed out", () => {
		for (const [category, level] of [
			[MapCategory.Function, NodeDetailLevel.Landscape],
			[MapCategory.System, NodeDetailLevel.Systems],
			[MapCategory.Container, NodeDetailLevel.Runtime],
			[MapCategory.Infrastructure, NodeDetailLevel.Runtime],
			[MapCategory.Component, NodeDetailLevel.Implementation],
			[MapCategory.Code, NodeDetailLevel.Implementation],
		] as const) {
			const graph: GraphSubset = {
				entities: [
					{ id: "baseline", category, label: "Baseline", kind: "subject" },
					{ id: "actor", category: MapCategory.Actor, label: "Actor", kind: "team" },
				],
				relationships: [],
				coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
			};
			expect(minimumDetailForGraph(graph)).toBe(level);
			const state = new MapViewportState(minimumDetailForGraph(graph)!);
			for (const zoom of [MAP_MIN_ZOOM, 2, MAP_MIN_ZOOM]) {
				state.updateViewport({ x: 100, y: 100, zoom });
				const projection = projectMap(graph, {
					detail: state.structuralDetail,
					nearbyEntityIds: graph.entities
						.filter((entity) => getMapCategoryDisplay(entity.category).level === level)
						.map((entity) => entity.id),
				}, displayOptions);
				expect(projection.nodes.map((node) => node.id)).toEqual(["baseline"]);
			}
			expect(state.structuralDetail).toBe(level);
			state.setMinimumDetail(NodeDetailLevel.Implementation);
			expect(state.structuralDetail).toBe(NodeDetailLevel.Implementation);
			state.setMinimumDetail(NodeDetailLevel.Landscape);
			expect(state.structuralDetail).toBe(NodeDetailLevel.Landscape);
		}
		expect(minimumDetailForGraph(relationshipExample.source)).toBe(NodeDetailLevel.Systems);
		expect(minimumDetailForGraph(sharedGroupsExample.source)).toBe(NodeDetailLevel.Landscape);
	});

	test("retains empty states only for empty or nonarchitectural input", () => {
		for (const category of [undefined, "future_category", MapCategory.Actor, MapCategory.Event]) {
			const graph: GraphSubset = {
				entities: category ? [{ id: "context", category, label: "Context", kind: "subject" }] : [],
				relationships: [],
				coverage: { parentMembership: Coverage.Unknown, relationships: Coverage.Unknown },
			};
			expect(minimumDetailForGraph(graph)).toBeUndefined();
			const projection = projectMap(graph, { detail: 0, nearbyEntityIds: [] }, displayOptions);
			expect(systemMapOverviewEmptyState(graph, projection)).toBe(category ? "no-architecture" : "truly-empty");
		}
	});
});
