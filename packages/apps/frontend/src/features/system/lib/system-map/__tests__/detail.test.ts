import { expect, test } from "bun:test";
import { detailBandFromZoom, resolveDetailBand, runtimeThreshold, type DetailBand } from "../detail";

const walk = (zooms: readonly number[], fitZoom: number) => {
	let band: DetailBand | undefined;
	return zooms.map((zoom) => {
		band = detailBandFromZoom(zoom, fitZoom, band);
		return band;
	});
};

test("the threshold is relative to the fit, with a floor for maps that fit far out", () => {
	expect(runtimeThreshold(1)).toBe(1.5);
	expect(runtimeThreshold(0.12)).toBe(0.7);
});

test("a map opens at the landscape band", () => {
	expect(detailBandFromZoom(1, 1)).toBe("landscape");
	expect(detailBandFromZoom(0.12, 0.12)).toBe("landscape");
});

test("a small map enters runtime above 1.5 and leaves below 1.35", () => {
	expect(walk([1, 1.5, 1.51, 1.4, 1.36, 1.34, 1.4], 1)).toEqual([
		"landscape",
		"landscape",
		"runtime",
		"runtime",
		"runtime",
		"landscape",
		"landscape",
	]);
});

test("a large map enters runtime above 0.7 and leaves below 0.63", () => {
	expect(walk([0.12, 0.7, 0.71, 0.64, 0.62], 0.12)).toEqual([
		"landscape",
		"landscape",
		"runtime",
		"runtime",
		"landscape",
	]);
});

test("a pinned band ignores zoom", () => {
	expect(resolveDetailBand("runtime", "landscape")).toBe("runtime");
	expect(resolveDetailBand("auto", "runtime")).toBe("runtime");
});
