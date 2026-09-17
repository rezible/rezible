import { getSmoothStepPath, Position, type GetSmoothStepPathParams } from "@xyflow/svelte";

import type { Bounds } from "$features/systems/lib/system-map/geometry";
import type { MapConnection } from "$features/systems/lib/system-map/presentation";

export type ConnectionHandleAssignment = {
	sourceHandle: string;
	targetHandle: string;
	sourcePosition: Position;
	targetPosition: Position;
};

const handleId = (type: "source" | "target", position: Position): string => `${type}-${position}`;

const centerOf = (bounds: Bounds) => ({
	x: bounds.x + bounds.width / 2,
	y: bounds.y + bounds.height / 2,
});

const assignmentForPositions = (
	sourcePosition: Position,
	targetPosition: Position
): ConnectionHandleAssignment => ({
	sourceHandle: handleId("source", sourcePosition),
	targetHandle: handleId("target", targetPosition),
	sourcePosition,
	targetPosition,
});

/** Chooses facing handles from settled world geometry without changing edge direction. */
export const connectionHandleAssignment = (
	sourceBounds: Bounds,
	targetBounds: Bounds
): ConnectionHandleAssignment => {
	const source = centerOf(sourceBounds);
	const target = centerOf(targetBounds);
	const deltaX = target.x - source.x;
	const deltaY = target.y - source.y;

	if (deltaX === 0 && deltaY === 0) {
		// A source self-relation needs two different sides to produce a visible loop.
		return assignmentForPositions(Position.Right, Position.Bottom);
	}

	if (Math.abs(deltaX) >= Math.abs(deltaY)) {
		return deltaX >= 0
			? assignmentForPositions(Position.Right, Position.Left)
			: assignmentForPositions(Position.Left, Position.Right);
	}

	return deltaY >= 0
		? assignmentForPositions(Position.Bottom, Position.Top)
		: assignmentForPositions(Position.Top, Position.Bottom);
};

export const CONNECTION_LANE_SPACING = 32;

export type ConnectionLane = {
	index: number;
	count: number;
	offset: number;
};

/** Groups both directions of a visible endpoint pair into one deterministic edge corridor. */
export const connectionLaneGroupKey = (connection: Pick<MapConnection, "endpoints">): string => {
	const [first, second] = connection.endpoints;
	return first < second ? `${first}\u0000${second}` : `${second}\u0000${first}`;
};

/** Assigns stable offsets without changing the source/projection connection order. */
export const buildConnectionLanes = (
	connections: readonly MapConnection[]
): ReadonlyMap<string, ConnectionLane> => {
	const groups = new Map<string, MapConnection[]>();
	for (const connection of connections) {
		const group = groups.get(connectionLaneGroupKey(connection)) ?? [];
		group.push(connection);
		groups.set(connectionLaneGroupKey(connection), group);
	}

	const lanes = new Map<string, ConnectionLane>();
	for (const group of groups.values()) {
		const ordered = group.slice().sort((left, right) => left.id.localeCompare(right.id));
		const count = ordered.length;
		for (const [index, connection] of ordered.entries()) {
			lanes.set(connection.id, {
				index,
				count,
				offset: (index - (count - 1) / 2) * CONNECTION_LANE_SPACING,
			});
		}
	}

	return lanes;
};

export type SystemMapConnectionPathParams = GetSmoothStepPathParams & {
	laneOffset?: number;
};

const DEFAULT_BORDER_RADIUS = 5;
const DEFAULT_EDGE_OFFSET = 20;

type CorridorAxis = "horizontal" | "vertical";

