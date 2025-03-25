import type { ParamMatcher } from "@sveltejs/kit";

export type OncallShiftViewRouteParam = undefined | "handover";
const params = new Set([undefined, "handover"]);

export const convertOncallShiftViewParam = (param?: string): OncallShiftViewRouteParam => {
	if (!params.has(param)) return undefined;
	return param as OncallShiftViewRouteParam;
};
export const match = ((param?: string): param is OncallShiftViewRouteParam => {
	return params.has(param);
}) satisfies ParamMatcher;
