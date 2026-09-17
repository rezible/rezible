import { describe, expect, test } from "bun:test";
import { Position } from "@xyflow/svelte";

import { MapCategory, NodeDetailLevel } from "../category";
import { Coverage, type GraphSubset } from "../graph";
import { projectMap } from "../projection";
import { nodePresentationForEntity } from "$features/systems/components/system-map/map-node/presentation";
import { worldBoundsByNodeId, worldCenter, type Bounds, type Point } from "../geometry";
import { createSystemMapLayoutEngine } from "$features/systems/components/system-map/layout-engine";

import { alignLayoutToPrevious, COMPACT_NODE_SIZE } from "$features/systems/components/system-map/layout";
import {
	buildConnectionLanes,
	connectionHandleAssignment,
	getSystemMapConnectionPath,
} from "$features/systems/components/system-map/map-connection/geometry";
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

type PathSegment = { kind: "line" | "quadratic"; start: Point; end: Point };

const pathSegments = (path: string): PathSegment[] => {
	const tokens = path.match(/[MLQ]|-?(?:\d+(?:\.\d+)?|\.\d+)/g) ?? [];
	const segments: PathSegment[] = [];
	let cursor = 0;
	let current: Point | undefined;
	const readPoint = (): Point => {
		const x = Number(tokens[cursor++]);
		const y = Number(tokens[cursor++]);
		if (!Number.isFinite(x) || !Number.isFinite(y)) throw new Error(`Invalid path point in ${path}`);
		return { x, y };
	};

	while (cursor < tokens.length) {
		const command = tokens[cursor++];
		if (!command || !["M", "L", "Q"].includes(command))
			throw new Error(`Invalid path command in ${path}`);

		if (command === "M") {
			current = readPoint();
			continue;
		}
		if (!current) throw new Error(`Path command has no starting point in ${path}`);

		if (command === "L") {
			const end = readPoint();
			segments.push({ kind: "line", start: current, end });
			current = end;
			continue;
		}

		readPoint();
		const end = readPoint();
		segments.push({ kind: "quadratic", start: current, end });
		current = end;
	}

	return segments;
};

const horizontalSegments = (path: string) =>
	pathSegments(path).filter(
		(segment) =>
			segment.kind === "line" && segment.start.y === segment.end.y && segment.start.x !== segment.end.x
	);

const verticalSegments = (path: string) =>
	pathSegments(path).filter(
		(segment) =>
			segment.kind === "line" && segment.start.x === segment.end.x && segment.start.y !== segment.end.y
	);

