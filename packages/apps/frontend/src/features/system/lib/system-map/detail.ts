export type DetailBand = "landscape" | "runtime";
export type DetailMode = "auto" | DetailBand;

/** The zoom above which runtime detail appears, relative to the zoom at which the whole map fits. */
export const runtimeThreshold = (fitZoom: number): number => Math.max(1.5 * fitZoom, 0.7);

/**
 * Enters the runtime band above the threshold and leaves it below 90% of it, so wheel zoom near the boundary
 * does not flicker. With no prior band the map is at the landscape band, where it opens.
 */
export function detailBandFromZoom(
	zoom: number,
	fitZoom: number,
	previous: DetailBand = "landscape"
): DetailBand {
	const threshold = runtimeThreshold(fitZoom);
	if (previous === "runtime") {
		if (zoom < 0.9 * threshold) {
			return "landscape";
		}
		return "runtime";
	}

	if (zoom > threshold) {
		return "runtime";
	}
	return "landscape";
}

export function resolveDetailBand(mode: DetailMode, autoBand: DetailBand): DetailBand {
	return mode === "auto" ? autoBand : mode;
}
