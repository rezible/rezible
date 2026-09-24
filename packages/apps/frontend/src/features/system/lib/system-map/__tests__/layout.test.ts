import { describe, expect, test } from "bun:test";

import { MapCategory } from "../category";
import { worldBoundsByNodeId, worldPositionByNodeId, type Bounds } from "../geometry";
import type { GraphSubset } from "../graph";
import { projectFullArchitecture, projectMap } from "../projection";
import type { MapConnection, MapProjection } from "../presentation";
import { nodePresentationForEntity } from "$features/system/components/system-map/map-node/presentation";
import {
	alignLayoutToPrevious,
	hintedStartingPositions,
	nodeMinSizeForCategory,
	presentMapProjection,
	systemMapLayoutInputKey,
} from "$features/system/components/system-map/layout";
import {
	CompactNodeDragSession,
	layoutRequestIsStale,
	nextLayoutRequestId,
	nodeCanBeDragged,
	nodesWithPendingDropPosition,
} from "$features/system/components/system-map/drag-state";
import { createSystemMapLayoutEngine } from "$features/system/components/system-map/layout-engine";
import type {
	ConnectionRoute,
	FlowNode,
	LayoutResult,
} from "$features/system/components/system-map/flow-model";
import {
	connectionRouteToSvgPath,
	connectionRouteWithNodeFeedback,
} from "$features/system/components/system-map/map-connection/geometry";
import { ProjectedRouteCache } from "$features/system/components/system-map/routing";
import {
	connectionEndpointIds,
	connectionLabel,
	connectionPresentationState,
} from "$features/system/components/system-map/map-connection/presentation";
import {
	edgeCasesExample,
	layoutProjection,
	relationshipExample,
	sharedGroupsExample,
} from "./test-fixtures";

const nodeById = <T extends { id: string }>(items: readonly T[], id: string): T => {
	const item = items.find((candidate) => candidate.id === id);
	if (!item) throw new Error(`Missing item ${id}`);
	return item;
};

const pointIsOnBounds = (point: { x: number; y: number }, bounds: Bounds) => {
	const horizontal =
		(point.y === bounds.y || point.y === bounds.y + bounds.height) &&
		point.x >= bounds.x &&
		point.x <= bounds.x + bounds.width;
	const vertical =
		(point.x === bounds.x || point.x === bounds.x + bounds.width) &&
		point.y >= bounds.y &&
		point.y <= bounds.y + bounds.height;
	return horizontal || vertical;
};

const flowNode = (
	id: string,
	category: string,
	position: { x: number; y: number },
	size: { width: number; height: number },
	options: { appearance?: "compact" | "group"; parentId?: string } = {}
): FlowNode => ({
	id,
	position,
	width: size.width,
	height: size.height,
	...(options.parentId ? { parentId: options.parentId } : {}),
	type: "system-map-node",
	data: {
		entity: { id, category, kind: id },
		appearance: options.appearance ?? "compact",
	},
});

const mapConnection = (
	id: string,
	endpoints: readonly [string, string],
	options: Partial<Pick<MapConnection, "classification" | "predicate" | "sourceRelationshipIds">> = {}
): MapConnection => ({
	id,
	endpoints,
	predicate: options.predicate ?? "calls",
	classification: options.classification ?? "direct",
	sourceRelationshipIds: options.sourceRelationshipIds ?? [id.replace(/^direct:/, "")],
});

const mapProjection = (
	nodes: MapProjection["nodes"],
	connections: readonly MapConnection[]
): MapProjection => ({
	nodes,
	connections,
	representativeByEntityId: new Map(nodes.map((node) => [node.id, node.id])),
});

const segmentCrossesBounds = (
	start: { x: number; y: number },
	end: { x: number; y: number },
	bounds: Bounds
) => {
	if (start.y === end.y) {
		return (
			start.y > bounds.y &&
			start.y < bounds.y + bounds.height &&
			Math.max(Math.min(start.x, end.x), bounds.x) <
				Math.min(Math.max(start.x, end.x), bounds.x + bounds.width)
		);
	}
	if (start.x === end.x) {
		return (
			start.x > bounds.x &&
			start.x < bounds.x + bounds.width &&
			Math.max(Math.min(start.y, end.y), bounds.y) <
				Math.min(Math.max(start.y, end.y), bounds.y + bounds.height)
		);
	}
	return true;
};

