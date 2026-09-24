import {
	detailFromZoom,
	INITIAL_VIEWPORT_ZOOM,
	MAP_MAX_ZOOM,
	MAP_MIN_ZOOM,
	initialStructuralDetail,
	revealZoomForDetail,
	structuralDetailForZoom,
} from "$features/system/lib/system-map/interaction";
import {
	viewportForZoomAtPoint,
	worldCenter,
	type Bounds,
	type Point,
	type Size,
	type Viewport,
} from "$features/system/lib/system-map/geometry";

export type RevealTransition = {
	viewport: Viewport;
	zoomChanged: boolean;
};

/** Keeps continuous label detail and discrete visible projection state for one map camera. */
export class MapViewportState {
	lastProcessedViewport: Viewport = { x: 0, y: 0, zoom: INITIAL_VIEWPORT_ZOOM };
	continuousDetail: number = detailFromZoom(INITIAL_VIEWPORT_ZOOM);
	structuralDetail: number;

	private minimumDetail: number;

	constructor(minimumDetail: number) {
		this.minimumDetail = initialStructuralDetail(minimumDetail);
		this.structuralDetail = this.minimumDetail;
	}

	setMinimumDetail(minimumDetail: number) {
		this.minimumDetail = initialStructuralDetail(minimumDetail);
		this.structuralDetail = this.detailForZoom(this.lastProcessedViewport.zoom);
	}

	private detailForZoom(zoom: number): number {
		return Math.max(this.minimumDetail, structuralDetailForZoom(zoom, this.structuralDetail));
	}

	/** Applies a viewport update and reports whether visible architecture must be reprojected. */
	updateViewport(viewport: Viewport): boolean {
		this.lastProcessedViewport = { ...viewport };
		this.continuousDetail = detailFromZoom(viewport.zoom);
		const nextStructuralDetail = this.detailForZoom(viewport.zoom);
		const structuralChanged = nextStructuralDetail !== this.structuralDetail;
		this.structuralDetail = nextStructuralDetail;
		return structuralChanged;
	}

	prepareKeyboardZoom(factor: number, focusBounds: Bounds | undefined, size: Size): Viewport {
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

		return viewportForZoomAtPoint(this.lastProcessedViewport, zoom, {
			x: worldPoint.x * this.lastProcessedViewport.zoom + this.lastProcessedViewport.x,
			y: worldPoint.y * this.lastProcessedViewport.zoom + this.lastProcessedViewport.y,
		}) as Viewport;
	}

	prepareReveal(nextDetail: number, screenPoint: Point): RevealTransition {
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
}
