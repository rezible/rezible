import { worldPositionByNodeId, type Point } from "$features/system/lib/system-map/geometry";
import type { MapNodeAppearance } from "$features/system/lib/system-map/presentation";
import type { FlowNode } from "./flow-model";

export type NodeMoveCallback = (entityId: string, position: Point) => void;
export type PendingNodeDrop = { entityId: string; position: Point };

export const nodeCanBeDragged = (appearance: MapNodeAppearance, hasMoveCallback: boolean): boolean =>
	hasMoveCallback && appearance === "compact";

export const nextLayoutRequestId = (currentId: number): number => currentId + 1;

export const layoutRequestIsStale = (
	requestId: number,
	currentRequestId: number,
	disposed: boolean
): boolean => disposed || requestId !== currentRequestId;

/** Reapplies a completed world-space drop over an older displayed layout. */
export const nodesWithPendingDropPosition = (
	nodes: readonly FlowNode[],
	drop: PendingNodeDrop | undefined
): FlowNode[] => {
	if (!drop) return [...nodes];
	const draggedNode = nodes.find((node) => node.id === drop.entityId);
	if (!draggedNode || draggedNode.data.appearance !== "compact") return [...nodes];

	const worldPositions = worldPositionByNodeId(nodes);
	const parentPosition = draggedNode.parentId
		? (worldPositions.get(draggedNode.parentId) ?? { x: 0, y: 0 })
		: { x: 0, y: 0 };
	return nodes.map((node) =>
		node.id === drop.entityId
			? {
					...node,
					position: {
						x: drop.position.x - parentPosition.x,
						y: drop.position.y - parentPosition.y,
					},
				}
			: node
	);
};

/** Tracks one pointer drag and emits only a completed compact-node move. */
export class CompactNodeDragSession {
	private active:
		| {
				entityId: string;
				start: Point;
				moved: boolean;
		  }
		| undefined;

	get isActive(): boolean {
		return this.active !== undefined;
	}

	start(
		entityId: string,
		appearance: MapNodeAppearance,
		position: Point,
		hasMoveCallback: boolean
	): boolean {
		if (!nodeCanBeDragged(appearance, hasMoveCallback)) return false;
		this.active = { entityId, start: position, moved: false };
		return true;
	}

	update(entityId: string, position: Point): Point | undefined {
		if (!this.active || this.active.entityId !== entityId) return undefined;
		this.active.moved =
			this.active.moved || position.x !== this.active.start.x || position.y !== this.active.start.y;
		return { x: position.x - this.active.start.x, y: position.y - this.active.start.y };
	}

	finish(entityId: string, position: Point, callback: NodeMoveCallback | undefined): boolean {
		const active = this.active;
		this.active = undefined;
		if (!active || active.entityId !== entityId || !active.moved || !callback) return false;
		callback(entityId, position);
		return true;
	}

	cancel() {
		this.active = undefined;
	}
}
