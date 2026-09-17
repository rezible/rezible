import {
	detailFromZoom,
	INITIAL_VIEWPORT_ZOOM,
	MAP_MAX_ZOOM,
	MAP_MIN_ZOOM,
	initialStructuralDetail,
	revealZoomForDetail,
	structuralDetailForZoom,
} from "$features/systems/lib/system-map/interaction";
import {
	viewportForZoomAtPoint,
	worldCenter,
	type Bounds,
	type Point,
	type Size,
	type Viewport,
} from "$features/systems/lib/system-map/geometry";

export type RevealTransition = {
	viewport: Viewport;
	zoomChanged: boolean;
};

export type LayoutChainSettlementOptions = {
	isCurrent: () => boolean;
	updateNearby: () => boolean;
	clearAnchor: () => void;
};

/** Runs post-layout nearby feedback and retains the anchor when a follow-up is scheduled. */
export const settleLayoutChain = ({
	isCurrent,
	updateNearby,
	clearAnchor,
}: LayoutChainSettlementOptions): boolean => {
	if (!isCurrent()) return false;

	const followupScheduled = updateNearby();
	if (!followupScheduled && isCurrent()) clearAnchor();
	return followupScheduled;
};

export const preferredLayoutAnchorId = (
	pendingAnchorId: string | undefined,
	selectedEntityIds: readonly string[]
): string | undefined => pendingAnchorId ?? selectedEntityIds[0];

/** Stateful, non-reactive viewport transitions shared by the Svelte controller and Bun tests. */
export class MapViewportState {
	/**
	 * Svelte Flow updates its bound viewport before calling `onmove`. Keep this last processed value
	 * separate so interaction decisions compare against the viewport that this state has actually seen.
	 */
	lastProcessedViewport: Viewport = { x: 0, y: 0, zoom: INITIAL_VIEWPORT_ZOOM };
	continuousDetail: number;
	structuralDetail: number;
	pendingLayoutAnchorId: string | undefined;

	private minimumDetail: number;

	constructor(minimumDetail: number) {
		this.minimumDetail = initialStructuralDetail(minimumDetail);
		this.continuousDetail = detailFromZoom(this.lastProcessedViewport.zoom);
		this.structuralDetail = this.minimumDetail;
	}

	setMinimumDetail(minimumDetail: number) {
		const next = initialStructuralDetail(minimumDetail);
		if (next === this.minimumDetail) return;
		this.minimumDetail = next;
		this.structuralDetail = this.detailForZoom(this.lastProcessedViewport.zoom);
	}

	private detailForZoom(zoom: number): number {
		return Math.max(this.minimumDetail, structuralDetailForZoom(zoom, this.structuralDetail));
	}

	/** Svelte Flow may already have updated its bound viewport by the time onmove runs. */
	hasZoomChanged(viewport: Viewport): boolean {
		return viewport.zoom !== this.lastProcessedViewport.zoom;
	}

	/** Applies a viewport update and reports whether structural projection must be rebuilt. */
	updateViewport(viewport: Viewport, pointerZoom = false, pointerAnchorId?: string): boolean {
		if (pointerZoom && pointerAnchorId) this.pendingLayoutAnchorId = pointerAnchorId;

		this.lastProcessedViewport = { ...viewport };
		this.continuousDetail = detailFromZoom(viewport.zoom);
		const nextStructuralDetail = this.detailForZoom(viewport.zoom);
		const structuralChanged = nextStructuralDetail !== this.structuralDetail;
		this.structuralDetail = nextStructuralDetail;
		return structuralChanged;
	}

	prepareKeyboardZoom(
		factor: number,
		focusEntityId: string | undefined,
		focusBounds: Bounds | undefined,
		size: Size
	): Viewport {
		this.pendingLayoutAnchorId = focusEntityId;
		const zoom = Math.max(MAP_MIN_ZOOM, Math.min(MAP_MAX_ZOOM, this.lastProcessedViewport.zoom * factor));
		const screenPoint = {
			x: size.width > 0 ? size.width / 2 : 0,
			y: size.height > 0 ? size.height / 2 : 0,
		};
		const worldPoint = focusBounds
			? worldCenter(focusBounds)
			: {
					x: (screenPoint.x - this.lastProcessedViewport.x) / this.lastProcessedViewport.zoom,
					y: (screenPoint.y - this.lastProcessedViewport.y) / this.lastProcessedViewport.zoom,
				};

		const nextViewport = viewportForZoomAtPoint(this.lastProcessedViewport, zoom, {
			x: worldPoint.x * this.lastProcessedViewport.zoom + this.lastProcessedViewport.x,
			y: worldPoint.y * this.lastProcessedViewport.zoom + this.lastProcessedViewport.y,
		}) as Viewport;
		if (this.detailForZoom(nextViewport.zoom) === this.structuralDetail) {
			this.pendingLayoutAnchorId = undefined;
		}
		return nextViewport;
	}

	prepareReveal(
		entityId: string,
		nextDetail: number,
		screenPoint: Point
	): RevealTransition {
		this.pendingLayoutAnchorId = entityId;
		const revealZoom = Math.max(this.lastProcessedViewport.zoom, revealZoomForDetail(nextDetail));
		if (revealZoom === this.lastProcessedViewport.zoom) {
			this.structuralDetail = Math.max(this.minimumDetail, nextDetail);
			this.continuousDetail = detailFromZoom(this.lastProcessedViewport.zoom);
			return { viewport: this.lastProcessedViewport, zoomChanged: false };
		}

		return {
			viewport: viewportForZoomAtPoint(this.lastProcessedViewport, revealZoom, screenPoint) as Viewport,
			zoomChanged: true,
		};
	}

	clearLayoutAnchor() {
		this.pendingLayoutAnchorId = undefined;
	}
}
