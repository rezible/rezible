import {
	boundsIntersect,
	type LayoutNodeLike,
	type Viewport,
	type Size,
	viewportWorldBounds,
	worldBoundsByNodeId,
} from "./geometry";
import {
	isArchitectureCategory,
	getMapCategoryDisplay,
	NodeDetailLevel,
} from "./category";
import type { GraphSubset } from "./graph";
import type { MapProjection } from "./presentation";

export const REVEAL_ZOOM_ANCHORS = [
	{ zoom: 0.45, detail: 0 },
	{ zoom: 0.82, detail: 1 },
	{ zoom: 1.25, detail: 2 },
	{ zoom: 1.68, detail: 3 },
] as const;

/** Prototype mapping only; browser review may tune these values. */
export const REVEAL_THRESHOLDS = {
	enter: [0.45, 0.86, 1.3, 1.74],
	exit: [0.45, 0.68, 1.1, 1.54],
} as const;

export const MAP_MIN_ZOOM = 0.1;
export const MAP_MAX_ZOOM = 2.5;
export const INITIAL_VIEWPORT_ZOOM = 0.55;
export const EXPLICIT_FIT_MAX_ZOOM = 1.25;
export const MAX_REVEAL_DETAIL = 3;

const INITIAL_FRAME_HEADROOM = 0.01;

const clamp = (value: number, minimum: number, maximum: number): number =>
	Math.max(minimum, Math.min(maximum, value));

const finiteOr = (value: number, fallback: number): number => (Number.isFinite(value) ? value : fallback);

/** Maps the viewport zoom to the continuous Landscape–Implementation detail position. */
export const detailFromZoom = (zoom: number): number => {
	const safeZoom = finiteOr(zoom, REVEAL_ZOOM_ANCHORS[0].zoom);
	if (safeZoom <= REVEAL_ZOOM_ANCHORS[0].zoom) return 0;
	if (safeZoom >= REVEAL_ZOOM_ANCHORS[REVEAL_ZOOM_ANCHORS.length - 1].zoom) return MAX_REVEAL_DETAIL;

	for (let index = 1; index < REVEAL_ZOOM_ANCHORS.length; index += 1) {
		const lower = REVEAL_ZOOM_ANCHORS[index - 1];
		const upper = REVEAL_ZOOM_ANCHORS[index];
		if (safeZoom <= upper.zoom) {
			const progress = (safeZoom - lower.zoom) / (upper.zoom - lower.zoom);
			return lower.detail + progress * (upper.detail - lower.detail);
		}
	}

	return MAX_REVEAL_DETAIL;
};

/** Keeps structural reveal discrete while allowing separate enter and exit thresholds. */
export const structuralDetailForZoom = (zoom: number, previousDetail: number): number => {
	const safeZoom = finiteOr(zoom, REVEAL_ZOOM_ANCHORS[0].zoom);
	let detail = Math.round(clamp(previousDetail, 0, MAX_REVEAL_DETAIL));

	while (detail < MAX_REVEAL_DETAIL && safeZoom >= REVEAL_THRESHOLDS.enter[detail + 1]) detail += 1;
	while (detail > 0 && safeZoom <= REVEAL_THRESHOLDS.exit[detail]) detail -= 1;

	return detail;
};

/** Returns the minimum zoom at which an explicit Reveal can show the requested detail. */
export const revealZoomForDetail = (detail: number): number => {
	const level = Math.round(clamp(finiteOr(detail, 0), 0, MAX_REVEAL_DETAIL));
	return REVEAL_THRESHOLDS.enter[level];
};

/** Converts an externally supplied initial detail into the structural state used by projection. */
export const initialStructuralDetail = (detail: number): number =>
	Math.round(clamp(finiteOr(detail, 0), 0, MAX_REVEAL_DETAIL));

/**
 * Caps the first automatic frame just below the next structural reveal threshold. Initial framing
 * preserves the requested level; explicit Fit uses its independent range and may frame more detail.
 */
export const initialFrameMaxZoom = (detail: number): number => {
	const structuralDetail = initialStructuralDetail(detail);
	if (structuralDetail >= MAX_REVEAL_DETAIL) return MAP_MAX_ZOOM;

	return Math.min(
		MAP_MAX_ZOOM,
		REVEAL_THRESHOLDS.enter[structuralDetail + 1] - INITIAL_FRAME_HEADROOM
	);
};

/**
 * Finds source entities eligible near the viewport using current representatives. Descendants are
 * checked through their source-to-representative mapping, so they do not need to be rendered first.
 */
export const deriveNearbyEntityIds = (
	graph: GraphSubset,
	projection: MapProjection,
	nodes: readonly LayoutNodeLike[],
	viewport: Viewport,
	size: Size,
	margin = 180
): readonly string[] => {
	const visibleWorld = viewportWorldBounds(viewport, size, margin);
	if (!visibleWorld) return [];

	const nodeBounds = worldBoundsByNodeId(nodes);
	const nodesById = new Map(nodes.map((node) => [node.id, node]));
	const appearanceById = new Map(projection.nodes.map((node) => [node.id, node.appearance]));
	const nearby = new Set<string>();

	for (const entity of graph.entities) {
		const representativeId = projection.representativeByEntityId.get(entity.id);
		if (!representativeId) continue;
		const representativeBounds = nodeBounds.get(representativeId);
		if (representativeBounds && boundsIntersect(representativeBounds, visibleWorld)) {
			nearby.add(entity.id);
			continue;
		}

		// A revealed descendant can sit outside its group's bounds while the group itself is still in the
		// viewport margin. Ancestor eligibility can retain that offscreen descendant until its local group leaves
		// the margin; otherwise
		// the next projection replaces it with the group and immediately asks for it again. The group still
		// leaves eligibility when its own bounds leave the margin, so panning across the root remains local.
		let parentId = nodesById.get(representativeId)?.parentId;
		while (parentId) {
			const parent = nodesById.get(parentId);
			if (!parent) break;

			const parentBounds = nodeBounds.get(parent.id);
			if (
				appearanceById.get(parent.id) === "group" &&
				parentBounds &&
				boundsIntersect(parentBounds, visibleWorld)
			) {
				nearby.add(entity.id);
				break;
			}

			parentId = parent.parentId;
		}
	}

	return [...nearby].sort();
};

/** Shallowest supplied architecture; context and unsupported categories have no level. */
export const minimumDetailForGraph = (graph: GraphSubset): NodeDetailLevel | undefined => {
	let minimum: NodeDetailLevel | undefined;
	for (const entity of graph.entities) {
		if (!isArchitectureCategory(entity.category)) continue;
		const level = getMapCategoryDisplay(entity.category).level;
		if (level === undefined) continue;
		if (minimum === undefined || level < minimum) minimum = level;
	}
	return minimum;
};

export type SystemMapOverviewEmptyState = "truly-empty" | "no-architecture";

export const systemMapOverviewEmptyState = (
	graph: GraphSubset | undefined,
	projection: MapProjection | undefined
): SystemMapOverviewEmptyState | undefined => {
	if (!graph || !projection || projection.nodes.length > 0) return undefined;
	if (graph.entities.length === 0) return "truly-empty";
	return minimumDetailForGraph(graph) === undefined ? "no-architecture" : undefined;
};
