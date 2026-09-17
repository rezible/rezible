import { describe, expect, test } from "bun:test";

import { MapCategory, NodeDetailLevel } from "../category";
import type { GraphSubset } from "../graph";
import { projectMap } from "../projection";
import { nodePresentationForEntity } from "$features/systems/components/system-map/map-node/presentation";
import { worldBoundsByNodeId, worldCenter, type Bounds, type Point } from "../geometry";
import { createSystemMapLayoutEngine } from "$features/systems/components/system-map/layout-engine";
import {
	alignLayoutToPrevious,
	COMPACT_NODE_SIZE,
} from "$features/systems/components/system-map/layout";
import type {
	ConnectionRoute,
	FlowEdge,
	FlowNode,
	LayoutResult,
} from "$features/systems/components/system-map/flow-model";
import { connectionRouteToSvgPath } from "$features/systems/components/system-map/map-connection/geometry";
import {
	connectionEndpointIds,
	connectionLabel,
	connectionPresentationState,
} from "$features/systems/components/system-map/map-connection/presentation";

import { edgeCasesExample, layoutProjection, relationshipExample, sharedGroupsExample } from "./test-fixtures";

const nodeById = <T extends { id: string }>(items: readonly T[], id: string): T => {
	const item = items.find((candidate) => candidate.id === id);
	if (!item) throw new Error(`Missing item ${id}`);
	return item;
};

const implementationReveal = (source: GraphSubset) => ({
	detail: NodeDetailLevel.Implementation,
	nearbyEntityIds: source.entities.map((entity) => entity.id),
});

const routePoints = (section: ConnectionRoute["sections"][number]): Point[] => [
	section.start,
	...section.bends,
	section.end,
];

const pointOnBoundary = (point: Point, bounds: Bounds): boolean => {
	const epsilon = 0.001;
	const onHorizontal =
		(Math.abs(point.y - bounds.y) <= epsilon || Math.abs(point.y - (bounds.y + bounds.height)) <= epsilon) &&
		point.x >= bounds.x - epsilon &&
		point.x <= bounds.x + bounds.width + epsilon;
	const onVertical =
		(Math.abs(point.x - bounds.x) <= epsilon || Math.abs(point.x - (bounds.x + bounds.width)) <= epsilon) &&
		point.y >= bounds.y - epsilon &&
		point.y <= bounds.y + bounds.height + epsilon;
	return onHorizontal || onVertical;
};

const pointOnSegment = (point: Point, start: Point, end: Point): boolean => {
	const epsilon = 0.001;
	const within = (value: number, left: number, right: number) =>
		value >= Math.min(left, right) - epsilon && value <= Math.max(left, right) + epsilon;
	return (
		(start.x === end.x && Math.abs(point.x - start.x) <= epsilon && within(point.y, start.y, end.y)) ||
		(start.y === end.y && Math.abs(point.y - start.y) <= epsilon && within(point.x, start.x, end.x))
	);
};

const expectUsableRoutes = (layout: LayoutResult, projection: ReturnType<typeof projectMap>) => {
	const boundsById = worldBoundsByNodeId(layout.nodes);
	expect(layout.edges.map((edge) => edge.id)).toEqual(projection.connections.map((connection) => connection.id));

	for (const edge of layout.edges) {
		const connection = edge.data?.connection;
		const route = edge.data?.route;
		expect(connection).toBeDefined();
		expect(route).toBeDefined();
		if (!connection || !route) continue;

		expect({ source: edge.source, target: edge.target }).toEqual({
			source: connection.endpoints[0],
			target: connection.endpoints[1],
		});
		expect(route.sections.length).toBeGreaterThan(0);
		expect(route.targetSectionIndex).toBeGreaterThanOrEqual(0);
		expect(route.targetSectionIndex).toBeLessThan(route.sections.length);
		expect(Number.isFinite(route.labelPosition.x)).toBe(true);
		expect(Number.isFinite(route.labelPosition.y)).toBe(true);
		expect(connectionRouteToSvgPath(route)).toMatch(/^M/);

		const allSegments = route.sections.flatMap((section) => {
			const points = routePoints(section);
			return points.slice(1).map((end, index) => ({ start: points[index], end }));
		});
		for (const segment of allSegments) {
			expect(segment.start.x === segment.end.x || segment.start.y === segment.end.y).toBe(true);
		}
		const hasHorizontalSegment = allSegments.some(
			({ start, end }) => start.y === end.y && start.x !== end.x
		);
		expect(
			allSegments.some(({ start, end }) => pointOnSegment(route.labelPosition, start, end))
		).toBe(true);
		if (hasHorizontalSegment) {
			expect(
				allSegments.some(
					({ start, end }) =>
						start.y === end.y && start.x !== end.x && pointOnSegment(route.labelPosition, start, end)
				)
			).toBe(true);
		}

		const sourceBounds = boundsById.get(connection.endpoints[0]);
		const targetBounds = boundsById.get(connection.endpoints[1]);
		expect(sourceBounds).toBeDefined();
		expect(targetBounds).toBeDefined();
		if (!sourceBounds || !targetBounds) continue;
		expect(route.sections.some((section) => pointOnBoundary(section.start, sourceBounds))).toBe(true);
		expect(pointOnBoundary(route.sections[route.targetSectionIndex].end, targetBounds)).toBe(true);
	}
};