const expectUsableRoutes = (layout: LayoutResult, projection: MapProjection) => {
	const boundsById = worldBoundsByNodeId(layout.nodes);
	expect(layout.edges.map((edge) => edge.id)).toEqual(
		projection.connections.map((connection) => connection.id)
	);

	for (const edge of layout.edges) {
		const connection = edge.data?.connection;
		const route = edge.data?.route;
		expect(connection).toBeDefined();
		expect(route).toBeDefined();
		if (!connection || !route) continue;
		expect([edge.source, edge.target]).toEqual([...connection.endpoints]);
		expect(route.sections.length).toBeGreaterThan(0);
		expect(route.targetSectionIndex).toBeGreaterThanOrEqual(0);
		expect(route.targetSectionIndex).toBeLessThan(route.sections.length);
		expect(Number.isFinite(route.labelPosition.x)).toBe(true);
		expect(Number.isFinite(route.labelPosition.y)).toBe(true);
		expect(connectionRouteToSvgPath(route)).toMatch(/^M/);
		for (const section of route.sections) {
			const points = [section.start, ...section.bends, section.end];
			for (let index = 1; index < points.length; index += 1) {
				const previous = points[index - 1];
				const current = points[index];
				expect(previous.x === current.x || previous.y === current.y).toBe(true);
			}
		}
		expect(boundsById.has(connection.endpoints[0])).toBe(true);
		expect(boundsById.has(connection.endpoints[1])).toBe(true);
	}
};