describe("system map ELK layout", () => {
	test("keeps the Function root visible and readable while collapsing to Landscape", async () => {
		const overviewProjection = projectMap(
			sharedGroupsExample.source,
			{ detail: NodeDetailLevel.Landscape, nearbyEntityIds: [] },
			sharedGroupsExample.displayOptions
		);
		const overviewLayout = await layoutProjection(sharedGroupsExample.source, overviewProjection);
		const overviewRoot = nodeById(overviewLayout.nodes, "root");

		expect(overviewProjection.nodes.map((node) => node.id)).toEqual(["root"]);
		expect(overviewRoot.data.entity.category).toBe(MapCategory.Function);
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
		const alignedOverview = alignLayoutToPrevious(
			overviewLayout,
			detailedLayout,
			"root",
			"root"
		);

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
		expect(layout.nodes.findIndex((node) => node.id === "root")).toBeLessThan(
			layout.nodes.findIndex((node) => node.id === "group-a")
		);
		expect(layout.nodes.findIndex((node) => node.id === "group-a")).toBeLessThan(
			layout.nodes.findIndex((node) => node.id === "member-a")
		);

		const sharedMembershipEdges = layout.edges.filter((edge) =>
			edge.data?.connection.sourceRelationshipIds.some((id) => id.startsWith("m-shared"))
		);
		expect(sharedMembershipEdges.map((edge) => [edge.source, edge.target])).toEqual([
			["group-a", "shared-resource"],
			["group-b", "shared-resource"],
		]);
	});

	test("retains directed cross-group connections between nested descendants", async () => {
		const reveal = implementationReveal(relationshipExample.source);
		const projection = projectMap(relationshipExample.source, reveal, relationshipExample.displayOptions);
		const layout = await layoutProjection(relationshipExample.source, projection);

		expect(
			layout.edges.some((edge) => edge.data?.connection.sourceRelationshipIds.includes("r-one"))
		).toBe(true);
		expect(
			layout.edges.some((edge) => edge.data?.connection.sourceRelationshipIds.includes("r-reverse"))
		).toBe(true);
		expect(
			layout.edges.some((edge) => edge.data?.connection.sourceRelationshipIds.includes("r-other"))
		).toBe(true);
		for (const edge of layout.edges) {
			expect(layout.nodes.some((node) => node.id === edge.source)).toBe(true);
			expect(layout.nodes.some((node) => node.id === edge.target)).toBe(true);
		}
		const reverseEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-reverse")
		);
		if (!reverseEdge) throw new Error("Missing reverse edge");
		expect({ source: reverseEdge.source, target: reverseEdge.target }).toEqual({
			source: "member-b",
			target: "member-a",
		});
		expect(reverseEdge.data?.connection.endpoints).toEqual(["member-b", "member-a"]);
		expect(reverseEdge.data?.connection.sourceRelationshipIds).toEqual(["r-reverse"]);
	});

	test("assigns facing handles without changing canonical reverse direction", async () => {
		const bounds = (x: number, y: number): Bounds => ({
			x,
			y,
			...COMPACT_NODE_SIZE,
		});
		const forward = connectionHandleAssignment(bounds(0, 0), bounds(320, 0));
		const reverse = connectionHandleAssignment(bounds(320, 0), bounds(0, 0));
		const vertical = connectionHandleAssignment(bounds(0, 0), bounds(0, 320));
		const self = connectionHandleAssignment(bounds(0, 0), bounds(0, 0));

		expect(forward).toMatchObject({
			sourceHandle: "source-right",
			targetHandle: "target-left",
			sourcePosition: Position.Right,
			targetPosition: Position.Left,
		});
		expect(reverse).toMatchObject({
			sourceHandle: "source-left",
			targetHandle: "target-right",
			sourcePosition: Position.Left,
			targetPosition: Position.Right,
		});
		expect(vertical).toMatchObject({
			sourceHandle: "source-bottom",
			targetHandle: "target-top",
		});
		expect(self).toMatchObject({
			sourceHandle: "source-right",
			targetHandle: "target-bottom",
		});

		const projection = projectMap(
			relationshipExample.source,
			implementationReveal(relationshipExample.source),
			relationshipExample.displayOptions
		);
		const summaryProjection = projectMap(
			relationshipExample.source,
			{
				detail: relationshipExample.detail,
				nearbyEntityIds: relationshipExample.source.entities.map((entity) => entity.id),
			},
			relationshipExample.displayOptions
		);
		const summary = summaryProjection.connections.find(
			(connection) => connection.predicate === "calls" && connection.classification === "summary"
		);
		expect(summary).toBeDefined();
		expect(summary!.sourceRelationshipIds.length).toBe(2);
		const layout = await layoutProjection(relationshipExample.source, projection);
		const worldBounds = worldBoundsByNodeId(layout.nodes);
		for (const edge of layout.edges) {
			const assignment = connectionHandleAssignment(
				worldBounds.get(edge.source)!,
				worldBounds.get(edge.target)!
			);
			expect({ sourceHandle: edge.sourceHandle, targetHandle: edge.targetHandle }).toEqual({
				sourceHandle: assignment.sourceHandle,
				targetHandle: assignment.targetHandle,
			});
		}
	});

	test("derives connection labels and applies hover/selection presentation precedence", async () => {
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
		const hoveredEdge = layout.edges.find((edge) =>
			edge.data?.connection.sourceRelationshipIds.includes("r-other")
		);
		if (!directEdge || !selectedEdge || !unrelatedEdge || !hoveredEdge) {
			throw new Error("Missing presentation fixture edges");
		}
		const summaryProjection = projectMap(
			relationshipExample.source,
			{
				detail: relationshipExample.detail,
				nearbyEntityIds: relationshipExample.source.entities.map((entity) => entity.id),
			},
			relationshipExample.displayOptions
		);
		const summaryCall = summaryProjection.connections.find((connection) =>
			connection.sourceRelationshipIds.includes("r-reverse")
		);
		const summaryDependency = summaryProjection.connections.find((connection) =>
			connection.sourceRelationshipIds.includes("r-other")
		);
		if (!summaryCall || !summaryDependency) throw new Error("Missing summary fixture connections");

		expect(connectionLabel(directEdge.data!.connection)).toBe("calls");
		expect(connectionLabel(summaryCall)).toBe("calls · 1");
		expect(connectionLabel(summaryDependency)).toBe("depends on · 1");

		expect(connectionPresentationState(directEdge, {})).toEqual({
			selected: false,
			hovered: false,
			highlighted: false,
			dimmed: false,
			endpointEmphasized: false,
			labelVisible: false,
		});
		expect(
			connectionPresentationState(directEdge, { showAllConnectionLabels: true })
		).toMatchObject({
			highlighted: false,
			dimmed: false,
			labelVisible: true,
		});

		const edgeSelection = connectionPresentationState(selectedEdge, {
			selection: {
				edgeId: selectedEdge.id,
				endpointIds: new Set([selectedEdge.source, selectedEdge.target]),
			},
		});
		const unrelatedPresentation = connectionPresentationState(unrelatedEdge, {
			selection: {
				edgeId: selectedEdge.id,
				endpointIds: new Set([selectedEdge.source, selectedEdge.target]),
			},
		});
		const nodeSelection = connectionPresentationState(selectedEdge, {
			selection: { endpointIds: new Set([selectedEdge.source]) },
		});
		const hoveredWithNodeSelection = connectionPresentationState(selectedEdge, {
			selection: { endpointIds: new Set([selectedEdge.source]) },
			hoveredConnectionId: hoveredEdge.id,
		});
		const hoveredWithEdgeSelection = connectionPresentationState(hoveredEdge, {
			selection: {
				edgeId: selectedEdge.id,
				endpointIds: new Set([selectedEdge.source, selectedEdge.target]),
			},
			hoveredConnectionId: hoveredEdge.id,
		});
		const selectedAndHovered = connectionPresentationState(selectedEdge, {
			selection: {
				edgeId: selectedEdge.id,
				endpointIds: new Set([selectedEdge.source, selectedEdge.target]),
			},
			hoveredConnectionId: hoveredEdge.id,
		});

		expect(edgeSelection).toMatchObject({
			selected: true,
			hovered: false,
			highlighted: true,
			dimmed: false,
			endpointEmphasized: true,
			labelVisible: true,
		});
		expect(unrelatedPresentation).toMatchObject({
			selected: false,
			hovered: false,
			highlighted: false,
			dimmed: true,
			endpointEmphasized: false,
			labelVisible: false,
		});
		expect(nodeSelection).toMatchObject({
			selected: false,
			hovered: false,
			highlighted: true,
			dimmed: false,
			endpointEmphasized: false,
			labelVisible: false,
		});
		expect(hoveredWithNodeSelection).toMatchObject({
			highlighted: false,
			dimmed: true,
			labelVisible: false,
		});
		expect(hoveredWithEdgeSelection).toMatchObject({
			selected: false,
			hovered: true,
			highlighted: true,
			dimmed: false,
			endpointEmphasized: true,
			labelVisible: true,
		});
		expect(selectedAndHovered).toMatchObject({
			selected: true,
			hovered: false,
			highlighted: true,
			dimmed: false,
			endpointEmphasized: true,
			labelVisible: true,
		});

		expect(
			connectionEndpointIds([selectedEdge, unrelatedEdge], {
				selection: {
					edgeId: selectedEdge.id,
					endpointIds: new Set([selectedEdge.source, selectedEdge.target]),
				},
				hoveredConnectionId: unrelatedEdge.id,
			})
		).toEqual(new Set([
			selectedEdge.source,
			selectedEdge.target,
			unrelatedEdge.source,
			unrelatedEdge.target,
		]));
	});

	test("assigns stable lanes to direct, summary, predicate, and reverse connections", () => {
		const projection = projectMap(
			relationshipExample.source,
			implementationReveal(relationshipExample.source),
			relationshipExample.displayOptions
		);
		const first = buildConnectionLanes(projection.connections);
		const reversed = buildConnectionLanes([...projection.connections].reverse());
		const laneOffsets = projection.connections
			.filter(
				(connection) =>
					connection.endpoints.includes("group-a") && connection.endpoints.includes("group-b")
			)
			.map((connection) => first.get(connection.id)?.offset);

		expect(new Set(laneOffsets).size).toBe(laneOffsets.length);
		for (const connection of projection.connections) {
			expect(reversed.get(connection.id)).toEqual(first.get(connection.id));
		}
	});

	test("separates aligned forward lanes through distinct middle corridors", () => {
		const projection = projectMap(
			relationshipExample.source,
			{
				detail: relationshipExample.detail,
				nearbyEntityIds: relationshipExample.source.entities.map((entity) => entity.id),
			},
			relationshipExample.displayOptions
		);
		const lanes = buildConnectionLanes(projection.connections);
		const forwardConnections = projection.connections.filter(
			(connection) => connection.endpoints[0] === "group-a" && connection.endpoints[1] === "group-b"
		);

		expect(forwardConnections.map((connection) => connection.classification)).toEqual([
			"direct",
			"summary",
			"summary",
		]);
		expect(forwardConnections.map((connection) => connection.predicate)).toEqual([
			"calls",
			"calls",
			"depends_on",
		]);

		const corridors = forwardConnections.map((connection) => {
			const laneOffset = lanes.get(connection.id)!.offset;
			const [path, labelX, labelY] = getSystemMapConnectionPath({
				sourceX: 0,
				sourceY: 120,
				sourcePosition: Position.Right,
				targetX: 320,
				targetY: 120,
				targetPosition: Position.Left,
				laneOffset,
			});
			const corridor = horizontalSegments(path).find(
				(segment) => segment.start.y !== 120 && Math.abs(segment.end.x - segment.start.x) > 100
			);

			expect(corridor).toBeDefined();
			if (!corridor) throw new Error(`Missing separated corridor for ${connection.id}`);
			expect(labelY).toBe(corridor.start.y);
			expect(labelX).toBeGreaterThan(Math.min(corridor.start.x, corridor.end.x));
			expect(labelX).toBeLessThan(Math.max(corridor.start.x, corridor.end.x));
			return corridor;
		});

		expect(new Set(corridors.map((corridor) => corridor.start.y)).size).toBe(forwardConnections.length);
		expect(corridors.every((corridor) => Math.abs(corridor.end.x - corridor.start.x) > 100)).toBe(true);
	});

	test("retains geometric separation for reverse and nonaligned forward lanes", () => {
		const projection = projectMap(
			relationshipExample.source,
			{
				detail: relationshipExample.detail,
				nearbyEntityIds: relationshipExample.source.entities.map((entity) => entity.id),
			},
			relationshipExample.displayOptions
		);
		const lanes = buildConnectionLanes(projection.connections);
		const reverse = projection.connections.find(
			(connection) => connection.endpoints[0] === "group-b" && connection.endpoints[1] === "group-a"
		);
		expect(reverse).toBeDefined();
		if (!reverse) throw new Error("Missing reverse connection");

		const reverseLaneOffset = lanes.get(reverse.id)!.offset;
		const [reversePath, reverseLabelX, reverseLabelY] = getSystemMapConnectionPath({
			sourceX: 320,
			sourceY: 120,
			sourcePosition: Position.Right,
			targetX: 0,
			targetY: 120,
			targetPosition: Position.Left,
			laneOffset: reverseLaneOffset,
		});
		const reverseCorridor = horizontalSegments(reversePath).find((segment) => segment.start.y !== 120);
		expect(reverseCorridor).toBeDefined();
		if (!reverseCorridor) throw new Error("Missing reverse lane corridor");
		expect(reverseCorridor.start.y).toBe(120 + reverseLaneOffset);
		expect(reverseLabelY).toBe(reverseCorridor.start.y);
		expect(reverseLabelX).toBe(160);

		const nonalignedPaths = projection.connections
			.filter(
				(connection) => connection.endpoints[0] === "group-a" && connection.endpoints[1] === "group-b"
			)
			.map(
				(connection) =>
					getSystemMapConnectionPath({
						sourceX: 0,
						sourceY: 120,
						sourcePosition: Position.Right,
						targetX: 320,
						targetY: 240,
						targetPosition: Position.Left,
						laneOffset: lanes.get(connection.id)!.offset,
					})[0]
			);
		const nonalignedCorridorXs = nonalignedPaths.map((path) => {
			const corridor = verticalSegments(path).find(
				(segment) =>
					segment.start.x !== 0 &&
					segment.start.x !== 320 &&
					Math.abs(segment.end.y - segment.start.y) > 40
			);
			expect(corridor).toBeDefined();
			if (!corridor) throw new Error("Missing nonaligned lane corridor");
			return corridor.start.x;
		});
		expect(new Set(nonalignedCorridorXs).size).toBe(nonalignedCorridorXs.length);
	});

	test("keeps both facing vertical lane directions on their geometric corridors", () => {
		const verticalDirections = [
			{
				name: "bottom to top",
				sourceX: 120,
				sourceY: 0,
				sourcePosition: Position.Bottom,
				targetX: 120,
				targetY: 320,
				targetPosition: Position.Top,
			},
			{
				name: "top to bottom",
				sourceX: 120,
				sourceY: 320,
				sourcePosition: Position.Top,
				targetX: 120,
				targetY: 0,
				targetPosition: Position.Bottom,
			},
		] as const;

		for (const direction of verticalDirections) {
			const aligned = getSystemMapConnectionPath({
				...direction,
				laneOffset: 32,
			});
			const alignedCorridor = verticalSegments(aligned[0]).find(
				(segment) =>
					segment.start.x === aligned[1] &&
					Math.min(segment.start.y, segment.end.y) < aligned[2] &&
					aligned[2] < Math.max(segment.start.y, segment.end.y) &&
					Math.abs(segment.end.y - segment.start.y) > 100
			);

			expect(alignedCorridor, `Missing aligned corridor for ${direction.name}`).toBeDefined();
			if (!alignedCorridor) throw new Error(`Missing aligned corridor for ${direction.name}`);
			expect(aligned[1]).not.toBe(direction.sourceX);
			expect(aligned[2]).toBeGreaterThan(
				Math.min(alignedCorridor.start.y, alignedCorridor.end.y)
			);
			expect(aligned[2]).toBeLessThan(Math.max(alignedCorridor.start.y, alignedCorridor.end.y));

			const nonaligned = getSystemMapConnectionPath({
				...direction,
				targetX: 240,
				laneOffset: 32,
			});
			const nonalignedCorridor = horizontalSegments(nonaligned[0]).find(
				(segment) =>
					segment.start.y === nonaligned[2] &&
					Math.min(segment.start.x, segment.end.x) < nonaligned[1] &&
					nonaligned[1] < Math.max(segment.start.x, segment.end.x) &&
					Math.abs(segment.end.x - segment.start.x) > 100
			);

			expect(nonalignedCorridor, `Missing nonaligned corridor for ${direction.name}`).toBeDefined();
			if (!nonalignedCorridor) throw new Error(`Missing nonaligned corridor for ${direction.name}`);
			expect(nonaligned[2]).toBe(nonalignedCorridor.start.y);
			expect(nonaligned[1]).toBeGreaterThan(
				Math.min(nonalignedCorridor.start.x, nonalignedCorridor.end.x)
			);
			expect(nonaligned[1]).toBeLessThan(Math.max(nonalignedCorridor.start.x, nonalignedCorridor.end.x));
		}
	});

	test("places optional actors after the architectural layout", async () => {
		const architectureProjection = projectMap(
			relationshipExample.source,
			implementationReveal(relationshipExample.source),
			relationshipExample.displayOptions
		);
		const architectureLayout = await layoutProjection(relationshipExample.source, architectureProjection);

		const withActor: GraphSubset = {
			...relationshipExample.source,
			entities: [
				...relationshipExample.source.entities,
				{ id: "team", category: MapCategory.Actor, label: "Team", kind: "team" },
			],
			relationships: [
				...relationshipExample.source.relationships,
				{ id: "r-team", source: "team", target: "group-a", predicate: "owns" },
			],
			coverage: { parentMembership: Coverage.Complete, relationships: Coverage.Complete },
		};
		const withActorProjection = projectMap(withActor, implementationReveal(withActor), {
			showActors: true,
			showAnnotations: false,
		});
		const withActorLayout = await layoutProjection(withActor, withActorProjection);

		for (const architectureNode of architectureLayout.nodes) {
			const next = nodeById(withActorLayout.nodes, architectureNode.id);
			expect(next.position).toEqual(architectureNode.position);
			expect(next.parentId).toBe(architectureNode.parentId);
		}
		expect(nodeById(withActorLayout.nodes, "team").position.x).toBeGreaterThan(
			Math.max(...architectureLayout.nodes.map((node) => node.position.x + (node.width ?? 0)))
		);
	});

	test("keeps architecture geometry stable when annotation context changes", async () => {
		const reveal = implementationReveal(edgeCasesExample.source);
		const withoutAnnotations = projectMap(edgeCasesExample.source, reveal, {
			showActors: false,
			showAnnotations: false,
		});
		const withAnnotations = projectMap(edgeCasesExample.source, reveal, {
			showActors: false,
			showAnnotations: true,
		});
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