describe("system map ELK layout", () => {
	test("keeps route sections separate and emits the target section last for the SVG marker", () => {
		const route: ConnectionRoute = {
			sections: [
				{ start: { x: 20, y: 20 }, bends: [], end: { x: 40, y: 20 } },
				{ start: { x: 60, y: 40 }, bends: [{ x: 80, y: 40 }], end: { x: 100, y: 40 } },
			],
			labelPosition: { x: 80, y: 40 },
			targetSectionIndex: 0,
		};

		const path = connectionRouteToSvgPath(route);
		expect(path.match(/M/g)).toHaveLength(2);
		expect(path.startsWith("M60 40L80 40L100 40M20 20L40 20")).toBe(true);
	});

	test("keeps the Function root visible and readable while collapsing to Landscape", async () => {
		const overviewProjection = projectMap(
			sharedGroupsExample.source,
			{ detail: NodeDetailLevel.Landscape, nearbyEntityIds: [] },
			sharedGroupsExample.displayOptions
		);
		const overviewLayout = await layoutProjection(sharedGroupsExample.source, overviewProjection);
		const overviewRoot = nodeById(overviewLayout.nodes, "root");

		expect(overviewProjection.nodes.map((node) => node.id)).toEqual(["root"]);
		expect(overviewRoot.data.entity.category).toBe(MapCategory.SystemFunction);
		expect(nodePresentationForEntity(overviewRoot.data.entity)).toMatchObject({
		label: "root",
		categoryLabel: "Function",
	});

		const detailedProjection = projectMap(
			sharedGroupsExample.source,
			implementationReveal(sharedGroupsExample.source),
			sharedGroupsExample.displayOptions
		);
		const detailedLayout = await layoutProjection(sharedGroupsExample.source, detailedProjection);
		const alignedOverview = alignLayoutToPrevious(overviewLayout, detailedLayout, "root", "root");

		expect(worldCenter(worldBoundsByNodeId(alignedOverview.nodes).get("root")!)).toEqual(
			worldCenter(worldBoundsByNodeId(detailedLayout.nodes).get("root")!)
		);
	});

	test("uses the shared readable dimensions for every compact node", async () => {
		const projection = projectMap(
			sharedGroupsExample.source,
			implementationReveal(sharedGroupsExample.source),
			sharedGroupsExample.displayOptions
		);
		const layout = await layoutProjection(sharedGroupsExample.source, projection);
		const compactNodes = layout.nodes.filter((node) => node.data.appearance === "compact");

		expect(compactNodes.length).toBeGreaterThan(0);
		for (const node of compactNodes) {
			expect({ width: node.width, height: node.height }).toEqual(COMPACT_NODE_SIZE);
		}
	});

	test("keeps shared nodes in common context while nesting only exclusive memberships", async () => {
		const projection = projectMap(
			sharedGroupsExample.source,
			implementationReveal(sharedGroupsExample.source),
			sharedGroupsExample.displayOptions
		);
		const layout = await layoutProjection(sharedGroupsExample.source, projection);

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
		const containedBy = (child: typeof memberA, parent: typeof groupA) => {
			expect(child.position.x + (child.width ?? 0)).toBeLessThanOrEqual(parent.width ?? 0);
			expect(child.position.y + (child.height ?? 0)).toBeLessThanOrEqual(parent.height ?? 0);
		};
		containedBy(memberA, groupA);
		containedBy(code, memberA);
		expect(groupA.position.x + (groupA.width ?? 0)).toBeLessThanOrEqual(root.width ?? 0);
		expect(groupA.position.y + (groupA.height ?? 0)).toBeLessThanOrEqual(root.height ?? 0);

		const sharedMembershipEdges = layout.edges.filter((edge) =>
			edge.data?.connection.sourceRelationshipIds.some((id) => id.startsWith("m-shared"))
		);
		expect(sharedMembershipEdges.map((edge) => [edge.source, edge.target])).toEqual([
			["group-a", "shared-resource"],
			["group-b", "shared-resource"],
		]);
		expectUsableRoutes(layout, projection);
	});

	test("retains one ELK route per direct, summary, predicate, and reverse connection", async () => {
		const projection = projectMap(
			relationshipExample.source,
			implementationReveal(relationshipExample.source),
			relationshipExample.displayOptions
		);
		const layout = await layoutProjection(relationshipExample.source, projection);
		const summaryProjection = projectMap(
			relationshipExample.source,
			{
				detail: relationshipExample.detail,
				nearbyEntityIds: relationshipExample.source.entities.map((entity) => entity.id),
			},
			relationshipExample.displayOptions
		);
		const summaryLayout = await layoutProjection(relationshipExample.source, summaryProjection);

		const parallel = summaryProjection.connections.filter(
			(connection) => connection.endpoints[0] === "group-a" && connection.endpoints[1] === "group-b"
		);
		expect(parallel.map((connection) => [connection.classification, connection.predicate])).toEqual([
			["direct", "calls"],
			["summary", "calls"],
			["summary", "depends_on"],
		]);
		expect(new Set(parallel.map((connection) => connection.id)).size).toBe(parallel.length);
		expect(new Set(summaryLayout.edges.map((edge) => edge.id)).size).toBe(summaryLayout.edges.length);
		expectUsableRoutes(summaryLayout, summaryProjection);
		expectUsableRoutes(layout, projection);

		const reverseEdge = summaryLayout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-reverse")
		);
		expect(reverseEdge).toBeDefined();
		if (!reverseEdge) return;
		expect({ source: reverseEdge.source, target: reverseEdge.target }).toEqual({
			source: "group-b",
			target: "group-a",
		});
		expect(reverseEdge.data?.connection.endpoints).toEqual(["group-b", "group-a"]);
		const detailedReverseEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-reverse")
		);
		expect(detailedReverseEdge?.data?.connection.endpoints).toEqual(["member-b", "member-a"]);
	});

	test("derives labels and preserves routed edges through hover and selection presentation", async () => {
		const projection = projectMap(
			relationshipExample.source,
			implementationReveal(relationshipExample.source),
			relationshipExample.displayOptions
		);
		const layout = await layoutProjection(relationshipExample.source, projection);
		const directEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-direct")
		);
		const selectedEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-one")
		);
		const unrelatedEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-reverse")
		);
		if (!directEdge || !selectedEdge || !unrelatedEdge) throw new Error("Missing presentation fixture edge");
		const routeBefore = directEdge.data!.route;

		expect(connectionLabel(directEdge.data!.connection)).toBe("calls");
		expect(connectionPresentationState(directEdge, {})).toEqual({
			selected: false,
			hovered: false,
			highlighted: false,
			dimmed: false,
			endpointEmphasized: false,
			labelVisible: false,
		});
		expect(connectionPresentationState(directEdge, { showAllConnectionLabels: true })).toMatchObject({
			labelVisible: true,
		});
		expect(connectionPresentationState(selectedEdge, {
			selection: { edgeId: selectedEdge.id, endpointIds: new Set([selectedEdge.source, selectedEdge.target]) },
		})).toMatchObject({ selected: true, highlighted: true, dimmed: false, labelVisible: true });
		expect(connectionPresentationState(unrelatedEdge, {
			selection: { edgeId: selectedEdge.id, endpointIds: new Set([selectedEdge.source, selectedEdge.target]) },
		})).toMatchObject({ selected: false, dimmed: true, labelVisible: false });
		expect(directEdge.data!.route).toBe(routeBefore);
		expect(connectionEndpointIds([selectedEdge, unrelatedEdge], {
			selection: { edgeId: selectedEdge.id, endpointIds: new Set([selectedEdge.source, selectedEdge.target]) },
		})).toEqual(new Set([selectedEdge.source, selectedEdge.target, unrelatedEdge.source, unrelatedEdge.target]));
	});

	test("translates root nodes and world route points while preserving child-relative positions", () => {
		const node = (id: string, x: number, y: number, parentId?: string): FlowNode => ({
			id,
			position: { x, y },
			...(parentId ? { parentId } : {}),
			width: 100,
			height: 60,
			data: {
				entity: { id, category: MapCategory.System, label: id, kind: "system" },
				appearance: "compact",
				annotationCount: 0,
			},
			type: "system-map-node",
		});
		const route: ConnectionRoute = {
			sections: [{ start: { x: 120, y: 130 }, bends: [{ x: 220, y: 130 }], end: { x: 320, y: 190 } }],
			labelPosition: { x: 220, y: 130 },
			targetSectionIndex: 0,
		};
		const edge = (edgeRoute: ConnectionRoute): FlowEdge => ({
			id: "edge",
			source: "anchor",
			target: "other",
			data: { connection: {} as never, route: edgeRoute },
			type: "system-map-connection",
		});
		const previous: LayoutResult = {
			nodes: [node("anchor", 100, 100), node("other", 300, 100), node("child", 10, 20, "anchor")],
			edges: [edge(route)],
		};
		const nextRoute: ConnectionRoute = {
			...route,
			sections: [{ start: { x: 40, y: 70 }, bends: [{ x: 140, y: 70 }], end: { x: 240, y: 130 } }],
			labelPosition: { x: 140, y: 70 },
		};
		const next: LayoutResult = {
			nodes: [node("anchor", 20, 40), node("other", 220, 40), node("child", 10, 20, "anchor")],
			edges: [edge(nextRoute)],
		};
		const aligned = alignLayoutToPrevious(next, previous, "anchor");

		expect(aligned.nodes.map((entry) => entry.position)).toEqual([
			{ x: 100, y: 100 },
			{ x: 300, y: 100 },
			{ x: 10, y: 20 },
		]);
		expect(aligned.edges[0].data!.route).toEqual({
			...nextRoute,
			sections: [{ start: { x: 120, y: 130 }, bends: [{ x: 220, y: 130 }], end: { x: 320, y: 190 } }],
			labelPosition: { x: 220, y: 130 },
		});

		const unanchored = alignLayoutToPrevious(next, previous);
		expect(unanchored).toBe(next);
	});

	test("keeps architecture geometry stable when annotation context changes", async () => {
		const reveal = implementationReveal(edgeCasesExample.source);
		const withoutAnnotations = projectMap(edgeCasesExample.source, reveal, { showAnnotations: false });
		const withAnnotations = projectMap(edgeCasesExample.source, reveal, { showAnnotations: true });
		const withoutLayout = await layoutProjection(edgeCasesExample.source, withoutAnnotations);
		const withLayout = await layoutProjection(edgeCasesExample.source, withAnnotations);

		expect(withAnnotations.annotations.length).toBeGreaterThan(0);
		for (const node of withoutLayout.nodes) {
			const next = nodeById(withLayout.nodes, node.id);
			expect(next.position).toEqual(node.position);
			expect(next.parentId).toBe(node.parentId);
			expect(next.width).toBe(node.width);
			expect(next.height).toBe(node.height);
		}
	});

	test("supports concurrent layouts and explicit engine disposal", async () => {
		const projection = projectMap(
			sharedGroupsExample.source,
			implementationReveal(sharedGroupsExample.source),
			sharedGroupsExample.displayOptions
		);
		const engine = createSystemMapLayoutEngine();

		try {
			const [first, second] = await Promise.all([
				engine.layout(sharedGroupsExample.source, projection),
				engine.layout(sharedGroupsExample.source, projection),
			]);

			expect(first.nodes.map((node) => node.id)).toEqual(second.nodes.map((node) => node.id));
			expect(first.edges.map((edge) => edge.id)).toEqual(second.edges.map((edge) => edge.id));
		} finally {
			engine.dispose();
		}

		await expect(engine.layout(sharedGroupsExample.source, projection)).rejects.toThrow("disposed");
	});
});