describe("system map ELK layout", () => {
	test("keeps multi-section routes separate and emits the marker target section last", () => {
		const route: ConnectionRoute = {
			sections: [
				{ start: { x: 20, y: 20 }, bends: [], end: { x: 40, y: 20 } },
				{ start: { x: 60, y: 40 }, bends: [{ x: 80, y: 40 }], end: { x: 100, y: 40 } },
			],
			labelPosition: { x: 80, y: 40 },
			sourceSectionIndex: 1,
			targetSectionIndex: 0,
		};
		const path = connectionRouteToSvgPath(route);
		expect(path.match(/M/g)).toHaveLength(2);
		expect(path.startsWith("M60 40L80 40L100 40M20 20L40 20")).toBe(true);
	});

	test("adds orthogonal route feedback for a moving compact endpoint", () => {
		const route: ConnectionRoute = {
			sections: [{ start: { x: 0, y: 0 }, bends: [{ x: 100, y: 0 }], end: { x: 100, y: 100 } }],
			labelPosition: { x: 100, y: 50 },
			sourceSectionIndex: 0,
			targetSectionIndex: 0,
		};
		const feedback = connectionRouteWithNodeFeedback(route, { x: 20, y: 30 }, { x: -40, y: 10 });
		expect(feedback.sections[0].start).toEqual({ x: 20, y: 30 });
		expect(feedback.sections[0].end).toEqual({ x: 60, y: 110 });
		const points = [feedback.sections[0].start, ...feedback.sections[0].bends, feedback.sections[0].end];
		for (let index = 1; index < points.length; index += 1) {
			const previous = points[index - 1];
			const current = points[index];
			expect(previous.x === current.x || previous.y === current.y).toBe(true);
		}
	});

	test("enables one callback per completed compact-node drag and invalidates stale layouts", () => {
		expect(nodeCanBeDragged("compact", true)).toBe(true);
		expect(nodeCanBeDragged("compact", false)).toBe(false);
		expect(nodeCanBeDragged("group", true)).toBe(false);

		const session = new CompactNodeDragSession();
		expect(session.start("group", "group", { x: 0, y: 0 }, true)).toBe(false);
		expect(session.start("node", "compact", { x: 0, y: 0 }, false)).toBe(false);
		expect(session.start("node", "compact", { x: 0, y: 0 }, true)).toBe(true);
		expect(session.isActive).toBe(true);
		session.update("node", { x: 20, y: 30 });
		const moves: Array<[string, { x: number; y: number }]> = [];
		const callback = (entityId: string, position: { x: number; y: number }) => {
			moves.push([entityId, position]);
		};
		expect(session.finish("node", { x: 20, y: 30 }, callback)).toBe(true);
		expect(session.isActive).toBe(false);
		expect(session.finish("node", { x: 20, y: 30 }, callback)).toBe(false);
		expect(moves).toEqual([["node", { x: 20, y: 30 }]]);

		expect(session.start("node", "compact", { x: 20, y: 30 }, true)).toBe(true);
		session.update("node", { x: 40, y: 50 });
		expect(session.finish("node", { x: 40, y: 50 }, undefined)).toBe(false);

		const pendingLayout = nextLayoutRequestId(0);
		expect(layoutRequestIsStale(pendingLayout, pendingLayout, false)).toBe(false);
		const afterDragStarts = nextLayoutRequestId(pendingLayout);
		expect(layoutRequestIsStale(pendingLayout, afterDragStarts, false)).toBe(true);
	});

	test("lays out supplied groups and compact nodes through the automatic ELK path", async () => {
		const graph = sharedGroupsExample.source;
		const projection = projectFullArchitecture(graph);
		const layout = await layoutProjection(graph);
		expect(layout.edges).toHaveLength(0);

		const root = nodeById(layout.nodes, "root");
		const groupA = nodeById(layout.nodes, "group-a");
		const memberA = nodeById(layout.nodes, "member-a");
		const code = nodeById(layout.nodes, "code");
		const shared = nodeById(layout.nodes, "shared-resource");

		expect(root.parentId).toBeUndefined();
		expect(groupA.parentId).toBe("root");
		expect(memberA.parentId).toBe("group-a");
		expect(code.parentId).toBe("member-a");
		expect(shared.parentId).toBeUndefined();
		expect(nodePresentationForEntity(root.data.entity)).toEqual({
			label: "order processing",
			categoryLabel: "Function",
		});
		for (const node of layout.nodes) {
			const minimum = nodeMinSizeForCategory(node.data.entity.category, node.data.appearance);
			expect(node.width).toBeGreaterThanOrEqual(minimum.width);
			expect(node.height).toBeGreaterThanOrEqual(minimum.height);
		}

		const presentation = presentMapProjection(layout, projection);
		const sharedMembershipEdges = presentation.edges.filter((edge) =>
			edge.data?.connection.sourceRelationshipIds.some((id) => id.startsWith("m-shared"))
		);
		expect(sharedMembershipEdges.map((edge) => [edge.source, edge.target])).toEqual([
			["group-a", "shared-resource"],
			["group-b", "shared-resource"],
		]);
		expectUsableRoutes(presentation, projection);
	});

	test("reserves the two-row group header and keeps category minimums visibly ordered", async () => {
		const graph = sharedGroupsExample.source;
		const layout = await layoutProjection(graph);
		const bounds = worldBoundsByNodeId(layout.nodes);
		const groups = layout.nodes.filter((node) => node.data.appearance === "group");
		let checkedChildren = 0;
		for (const group of groups) {
			const groupBounds = bounds.get(group.id)!;
			for (const child of layout.nodes.filter((node) => node.parentId === group.id)) {
				const childBounds = bounds.get(child.id)!;
				// 12px shell inset + a 16px category row + 8px row gap + two 18px title lines + 16px child gap.
				expect(childBounds.y - groupBounds.y).toBeGreaterThanOrEqual(88);
				checkedChildren += 1;
			}
		}
		expect(checkedChildren).toBeGreaterThan(0);

		for (const appearance of ["compact", "group"] as const) {
			const sizes = ["system_function", "system", "container", "component"].map((category) =>
				nodeMinSizeForCategory(category, appearance)
			);
			for (let index = 1; index < sizes.length; index += 1) {
				expect(sizes[index - 1].width).toBeGreaterThan(sizes[index].width);
				expect(sizes[index - 1].height).toBeGreaterThan(sizes[index].height);
			}
		}
		const system = nodeMinSizeForCategory("system", "compact");
		const runtime = nodeMinSizeForCategory("container", "compact");
		expect(system.width - runtime.width).toBeGreaterThanOrEqual(30);
		expect(system.height - runtime.height).toBeGreaterThanOrEqual(12);
	});

	test("keeps nested category minima in ELK placement without overlapping sibling groups", async () => {
		const graph = {
			entities: [
				{ id: "landscape", category: "system_function", kind: "landscape" },
				{ id: "system-a", category: "system", kind: "system a" },
				{ id: "system-b", category: "system", kind: "system b" },
				{ id: "runtime-a1", category: "container", kind: "runtime a1" },
				{ id: "runtime-a2", category: "container", kind: "runtime a2" },
				{ id: "runtime-b", category: "container", kind: "runtime b" },
				{ id: "code-a1", category: "code", kind: "implementation a1" },
				{ id: "code-a2", category: "code", kind: "implementation a2" },
				{ id: "code-b", category: "code", kind: "implementation b" },
			],
			relationships: [
				{ id: "contains-system-a", source: "landscape", target: "system-a", predicate: "contains" },
				{ id: "contains-system-b", source: "landscape", target: "system-b", predicate: "contains" },
				{
					id: "contains-runtime-a1",
					source: "system-a",
					target: "runtime-a1",
					predicate: "contains",
				},
				{
					id: "contains-runtime-a2",
					source: "system-a",
					target: "runtime-a2",
					predicate: "contains",
				},
				{ id: "contains-runtime-b", source: "system-b", target: "runtime-b", predicate: "contains" },
				{ id: "contains-code-a1", source: "runtime-a1", target: "code-a1", predicate: "contains" },
				{ id: "contains-code-a2", source: "runtime-a2", target: "code-a2", predicate: "contains" },
				{ id: "contains-code-b", source: "runtime-b", target: "code-b", predicate: "contains" },
			],
			unresolvedRelationships: [],
			enumeration: { scope: "fixture", stopReason: "exhausted" },
		} satisfies GraphSubset;
		const layout = await layoutProjection(graph);
		const bounds = worldBoundsByNodeId(layout.nodes);

		for (const node of layout.nodes) {
			const minimum = nodeMinSizeForCategory(node.data.entity.category, node.data.appearance);
			expect(node.width).toBeGreaterThanOrEqual(minimum.width);
			expect(node.height).toBeGreaterThanOrEqual(minimum.height);
		}

		const parents = layout.nodes.filter((node) => node.data.appearance === "group");
		expect(parents.map((node) => node.id).sort()).toEqual([
			"landscape",
			"runtime-a1",
			"runtime-a2",
			"runtime-b",
			"system-a",
			"system-b",
		]);
		const checkedSiblingPairs: string[] = [];
		for (const parent of parents) {
			const siblings = layout.nodes.filter((node) => node.parentId === parent.id);
			for (let firstIndex = 0; firstIndex < siblings.length; firstIndex += 1) {
				for (let secondIndex = firstIndex + 1; secondIndex < siblings.length; secondIndex += 1) {
					checkedSiblingPairs.push(
						[siblings[firstIndex].id, siblings[secondIndex].id].sort().join("|")
					);
					const first = bounds.get(siblings[firstIndex].id)!;
					const second = bounds.get(siblings[secondIndex].id)!;
					const overlap =
						first.x < second.x + second.width &&
						second.x < first.x + first.width &&
						first.y < second.y + second.height &&
						second.y < first.y + first.height;
					expect(overlap).toBe(false);
				}
			}
		}
		expect(checkedSiblingPairs.sort()).toEqual(["runtime-a1|runtime-a2", "system-a|system-b"]);
	});

	test("uses sparse world hints without feeding computed positions back into ELK", () => {
		const graph = sharedGroupsExample.source;
		const projection = projectFullArchitecture(graph);
		expect(hintedStartingPositions(graph, projection).size).toBe(0);

		const partial = hintedStartingPositions(graph, projection, {
			positions: { "shared-resource": { x: -640, y: 320 }, code: { x: 520, y: 450 } },
			previousWorldPositions: { "member-a": { x: 120, y: 180 } },
		});
		expect(partial.get("shared-resource")).toEqual({ x: -640, y: 320 });
		expect(partial.get("code")).toEqual({ x: 400, y: 270 });
		expect(partial.has("member-a")).toBe(false);

		const changed = hintedStartingPositions(graph, projection, {
			positions: { "shared-resource": { x: 910, y: -210 } },
		});
		expect(changed.get("shared-resource")).toEqual({ x: 910, y: -210 });
	});

	test("validates only supplied architectural hints", () => {
		const graph = edgeCasesExample.source;
		const projection = projectFullArchitecture(graph);
		expect(() =>
			hintedStartingPositions(graph, projection, {
				positions: {
					child: { x: Number.NaN, y: 0 },
					context: { x: Number.POSITIVE_INFINITY, y: 0 },
					absent: { x: Number.NaN, y: 0 },
				} as never,
			})
		).toThrow("child must have finite coordinates");
		const ignored = hintedStartingPositions(graph, projection, {
			positions: {
				context: { x: Number.POSITIVE_INFINITY, y: 0 },
				absent: { x: Number.NaN, y: 0 },
			} as never,
		});
		expect(ignored.size).toBe(0);
	});

	test("lets ELK use a single hint as a starting position without exact pinning", async () => {
		const graph = {
			entities: [{ id: "function", category: "system_function", kind: "orders" }],
			relationships: [],
			unresolvedRelationships: [],
			enumeration: { scope: "fixture", stopReason: "exhausted" },
		} as const;
		const projection = projectFullArchitecture(graph);
		const automatic = await layoutProjection(graph);
		const starts = hintedStartingPositions(graph, projection, {
			positions: { function: { x: 740, y: 510 } },
		});
		const hinted = await layoutProjection(graph, {
			positions: { function: { x: 740, y: 510 } },
		});
		const automaticPosition = worldPositionByNodeId(automatic.nodes).get("function");
		const hintedPosition = worldPositionByNodeId(hinted.nodes).get("function");
		expect(starts.get("function")).toEqual({ x: 740, y: 510 });
		expect(automaticPosition).toBeDefined();
		expect(hintedPosition).toBeDefined();
		expect(hinted.nodes[0].position).not.toEqual({ x: 740, y: 510 });
	});

	test("uses sparse hints to influence ELK placement without pinning exact coordinates", async () => {
		const graph = {
			entities: [
				{ id: "a", category: "system", kind: "a" },
				{ id: "b", category: "system", kind: "b" },
				{ id: "c", category: "system", kind: "c" },
			],
			relationships: [],
			unresolvedRelationships: [],
			enumeration: { scope: "fixture", stopReason: "exhausted" },
		} as const;
		const automatic = await layoutProjection(graph);
		const hinted = await layoutProjection(graph, {
			positions: {
				a: { x: 1000, y: 700 },
				c: { x: 1000, y: -700 },
			},
		});
		const automaticPositions = worldPositionByNodeId(automatic.nodes);
		const hintedPositions = worldPositionByNodeId(hinted.nodes);
		expect(hinted.nodes.length).toBe(3);
		expect(hintedPositions.get("a")).not.toEqual(automaticPositions.get("a"));
		expect(hintedPositions.get("c")).not.toEqual(automaticPositions.get("c"));
		expect(hintedPositions.get("a")).not.toEqual({ x: 1000, y: 700 });
	});

	test("keeps complete geometry fixed through detail projection changes and repeated layout", async () => {
		const graph = sharedGroupsExample.source;
		const fullProjection = projectFullArchitecture(graph);
		const firstLayout = await layoutProjection(graph);
		const repeatedLayout = await layoutProjection({
			...graph,
			entities: graph.entities.map((entity) => ({ ...entity })),
			relationships: graph.relationships.map((relationship) => ({ ...relationship })),
			unresolvedRelationships: graph.unresolvedRelationships.map((relationship) => ({
				...relationship,
			})),
		});
		const detailOne = projectMap(graph, {
			detail: 1,
			nearbyEntityIds: ["root", "group-a", "group-b"],
		});
		const detailThree = projectMap(graph, {
			detail: 3,
			nearbyEntityIds: fullProjection.nodes.map((node) => node.id),
		});
		const fadingDetail = projectMap(graph, {
			detail: 1,
			visualDetail: 1.9,
			nearbyEntityIds: ["root", "group-a", "group-b", "member-a"],
		});
		const lowDetail = presentMapProjection(firstLayout, detailOne);
		const fadingLayout = presentMapProjection(firstLayout, fadingDetail);
		const highDetail = presentMapProjection(firstLayout, detailThree);
		const firstBounds = worldBoundsByNodeId(firstLayout.nodes);
		const repeatedBounds = worldBoundsByNodeId(repeatedLayout.nodes);
		const lowBounds = worldBoundsByNodeId(lowDetail.nodes);
		const fadingBounds = worldBoundsByNodeId(fadingLayout.nodes);
		const highBounds = worldBoundsByNodeId(highDetail.nodes);

		expect(firstLayout.nodes.map((node) => node.id).sort()).toEqual(
			graph.entities
				.filter((entity) => entity.category !== "actor")
				.map((entity) => entity.id)
				.sort()
		);
		for (const id of ["root", "group-a", "group-b", "member-a", "code"]) {
			expect(repeatedBounds.get(id)).toEqual(firstBounds.get(id));
		}
		expect(detailOne.nodes.some((node) => node.id === "code")).toBe(false);
		expect(lowBounds.get("group-a")).toEqual(firstBounds.get("group-a"));
		expect(nodeById(lowDetail.nodes, "group-a").data.appearance).toBe("compact");
		expect(nodeById(lowDetail.nodes, "group-a").width).toBe(nodeById(firstLayout.nodes, "group-a").width);
		expect(fadingBounds.get("member-a")).toEqual(firstBounds.get("member-a"));
		expect(highBounds.get("code")).toEqual(firstBounds.get("code"));
		expect(highBounds.get("group-a")).toEqual(firstBounds.get("group-a"));
		expect(fullProjection.nodes).toHaveLength(firstLayout.nodes.length);
		expect(fadingLayout.nodes.some(({ id }) => id === "member-a")).toBe(true);
		for (const node of fadingLayout.nodes) {
			if (node.parentId)
				expect(fadingLayout.nodes.some((parent) => parent.id === node.parentId)).toBe(true);
		}

		const sameTopologyNewLabels = {
			...graph,
			entities: graph.entities.map((entity) => ({ ...entity, kind: `${entity.kind} renamed` })),
		};
		const sameProjection = projectFullArchitecture(sameTopologyNewLabels);
		expect(systemMapLayoutInputKey(graph, fullProjection)).toBe(
			systemMapLayoutInputKey(sameTopologyNewLabels, sameProjection)
		);
		expect(systemMapLayoutInputKey(graph, fullProjection, { code: { x: 50, y: 60 } })).not.toBe(
			systemMapLayoutInputKey(graph, fullProjection)
		);
	});

	test("routes through space reserved by hidden nodes and around visible nodes", () => {
		const nodes = [
			flowNode("source", "system", { x: 0, y: 0 }, { width: 80, height: 60 }),
			flowNode("hidden", "container", { x: 140, y: -20 }, { width: 100, height: 100 }),
			flowNode("target", "system", { x: 300, y: 0 }, { width: 80, height: 60 }),
		];
		const fullLayout: LayoutResult = { nodes, edges: [] };
		const connection = mapConnection("direct:r", ["source", "target"]);
		const hiddenProjection = mapProjection(
			[
				{ id: "source", appearance: "compact" },
				{ id: "target", appearance: "compact" },
			],
			[connection]
		);
		const hiddenRoute = presentMapProjection(fullLayout, hiddenProjection).edges[0].data!.route;
		expect(hiddenRoute.sections[0].bends).toHaveLength(0);

		const visibleProjection = mapProjection(
			[
				{ id: "source", appearance: "compact" },
				{ id: "hidden", appearance: "compact" },
				{ id: "target", appearance: "compact" },
			],
			[connection]
		);
		const visibleLayout = presentMapProjection(fullLayout, visibleProjection);
		const route = visibleLayout.edges[0].data!.route.sections[0];
		const obstacle = worldBoundsByNodeId(visibleLayout.nodes).get("hidden")!;
		const points = [route.start, ...route.bends, route.end];
		expect(route.bends.length).toBeGreaterThan(0);
		for (let index = 1; index < points.length; index += 1) {
			expect(segmentCrossesBounds(points[index - 1], points[index], obstacle)).toBe(false);
		}
	});

	test("routes nested endpoints in world coordinates through their enclosing group", () => {
		const nodes = [
			flowNode(
				"group",
				"system",
				{ x: 100, y: 100 },
				{ width: 300, height: 200 },
				{ appearance: "group" }
			),
			flowNode(
				"source",
				"container",
				{ x: 30, y: 70 },
				{ width: 80, height: 40 },
				{ parentId: "group" }
			),
			flowNode("target", "system", { x: 500, y: 170 }, { width: 80, height: 40 }),
		];
		const projection = mapProjection(
			[
				{ id: "group", appearance: "group" },
				{ id: "source", appearance: "compact", enclosure: { parentId: "group", membershipId: "m" } },
				{ id: "target", appearance: "compact" },
			],
			[mapConnection("direct:r", ["source", "target"])]
		);
		const layout = presentMapProjection({ nodes, edges: [] }, projection);
		const bounds = worldBoundsByNodeId(layout.nodes);
		const route = layout.edges[0].data!.route.sections[0];

		expect(bounds.get("source")).toMatchObject({ x: 130, y: 170 });
		expect(route.start).toEqual({ x: 210, y: 190 });
		expect(route.end).toEqual({ x: 500, y: 190 });
		expect(pointIsOnBounds(route.start, bounds.get("source")!)).toBe(true);
		expect(pointIsOnBounds(route.end, bounds.get("target")!)).toBe(true);
	});

	test("separates shared ports and corridors deterministically without moving nodes", () => {
		const nodes = [
			flowNode("source", "system", { x: 0, y: 0 }, { width: 120, height: 80 }),
			flowNode("target", "container", { x: 320, y: 0 }, { width: 120, height: 80 }),
		];
		const connections = [
			mapConnection("direct:a", ["source", "target"]),
			mapConnection("direct:b", ["source", "target"], { predicate: "runs_on" }),
			mapConnection("direct:c", ["source", "target"], { predicate: "depends_on" }),
		];
		const projection = mapProjection(
			nodes.map((node) => ({ id: node.id, appearance: "compact" as const })),
			connections
		);
		const before = nodes.map(({ id, position, parentId }) => [id, position.x, position.y, parentId]);
		const firstRoutes = new ProjectedRouteCache().routesFor(nodes, projection);
		const presented = presentMapProjection({ nodes, edges: [] }, projection);
		const reorderedRoutes = new ProjectedRouteCache().routesFor(nodes, {
			...projection,
			connections: [...connections].reverse(),
		});
		const paths = connections.map((connection) => {
			const route = firstRoutes.get(connection.id)!;
			const firstSection = route.sections[route.sourceSectionIndex];
			const lastSection = route.sections[route.targetSectionIndex];
			expect(pointIsOnBounds(firstSection.start, worldBoundsByNodeId(nodes).get("source")!)).toBe(true);
			expect(pointIsOnBounds(lastSection.end, worldBoundsByNodeId(nodes).get("target")!)).toBe(true);
			return connectionRouteToSvgPath(route);
		});
		expect(new Set(connections.map(({ id }) => firstRoutes.get(id)!.sections[0].start.y)).size).toBe(3);
		expect(new Set(connections.map(({ id }) => firstRoutes.get(id)!.sections[0].end.y)).size).toBe(3);
		const corridorY = connections.map(({ id }) => {
			const section = firstRoutes.get(id)!.sections[0];
			const points = [section.start, ...section.bends, section.end];
			const horizontalSegments = points.slice(1).flatMap((end, index) => {
				const start = points[index];
				return start.y === end.y ? [{ y: start.y, length: Math.abs(end.x - start.x) }] : [];
			});
			return horizontalSegments.reduce((longest, segment) =>
				segment.length > longest.length ? segment : longest
			).y;
		});
		expect(new Set(corridorY).size).toBe(3);
		expect(connections.map(({ id }) => connectionRouteToSvgPath(reorderedRoutes.get(id)!))).toEqual(
			paths
		);
		expect(
			presented.nodes.map(({ id, position, parentId }) => [id, position.x, position.y, parentId])
		).toEqual(before);
		expect(nodes.map(({ id, position, parentId }) => [id, position.x, position.y, parentId])).toEqual(
			before
		);
	});

	test("keeps both path geometries for a canonical edge during a layer route transition", () => {
		const nodes = [
			flowNode("source", "system_function", { x: 0, y: 0 }, { width: 80, height: 60 }),
			flowNode("obstacle", "system", { x: 140, y: -20 }, { width: 100, height: 100 }),
			flowNode("target", "system_function", { x: 300, y: 0 }, { width: 80, height: 60 }),
		];
		const connection = mapConnection("direct:r", ["source", "target"]);
		const lowerProjection = mapProjection(
			[
				{ id: "source", appearance: "compact" },
				{ id: "target", appearance: "compact" },
			],
			[connection]
		);
		const upperProjection = mapProjection(
			[
				{ id: "source", appearance: "compact" },
				{ id: "obstacle", appearance: "compact" },
				{ id: "target", appearance: "compact" },
			],
			[connection]
		);
		const layout = presentMapProjection(
			{ nodes, edges: [] },
			upperProjection,
			new ProjectedRouteCache(),
			{ lowerProjection, upperProjection, upperLevel: 1, progress: 0.4 }
		);
		const edge = layout.edges[0];
		const variants = edge.data!.routeVariants!;

		expect(edge.id).toBe("direct:r");
		expect(variants.map((variant) => variant.opacity)).toEqual([0.6, 0.4]);
		expect(new Set(variants.map((variant) => connectionRouteToSvgPath(variant.route))).size).toBe(2);
	});

	test("crossfades cached route geometries and canonical summary/detail connections", async () => {
		const graph = relationshipExample.source;
		const allNearby = ["group-a", "group-b", "member-a", "member-a2", "member-b"];
		const lowerProjection = projectMap(graph, { detail: 1, nearbyEntityIds: allNearby });
		const upperProjection = projectMap(graph, { detail: 2, nearbyEntityIds: allNearby });
		const currentProjection = projectMap(graph, {
			detail: 1,
			visualDetail: 1.9,
			nearbyEntityIds: allNearby,
		});
		const routeCache = new ProjectedRouteCache();
		const layout = presentMapProjection(await layoutProjection(graph), currentProjection, routeCache, {
			lowerProjection,
			upperProjection,
			upperLevel: 2,
			progress: 0.5,
		});
		const summary = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-one")
		);
		const direct = layout.edges.find((edge) => edge.id === "direct:r-one");

		expect(summary?.id.startsWith("summary:")).toBe(true);
		expect(summary?.data?.connection.sourceRelationshipIds).toEqual(["r-one", "r-two"]);
		expect(summary?.data?.opacity).toBeGreaterThan(0);
		expect(summary?.data?.opacity).toBeLessThan(0.5);
		expect(direct?.id).toBe("direct:r-one");
		expect(direct?.data?.connection.sourceRelationshipIds).toEqual(["r-one"]);
		expect(direct?.data?.opacity).toBeCloseTo(0.5);
		expect(layout.edges.every((edge) => !edge.id.includes("--route-"))).toBe(true);
		const visibleNodeIds = new Set(layout.nodes.map((node) => node.id));
		expect(
			layout.edges.every((edge) => visibleNodeIds.has(edge.source) && visibleNodeIds.has(edge.target))
		).toBe(true);
	});

	test("reuses opacity-independent routes and bounds topology cache to four entries", () => {
		const nodes = [
			flowNode("source", "system", { x: 0, y: 0 }, { width: 80, height: 60 }),
			flowNode("target", "system", { x: 300, y: 0 }, { width: 80, height: 60 }),
		];
		const connection = mapConnection("direct:r", ["source", "target"]);
		const projection = mapProjection(
			[
				{ id: "source", appearance: "compact", opacity: 1 },
				{ id: "target", appearance: "compact", opacity: 1 },
			],
			[connection]
		);
		const cache = new ProjectedRouteCache();
		const initial = cache.routesFor(nodes, projection);
		const opacityOnly = {
			...projection,
			nodes: projection.nodes.map((node) => ({ ...node, opacity: 0.4 })),
			connections: projection.connections.map((edge) => ({ ...edge, opacity: 0.4 })),
		};
		expect(cache.routesFor(nodes, opacityOnly)).toBe(initial);
		expect(cache.size).toBe(1);

		for (let index = 0; index < 5; index += 1) {
			const variant = mapProjection(projection.nodes, [
				mapConnection(`direct:r-${index}`, ["source", "target"]),
			]);
			cache.routesFor(nodes, variant);
		}
		expect(cache.size).toBe(4);
	});

	test("keeps a dropped nested node over the pre-drop layout used during host adoption", async () => {
		const graph = sharedGroupsExample.source;
		const displayedLayout = await layoutProjection(graph);
		const original = worldPositionByNodeId(displayedLayout.nodes).get("code");
		expect(original).toBeDefined();
		if (!original) return;

		const drop = { x: original.x + 317, y: original.y - 143 };
		const pendingDrop = { entityId: "code", position: drop };
		const refreshedPresentation = nodesWithPendingDropPosition(displayedLayout.nodes, pendingDrop);
		const repeatedPresentation = nodesWithPendingDropPosition(displayedLayout.nodes, pendingDrop);
		const refreshedNode = nodeById(refreshedPresentation, "code");
		const aligned = alignLayoutToPrevious(
			displayedLayout,
			{ ...displayedLayout, nodes: refreshedPresentation },
			"code",
			"code"
		);

		expect(refreshedNode.parentId).toBeDefined();
		expect(worldPositionByNodeId(refreshedPresentation).get("code")).toEqual(drop);
		expect(worldPositionByNodeId(repeatedPresentation).get("code")).toEqual(drop);
		expect(worldBoundsByNodeId(refreshedPresentation).get("code")?.x).toBe(drop.x);
		expect(worldPositionByNodeId(aligned.nodes).get("code")).toEqual(drop);
		expect(nodeById(refreshedPresentation, "member-a").position).toEqual(
			nodeById(displayedLayout.nodes, "member-a").position
		);
	});

	test("creates distinct routes for direct, summary, reverse, and predicate-separated connections", async () => {
		const graph = relationshipExample.source;
		const projection = projectMap(graph, { detail: relationshipExample.detail, nearbyEntityIds: [] });
		const fullLayout = await layoutProjection(graph);
		const layout = presentMapProjection(fullLayout, projection);
		const parallel = projection.connections.filter(
			(connection) => connection.endpoints[0] === "group-a" && connection.endpoints[1] === "group-b"
		);

		expect(parallel.map((connection) => [connection.classification, connection.predicate])).toEqual([
			["direct", "calls"],
			["summary", "calls"],
			["summary", "depends_on"],
		]);
		expect(new Set(layout.edges.map((edge) => edge.id)).size).toBe(layout.edges.length);
		expectUsableRoutes(layout, projection);
		for (const edge of layout.edges.filter(
			(entry) => entry.data?.connection.classification === "summary"
		)) {
			const route = edge.data?.route;
			const bounds = worldBoundsByNodeId(layout.nodes);
			if (!route) throw new Error(`Missing summary route ${edge.id}`);
			expect(
				pointIsOnBounds(route.sections[route.sourceSectionIndex].start, bounds.get(edge.source)!)
			).toBe(true);
			expect(
				pointIsOnBounds(route.sections[route.targetSectionIndex].end, bounds.get(edge.target)!)
			).toBe(true);
		}

		const collapsed = presentMapProjection(
			fullLayout,
			projectMap(graph, { detail: 1, nearbyEntityIds: [] })
		);
		expect(worldBoundsByNodeId(collapsed.nodes).get("group-a")).toEqual(
			worldBoundsByNodeId(fullLayout.nodes).get("group-a")
		);

		const reverseEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-reverse")
		);
		expect(reverseEdge && [reverseEdge.source, reverseEdge.target]).toEqual(["group-b", "group-a"]);
	});

	test("applies controlled relationship highlights by canonical contributors after regrouping", async () => {
		const graph = relationshipExample.source;
		const projection = projectMap(graph, { detail: relationshipExample.detail, nearbyEntityIds: [] });
		const layout = presentMapProjection(await layoutProjection(graph), projection);
		const selected = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-one")
		);
		const unrelated = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-reverse")
		);
		if (!selected || !unrelated) throw new Error("Missing connection presentation fixture");

		expect(connectionLabel(selected.data!.connection)).toBe("calls · 2");
		expect(
			connectionPresentationState(selected, {
				selection: { relationshipIds: new Set(["r-one"]), endpointIds: new Set() },
			})
		).toMatchObject({ selected: true, highlighted: true, labelVisible: true });
		expect(
			connectionPresentationState(unrelated, {
				selection: { relationshipIds: new Set(["r-one"]), endpointIds: new Set() },
			}).dimmed
		).toBe(true);
		expect(
			connectionPresentationState(selected, {
				highlightedRelationshipIds: new Set(["r-two"]),
			}).highlighted
		).toBe(true);
		expect(
			connectionEndpointIds([selected, unrelated], {
				selection: { relationshipIds: new Set(["r-one"]), endpointIds: new Set() },
			})
		).toEqual(new Set([selected.source, selected.target]));
	});

	test("translates only root positions when aligning a full-subset layout", () => {
		const node = (id: string, x: number, y: number, parentId?: string): FlowNode => ({
			id,
			position: { x, y },
			...(parentId ? { parentId } : {}),
			width: 100,
			height: 60,
			data: {
				entity: { id, category: MapCategory.System, kind: "system" },
				appearance: "compact",
			},
			type: "system-map-node",
		});
		const previous: LayoutResult = {
			nodes: [node("anchor", 100, 100), node("other", 300, 100), node("child", 10, 20, "anchor")],
			edges: [],
		};
		const next: LayoutResult = {
			nodes: [node("anchor", 20, 40), node("other", 220, 40), node("child", 10, 20, "anchor")],
			edges: [],
		};
		const aligned = alignLayoutToPrevious(next, previous, "anchor");

		expect(aligned.nodes.map((entry) => entry.position)).toEqual([
			{ x: 100, y: 100 },
			{ x: 300, y: 100 },
			{ x: 10, y: 20 },
		]);
		expect(aligned.edges).toEqual([]);
		expect(alignLayoutToPrevious(next, previous)).toBe(next);
	});

	test("supports concurrent ELK layouts and rejects work after disposal", async () => {
		const graph = sharedGroupsExample.source;
		const projection = projectFullArchitecture(graph);
		const engine = createSystemMapLayoutEngine();
		try {
			const [first, second] = await Promise.all([
				engine.layout(graph, projection),
				engine.layout(graph, projection),
			]);
			expect(first.nodes.map((node) => node.id)).toEqual(second.nodes.map((node) => node.id));
			expect(first.edges.map((edge) => edge.id)).toEqual(second.edges.map((edge) => edge.id));
		} finally {
			engine.dispose();
		}
		await expect(engine.layout(graph, projection)).rejects.toThrow("disposed");
	});
});
