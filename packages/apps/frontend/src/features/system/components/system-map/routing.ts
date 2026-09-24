import {
	worldBoundsByNodeId,
	worldCenter,
	type Bounds,
	type Point,
} from "$features/system/lib/system-map/geometry";
import type { MapConnection, MapProjection } from "$features/system/lib/system-map/presentation";
import type { ConnectionRoute, FlowNode } from "./flow-model";

const MAX_CACHED_ROUTE_TOPOLOGIES = 4;
const ROUTE_CLEARANCE = 24;
const ROUTE_LANE_OFFSET = 16;
const ROUTE_PORT_SPACING = 14;
const ROUTE_PORT_CORNER_CLEARANCE = 16;
const ROUTE_SHARED_SEGMENT_PENALTY = 8;
const ROUTE_NEAR_CORRIDOR_PENALTY = 2;
const ROUTE_CROSSING_PENALTY = 36;
const EPSILON = 0.001;

type RouteObstacle = Bounds;
type RouteSide = "left" | "right" | "top" | "bottom";
type RouteSegment = { start: Point; end: Point };

const ROUTE_SIDES: readonly RouteSide[] = ["left", "right", "top", "bottom"];

const portAssignmentKey = (connectionId: string, nodeId: string, side: RouteSide): string =>
	JSON.stringify([connectionId, nodeId, side]);

const topologyKey = (nodes: readonly FlowNode[], projection: MapProjection): string =>
	JSON.stringify([
		nodes
			.map((node) => [node.id, node.parentId ?? "", node.data.appearance])
			.sort(([left], [right]) => String(left).localeCompare(String(right))),
		projection.connections
			.map((connection) => [connection.id, connection.endpoints])
			.sort(([left], [right]) => String(left).localeCompare(String(right))),
	]);

const pointOnSide = (bounds: Bounds, side: RouteSide): Point => {
	const center = worldCenter(bounds);
	switch (side) {
		case "left":
			return { x: bounds.x, y: center.y };
		case "right":
			return { x: bounds.x + bounds.width, y: center.y };
		case "top":
			return { x: center.x, y: bounds.y };
		case "bottom":
			return { x: center.x, y: bounds.y + bounds.height };
	}
};

const preferredSides = (source: Bounds, target: Bounds): [RouteSide, RouteSide] => {
	const sourceCenter = worldCenter(source);
	const targetCenter = worldCenter(target);
	const dx = targetCenter.x - sourceCenter.x;
	const dy = targetCenter.y - sourceCenter.y;
	if (Math.abs(dx) >= Math.abs(dy)) return dx >= 0 ? ["right", "left"] : ["left", "right"];
	return dy >= 0 ? ["bottom", "top"] : ["top", "bottom"];
};

const isSamePoint = (left: Point, right: Point): boolean => left.x === right.x && left.y === right.y;

const compactPoints = (points: readonly Point[]): Point[] => {
	const compact: Point[] = [];
	for (const point of points) {
		if (compact.length && isSamePoint(compact[compact.length - 1], point)) continue;
		while (compact.length >= 2) {
			const previous = compact[compact.length - 2];
			const current = compact[compact.length - 1];
			if (
				(previous.x === current.x && current.x === point.x) ||
				(previous.y === current.y && current.y === point.y)
			) {
				compact.pop();
			} else break;
		}
		compact.push(point);
	}
	return compact;
};

const segmentHitsObstacle = (start: Point, end: Point, obstacle: RouteObstacle): boolean => {
	if (start.y === end.y) {
		const insideY = start.y > obstacle.y + EPSILON && start.y < obstacle.y + obstacle.height - EPSILON;
		const overlapsX =
			Math.max(Math.min(start.x, end.x), obstacle.x) <
			Math.min(Math.max(start.x, end.x), obstacle.x + obstacle.width) - EPSILON;
		return insideY && overlapsX;
	}
	if (start.x === end.x) {
		const insideX = start.x > obstacle.x + EPSILON && start.x < obstacle.x + obstacle.width - EPSILON;
		const overlapsY =
			Math.max(Math.min(start.y, end.y), obstacle.y) <
			Math.min(Math.max(start.y, end.y), obstacle.y + obstacle.height) - EPSILON;
		return insideX && overlapsY;
	}
	return true;
};

const pathHitsObstacle = (points: readonly Point[], obstacles: readonly RouteObstacle[]): boolean => {
	for (let index = 1; index < points.length; index += 1) {
		for (const obstacle of obstacles) {
			if (segmentHitsObstacle(points[index - 1], points[index], obstacle)) return true;
		}
	}
	return false;
};