const alignedForwardCorridorPath = (
	params: Omit<SystemMapConnectionPathParams, "laneOffset">,
	laneOffset: number,
	axis: CorridorAxis
) => {
	const { sourceX, sourceY, targetX, targetY } = params;
	const isVertical = axis === "vertical";
	const sourceAlong = isVertical ? sourceY : sourceX;
	const targetAlong = isVertical ? targetY : targetX;
	const sourcePerpendicular = isVertical ? sourceX : sourceY;
	const direction = targetAlong > sourceAlong ? 1 : -1;
	const distance = Math.abs(targetAlong - sourceAlong);
	const edgeOffset = Math.max(0, params.offset ?? DEFAULT_EDGE_OFFSET);
	const gap = Math.min(edgeOffset, distance / 4);
	const corridorStart = sourceAlong + direction * gap;
	const corridorEnd = targetAlong - direction * gap;
	const lane = sourcePerpendicular + laneOffset;
	const corridorLength = Math.abs(corridorEnd - corridorStart);
	const bendSize = Math.min(
		params.borderRadius ?? DEFAULT_BORDER_RADIUS,
		Math.abs(laneOffset) / 2,
		corridorLength / 4
	);

	if (bendSize === 0 && isVertical) {
		const path = [
			`M${sourceX} ${sourceY}`,
			`L${sourceX} ${corridorStart}`,
			`L${lane} ${corridorStart}`,
			`L${lane} ${corridorEnd}`,
			`L${targetX} ${corridorEnd}`,
			`L${targetX} ${targetY}`,
		].join("");
		const labelAlong = (corridorStart + corridorEnd) / 2;
		return [path, lane, labelAlong, Math.abs(lane - sourceX), Math.abs(labelAlong - sourceY)] as const;
	}

	if (bendSize === 0) {
		const path = [
			`M${sourceX} ${sourceY}`,
			`L${corridorStart} ${sourceY}`,
			`L${corridorStart} ${lane}`,
			`L${corridorEnd} ${lane}`,
			`L${corridorEnd} ${targetY}`,
			`L${targetX} ${targetY}`,
		].join("");
		const labelAlong = (corridorStart + corridorEnd) / 2;
		return [path, labelAlong, lane, Math.abs(labelAlong - sourceX), Math.abs(lane - sourceY)] as const;
	}

	if (isVertical) {
		const horizontalDirection = lane > sourceX ? 1 : -1;
		const corridorStartY = corridorStart + direction * bendSize;
		const corridorEndY = corridorEnd - direction * bendSize;
		const path = [
			`M${sourceX} ${sourceY}`,
			`L${sourceX} ${corridorStart - direction * bendSize}`,
			`Q${sourceX} ${corridorStart} ${sourceX + horizontalDirection * bendSize} ${corridorStart}`,
			`L${lane - horizontalDirection * bendSize} ${corridorStart}`,
			`Q${lane} ${corridorStart} ${lane} ${corridorStartY}`,
			`L${lane} ${corridorEndY}`,
			`Q${lane} ${corridorEnd} ${lane - horizontalDirection * bendSize} ${corridorEnd}`,
			`L${targetX + horizontalDirection * bendSize} ${corridorEnd}`,
			`Q${targetX} ${corridorEnd} ${targetX} ${corridorEnd + direction * bendSize}`,
			`L${targetX} ${targetY}`,
		].join("");
		const labelY = (corridorStart + corridorEnd) / 2;
		return [path, lane, labelY, Math.abs(lane - sourceX), Math.abs(labelY - sourceY)] as const;
	}

	const verticalDirection = lane > sourceY ? 1 : -1;
	const corridorStartX = corridorStart;
	const corridorEndX = corridorEnd;
	const laneY = lane;
	const path = [
		`M${sourceX} ${sourceY}`,
		`L${corridorStartX - direction * bendSize} ${sourceY}`,
		`Q${corridorStartX} ${sourceY} ${corridorStartX} ${sourceY + verticalDirection * bendSize}`,
		`L${corridorStartX} ${laneY - verticalDirection * bendSize}`,
		`Q${corridorStartX} ${laneY} ${corridorStart} ${laneY}`,
		`L${corridorEnd} ${laneY}`,
		`Q${corridorEndX} ${laneY} ${corridorEndX} ${laneY - verticalDirection * bendSize}`,
		`L${corridorEndX} ${sourceY + verticalDirection * bendSize}`,
		`Q${corridorEndX} ${sourceY} ${corridorEndX + direction * bendSize} ${sourceY}`,
		`L${targetX} ${targetY}`,
	].join("");
	const labelX = (corridorStart + corridorEnd) / 2;
	return [path, labelX, laneY, Math.abs(labelX - sourceX), Math.abs(laneY - sourceY)] as const;
};

/** Applies lanes to the coordinate that changes the installed smooth-step route for each direction. */
export const getSystemMapConnectionPath = ({ laneOffset = 0, ...params }: SystemMapConnectionPathParams) => {
	const sourcePosition = params.sourcePosition ?? Position.Bottom;
	const targetPosition = params.targetPosition ?? Position.Top;
	const isOppositeHorizontal =
		(sourcePosition === Position.Right && targetPosition === Position.Left) ||
		(sourcePosition === Position.Left && targetPosition === Position.Right);

	if (isOppositeHorizontal) {
		const usesVerticalCorridor =
			(sourcePosition === Position.Right && params.sourceX < params.targetX) ||
			(sourcePosition === Position.Left && params.sourceX > params.targetX);
		const isAligned = params.sourceY === params.targetY;

		if (usesVerticalCorridor && isAligned && laneOffset !== 0) {
			return alignedForwardCorridorPath(params, laneOffset, "horizontal");
		}

		return getSmoothStepPath({
			...params,
			...(usesVerticalCorridor
				? { centerX: (params.sourceX + params.targetX) / 2 + laneOffset }
				: { centerY: (params.sourceY + params.targetY) / 2 + laneOffset }),
		});
	}

	const isOppositeVertical =
		(sourcePosition === Position.Bottom && targetPosition === Position.Top) ||
		(sourcePosition === Position.Top && targetPosition === Position.Bottom);

	if (isOppositeVertical) {
		const usesHorizontalCorridor =
			(sourcePosition === Position.Bottom && params.sourceY < params.targetY) ||
			(sourcePosition === Position.Top && params.sourceY > params.targetY);
		const isAligned = params.sourceX === params.targetX;

		if (usesHorizontalCorridor && isAligned && laneOffset !== 0) {
			return alignedForwardCorridorPath(params, laneOffset, "vertical");
		}

		return getSmoothStepPath({
			...params,
			...(usesHorizontalCorridor
				? { centerY: (params.sourceY + params.targetY) / 2 + laneOffset }
				: { centerX: (params.sourceX + params.targetX) / 2 + laneOffset }),
		});
	}

	const isHorizontal = sourcePosition === Position.Left || sourcePosition === Position.Right;
	return getSmoothStepPath({
		...params,
		...(isHorizontal
			? { centerY: (params.sourceY + params.targetY) / 2 + laneOffset }
			: { centerX: (params.sourceX + params.targetX) / 2 + laneOffset }),
	});
};
