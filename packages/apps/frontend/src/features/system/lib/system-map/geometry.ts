export type Point = {
	x: number;
	y: number;
};

export type Size = {
	width: number;
	height: number;
};

export type Bounds = Point & Size;

export type Viewport = {
	x: number;
	y: number;
	zoom: number;
};

export type LayoutNodeLike = {
	id: string;
	position: Point;
	parentId?: string;
	width?: number;
	height?: number;
};

const finiteOr = (value: number, fallback: number): number => (Number.isFinite(value) ? value : fallback);

const parentRelativePosition = (
	node: LayoutNodeLike,
	nodesById: ReadonlyMap<string, LayoutNodeLike>,
	cache: Map<string, Point>,
	visiting: Set<string>
): Point => {
	const cached = cache.get(node.id);
	if (cached) return cached;
	if (visiting.has(node.id)) return node.position;

	visiting.add(node.id);
	const parent = node.parentId ? nodesById.get(node.parentId) : undefined;
	const parentPosition = parent
		? parentRelativePosition(parent, nodesById, cache, visiting)
		: { x: 0, y: 0 };
	const position = { x: parentPosition.x + node.position.x, y: parentPosition.y + node.position.y };
	visiting.delete(node.id);
	cache.set(node.id, position);
	return position;
};

/** Resolves Svelte Flow's parent-relative positions into world-space positions. */
export const worldPositionByNodeId = (nodes: readonly LayoutNodeLike[]): ReadonlyMap<string, Point> => {
	const nodesById = new Map(nodes.map((node) => [node.id, node]));
	const positions = new Map<string, Point>();

	for (const node of nodes) {
		parentRelativePosition(node, nodesById, positions, new Set());
	}

	return positions;
};

/** Returns world-space bounds for nodes, resolving parent-relative renderer coordinates. */
export const worldBoundsByNodeId = (nodes: readonly LayoutNodeLike[]): ReadonlyMap<string, Bounds> => {
	const positions = worldPositionByNodeId(nodes);
	const bounds = new Map<string, Bounds>();

	for (const node of nodes) {
		const position = positions.get(node.id);
		if (!position) continue;

		bounds.set(node.id, {
			x: position.x,
			y: position.y,
			width: Math.max(1, node.width ?? 1),
			height: Math.max(1, node.height ?? 1),
		});
	}

	return bounds;
};

export const worldPointFromScreen = (viewport: Viewport, screenPoint: Point): Point => {
	const zoom = Math.max(0.01, finiteOr(viewport.zoom, 1));
	return {
		x: (screenPoint.x - viewport.x) / zoom,
		y: (screenPoint.y - viewport.y) / zoom,
	};
};

export const screenPointFromWorld = (viewport: Viewport, worldPoint: Point): Point => ({
	x: worldPoint.x * viewport.zoom + viewport.x,
	y: worldPoint.y * viewport.zoom + viewport.y,
});

const distanceToBoundsSquared = (worldPoint: Point, bounds: Bounds): number => {
	const dx = worldPoint.x < bounds.x ? bounds.x - worldPoint.x : worldPoint.x - (bounds.x + bounds.width);
	const dy = worldPoint.y < bounds.y ? bounds.y - worldPoint.y : worldPoint.y - (bounds.y + bounds.height);
	return (dx > 0 ? dx : 0) ** 2 + (dy > 0 ? dy : 0) ** 2;
};

/** Finds the closest displayed node to a screen point, preferring the smallest containing node. */
export const nearestNodeIdAtScreenPoint = (
	nodes: readonly LayoutNodeLike[],
	viewport: Viewport,
	screenPoint: Point
): string | undefined => {
	const worldPoint = worldPointFromScreen(viewport, screenPoint);
	const bounds = worldBoundsByNodeId(nodes);
	let nearest: { id: string; distance: number; area: number } | undefined;

	for (const node of nodes) {
		const nodeBounds = bounds.get(node.id);
		if (!nodeBounds) continue;

		const candidate = {
			id: node.id,
			distance: distanceToBoundsSquared(worldPoint, nodeBounds),
			area: nodeBounds.width * nodeBounds.height,
		};
		if (
			!nearest ||
			candidate.distance < nearest.distance ||
			(candidate.distance === nearest.distance && candidate.area < nearest.area) ||
			(candidate.distance === nearest.distance &&
				candidate.area === nearest.area &&
				candidate.id < nearest.id)
		) {
			nearest = candidate;
		}
	}

	return nearest?.id;
};

export const viewportForZoomAtPoint = (
	viewport: Viewport,
	zoom: number,
	screenPoint: Point
): Viewport => {
	const safeZoom = Math.max(0.01, finiteOr(zoom, viewport.zoom));
	const worldPoint = worldPointFromScreen(viewport, screenPoint);

	return {
		x: screenPoint.x - worldPoint.x * safeZoom,
		y: screenPoint.y - worldPoint.y * safeZoom,
		zoom: safeZoom,
	};
};

/** Centers a world-space point without altering the current zoom. */
export const viewportCenteredOn = (
	viewport: Viewport,
	worldPoint: Point,
	size: Size
): Viewport => ({
	x: size.width / 2 - worldPoint.x * viewport.zoom,
	y: size.height / 2 - worldPoint.y * viewport.zoom,
	zoom: viewport.zoom,
});

export const worldCenter = (bounds: Bounds): Point => ({
	x: bounds.x + bounds.width / 2,
	y: bounds.y + bounds.height / 2,
});

export const boundsForNodes = (
	nodes: readonly LayoutNodeLike[],
	ids?: ReadonlySet<string>
): Bounds | undefined => {
	const bounds = [...worldBoundsByNodeId(nodes).entries()]
		.filter(([id]) => !ids || ids.has(id))
		.map(([, nodeBounds]) => nodeBounds);
	if (!bounds.length) return undefined;

	const left = Math.min(...bounds.map((entry) => entry.x));
	const top = Math.min(...bounds.map((entry) => entry.y));
	const right = Math.max(...bounds.map((entry) => entry.x + entry.width));
	const bottom = Math.max(...bounds.map((entry) => entry.y + entry.height));
	return { x: left, y: top, width: right - left, height: bottom - top };
};

export const viewportWorldBounds = (
	viewport: Viewport,
	size: Size,
	margin: number
): Bounds | undefined => {
	const zoom = finiteOr(viewport.zoom, 1);
	if (zoom <= 0 || size.width <= 0 || size.height <= 0) return undefined;
	const worldMargin = margin / zoom;
	return {
		x: -viewport.x / zoom - worldMargin,
		y: -viewport.y / zoom - worldMargin,
		width: size.width / zoom + worldMargin * 2,
		height: size.height / zoom + worldMargin * 2,
	};
};

export const boundsIntersect = (left: Bounds, right: Bounds): boolean =>
	left.x < right.x + right.width &&
	left.x + left.width > right.x &&
	left.y < right.y + right.height &&
	left.y + left.height > right.y;