const uniqueCoordinates = (values: readonly number[]): number[] => [...new Set(values)].sort((a, b) => a - b);

const portAssignmentsFor = (
	connections: readonly MapConnection[],
	boundsById: ReadonlyMap<string, Bounds>
): ReadonlyMap<string, Point> => {
	const groups = new Map<
		string,
		Array<{ connectionId: string; nodeId: string; side: RouteSide; otherCenter: Point; bounds: Bounds }>
	>();
	for (const connection of connections) {
		for (const [nodeId, otherId] of [
			[connection.endpoints[0], connection.endpoints[1]],
			[connection.endpoints[1], connection.endpoints[0]],
		] as const) {
			const bounds = boundsById.get(nodeId);
			const otherBounds = boundsById.get(otherId);
			if (!bounds || !otherBounds) continue;
			const otherCenter = worldCenter(otherBounds);
			for (const side of ROUTE_SIDES) {
				const key = JSON.stringify([nodeId, side]);
				const entries = groups.get(key) ?? [];
				entries.push({ connectionId: connection.id, nodeId, side, otherCenter, bounds });
				groups.set(key, entries);
			}
		}
	}

	const assignments = new Map<string, Point>();
	for (const entries of groups.values()) {
		const side = entries[0].side;
		const horizontalSide = side === "top" || side === "bottom";
		entries.sort((left, right) => {
			const coordinateDifference = horizontalSide
				? left.otherCenter.x - right.otherCenter.x
				: left.otherCenter.y - right.otherCenter.y;
			return coordinateDifference || left.connectionId.localeCompare(right.connectionId);
		});
		const count = entries.length;
		const extent = horizontalSide ? entries[0].bounds.width : entries[0].bounds.height;
		const maximumOffset = Math.max(0, extent / 2 - ROUTE_PORT_CORNER_CLEARANCE);
		const spacing = count < 2 ? 0 : Math.min(ROUTE_PORT_SPACING, (2 * maximumOffset) / (count - 1));
		for (let index = 0; index < count; index += 1) {
			const entry = entries[index];
			const offset = (index - (count - 1) / 2) * spacing;
			const center = pointOnSide(entry.bounds, entry.side);
			assignments.set(portAssignmentKey(entry.connectionId, entry.nodeId, entry.side), {
				x: center.x + (horizontalSide ? offset : 0),
				y: center.y + (horizontalSide ? 0 : offset),
			});
		}
	}
	return assignments;
};

const simpleRouteCandidates = (start: Point, end: Point): Point[][] => {
	const candidates: Point[][] = [];
	if (start.x === end.x || start.y === end.y) candidates.push([start, end]);
	candidates.push([start, { x: end.x, y: start.y }, end], [start, { x: start.x, y: end.y }, end]);
	return candidates.map(compactPoints);
};

const routeCandidates = (
	start: Point,
	end: Point,
	obstacles: readonly RouteObstacle[],
	occupiedSegments: readonly RouteSegment[]
): Point[][] => {
	const xLanes = uniqueCoordinates([
		(start.x + end.x) / 2,
		start.x,
		end.x,
		...occupiedSegments.flatMap((segment) =>
			segment.start.x === segment.end.x
				? [segment.start.x - ROUTE_LANE_OFFSET, segment.start.x + ROUTE_LANE_OFFSET]
				: []
		),
		...obstacles.flatMap((obstacle) => [
			obstacle.x - ROUTE_CLEARANCE,
			obstacle.x + obstacle.width + ROUTE_CLEARANCE,
		]),
	]);
	const yLanes = uniqueCoordinates([
		(start.y + end.y) / 2,
		start.y,
		end.y,
		...occupiedSegments.flatMap((segment) =>
			segment.start.y === segment.end.y
				? [segment.start.y - ROUTE_LANE_OFFSET, segment.start.y + ROUTE_LANE_OFFSET]
				: []
		),
		...obstacles.flatMap((obstacle) => [
			obstacle.y - ROUTE_CLEARANCE,
			obstacle.y + obstacle.height + ROUTE_CLEARANCE,
		]),
	]);
	const candidates = simpleRouteCandidates(start, end);
	for (const x of xLanes) {
		candidates.push([start, { x, y: start.y }, { x, y: end.y }, end]);
	}
	for (const y of yLanes) {
		candidates.push([start, { x: start.x, y }, { x: end.x, y }, end]);
	}
	return candidates.map(compactPoints);
};

const overlapLength = (leftStart: number, leftEnd: number, rightStart: number, rightEnd: number): number =>
	Math.max(
		0,
		Math.min(Math.max(leftStart, leftEnd), Math.max(rightStart, rightEnd)) -
			Math.max(Math.min(leftStart, leftEnd), Math.min(rightStart, rightEnd))
	);

const routeTrafficScore = (points: readonly Point[], occupiedSegments: readonly RouteSegment[]): number => {
	let score = 0;
	for (let index = 1; index < points.length; index += 1) {
		const start = points[index - 1];
		const end = points[index];
		const horizontal = start.y === end.y;
		for (const occupied of occupiedSegments) {
			const occupiedHorizontal = occupied.start.y === occupied.end.y;
			if (horizontal === occupiedHorizontal) {
				const leftLane = horizontal ? start.y : start.x;
				const rightLane = horizontal ? occupied.start.y : occupied.start.x;
				const overlap = horizontal
					? overlapLength(start.x, end.x, occupied.start.x, occupied.end.x)
					: overlapLength(start.y, end.y, occupied.start.y, occupied.end.y);
				if (overlap <= EPSILON) continue;
				const laneDistance = Math.abs(leftLane - rightLane);
				score +=
					laneDistance <= EPSILON
						? overlap * ROUTE_SHARED_SEGMENT_PENALTY
						: overlap *
							ROUTE_NEAR_CORRIDOR_PENALTY *
							Math.max(0, 1 - laneDistance / ROUTE_LANE_OFFSET);
				continue;
			}

			const horizontalSegment = horizontal ? { start, end } : occupied;
			const verticalSegment = horizontal ? occupied : { start, end };
			const crossingX = verticalSegment.start.x;
			const crossingY = horizontalSegment.start.y;
			if (
				crossingX > Math.min(horizontalSegment.start.x, horizontalSegment.end.x) + EPSILON &&
				crossingX < Math.max(horizontalSegment.start.x, horizontalSegment.end.x) - EPSILON &&
				crossingY > Math.min(verticalSegment.start.y, verticalSegment.end.y) + EPSILON &&
				crossingY < Math.max(verticalSegment.start.y, verticalSegment.end.y) - EPSILON
			) {
				score += ROUTE_CROSSING_PENALTY;
			}
		}
	}
	return score;
};

const candidateScore = (points: readonly Point[], occupiedSegments: readonly RouteSegment[]): number => {
	let distance = 0;
	for (let index = 1; index < points.length; index += 1) {
		distance += Math.abs(points[index].x - points[index - 1].x);
		distance += Math.abs(points[index].y - points[index - 1].y);
	}
	return distance + Math.max(0, points.length - 2) * 28 + routeTrafficScore(points, occupiedSegments);
};

const enclosingGroupIds = (endpointIds: readonly string[], nodesById: ReadonlyMap<string, FlowNode>) => {
	const groupIds = new Set<string>();
	for (const endpointId of endpointIds) {
		let parentId = nodesById.get(endpointId)?.parentId;
		while (parentId) {
			const parent = nodesById.get(parentId);
			if (!parent) break;
			if (parent.data.appearance === "group") groupIds.add(parent.id);
			parentId = parent.parentId;
		}
	}
	return groupIds;
};

const labelPositionForRoute = (points: readonly Point[]): Point => {
	const segments = points.slice(1).map((end, index) => {
		const start = points[index];
		return {
			start,
			end,
			length: Math.abs(end.x - start.x) + Math.abs(end.y - start.y),
			horizontal: start.y === end.y,
		};
	});
	const horizontal = segments.filter((segment) => segment.horizontal);
	const candidates = horizontal.length ? horizontal : segments;
	if (!candidates.length) return points[0] ?? { x: 0, y: 0 };
	const longest = candidates.reduce((best, segment) => (segment.length > best.length ? segment : best));
	return { x: (longest.start.x + longest.end.x) / 2, y: (longest.start.y + longest.end.y) / 2 };
};

const routeConnection = (
	connection: MapConnection,
	nodes: readonly FlowNode[],
	boundsById: ReadonlyMap<string, Bounds>,
	portAssignments: ReadonlyMap<string, Point>,
	occupiedSegments: readonly RouteSegment[]
): ConnectionRoute => {
	const source = boundsById.get(connection.endpoints[0]);
	const target = boundsById.get(connection.endpoints[1]);
	if (!source || !target) throw new Error(`Missing visible route endpoint for ${connection.id}`);

	const nodesById = new Map(nodes.map((node) => [node.id, node]));
	const traversableGroupIds = enclosingGroupIds(connection.endpoints, nodesById);
	const obstacles = nodes
		.filter(
			(node) =>
				node.id !== connection.endpoints[0] &&
				node.id !== connection.endpoints[1] &&
				!(node.data.appearance === "group" && traversableGroupIds.has(node.id))
		)
		.map((node) => boundsById.get(node.id))
		.filter((bounds): bounds is Bounds => Boolean(bounds));

	const [preferredSourceSide, preferredTargetSide] = preferredSides(source, target);
	const portFor = (connectionNodeId: string, side: RouteSide, bounds: Bounds) =>
		portAssignments.get(portAssignmentKey(connection.id, connectionNodeId, side)) ??
		pointOnSide(bounds, side);
	const preferredSourcePort = portFor(connection.endpoints[0], preferredSourceSide, source);
	const preferredTargetPort = portFor(connection.endpoints[1], preferredTargetSide, target);
	const portPairs: Array<[Point, Point]> = [[preferredSourcePort, preferredTargetPort]];
	const scoreUsable = (candidates: readonly Point[][]) =>
		candidates
			.filter((points) => !pathHitsObstacle(points, obstacles))
			.map((points) => ({ points, score: candidateScore(points, occupiedSegments) }));
	// Try the facing ports first for the common direct case; consider alternate sides only when blocked.
	const preferredSimpleCandidates = simpleRouteCandidates(preferredSourcePort, preferredTargetPort);
	let preferredCandidates = preferredSimpleCandidates;
	let usable = scoreUsable(preferredSimpleCandidates);
	if (!usable.length || occupiedSegments.length) {
		preferredCandidates = routeCandidates(
			preferredSourcePort,
			preferredTargetPort,
			obstacles,
			occupiedSegments
		);
		usable = scoreUsable([...preferredSimpleCandidates, ...preferredCandidates]);
	}
	if (!usable.length) {
		for (const sourceSide of ROUTE_SIDES) {
			for (const targetSide of ROUTE_SIDES) {
				if (sourceSide === preferredSourceSide && targetSide === preferredTargetSide) continue;
				portPairs.push([
					portFor(connection.endpoints[0], sourceSide, source),
					portFor(connection.endpoints[1], targetSide, target),
				]);
			}
		}
		usable = portPairs.slice(1).flatMap(([start, end]) => {
			const simple = scoreUsable(simpleRouteCandidates(start, end));
			return simple.length
				? simple
				: scoreUsable(routeCandidates(start, end, obstacles, occupiedSegments));
		});
	}
	if (!usable.length) {
		// Overlapping rectangles can leave no clear orthogonal corridor; keep a deterministic direct route.
		usable = preferredCandidates.map((points) => ({
			points,
			score: candidateScore(points, occupiedSegments),
		}));
	}
	usable.sort((left, right) => left.score - right.score);
	const points = usable[0].points;
	const section = { start: points[0], bends: points.slice(1, -1), end: points.at(-1)! };
	return {
		sections: [section],
		labelPosition: labelPositionForRoute(points),
		sourceSectionIndex: 0,
		targetSectionIndex: 0,
	};
};

/** Routes visible connections against only the rectangles in that visible projection. */
export class ProjectedRouteCache {
	private readonly routesByTopology = new Map<string, ReadonlyMap<string, ConnectionRoute>>();

	get size(): number {
		return this.routesByTopology.size;
	}

	routesFor(nodes: readonly FlowNode[], projection: MapProjection): ReadonlyMap<string, ConnectionRoute> {
		const key = topologyKey(nodes, projection);
		const cached = this.routesByTopology.get(key);
		if (cached) {
			this.routesByTopology.delete(key);
			this.routesByTopology.set(key, cached);
			return cached;
		}

		const boundsById = worldBoundsByNodeId(nodes);
		const routes = new Map<string, ConnectionRoute>();
		const orderedConnections = [...projection.connections].sort((left, right) =>
			left.id.localeCompare(right.id)
		);
		const portAssignments = portAssignmentsFor(orderedConnections, boundsById);
		const occupiedSegments: RouteSegment[] = [];
		for (const connection of orderedConnections) {
			const route = routeConnection(connection, nodes, boundsById, portAssignments, occupiedSegments);
			routes.set(connection.id, route);
			for (const section of route.sections) {
				const points = [section.start, ...section.bends, section.end];
				for (let index = 1; index < points.length; index += 1) {
					occupiedSegments.push({ start: points[index - 1], end: points[index] });
				}
			}
		}
		this.routesByTopology.set(key, routes);
		while (this.routesByTopology.size > MAX_CACHED_ROUTE_TOPOLOGIES) {
			const oldestKey = this.routesByTopology.keys().next().value;
			if (oldestKey === undefined) break;
			this.routesByTopology.delete(oldestKey);
		}
		return routes;
	}
}
